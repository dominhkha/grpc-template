package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"time"

	heath_rpc "github.com/dominhkha/grpc-template/rpc/backend/v1"
	backend_health "github.com/dominhkha/grpc-template/service/health"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func methodsDecider(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
	return true
}

func initServer(logger *zap.Logger) *grpc.Server {
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	allMethodDecider := methodsDecider

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			logger.Error("recovery err: "+string(debug.Stack()), zap.Any("error", p))
			return status.Error(codes.Unknown, "panic triggered: ")
		}),
	}

	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(opts...),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_zap.PayloadUnaryServerInterceptor(logger, allMethodDecider),
		), grpc_middleware.WithStreamServerChain(),
	)

	//helloServer := demo_hello.NewServer()
	//hello_rpc.RegisterHelloServiceServer(server, helloServer)

	//eatingServer := demo_eating.NewServer()
	//eating_rpc.RegisterEatingServiceServer(server, eatingServer)
	healthServer := backend_health.NewServer()
	heath_rpc.RegisterHealthServiceServer(server, healthServer)

	grpc_prometheus.Register(server)
	grpc_prometheus.EnableHandlingTimeHistogram()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	ctx := context.Background()
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	//hello_rpc.RegisterHelloServiceHandlerFromEndpoint(ctx, mux, "localhost:5000", dialOpts)
	//letme_engine_rpc.RegisterHealthHandlerFromEndpoint(ctx, mux, "localhost:5000", dialOpts)
	heath_rpc.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, "localhost:5000", dialOpts)
	http.Handle("/api/", mux)
	http.Handle("/metrics", promhttp.Handler())

	return server
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	server := initServer(logger)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill)

	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		err = server.Serve(lis)
		if err != nil {
			logger.Error("serve", zap.Error(err))
		}
	}()

	httpServer := http.Server{
		Addr: ":5001",
	}

	go func() {
		defer wg.Done()

		err := httpServer.ListenAndServe()
		if err == http.ErrServerClosed {
			return
		}
		if err != nil {
			logger.Error("httpServer Listen", zap.Error(err))
		}
	}()

	signal := <-exit
	fmt.Println("SIGNAL", signal)

	server.GracefulStop()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		logger.Error("httpServer Shutdown", zap.Error(err))
	}

	wg.Wait()

	fmt.Println("Stop successfully")

}
