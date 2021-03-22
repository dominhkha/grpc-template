

TARGET = bin
TARGET_BIN = letme
CLIENT = cmd/client/client.go
CMD_MAIN := cmd/server/main.go

PROTO_DIR := proto
RPC_DIR := rpc


GRPC_GATEWAY := $(shell go list -m -f "{{.Dir}}" github.com/grpc-ecosystem/grpc-gateway)

GOOGLE_API_PATH := ${GRPC_GATEWAY}/third_party/googleapis/

GO_TOOLS = golang.org/x/tools/cmd/stringer \
github.com/stretchr/testify/mock \
github.com/vektra/mockery/.../ \
github.com/google/wire/cmd/wire \
github.com/golang/protobuf/protoc-gen-go \
github.com/gogo/protobuf/protoc-gen-gogo \
github.com/gogo/protobuf/protoc-gen-gofast \
github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
google.golang.org/grpc/cmd/protoc-gen-go-grpc \
github.com/envoyproxy/protoc-gen-validate \
github.com/mwitkow/go-proto-validators/protoc-gen-govalidators \
github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
github.com/golang/mock/mockgen@v1.4.4 \
github.com/grpc-ecosystem/go-grpc-middleware


define generate
	mkdir -p ${RPC_DIR}/$(1) && \
		protoc -I ${PROTO_DIR} \
			-I ${GOOGLE_API_PATH} \
			--go_out=paths=source_relative:${RPC_DIR} \
			--go-grpc_out=paths=source_relative:${RPC_DIR} \
			--grpc-gateway_out=logtostderr=true,paths=source_relative:${RPC_DIR} \
			${PROTO_DIR}/$(1)/$(2)
endef


.PHONY: all


$(GO_TOOLS):
	GOSUMDB=off go get -u $@

install-go-tools: $(GO_TOOLS)
	@echo \# installed go tools

generate-rpc:
	rm -rf $(RPC_DIR)
	$(call generate,backend/v1,health.proto)

build-only:
	@go build -o ${TARGET}/${TARGET_BIN} ${CMD_MAIN}

build: generate-rpc build-only

run-server: build
	@${TARGET}/${TARGET_BIN}