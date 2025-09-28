package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/cmrd-a/GophKeeper/gen/go/proto/user/v1"
	"github.com/cmrd-a/GophKeeper/insecure"

	"github.com/cmrd-a/GophKeeper/gateway"
	"github.com/cmrd-a/GophKeeper/server"

	"google.golang.org/grpc/credentials"
)

func main() {
	addr := "0.0.0.0:8082"

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)))
	pb.RegisterUserServiceServer(s, &server.Server{})
	reflection.Register(s)

	log.Printf("Serving gRPC on https://%s", addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	err = gateway.Run("dns:///" + addr)
	log.Fatalln(err)
}
