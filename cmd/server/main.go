package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/cmrd-a/GophKeeper/pb/proto"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement GophKeeper.EchoHandlerServer.
type server struct {
	pb.UnimplementedEchoHandlerServer
}

// Echo implements EchoHandlerServer.Echo.
func (s *server) Echo(_ context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	m := in.GetMessage()
	log.Printf("Received: %v", m)
	return &pb.EchoResponse{Message: &m}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterEchoHandlerServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
