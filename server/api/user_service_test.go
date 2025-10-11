package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/server/repository"
)

func TestUserServer_Creation(t *testing.T) {
	repo := &repository.Repository{}
	server := &UserServer{Repository: repo}

	assert.NotNil(t, server)
	assert.Equal(t, repo, server.Repository)
}

func TestUserServer_RequestValidation(t *testing.T) {
	server := &UserServer{Repository: nil}

	// Verify server is properly initialized
	assert.NotNil(t, server)
	assert.Nil(t, server.Repository)

	t.Run("RegisterRequestFields", func(t *testing.T) {
		req := &user.RegisterRequest{
			Login:    "testuser",
			Password: "testpass",
		}

		assert.Equal(t, "testuser", req.GetLogin())
		assert.Equal(t, "testpass", req.GetPassword())
	})

	t.Run("LoginRequestFields", func(t *testing.T) {
		req := &user.LoginRequest{
			Login:    "testuser",
			Password: "testpass",
		}

		assert.Equal(t, "testuser", req.GetLogin())
		assert.Equal(t, "testpass", req.GetPassword())
	})

	t.Run("RegisterResponseCreation", func(t *testing.T) {
		resp := &user.RegisterResponse{}
		assert.NotNil(t, resp)
	})

	t.Run("LoginResponseCreation", func(t *testing.T) {
		resp := &user.LoginResponse{
			Token: "test-token",
		}
		assert.Equal(t, "test-token", resp.GetToken())
	})
}
