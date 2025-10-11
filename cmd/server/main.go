package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
	"github.com/cmrd-a/GophKeeper/server/api"
	"github.com/cmrd-a/GophKeeper/server/config"
	"github.com/cmrd-a/GophKeeper/server/gateway"
	"github.com/cmrd-a/GophKeeper/server/insecure"
	"github.com/cmrd-a/GophKeeper/server/interceptor"
	"github.com/cmrd-a/GophKeeper/server/logger"
	"github.com/cmrd-a/GophKeeper/server/repository"
	"github.com/cmrd-a/GophKeeper/server/service"
)

func main() {
	log, lvl := logger.NewLogger()
	cfg, err := config.NewConfig(log, lvl)
	if err != nil {
		log.Error("failed to make config", "error", err)
		os.Exit(1)
	}
	startServers(log, cfg)
}

func startServers(log *slog.Logger, cfg *config.Config) {
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	// Create repository
	repo, err := repository.NewRepository(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Error("failed to create repository", "error", err)
		os.Exit(1)
	}

	// Create vault service
	vaultService := service.NewService(repo)

	// Create chained interceptors with configuration
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// Add logging interceptors if enabled
	if cfg.LogGRPCRequests {
		loggingConfig := interceptor.LoggingConfig{
			LogPayloads:    cfg.LogGRPCPayloads,
			LogLevel:       slog.LevelInfo,
			MaxPayloadSize: cfg.MaxLogPayloadSize,
		}
		unaryInterceptors = append(
			unaryInterceptors,
			interceptor.ConfigurableLoggingUnaryInterceptor(log, loggingConfig),
		)
		streamInterceptors = append(streamInterceptors, interceptor.LoggingStreamInterceptor(log))
	}

	// Add auth interceptors
	unaryInterceptors = append(unaryInterceptors, interceptor.AuthInterceptor)
	streamInterceptors = append(streamInterceptors, interceptor.StreamAuthInterceptor)

	// Create chained interceptors
	unaryChain := chainUnaryInterceptors(unaryInterceptors...)
	streamChain := chainStreamInterceptors(streamInterceptors...)

	// Create server with chained interceptors
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)),
		grpc.UnaryInterceptor(unaryChain),
		grpc.StreamInterceptor(streamChain),
	}
	s := grpc.NewServer(opts...)

	// Register services
	userServer := &api.UserServer{Repository: repo}
	user.RegisterUserServiceServer(s, userServer)

	vaultServer := api.NewVaultServer(vaultService)
	vault.RegisterVaultServiceServer(s, vaultServer)

	reflection.Register(s)

	log.Info("Serving gRPC on ", "addr", addr)
	go func() {
		err := s.Serve(lis)
		if err != nil {
			log.Error("failed to serve grpc", "error", err)
			os.Exit(1)
		}
	}()

	err = gateway.Run(log, addr, cfg.HTTPPort)
	if err != nil {
		log.Error("failed to serve http", "error", err)
		os.Exit(1)
	}
}

// chainUnaryInterceptors chains multiple unary interceptors.
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chained := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chained
			chained = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}
		return chained(ctx, req)
	}
}

// chainStreamInterceptors chains multiple stream interceptors.
func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chained := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chained
			chained = func(currentSrv interface{}, currentStream grpc.ServerStream) error {
				return interceptor(currentSrv, currentStream, info, next)
			}
		}
		return chained(srv, stream)
	}
}
