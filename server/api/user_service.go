package api

import (
	"context"
	"time"

	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/server/auth"
	"github.com/cmrd-a/GophKeeper/server/repository"
)

// UserServer implements UserService.
type UserServer struct {
	user.UnimplementedUserServiceServer

	Repository *repository.Repository
}

// Register creates a new user with hashed password.
func (s *UserServer) Register(ctx context.Context, in *user.RegisterRequest) (*user.RegisterResponse, error) {
	login := in.GetLogin()
	pw := in.GetPassword()
	log.Printf("register login: %v", login)

	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	_, err = s.Repository.InsertUser(ctx, login, hashed)
	if err != nil {
		return nil, err
	}
	return &user.RegisterResponse{}, nil
}

// Login authenticates a user and returns a JWT token.
func (s *UserServer) Login(ctx context.Context, in *user.LoginRequest) (*user.LoginResponse, error) {
	login := in.GetLogin()
	pw := in.GetPassword()
	id, hashed, err := s.Repository.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(pw)); err != nil {
		return nil, err
	}
	token, err := auth.CreateToken(id, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &user.LoginResponse{Token: token}, nil
}
