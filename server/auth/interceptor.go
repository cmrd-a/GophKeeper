package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userIDKeyType string

const UserIDKey userIDKeyType = "user_id"

// UnaryServerInterceptor returns a new unary server interceptor for auth.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Skip auth for Register and Login
		if info.FullMethod == "/user.UserService/Register" || info.FullMethod == "/user.UserService/Login" {
			return handler(ctx, req)
		}

		// Get token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token := md.Get("authorization")
		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		// Validate token and get user ID
		userID, err := ParseAndValidate(token[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add user ID to context
		newCtx := context.WithValue(ctx, UserIDKey, userID)
		return handler(newCtx, req)
	}
}

// GetUserIDFromContext extracts user ID from context.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", status.Error(codes.Internal, "user ID not found in context")
	}
	return userID, nil
}
