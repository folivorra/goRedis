package grpc

import (
	"context"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/storage"
	goredis_v1 "github.com/folivorra/goRedis/pkg/proto/goredis/v1"
	rpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	grpcServer *rpc.Server
	listener   net.Listener
}

func NewServer(cfg *config.Config, store storage.Storager) (*Server, error) {
	lis, err := net.Listen("tcp", cfg.Server.GrpcPort)
	if err != nil {
		logger.ErrorLogger.Printf("failed to listen: %v", err)
		return nil, err
	}

	grpcServer := rpc.NewServer()

	service := NewItemController(store)

	goredis_v1.RegisterGoRedisServiceServer(grpcServer, service)

	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

func (s *Server) Start() error {
	logger.InfoLogger.Printf("Starting gRPC server on port 50051")
	if err := s.grpcServer.Serve(s.listener); err != nil {
		logger.ErrorLogger.Printf("failed to serve: %v", err)
		return err
	}
	return nil
}

func (s *Server) Shutdown(_ context.Context) error {
	logger.InfoLogger.Printf("Shutting down gRPC server on port 50051")
	s.grpcServer.GracefulStop()
	logger.InfoLogger.Println("gRPC server shutdown complete")
	return nil
}
