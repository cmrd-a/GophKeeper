# gRPC Logging Interceptor Examples

This document provides practical examples of how the gRPC logging interceptor works in the GophKeeper application.

## Configuration Examples

### Environment Variables (.env file)

```bash
# Basic configuration
LOG_GRPC_REQUESTS=true
LOG_GRPC_PAYLOADS=false
MAX_LOG_PAYLOAD_SIZE=500

# Development configuration (verbose)
LOG_GRPC_REQUESTS=true
LOG_GRPC_PAYLOADS=true
MAX_LOG_PAYLOAD_SIZE=2000

# Production configuration (minimal)
LOG_GRPC_REQUESTS=true
LOG_GRPC_PAYLOADS=false
MAX_LOG_PAYLOAD_SIZE=100
```

## Example Log Output

### User Registration Request

**Request:**
```json
{
  "time": "2025-10-11T21:05:21.036+03:00",
  "level": "INFO",
  "msg": "gRPC request",
  "method": "/gophkeeper.v1.user.UserService/Register",
  "request": {
    "login": "john.doe@example.com",
    "password": "********"
  }
}
```

**Response (Success):**
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

**Response (Error):**
```json
{
  "time": "2025-10-11T21:05:21.256+03:00",
  "level": "INFO",
  "msg": "gRPC response",
  "method": "/gophkeeper.v1.user.UserService/Register",
  "duration": "45ms",
  "code": "ALREADY_EXISTS",
  "error": "user already exists",
  "response": null
}
```

### Vault Operations

**Save Login Password Request:**
```json
{
  "time": "2025-10-11T21:06:15.234+03:00",
  "level": "INFO",
  "msg": "gRPC request",
  "method": "/gophkeeper.v1.vault.VaultService/SaveLoginPassword",
  "request": {
    "login": "admin@company.com",
    "password": "securePassword123"
  }
}
```

**Save Login Password Response:**
```json
{
  "time": "2025-10-11T21:06:15.289+03:00",
  "level": "INFO",
  "msg": "gRPC response",
  "method": "/gophkeeper.v1.vault.VaultService/SaveLoginPassword",
  "duration": "55ms",
  "code": "OK",
  "error": "",
  "response": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Get Vault Items Request:**
```json
{
  "time": "2025-10-11T21:07:30.123+03:00",
  "level": "INFO",
  "msg": "gRPC request",
  "method": "/gophkeeper.v1.vault.VaultService/GetVaultItems",
  "request": {}
}
```

**Get Vault Items Response:**
```json
{
  "time": "2025-10-11T21:07:30.234+03:00",
  "level": "INFO",
  "msg": "gRPC response",
  "method": "/gophkeeper.v1.vault.VaultService/GetVaultItems",
  "duration": "111ms",
  "code": "OK",
  "error": "",
  "response": {
    "loginPasswords": [
      {
        "base": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "createdAt": "2025-10-11T18:06:15Z",
          "updatedAt": "2025-10-11T18:06:15Z"
        },
        "login": "admin@company.com",
        "password": "securePassword123"
      }
    ],
    "textData": [],
    "binaryData": [],
    "cardData": []
  }
}
```

### Authentication Failures

**Missing Token:**
```json
{
  "time": "2025-10-11T21:08:45.567+03:00",
  "level": "INFO",
  "msg": "gRPC request",
  "method": "/gophkeeper.v1.vault.VaultService/GetVaultItems",
  "request": {}
}
```

```json
{
  "time": "2025-10-11T21:08:45.568+03:00",
  "level": "INFO",
  "msg": "gRPC response",
  "method": "/gophkeeper.v1.vault.VaultService/GetVaultItems",
  "duration": "1ms",
  "code": "UNAUTHENTICATED",
  "error": "no token provided",
  "response": null
}
```

### Large Payload Truncation

**Request with Large Payload (truncated):**
```json
{
  "time": "2025-10-11T21:09:12.345+03:00",
  "level": "INFO",
  "msg": "gRPC request",
  "method": "/gophkeeper.v1.vault.VaultService/SaveBinaryData",
  "request": {
    "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==VBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==VBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==VBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ....[truncated]"
  }
}
```

## Stream Logging Examples

### Stream Start
```json
{
  "time": "2025-10-11T21:10:00.000+03:00",
  "level": "INFO",
  "msg": "gRPC stream started",
  "method": "/gophkeeper.v1.vault.VaultService/SyncItems",
  "is_client_stream": true,
  "is_server_stream": true
}
```

### Stream Messages (Debug Level)
```json
{
  "time": "2025-10-11T21:10:00.100+03:00",
  "level": "DEBUG",
  "msg": "gRPC stream received message",
  "method": "/gophkeeper.v1.vault.VaultService/SyncItems",
  "message": {
    "item": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "type": "login_password"
    }
  }
}
```

### Stream Completion
```json
{
  "time": "2025-10-11T21:10:05.500+03:00",
  "level": "INFO",
  "msg": "gRPC stream completed",
  "method": "/gophkeeper.v1.vault.VaultService/SyncItems",
  "duration": "5.5s",
  "code": "OK",
  "error": ""
}
```

## Performance Impact Analysis

### Typical Request Processing Times

**Without Logging:**
- Simple requests: 1-5ms
- Database operations: 10-50ms
- Complex operations: 100-500ms

**With Request/Response Logging (no payloads):**
- Overhead: +0.1-0.5ms
- Impact: Negligible for most operations

**With Full Payload Logging:**
- Small payloads (<1KB): +0.5-1ms
- Medium payloads (1-10KB): +1-5ms
- Large payloads (>10KB): +5-20ms

### Recommendations by Environment

**Development:**
```bash
LOG_GRPC_REQUESTS=true
LOG_GRPC_PAYLOADS=true
MAX_LOG_PAYLOAD_SIZE=0  # Unlimited
```

**Staging:**
```bash
LOG_GRPC_REQUESTS=true
LOG_GRPC_PAYLOADS=true
MAX_LOG_PAYLOAD_SIZE=1000
```

**Production:**
```bash
LOG_GRPC_REQUESTS=true
LOG_GRPC_PAYLOADS=false
MAX_LOG_PAYLOAD_SIZE=100
```

## Security Considerations

### Sensitive Data in Logs

Be aware that the following methods may log sensitive data:

1. **User Registration/Login:**
   - Passwords (consider masking)
   - Email addresses
   - Personal information

2. **Vault Operations:**
   - Stored passwords
   - Credit card numbers
   - Private notes
   - Binary data (files, images)

### Recommended Filtering

Consider implementing custom formatters to mask sensitive fields:

```go
// Example of custom message formatter
func secureFormatMessage(msg any) string {
    formatted := formatMessage(msg)
    // Mask password fields
    re := regexp.MustCompile(`"password":\s*"[^"]*"`)
    formatted = re.ReplaceAllString(formatted, `"password": "[REDACTED]"`)
    return formatted
}
```

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Request Rate:** Requests per second by method
2. **Error Rate:** Percentage of failed requests
3. **Response Time:** P50, P95, P99 latencies
4. **Error Codes:** Distribution of gRPC status codes

### Sample Log Queries

**Find slow requests:**
```bash
grep "gRPC response" logs.json | jq 'select(.duration | tonumber > 1000)'
```

**Count errors by method:**
```bash
grep "gRPC response" logs.json | jq 'select(.code != "OK")' | jq -r '.method' | sort | uniq -c
```

**Average response times:**
```bash
grep "gRPC response" logs.json | jq '.duration' | sed 's/ms//' | awk '{sum+=$1; count++} END {print "Average:", sum/count, "ms"}'
```
