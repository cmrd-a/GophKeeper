package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
	"github.com/cmrd-a/GophKeeper/server/insecure"
)

// GophKeeperClient defines the interface for interacting with the GophKeeper server
type GophKeeperClient interface {
	// Authentication
	Login(ctx context.Context, login, password string) error
	Register(ctx context.Context, login, password string) error
	GetToken() string
	IsAuthenticated() bool

	// Vault operations
	GetVaultItems(ctx context.Context) (*vault.GetVaultItemsResponse, error)
	SaveLoginPassword(ctx context.Context, login, password string) (string, error)
	SaveTextData(ctx context.Context, text string) (string, error)
	SaveCardData(ctx context.Context, number, holder, expire, cvv string) (string, error)
	SaveBinaryData(ctx context.Context, data []byte) (string, error)
	SaveMeta(ctx context.Context, meta []*vault.Meta) error
	DeleteVaultItem(ctx context.Context, id, itemType string) error

	// Connection management
	Close() error
}

// Client implements the GophKeeperClient interface
type Client struct {
	conn        *grpc.ClientConn
	userClient  user.UserServiceClient
	vaultClient vault.VaultServiceClient
	token       string
	serverAddr  string
}

// ClientConfig holds configuration for the client
type ClientConfig struct {
	ServerAddr     string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	SkipConnTest   bool
	TLSConfig      *TLSConfig
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertPool   *credentials.TransportCredentials
	ServerName string
}

// DefaultConfig returns a default client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		ServerAddr:     "localhost:8082",
		ConnectTimeout: 30 * time.Second,
		RequestTimeout: 30 * time.Second,
		SkipConnTest:   false,
		TLSConfig: &TLSConfig{
			CertPool:   nil, // Will use insecure creds
			ServerName: "",
		},
	}
}

// NewClient creates a new GophKeeper client
func NewClient(config *ClientConfig) (GophKeeperClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Test basic network connectivity if not skipped
	if !config.SkipConnTest {
		if err := testConnectivity(config.ServerAddr, 5*time.Second); err != nil {
			return nil, fmt.Errorf("server not reachable at %s: %w", config.ServerAddr, err)
		}
	}

	// Setup TLS credentials
	var creds credentials.TransportCredentials
	if config.TLSConfig != nil && config.TLSConfig.CertPool != nil {
		creds = *config.TLSConfig.CertPool
	} else {
		// Use insecure credentials for development
		serverName := ""
		if config.TLSConfig != nil {
			serverName = config.TLSConfig.ServerName
		}
		creds = credentials.NewClientTLSFromCert(insecure.CertPool, serverName)
	}

	// Setup dial options
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	// TODO: Replace deprecated grpc.DialContext when stable alternative is available
	grpcConn, err := grpc.NewClient( config.ServerAddr, opts...) //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("failed to dial server via gRPC: %w", err)
	}

	return &Client{
		conn:        grpcConn,
		userClient:  user.NewUserServiceClient(grpcConn),
		vaultClient: vault.NewVaultServiceClient(grpcConn),
		serverAddr:  config.ServerAddr,
	}, nil
}

// NewClientWithConn creates a client with an existing gRPC connection (useful for testing)
func NewClientWithConn(conn *grpc.ClientConn, serverAddr string) GophKeeperClient {
	return &Client{
		conn:        conn,
		userClient:  user.NewUserServiceClient(conn),
		vaultClient: vault.NewVaultServiceClient(conn),
		serverAddr:  serverAddr,
	}
}

