# gRPC Interceptors

This package provides gRPC interceptors for the GophKeeper server, including authentication and request/response logging capabilities.

## Logging Interceptor

The logging interceptor provides comprehensive request and response logging for gRPC services.

### Features

- **Request/Response Logging**: Logs all incoming requests and outgoing responses
- **Payload Logging**: Optional logging of request/response payloads
- **Duration Tracking**: Measures and logs request processing time
- **Error Logging**: Captures and logs gRPC errors with proper status codes
- **Stream Support**: Handles both unary and streaming RPCs
- **Configurable**: Multiple configuration options for different environments

### Basic Usage

```go
import (
    "log/slog"
    "github.com/cmrd-a/GophKeeper/server/interceptor"
)

// Simple logging interceptor
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
loggingInterceptor := interceptor.LoggingUnaryInterceptor(logger)

// Configure gRPC server
opts := []grpc.ServerOption{
    grpc.UnaryInterceptor(loggingInterceptor),
}
server := grpc.NewServer(opts...)
```

### Advanced Configuration

```go
// Configurable logging interceptor
config := interceptor.LoggingConfig{
    LogPayloads:    true,              // Log request/response payloads
    LogLevel:       slog.LevelInfo,    // Set log level
    MaxPayloadSize: 1000,              // Limit payload size (0 = unlimited)
}

loggingInterceptor := interceptor.ConfigurableLoggingUnaryInterceptor(logger, config)
```

### Configuration Options

The `LoggingConfig` struct supports the following options:

- `LogPayloads` (bool): Whether to include request/response payloads in logs
- `LogLevel` (slog.Level): The log level to use for request/response logging
- `MaxPayloadSize` (int): Maximum size of logged payloads in characters (0 = unlimited)

### Environment Variables

When using the GophKeeper server, you can configure logging through environment variables:

```bash
# Enable gRPC request logging
LOG_GRPC_REQUESTS=true

# Enable payload logging (be careful with sensitive data)
LOG_GRPC_PAYLOADS=false

# Set maximum payload size for logging
MAX_LOG_PAYLOAD_SIZE=500
```

### Example Log Output

#### Request Log
```json
{
  "time": "2025-10-11T21:05:21.036+03:00",
  "level": "INFO",
  "msg": "gRPC request",
  "method": "/gophkeeper.v1.user.UserService/Register",
  "request": {
    "login": "user@example.com",
    "password": "[REDACTED]"
  }
}
```

#### Response Log
```json
{
  "time": "2025-10-11T21:05:21.156+03:00",
  "level": "INFO",
  "msg": "gRPC response",
  "method": "/gophkeeper.v1.user.UserService/Register",
  "duration": "120ms",
  "code": "OK",
  "error": "",
  "response": {}
}
```

#### Error Log
```json
{
  "time": "2025-10-11T21:05:21.256+03:00",
  "level": "INFO",
  "msg": "gRPC response",
  "method": "/gophkeeper.v1.user.UserService/Login",
  "duration": "45ms",
  "code": "UNAUTHENTICATED",
  "error": "invalid credentials",
  "response": null
}
```

### Stream Logging

For streaming RPCs, the interceptor provides:

- Stream start/completion logging
- Individual message logging (at DEBUG level)
- Duration tracking for entire stream lifetime

```json
{
  "time": "2025-10-11T21:05:21.036+03:00",
  "level": "INFO",
  "msg": "gRPC stream started",
  "method": "/gophkeeper.v1.vault.VaultService/SyncItems",
  "is_client_stream": true,
  "is_server_stream": true
}
```

### Security Considerations

**Important**: Be careful when enabling payload logging in production environments:

1. **Sensitive Data**: Payloads may contain passwords, tokens, or other sensitive information
2. **Performance**: Logging large payloads can impact performance
3. **Storage**: Detailed logs require more storage space
4. **Privacy**: Consider data privacy regulations when logging user data

### Recommendations

- **Development**: Enable full payload logging for debugging
- **Staging**: Enable request logging without payloads
- **Production**: Use minimal logging or disable payload logging entirely
- **Monitoring**: Use structured logging for better observability

### Integration with Auth Interceptor

The logging and authentication interceptors work together using interceptor chaining:

```go
// Logging runs first, then authentication
unaryChain := chainUnaryInterceptors(
    interceptor.LoggingUnaryInterceptor(log),
    interceptor.AuthInterceptor,
)
```

This ensures that all requests are logged, including authentication failures.

### Testing

The interceptor includes comprehensive tests covering:

- Successful requests/responses
- Error handling
- Stream operations
- Message formatting
- Configuration options

Run tests with:
```bash
go test -tags=unit ./server/interceptor
```

### Message Formatting

The interceptor automatically formats different types of messages:

1. **Protocol Buffer Messages**: Formatted as pretty-printed JSON
2. **Regular Structs**: JSON marshaled with indentation
3. **Primitive Types**: String representation
4. **Nil Values**: Logged as "null"

Large messages are automatically truncated based on the `MaxPayloadSize` configuration.