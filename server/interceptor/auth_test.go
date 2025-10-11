package interceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		token         string
		expectedError bool
	}{
		{
			name:          "NoAuth_WhitelistedMethod",
			method:        "/v1.user.UserService/Register",
			token:         "",
			expectedError: false,
		},
		{
			name:          "NoAuth_NonWhitelistedMethod",
			method:        "/v1.vault.VaultService/GetVaultItems",
			token:         "",
			expectedError: true,
		},
		{
			name:          "InvalidToken",
			method:        "/v1.vault.VaultService/GetVaultItems",
			token:         "Bearer invalid-token",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthInterceptor

			ctx := context.Background()
			if tt.token != "" {
				md := metadata.New(map[string]string{
					"authorization": tt.token,
				})
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: tt.method,
			}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "ok", nil
			}

			// Test
			resp, err := interceptor(ctx, nil, info, handler)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "ok", resp)
			}
		})
	}
}

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		expectedError bool
	}{
		{
			name:          "ValidUserID",
			userID:        "12345",
			expectedError: false,
		},
		{
			name:          "NoUserID",
			userID:        "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.userID != "" {
				ctx = ContextWithUserID(ctx, tt.userID)
			}

			userID, err := UserIDFromContext(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, userID)
			}
		})
	}
}

func TestStreamAuthInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedError bool
	}{
		{
			name:          "NoToken",
			token:         "",
			expectedError: true,
		},
		{
			name:          "InvalidToken",
			token:         "Bearer invalid-token",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.token != "" {
				md := metadata.New(map[string]string{
					"authorization": tt.token,
				})
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			stream := &mockServerStream{ctx: ctx}
			info := &grpc.StreamServerInfo{}
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}

			err := StreamAuthInterceptor(nil, stream, info, handler)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type mockServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (s *mockServerStream) Context() context.Context {
	return s.ctx
}