// testConnectivity tests basic TCP connectivity to the server
func testConnectivity(serverAddr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", serverAddr, timeout)
	if err != nil {
		return fmt.Errorf("connection failed: %w (is the server running?)", err)
	}
	conn.Close()
	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// GetToken returns the current authentication token
func (c *Client) GetToken() string {
	return c.token
}

// IsAuthenticated returns true if the client has a valid token
func (c *Client) IsAuthenticated() bool {
	return c.token != ""
}

// Login authenticates the user and stores the token
func (c *Client) Login(ctx context.Context, login, password string) error {
	if login == "" || password == "" {
		return fmt.Errorf("login and password cannot be empty")
	}

	// Use provided context or create one with timeout
	loginCtx := ctx
	if ctx == context.Background() || ctx == nil {
		var cancel context.CancelFunc
		loginCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	resp, err := c.userClient.Login(loginCtx, &user.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if resp.Token == "" {
		return fmt.Errorf("login failed: received empty token")
	}

	c.token = resp.Token
	return nil
}

// Register creates a new user account
func (c *Client) Register(ctx context.Context, login, password string) error {
	if login == "" || password == "" {
		return fmt.Errorf("login and password cannot be empty")
	}

	// Use provided context or create one with timeout
	regCtx := ctx
	if ctx == context.Background() || ctx == nil {
		var cancel context.CancelFunc
		regCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	_, err := c.userClient.Register(regCtx, &user.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	return nil
}

// GetAuthContext returns a context with authentication metadata
func (c *Client) GetAuthContext(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}

	md := metadata.Pairs("authorization", "Bearer "+c.token)
	return metadata.NewOutgoingContext(ctx, md)
}

// GetVaultItems retrieves all vault items
func (c *Client) GetVaultItems(ctx context.Context) (*vault.GetVaultItemsResponse, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	authCtx := c.GetAuthContext(ctx)
	resp, err := c.vaultClient.GetVaultItems(authCtx, &vault.GetVaultItemsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get vault items: %w", err)
	}

	return resp, nil
}

// SaveLoginPassword saves login/password data
func (c *Client) SaveLoginPassword(ctx context.Context, login, password string) (string, error) {
	if !c.IsAuthenticated() {
		return "", fmt.Errorf("not authenticated")
	}

	if login == "" || password == "" {
		return "", fmt.Errorf("login and password cannot be empty")
	}

	authCtx := c.GetAuthContext(ctx)
	resp, err := c.vaultClient.SaveLoginPassword(authCtx, &vault.SaveLoginPasswordRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save login password: %w", err)
	}

	return resp.Id, nil
}

// SaveTextData saves text data
func (c *Client) SaveTextData(ctx context.Context, text string) (string, error) {
	if !c.IsAuthenticated() {
		return "", fmt.Errorf("not authenticated")
	}

	if text == "" {
		return "", fmt.Errorf("text cannot be empty")
	}

	authCtx := c.GetAuthContext(ctx)
	resp, err := c.vaultClient.SaveTextData(authCtx, &vault.SaveTextDataRequest{
		Text: text,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save text data: %w", err)
	}

	return resp.Id, nil
}

// SaveCardData saves credit card data
func (c *Client) SaveCardData(ctx context.Context, number, holder, expire, cvv string) (string, error) {
	if !c.IsAuthenticated() {
		return "", fmt.Errorf("not authenticated")
	}

	if number == "" || holder == "" || expire == "" || cvv == "" {
		return "", fmt.Errorf("all card fields are required")
	}

	authCtx := c.GetAuthContext(ctx)
	resp, err := c.vaultClient.SaveCardData(authCtx, &vault.SaveCardDataRequest{
		Number: number,
		Holder: holder,
		Expire: expire,
		Cvv:    cvv,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save card data: %w", err)
	}

	return resp.Id, nil
}

// SaveBinaryData saves binary data
func (c *Client) SaveBinaryData(ctx context.Context, data []byte) (string, error) {
	if !c.IsAuthenticated() {
		return "", fmt.Errorf("not authenticated")
	}

	if len(data) == 0 {
		return "", fmt.Errorf("data cannot be empty")
	}

	authCtx := c.GetAuthContext(ctx)
	resp, err := c.vaultClient.SaveBinaryData(authCtx, &vault.SaveBinaryDataRequest{
		Data: data,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save binary data: %w", err)
	}

	return resp.Id, nil
}

// SaveMeta saves metadata
func (c *Client) SaveMeta(ctx context.Context, meta []*vault.Meta) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	if len(meta) == 0 {
		return fmt.Errorf("meta cannot be empty")
	}

	authCtx := c.GetAuthContext(ctx)
	_, err := c.vaultClient.SaveMeta(authCtx, &vault.SaveMetaRequest{
		Meta: meta,
	})
	if err != nil {
		return fmt.Errorf("failed to save meta: %w", err)
	}

	return nil
}

// DeleteVaultItem deletes a vault item
func (c *Client) DeleteVaultItem(ctx context.Context, id, itemType string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	if id == "" || itemType == "" {
		return fmt.Errorf("id and itemType cannot be empty")
	}

	authCtx := c.GetAuthContext(ctx)
	_, err := c.vaultClient.DeleteVaultItem(authCtx, &vault.DeleteVaultItemRequest{
		Id:   id,
		Type: itemType,
	})
	if err != nil {
		return fmt.Errorf("failed to delete vault item: %w", err)
	}

	return nil
}
