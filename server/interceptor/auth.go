package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/cmrd-a/GophKeeper/server/auth"
)

type userIDKey struct{}

// UserIDFromContext gets user ID from context.
func UserIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no user id in context")
	}
	return id, nil
}

// AuthInterceptor validates JWT tokens and adds user ID to context.
func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Skip auth for user service methods
	switch info.FullMethod {
	case "/v1.user.UserService/Register", "/v1.user.UserService/Login":
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata in context")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no token provided")
	}

	userID, err := auth.ParseAndValidate(tokens[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Add user ID to context
	newCtx := context.WithValue(ctx, userIDKey{}, userID)
	return handler(newCtx, req)
}

// StreamAuthInterceptor validates JWT for streaming RPCs.
func StreamAuthInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// Extract token from metadata
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no metadata in context")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return status.Error(codes.Unauthenticated, "no token provided")
	}

	userID, err := auth.ParseAndValidate(tokens[0])
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Wrap stream with user ID context
	wrapped := &wrappedStream{
		ServerStream: ss,
		ctx:          context.WithValue(ctx, userIDKey{}, userID),
	}
	return handler(srv, wrapped)
}

type wrappedStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// ContextWithUserID creates a context with user ID for testing purposes.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}
