package grpc

import (
	"context"
	"github.com/folivorra/dumpd/application"
	"github.com/folivorra/dumpd/internal/config"
	"github.com/folivorra/dumpd/internal/logger"
	"github.com/folivorra/dumpd/internal/storage"
	dumpd_v1 "github.com/folivorra/dumpd/pkg/proto/dumpd/v1"
	rpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	grpcServer *rpc.Server
	listener   net.Listener
}

func NewServer(cfg *config.Config, app *application.App, store storage.Storager) (*Server, error) {
	lis, err := net.Listen("tcp", cfg.Server.GrpcPort)
	if err != nil {
		logger.ErrorLogger.Printf("failed to listen: %v", err)
		return nil, err
	}

	grpcServer := rpc.NewServer()

	service := NewItemController(store)

	dumpd_v1.RegisterDumpdServiceServer(grpcServer, service)

	reflection.Register(grpcServer)

	s := &Server{
		grpcServer: grpcServer,
		listener:   lis,
	}

	app.RegisterCleanup(func(ctx context.Context) {
		_ = s.Shutdown(context.Background())
	})

	return s, nil
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
	return nil
}
