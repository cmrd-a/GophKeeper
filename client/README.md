# GophKeeper Client Package

A robust, testable Go client library for interacting with the GophKeeper secure vault server.

## Overview

The GophKeeper client package provides a clean, type-safe interface for authenticating with and managing vault items on a GophKeeper server. It includes comprehensive error handling, configurable timeouts, and extensive test coverage.

## Features

- 🔐 **Authentication**: User registration and login with JWT token management
- 📝 **Vault Operations**: Store and retrieve text data, login credentials, card data, and binary files
- 🏷️ **Metadata**: Associate custom metadata with vault items
- ⚡ **Configurable**: Customizable timeouts, TLS settings, and connection options
- 🧪 **Testable**: Interface-based design with comprehensive mocks and tests
- 🛡️ **Secure**: TLS encryption with support for custom certificates
- 🚀 **Concurrent**: Thread-safe operations with context support

## Installation

```bash
go get github.com/cmrd-a/GophKeeper/client
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/cmrd-a/GophKeeper/client"
)

func main() {
    // Create client with default configuration
    gophClient, err := client.NewClient(nil)
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }
    defer gophClient.Close()
    
    ctx := context.Background()
    
    // Register a new user
    err = gophClient.Register(ctx, "myusername", "mypassword")
    if err != nil {
        log.Fatal("Registration failed:", err)
    }
    
    // Login
    err = gophClient.Login(ctx, "myusername", "mypassword")
    if err != nil {
        log.Fatal("Login failed:", err)
    }
    
    // Save some text data
    textID, err := gophClient.SaveTextData(ctx, "My secret notes")
    if err != nil {
        log.Fatal("Failed to save text:", err)
    }
    
    log.Printf("Saved text data with ID: %s", textID)
    
    // Retrieve all vault items
    items, err := gophClient.GetVaultItems(ctx)
    if err != nil {
        log.Fatal("Failed to get vault items:", err)
    }
    
    log.Printf("Found %d text items", len(items.TextData))
}
```

## Configuration

### Basic Configuration

```go
config := &client.ClientConfig{
    ServerAddr:     "localhost:8082",
    ConnectTimeout: 30 * time.Second,
    RequestTimeout: 30 * time.Second,
    SkipConnTest:   false,
}

gophClient, err := client.NewClient(config)
```

### Advanced TLS Configuration

```go
config := &client.ClientConfig{
    ServerAddr:     "myserver.com:8082",
    ConnectTimeout: 10 * time.Second,
    RequestTimeout: 15 * time.Second,
    TLSConfig: &client.TLSConfig{
        CertPool:   &customCreds,  // Custom TLS credentials
        ServerName: "myserver.com", // Override server name for certificate validation
    },
}
```

### Default Configuration

```go
// Uses default settings:
// - ServerAddr: "localhost:8082"
// - ConnectTimeout: 30 seconds  
// - RequestTimeout: 30 seconds
// - SkipConnTest: false
// - TLS: Insecure development certificates
gophClient, err := client.NewClient(client.DefaultConfig())
```

## API Reference

### Authentication

#### Register
```go
err := client.Register(ctx, "username", "password")
```
Creates a new user account.

#### Login  
```go
err := client.Login(ctx, "username", "password")
```
Authenticates user and stores JWT token for subsequent requests.

#### Authentication Status
```go
if client.IsAuthenticated() {
    token := client.GetToken()
    // Use token as needed
}
```

### Vault Operations

#### Text Data
```go
// Save text
id, err := client.SaveTextData(ctx, "My important text")

// Text is retrieved via GetVaultItems()
```

#### Login Credentials
```go
// Save login/password pair
id, err := client.SaveLoginPassword(ctx, "mylogin", "mypassword")
```

#### Credit Card Data
```go
// Save card information (encrypted on server)
id, err := client.SaveCardData(ctx, "4111111111111111", "John Doe", "12/25", "123")
```

#### Binary Data
```go
// Save files or binary data
data, _ := os.ReadFile("document.pdf")
id, err := client.SaveBinaryData(ctx, data)
```

