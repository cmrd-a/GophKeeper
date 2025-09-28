package server

import (
	"context"

	"log"

	pb "github.com/cmrd-a/GophKeeper/gen/go/proto/user/v1"
)

// server is used to implement GophKeeper.EchoHandlerServer.
type Server struct {
	pb.UnimplementedUserServiceServer
}

// Echo implements EchoHandlerServer.Echo.
func (s *Server) Echo(_ context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	m := in.GetIn()
	log.Printf("Received: %v", m)
	return &pb.EchoResponse{Out: m}, nil
}
