package api

import (
	"context"
	"os"

	"log"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/server/repository"
)

// UserServer implements UserService.
type UserServer struct {
	user.UnimplementedUserServiceServer
}

// Register implements EchoHandlerServer.Echo.
func (s *UserServer) Register(_ context.Context, in *user.RegisterRequest) (*user.RegisterResponse, error) {
	email := in.GetEmail()
	password := in.GetPassword()
	log.Printf("email: %v", email)
	log.Printf("password: %v", password)
	r, err := repository.NewRepository(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	r.InsertUser("1")
	return &user.RegisterResponse{}, nil
}
