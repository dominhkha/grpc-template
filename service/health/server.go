package backend_health

import (
	"context"
	health_rpc "github.com/dominhkha/grpc-template/rpc/backend/v1"
)

type (
	Server struct {
		health_rpc.UnimplementedHealthServiceServer
	}
)

func NewServer() *Server {
	return &Server{}
}

func (s Server) CheckHealth(context.Context, *health_rpc.CheckHealthRequest) (*health_rpc.CheckHealthResponse, error) {
	return &health_rpc.CheckHealthResponse{
		Response: "ok!",
	}, nil
}
