package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// LoggingUnaryInterceptor logs gRPC unary requests and responses.
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Log incoming request
		reqJSON := formatMessage(req)
		logger.Info("gRPC request",
			"method", info.FullMethod,
			"request", reqJSON,
		)

		// Call the handler
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log response
		respJSON := formatMessage(resp)
		grpcCode := codes.OK
		errMsg := ""
		if err != nil {
			if st, ok := status.FromError(err); ok {
				grpcCode = st.Code()
				errMsg = st.Message()
			} else {
				grpcCode = codes.Internal
				errMsg = err.Error()
			}
		}

		logger.Info("gRPC response",
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", grpcCode.String(),
			"error", errMsg,
			"response", respJSON,
		)

		return resp, err
	}
}

// LoggingStreamInterceptor logs gRPC streaming requests.
func LoggingStreamInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		logger.Info("gRPC stream started",
			"method", info.FullMethod,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		// Wrap the server stream to log messages
		wrappedStream := &loggingServerStream{
			ServerStream: ss,
			logger:       logger,
			method:       info.FullMethod,
		}

		// Call the handler
		err := handler(srv, wrappedStream)
		duration := time.Since(start)

		grpcCode := codes.OK
		errMsg := ""
		if err != nil {
			if st, ok := status.FromError(err); ok {
				grpcCode = st.Code()
				errMsg = st.Message()
			} else {
				grpcCode = codes.Internal
				errMsg = err.Error()
			}
		}

		logger.Info("gRPC stream completed",
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", grpcCode.String(),
			"error", errMsg,
		)

		return err
	}
}

// loggingServerStream wraps grpc.ServerStream to log messages.
type loggingServerStream struct {
	grpc.ServerStream

	logger *slog.Logger
	method string
}

func (s *loggingServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		msgJSON := formatMessage(m)
		s.logger.Debug("gRPC stream received message",
			"method", s.method,
			"message", msgJSON,
		)
	}
	return err
}

func (s *loggingServerStream) SendMsg(m interface{}) error {
	msgJSON := formatMessage(m)
	s.logger.Debug("gRPC stream sending message",
		"method", s.method,
		"message", msgJSON,
	)
	return s.ServerStream.SendMsg(m)
}

// formatMessage formats a protobuf message or any interface{} for logging.
func formatMessage(msg interface{}) string {
	if msg == nil {
		return "null"
	}

	// Try to format as protobuf message first
	if pbMsg, ok := msg.(proto.Message); ok {
		if jsonBytes, err := protojson.Marshal(pbMsg); err == nil {
			// Pretty print JSON for better readability
			var prettyJSON interface{}
			if err := json.Unmarshal(jsonBytes, &prettyJSON); err == nil {
				if formatted, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
					return string(formatted)
				}
			}
			return string(jsonBytes)
		}
	}

	// Fallback to regular JSON marshaling
	if jsonBytes, err := json.MarshalIndent(msg, "", "  "); err == nil {
		return string(jsonBytes)
	}

	// Last resort: string representation
	return fmt.Sprintf("%+v", msg)
}

// LoggingConfig holds configuration for the logging interceptor.
type LoggingConfig struct {
	// LogPayloads controls whether request/response payloads are logged
	LogPayloads bool
	// LogLevel sets the log level for request/response logging
	LogLevel slog.Level
	// MaxPayloadSize limits the size of logged payloads (0 = unlimited)
	MaxPayloadSize int
}

// ConfigurableLoggingUnaryInterceptor creates a logging interceptor with custom configuration.
func ConfigurableLoggingUnaryInterceptor(logger *slog.Logger, config LoggingConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Log incoming request
		if config.LogPayloads {
			reqJSON := formatMessageWithLimit(req, config.MaxPayloadSize)
			logger.Log(ctx, config.LogLevel, "gRPC request",
				"method", info.FullMethod,
				"request", reqJSON,
			)
		} else {
			logger.Log(ctx, config.LogLevel, "gRPC request",
				"method", info.FullMethod,
			)
		}

		// Call the handler
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log response
		grpcCode := codes.OK
		errMsg := ""
		if err != nil {
			if st, ok := status.FromError(err); ok {
				grpcCode = st.Code()
				errMsg = st.Message()
			} else {
				grpcCode = codes.Internal
				errMsg = err.Error()
			}
		}

		if config.LogPayloads {
			respJSON := formatMessageWithLimit(resp, config.MaxPayloadSize)
			logger.Log(ctx, config.LogLevel, "gRPC response",
				"method", info.FullMethod,
				"duration", duration.String(),
				"code", grpcCode.String(),
				"error", errMsg,
				"response", respJSON,
			)
		} else {
			logger.Log(ctx, config.LogLevel, "gRPC response",
				"method", info.FullMethod,
				"duration", duration.String(),
				"code", grpcCode.String(),
				"error", errMsg,
			)
		}

		return resp, err
	}
}

// formatMessageWithLimit formats a message with size limit.
func formatMessageWithLimit(msg interface{}, maxSize int) string {
	formatted := formatMessage(msg)
	if maxSize > 0 && len(formatted) > maxSize {
		return formatted[:maxSize] + "...[truncated]"
	}
	return formatted
}
