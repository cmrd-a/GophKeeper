package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
	"github.com/cmrd-a/GophKeeper/server/insecure"
	"github.com/cmrd-a/GophKeeper/server/logger"

	"github.com/cmrd-a/GophKeeper/server/api"
	"github.com/cmrd-a/GophKeeper/server/config"
	"github.com/cmrd-a/GophKeeper/server/gateway"

	"google.golang.org/grpc/credentials"
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

	s := grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)))
	user.RegisterUserServiceServer(s, &api.UserServer{})
	vault.RegisterVaultServiceServer(s, &api.VaultServer{})
	reflection.Register(s)

	log.Info("Serving gRPC on ", "addr", addr)
	go func() {
		err := s.Serve(lis)
		if err != nil {
			log.Error("failed to serve grpc", "error", err)
			os.Exit(1)
		}
	}()

	err = gateway.Run(addr, cfg.HTTPPort)
	if err != nil {
		log.Error("failed to serve http", "error", err)
		os.Exit(1)
	}
}