#### Metadata
```go
// Associate metadata with vault items
meta := []*vault.Meta{
    {Key: "category", Value: "personal"},
    {Key: "tags", Value: "important,secret"},
}
err := client.SaveMeta(ctx, meta)
```

#### Retrieve All Items
```go
items, err := client.GetVaultItems(ctx)
if err != nil {
    return err
}

// Access different types of data
for _, textItem := range items.TextData {
    fmt.Printf("Text ID: %s, Content: %s\n", textItem.Base.Id, textItem.Text)
}

for _, loginItem := range items.LoginPasswords {
    fmt.Printf("Login ID: %s, User: %s\n", loginItem.Base.Id, loginItem.Login)
}
```

#### Delete Items
```go
err := client.DeleteVaultItem(ctx, itemID, "text")
```

## Error Handling

The client returns detailed, wrapped errors for better debugging:

```go
err := client.Login(ctx, "user", "wrongpass")
if err != nil {
    if strings.Contains(err.Error(), "unauthenticated") {
        // Handle authentication error
    } else if strings.Contains(err.Error(), "context canceled") {
        // Handle timeout or cancellation  
    } else {
        // Handle other errors
    }
}
```

## Testing

The client package includes comprehensive test coverage with both unit and integration tests.

### Running Unit Tests

```bash
# Run all unit tests
go test ./client/

# Run with coverage
go test -cover ./client/

# Run specific tests
go test -run TestLogin ./client/
```

### Running Integration Tests

Integration tests require a running GophKeeper server:

```bash
# Start server first
./scripts/start-dev.sh

# Run integration tests (in another terminal)
go test -tags=integration ./client/

# Or use the test runner script
./scripts/run-tests.sh --integration-only
```

### Using Test Scripts

The package includes helper scripts for testing:

```bash
# Run all tests with coverage
./scripts/run-tests.sh

# Run only unit tests  
./scripts/run-tests.sh --unit-only

# Run only integration tests
./scripts/run-tests.sh --integration-only

# Verbose output
./scripts/run-tests.sh --verbose
```

### Mocking for Tests

The client implements an interface for easy mocking:

```go
type MockGophKeeperClient struct {
    mock.Mock
}

func (m *MockGophKeeperClient) Login(ctx context.Context, login, password string) error {
    args := m.Called(ctx, login, password)
    return args.Error(0)
}

// Use in tests
mockClient := &MockGophKeeperClient{}
mockClient.On("Login", mock.Anything, "user", "pass").Return(nil)
```

## Context and Cancellation

All operations support context for timeouts and cancellation:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := client.Login(ctx, "user", "pass")
if err == context.DeadlineExceeded {
    // Handle timeout
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(2 * time.Second)
    cancel() // Cancel the operation
}()

err := client.GetVaultItems(ctx)
```

## Security Considerations

### TLS Configuration
- The client uses TLS encryption for all communications
- Development mode uses self-signed certificates (insecure)
- Production should use proper TLS certificates

### Authentication
- JWT tokens are stored in memory only
- Tokens are automatically included in authenticated requests
- Re-authentication required if token expires

### Data Encryption  
- Sensitive data (passwords, card numbers, CVV) is encrypted on the server
- Binary data and text are transmitted securely over TLS
- Client does not perform additional encryption

## Performance

### Connection Pooling
- gRPC connections are reused across requests
- Single connection per client instance
- Close client when done to free resources

### Timeouts
- Configure appropriate timeouts based on network conditions
- Default timeouts are 30 seconds for most operations
- Use context timeouts for request-specific limits

### Concurrent Usage
- Client is thread-safe for concurrent operations
- Each goroutine can use the same client instance
- Authentication state is shared across goroutines

## Troubleshooting

### Connection Issues
```bash
# Check server connectivity
nc -zv localhost 8082

# Test with diagnostics
./scripts/run-client-debug.sh diag
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `server not reachable` | Server not running | Start server with `./scripts/start-dev.sh` |
| `context canceled` | Timeout or cancellation | Check network connectivity, increase timeout |
| `not authenticated` | No valid token | Call `Login()` first |
| `login failed` | Invalid credentials | Check username/password |
| `connection refused` | Wrong port/address | Verify server address configuration |

