package interceptor

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

func TestLoggingUnaryInterceptor(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	interceptor := LoggingUnaryInterceptor(logger)

	tests := []struct {
		name          string
		handler       grpc.UnaryHandler
		request       interface{}
		expectedError error
	}{
		{
			name: "Success",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return &vault.GetVaultItemsResponse{}, nil
			},
			request:       &vault.GetVaultItemsRequest{},
			expectedError: nil,
		},
		{
			name: "Error",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, status.Error(codes.Internal, "test error")
			},
			request:       &vault.GetVaultItemsRequest{},
			expectedError: status.Error(codes.Internal, "test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset() // Clear buffer for each test

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			resp, err := interceptor(context.Background(), tt.request, info, tt.handler)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			// Check that logs were written
			logOutput := buf.String()
			assert.Contains(t, logOutput, "gRPC request")
			assert.Contains(t, logOutput, "gRPC response")
			assert.Contains(t, logOutput, "/test.Service/Method")
		})
	}
}

func TestLoggingStreamInterceptor(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	interceptor := LoggingStreamInterceptor(logger)

	tests := []struct {
		name          string
		handler       grpc.StreamHandler
		expectedError error
	}{
		{
			name: "Success",
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			},
			expectedError: nil,
		},
		{
			name: "Error",
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				return status.Error(codes.Internal, "stream error")
			},
			expectedError: status.Error(codes.Internal, "stream error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			mockStream := &mockServerStreamWithMethods{ctx: context.Background()}
			info := &grpc.StreamServerInfo{
				FullMethod:     "/test.Service/StreamMethod",
				IsClientStream: true,
				IsServerStream: true,
			}

			err := interceptor(nil, mockStream, info, tt.handler)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			logOutput := buf.String()
			assert.Contains(t, logOutput, "gRPC stream started")
			assert.Contains(t, logOutput, "gRPC stream completed")
			assert.Contains(t, logOutput, "/test.Service/StreamMethod")
		})
	}
}

func TestLoggingServerStream(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	mockStream := &mockServerStreamWithMethods{ctx: context.Background()}
	loggingStream := &loggingServerStream{
		ServerStream: mockStream,
		logger:       logger,
		method:       "/test.Service/StreamMethod",
	}

	t.Run("RecvMsg", func(t *testing.T) {
		buf.Reset()

		var msg vault.GetVaultItemsRequest
		err := loggingStream.RecvMsg(&msg)

		assert.NoError(t, err)
		logOutput := buf.String()
		assert.Contains(t, logOutput, "gRPC stream received message")
	})

	t.Run("SendMsg", func(t *testing.T) {
		buf.Reset()

		msg := &vault.GetVaultItemsResponse{}
		err := loggingStream.SendMsg(msg)

		assert.NoError(t, err)
		logOutput := buf.String()
		assert.Contains(t, logOutput, "gRPC stream sending message")
	})
}

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      interface{}
		contains []string
	}{
		{
			name:     "nil message",
			msg:      nil,
			contains: []string{"null"},
		},
		{
			name:     "proto message",
			msg:      &vault.GetVaultItemsRequest{},
			contains: []string{"{"},
		},
		{
			name:     "regular struct",
			msg:      struct{ Name string }{Name: "test"},
			contains: []string{"test"},
		},
		{
			name:     "string",
			msg:      "test string",
			contains: []string{"test string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMessage(tt.msg)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestConfigurableLoggingUnaryInterceptor(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	tests := []struct {
		name   string
		config LoggingConfig
	}{
		{
			name: "WithPayloads",
			config: LoggingConfig{
				LogPayloads:    true,
				LogLevel:       slog.LevelInfo,
				MaxPayloadSize: 0,
			},
		},
		{
			name: "WithoutPayloads",
			config: LoggingConfig{
				LogPayloads:    false,
				LogLevel:       slog.LevelInfo,
				MaxPayloadSize: 0,
			},
		},
		{
			name: "WithSizeLimit",
			config: LoggingConfig{
				LogPayloads:    true,
				LogLevel:       slog.LevelInfo,
				MaxPayloadSize: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			interceptor := ConfigurableLoggingUnaryInterceptor(logger, tt.config)
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return &vault.GetVaultItemsResponse{}, nil
			}

			resp, err := interceptor(context.Background(), &vault.GetVaultItemsRequest{}, info, handler)

			assert.NoError(t, err)
			assert.NotNil(t, resp)

			logOutput := buf.String()
			assert.Contains(t, logOutput, "gRPC request")
			assert.Contains(t, logOutput, "gRPC response")

			if tt.config.LogPayloads {
				assert.Contains(t, logOutput, "request")
				assert.Contains(t, logOutput, "response")
			}
		})
	}
}

func TestFormatMessageWithLimit(t *testing.T) {
	tests := []struct {
		name     string
		msg      interface{}
		maxSize  int
		expected string
	}{
		{
			name:     "no limit",
			msg:      "test message",
			maxSize:  0,
			expected: "test message",
		},
		{
			name:     "under limit",
			msg:      "test",
			maxSize:  10,
			expected: "test",
		},
		{
			name:     "over limit",
			msg:      "this is a very long test message",
			maxSize:  10,
			expected: "this is a ...[truncated]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMessageWithLimit(tt.msg, tt.maxSize)
			if tt.maxSize > 0 && len(formatMessage(tt.msg)) > tt.maxSize {
				assert.Contains(t, result, "...[truncated]")
				assert.True(t, len(result) > tt.maxSize) // Should be maxSize + len("...[truncated]")
			} else {
				assert.Equal(t, formatMessage(tt.msg), result)
			}
		})
	}
}

func TestLoggingInterceptorTiming(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	interceptor := LoggingUnaryInterceptor(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	// Handler that takes some time
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return &vault.GetVaultItemsResponse{}, nil
	}

	_, err := interceptor(context.Background(), &vault.GetVaultItemsRequest{}, info, handler)

	assert.NoError(t, err)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "duration")
	assert.Contains(t, logOutput, "ms") // Should contain milliseconds in duration
}

// mockServerStreamWithMethods implements the missing methods for ServerStream
type mockServerStreamWithMethods struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *mockServerStreamWithMethods) Context() context.Context {
	return s.ctx
}

func (s *mockServerStreamWithMethods) RecvMsg(m interface{}) error {
	// Mock implementation - just return nil to simulate successful receive
	return nil
}

func (s *mockServerStreamWithMethods) SendMsg(m interface{}) error {
	// Mock implementation - just return nil to simulate successful send
	return nil
}
