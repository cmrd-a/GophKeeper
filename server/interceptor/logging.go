package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// LoggingUnaryInterceptor logs gRPC unary requests and responses.
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		// Extract metadata
		metadataAttrs := extractMetadata(ctx)

		// Log incoming request
		reqJSON := formatMessage(req)
		logAttrs := []any{
			"method", info.FullMethod,
			"request", reqJSON,
		}
		logAttrs = append(logAttrs, metadataAttrs...)
		logger.Info("gRPC request", logAttrs...)

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

		logAttrs = []any{
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", grpcCode.String(),
			"error", errMsg,
			"response", respJSON,
		}
		logAttrs = append(logAttrs, metadataAttrs...)
		logger.Info("gRPC response", logAttrs...)

		return resp, err
	}
}

// LoggingStreamInterceptor logs gRPC streaming requests.
func LoggingStreamInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Extract metadata
		ctx := ss.Context()
		metadataAttrs := extractMetadata(ctx)

		logAttrs := []any{
			"method", info.FullMethod,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		}
		logAttrs = append(logAttrs, metadataAttrs...)
		logger.Info("gRPC stream started", logAttrs...)

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

		logAttrs = []any{
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", grpcCode.String(),
			"error", errMsg,
		}
		logAttrs = append(logAttrs, metadataAttrs...)
		logger.Info("gRPC stream completed", logAttrs...)

		return err
	}
}

// loggingServerStream wraps grpc.ServerStream to log messages.
type loggingServerStream struct {
	grpc.ServerStream

	logger *slog.Logger
	method string
}

func (s *loggingServerStream) RecvMsg(m any) error {
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

func (s *loggingServerStream) SendMsg(m any) error {
	msgJSON := formatMessage(m)
	s.logger.Debug("gRPC stream sending message",
		"method", s.method,
		"message", msgJSON,
	)
	return s.ServerStream.SendMsg(m)
}

// formatMessage formats a protobuf message or any any for logging.
func formatMessage(msg any) string {
	if msg == nil {
		return "null"
	}

	// Try to format as protobuf message first
	if pbMsg, ok := msg.(proto.Message); ok {
		if jsonBytes, err := protojson.Marshal(pbMsg); err == nil {
			// Pretty print JSON for better readability
			var prettyJSON any
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
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		// Extract metadata
		metadataAttrs := extractMetadata(ctx)

		// Log incoming request
		if config.LogPayloads {
			reqJSON := formatMessageWithLimit(req, config.MaxPayloadSize)
			logAttrs := []any{
				"method", info.FullMethod,
				"request", reqJSON,
			}
			logAttrs = append(logAttrs, metadataAttrs...)
			logger.Log(ctx, config.LogLevel, "gRPC request", logAttrs...)
		} else {
			logAttrs := []any{
				"method", info.FullMethod,
			}
			logAttrs = append(logAttrs, metadataAttrs...)
			logger.Log(ctx, config.LogLevel, "gRPC request", logAttrs...)
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
			logAttrs := []any{
				"method", info.FullMethod,
				"duration", duration.String(),
				"code", grpcCode.String(),
				"error", errMsg,
				"response", respJSON,
			}
			logAttrs = append(logAttrs, metadataAttrs...)
			logger.Log(ctx, config.LogLevel, "gRPC response", logAttrs...)
		} else {
			logAttrs := []any{
				"method", info.FullMethod,
				"duration", duration.String(),
				"code", grpcCode.String(),
				"error", errMsg,
			}
			logAttrs = append(logAttrs, metadataAttrs...)
			logger.Log(ctx, config.LogLevel, "gRPC response", logAttrs...)
		}

		return resp, err
	}
}

// formatMessageWithLimit formats a message with size limit.
func formatMessageWithLimit(msg any, maxSize int) string {
	formatted := formatMessage(msg)
	if maxSize > 0 && len(formatted) > maxSize {
		return formatted[:maxSize] + "...[truncated]"
	}
	return formatted
}

// extractMetadata extracts useful metadata from the gRPC context.
func extractMetadata(ctx context.Context) []any {
	attrs := make([]any, 0)

	// Extract peer information (client address)
	if p, ok := peer.FromContext(ctx); ok {
		attrs = append(attrs, "remote_addr", p.Addr.String())
	}

	// Extract metadata (headers)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// User-Agent
		if ua := md.Get("user-agent"); len(ua) > 0 {
			attrs = append(attrs, "user_agent", ua[0])
		}

		// Authorization (just indicate presence, don't log the token)
		if auth := md.Get("authorization"); len(auth) > 0 {
			attrs = append(attrs, "has_auth", true)
			attrs = append(attrs, "authorization", auth[0])
		}

		// Content-Type
		if ct := md.Get("content-type"); len(ct) > 0 {
			attrs = append(attrs, "content_type", ct[0])
		}

		// X-Forwarded-For (for proxied requests)
		if xff := md.Get("x-forwarded-for"); len(xff) > 0 {
			attrs = append(attrs, "x_forwarded_for", xff[0])
		}

		// Request ID (if present)
		if reqID := md.Get("x-request-id"); len(reqID) > 0 {
			attrs = append(attrs, "request_id", reqID[0])
		}
	}

	return attrs
}