### Debug Logging

Enable debug logging for troubleshooting:

```go
// Set log level before creating client
os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "99")
os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
```

## Examples

### Complete Application Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/cmrd-a/GophKeeper/client"
    "github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

func main() {
    // Create client
    config := &client.ClientConfig{
        ServerAddr: os.Getenv("GOPHKEEPER_SERVER"),
    }
    if config.ServerAddr == "" {
        config.ServerAddr = "localhost:8082"
    }
    
    gophClient, err := client.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer gophClient.Close()
    
    ctx := context.Background()
    
    // Register and login
    username := "demo_user"
    password := "secure_password_123"
    
    if err := gophClient.Register(ctx, username, password); err != nil {
        log.Printf("Registration failed (user may already exist): %v", err)
    }
    
    if err := gophClient.Login(ctx, username, password); err != nil {
        log.Fatal("Login failed:", err)
    }
    
    fmt.Printf("✅ Authenticated as %s\n", username)
    
    // Save various types of data
    textID, _ := gophClient.SaveTextData(ctx, "My secret notes")
    loginID, _ := gophClient.SaveLoginPassword(ctx, "gmail", "my_gmail_pass")
    cardID, _ := gophClient.SaveCardData(ctx, "4111111111111111", "John Doe", "12/25", "123")
    
    // Add metadata
    meta := []*vault.Meta{
        {Key: "category", Value: "personal", ItemId: textID},
        {Key: "service", Value: "gmail", ItemId: loginID},
        {Key: "bank", Value: "Chase", ItemId: cardID},
    }
    gophClient.SaveMeta(ctx, meta)
    
    // Retrieve and display all items
    items, err := gophClient.GetVaultItems(ctx)
    if err != nil {
        log.Fatal("Failed to get vault items:", err)
    }
    
    fmt.Printf("\n📋 Vault Summary:\n")
    fmt.Printf("  Text items: %d\n", len(items.TextData))
    fmt.Printf("  Login credentials: %d\n", len(items.LoginPasswords))
    fmt.Printf("  Credit cards: %d\n", len(items.CardData))
    fmt.Printf("  Binary files: %d\n", len(items.BinaryData))
    
    // Clean up - delete a text item
    if len(items.TextData) > 0 {
        err := gophClient.DeleteVaultItem(ctx, items.TextData[0].Base.Id, "text")
        if err != nil {
            log.Printf("Failed to delete item: %v", err)
        } else {
            fmt.Println("🗑️  Deleted text item")
        }
    }
}
```

### Batch Operations Example

```go
func batchSaveExample(client client.GophKeeperClient) error {
    ctx := context.Background()
    
    // Save multiple items concurrently
    const numItems = 10
    errChan := make(chan error, numItems)
    idChan := make(chan string, numItems)
    
    for i := 0; i < numItems; i++ {
        go func(index int) {
            text := fmt.Sprintf("Batch item %d", index)
            id, err := client.SaveTextData(ctx, text)
            errChan <- err
            if err == nil {
                idChan <- id
            }
        }(i)
    }
    
    // Collect results
    var ids []string
    for i := 0; i < numItems; i++ {
        if err := <-errChan; err != nil {
            return fmt.Errorf("batch save failed: %w", err)
        }
        ids = append(ids, <-idChan)
    }
    
    fmt.Printf("Successfully saved %d items\n", len(ids))
    return nil
}
```

## Development

### Building

```bash
# Build client package
go build ./client/

# Build with race detection
go build -race ./client/

# Cross-compile
GOOS=windows GOARCH=amd64 go build ./client/
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality  
4. Ensure all tests pass: `./scripts/run-tests.sh`
5. Submit a pull request

### Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comprehensive error handling
- Include unit tests for new features
- Document public APIs with comments

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- 📖 [Documentation](./TROUBLESHOOTING.md)
- 🐛 [Issue Tracker](https://github.com/cmrd-a/GophKeeper/issues)  
- 💬 [Discussions](https://github.com/cmrd-a/GophKeeper/discussions)

---

**Built with ❤️ for secure credential management**