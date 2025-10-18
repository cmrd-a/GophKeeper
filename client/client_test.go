package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

// MockUserServiceClient implements user.UserServiceClient for testing
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) Login(
	ctx context.Context, req *user.LoginRequest, opts ...grpc.CallOption,
) (*user.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func (m *MockUserServiceClient) Register(
	ctx context.Context, req *user.RegisterRequest, opts ...grpc.CallOption,
) (*user.RegisterResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*user.RegisterResponse), args.Error(1)
}

// MockVaultServiceClient implements vault.VaultServiceClient for testing
type MockVaultServiceClient struct {
	mock.Mock
}

func (m *MockVaultServiceClient) GetVaultItems(
	ctx context.Context, req *vault.GetVaultItemsRequest, opts ...grpc.CallOption,
) (*vault.GetVaultItemsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.GetVaultItemsResponse), args.Error(1)
}

func (m *MockVaultServiceClient) SaveLoginPassword(
	ctx context.Context, req *vault.SaveLoginPasswordRequest, opts ...grpc.CallOption,
) (*vault.SaveLoginPasswordResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.SaveLoginPasswordResponse), args.Error(1)
}

func (m *MockVaultServiceClient) SaveTextData(
	ctx context.Context, req *vault.SaveTextDataRequest, opts ...grpc.CallOption,
) (*vault.SaveTextDataResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.SaveTextDataResponse), args.Error(1)
}

func (m *MockVaultServiceClient) SaveCardData(
	ctx context.Context, req *vault.SaveCardDataRequest, opts ...grpc.CallOption,
) (*vault.SaveCardDataResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.SaveCardDataResponse), args.Error(1)
}

func (m *MockVaultServiceClient) SaveBinaryData(
	ctx context.Context, req *vault.SaveBinaryDataRequest, opts ...grpc.CallOption,
) (*vault.SaveBinaryDataResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.SaveBinaryDataResponse), args.Error(1)
}

func (m *MockVaultServiceClient) SaveMeta(
	ctx context.Context, req *vault.SaveMetaRequest, opts ...grpc.CallOption,
) (*vault.SaveMetaResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.SaveMetaResponse), args.Error(1)
}

func (m *MockVaultServiceClient) DeleteVaultItem(
	ctx context.Context, req *vault.DeleteVaultItemRequest, opts ...grpc.CallOption,
) (*vault.DeleteVaultItemResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*vault.DeleteVaultItemResponse), args.Error(1)
}

// TestClient wraps Client to allow injection of mock services
type TestClient struct {
	*Client
	mockUserClient  *MockUserServiceClient
	mockVaultClient *MockVaultServiceClient
}

func newTestClient() *TestClient {
	mockUserClient := &MockUserServiceClient{}
	mockVaultClient := &MockVaultServiceClient{}

	client := &Client{
		userClient:  mockUserClient,
		vaultClient: mockVaultClient,
		serverAddr:  "test:8082",
	}

	return &TestClient{
		Client:          client,
		mockUserClient:  mockUserClient,
		mockVaultClient: mockVaultClient,
	}
}

// Test Configuration

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost:8082", config.ServerAddr)
	assert.Equal(t, 30*time.Second, config.ConnectTimeout)
	assert.Equal(t, 30*time.Second, config.RequestTimeout)
	assert.False(t, config.SkipConnTest)
	assert.NotNil(t, config.TLSConfig)
	assert.Equal(t, "", config.TLSConfig.ServerName)
}

// Test Client Creation

func TestNewClient_InvalidAddress(t *testing.T) {
	// Test with clearly invalid address format to avoid actual network calls
	config := &ClientConfig{
		ServerAddr:     "invalid-address-format:99999",
		ConnectTimeout: 100 * time.Millisecond,
		RequestTimeout: 100 * time.Millisecond,
		SkipConnTest:   false,
	}

	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "server not reachable")
}

func TestNewClient_SkipConnectivityTest(t *testing.T) {
	// Test configuration validation without network calls
	config := &ClientConfig{
		ServerAddr:     "test:8082",
		ConnectTimeout: 1 * time.Second,
		RequestTimeout: 1 * time.Second,
		SkipConnTest:   true,
	}

	// This will still fail at gRPC dial, but validates config processing
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClientWithConn(t *testing.T) {
	// Test the mock-friendly constructor
	testClient := newTestClient()

	assert.NotNil(t, testClient)
	assert.NotNil(t, testClient.Client)
	assert.Equal(t, "test:8082", testClient.serverAddr)
	assert.False(t, testClient.IsAuthenticated())
}

// Test connectivity check function
func TestConnectivity(t *testing.T) {
	// Test with unreachable address
	err := testConnectivity("192.0.2.1:65535", 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection failed")
}

// Test Authentication Methods

func TestClient_Login_Success(t *testing.T) {
	testClient := newTestClient()

	expectedReq := &user.LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}
	expectedResp := &user.LoginResponse{
		Token: "test-token-123",
	}

	testClient.mockUserClient.On("Login", mock.Anything, expectedReq).Return(expectedResp, nil)

	err := testClient.Login(context.Background(), "testuser", "testpass")

	assert.NoError(t, err)
	assert.Equal(t, "test-token-123", testClient.GetToken())
	assert.True(t, testClient.IsAuthenticated())
	testClient.mockUserClient.AssertExpectations(t)
}

func TestClient_Login_EmptyCredentials(t *testing.T) {
	testClient := newTestClient()

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"empty login", "", "password"},
		{"empty password", "login", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testClient.Login(context.Background(), tt.login, tt.password)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be empty")
			assert.False(t, testClient.IsAuthenticated())
		})
	}
}

func TestClient_Login_ServerError(t *testing.T) {
	testClient := newTestClient()

	expectedReq := &user.LoginRequest{
		Login:    "testuser",
		Password: "wrongpass",
	}

	testClient.mockUserClient.On("Login", mock.Anything, expectedReq).
		Return((*user.LoginResponse)(nil), status.Error(codes.Unauthenticated, "invalid credentials"))

	err := testClient.Login(context.Background(), "testuser", "wrongpass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
	assert.Contains(t, err.Error(), "invalid credentials")
	assert.False(t, testClient.IsAuthenticated())
}

func TestClient_Login_EmptyToken(t *testing.T) {
	testClient := newTestClient()

	expectedReq := &user.LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}
	expectedResp := &user.LoginResponse{
		Token: "", // Empty token
	}

	testClient.mockUserClient.On("Login", mock.Anything, expectedReq).Return(expectedResp, nil)

	err := testClient.Login(context.Background(), "testuser", "testpass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "received empty token")
	assert.False(t, testClient.IsAuthenticated())
}

func TestClient_Register_Success(t *testing.T) {
	testClient := newTestClient()

	expectedReq := &user.RegisterRequest{
		Login:    "newuser",
		Password: "newpass",
	}
	expectedResp := &user.RegisterResponse{}

	testClient.mockUserClient.On("Register", mock.Anything, expectedReq).Return(expectedResp, nil)

	err := testClient.Register(context.Background(), "newuser", "newpass")

	assert.NoError(t, err)
	testClient.mockUserClient.AssertExpectations(t)
}

func TestClient_Register_EmptyCredentials(t *testing.T) {
	testClient := newTestClient()

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"empty login", "", "password"},
		{"empty password", "login", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testClient.Register(context.Background(), tt.login, tt.password)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be empty")
		})
	}
}

func TestClient_Register_ServerError(t *testing.T) {
	testClient := newTestClient()

	expectedReq := &user.RegisterRequest{
		Login:    "existinguser",
		Password: "password",
	}

	testClient.mockUserClient.On("Register", mock.Anything, expectedReq).
		Return((*user.RegisterResponse)(nil), status.Error(codes.AlreadyExists, "user already exists"))

	err := testClient.Register(context.Background(), "existinguser", "password")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
	assert.Contains(t, err.Error(), "user already exists")
}

// Test Vault Operations

func TestClient_GetVaultItems_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token" // Simulate authenticated state

	expectedResp := &vault.GetVaultItemsResponse{
		TextData: []*vault.TextData{
			{Base: &vault.VaultItem{Id: "text1"}, Text: "sample text"},
		},
		LoginPasswords: []*vault.LoginPassword{
			{Base: &vault.VaultItem{Id: "login1"}, Login: "user", Password: "pass"},
		},
	}

	testClient.mockVaultClient.On("GetVaultItems", mock.Anything, &vault.GetVaultItemsRequest{}).
		Return(expectedResp, nil)

	resp, err := testClient.GetVaultItems(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.TextData, 1)
	assert.Len(t, resp.LoginPasswords, 1)
	assert.Equal(t, "text1", resp.TextData[0].Base.Id)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_GetVaultItems_NotAuthenticated(t *testing.T) {
	testClient := newTestClient()
	// Don't set token - not authenticated

	resp, err := testClient.GetVaultItems(context.Background())

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestClient_SaveLoginPassword_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	expectedReq := &vault.SaveLoginPasswordRequest{
		Login:    "mylogin",
		Password: "mypassword",
	}
	expectedResp := &vault.SaveLoginPasswordResponse{
		Id: "generated-id-123",
	}

	testClient.mockVaultClient.On("SaveLoginPassword", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	id, err := testClient.SaveLoginPassword(context.Background(), "mylogin", "mypassword")

	assert.NoError(t, err)
	assert.Equal(t, "generated-id-123", id)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_SaveLoginPassword_EmptyData(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"empty login", "", "password"},
		{"empty password", "login", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := testClient.SaveLoginPassword(context.Background(), tt.login, tt.password)
			assert.Error(t, err)
			assert.Empty(t, id)
			assert.Contains(t, err.Error(), "cannot be empty")
		})
	}
}

func TestClient_SaveTextData_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	expectedReq := &vault.SaveTextDataRequest{
		Text: "my important text",
	}
	expectedResp := &vault.SaveTextDataResponse{
		Id: "text-id-456",
	}

	testClient.mockVaultClient.On("SaveTextData", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	id, err := testClient.SaveTextData(context.Background(), "my important text")

	assert.NoError(t, err)
	assert.Equal(t, "text-id-456", id)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_SaveTextData_EmptyText(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	id, err := testClient.SaveTextData(context.Background(), "")

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Contains(t, err.Error(), "text cannot be empty")
}

func TestClient_SaveCardData_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	expectedReq := &vault.SaveCardDataRequest{
		Number: "4111111111111111",
		Holder: "John Doe",
		Expire: "12/25",
		Cvv:    "123",
	}
	expectedResp := &vault.SaveCardDataResponse{
		Id: "card-id-789",
	}

	testClient.mockVaultClient.On("SaveCardData", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	id, err := testClient.SaveCardData(
		context.Background(), "4111111111111111", "John Doe", "12/25", "123",
	)

	assert.NoError(t, err)
	assert.Equal(t, "card-id-789", id)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_SaveCardData_MissingFields(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	tests := []struct {
		name   string
		number string
		holder string
		expire string
		cvv    string
	}{
		{"empty number", "", "John Doe", "12/25", "123"},
		{"empty holder", "4111111111111111", "", "12/25", "123"},
		{"empty expire", "4111111111111111", "John Doe", "", "123"},
		{"empty cvv", "4111111111111111", "John Doe", "12/25", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := testClient.SaveCardData(
				context.Background(), tt.number, tt.holder, tt.expire, tt.cvv,
			)
			assert.Error(t, err)
			assert.Empty(t, id)
			assert.Contains(t, err.Error(), "all card fields are required")
		})
	}
}

func TestClient_SaveBinaryData_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	testData := []byte("binary test data")
	expectedReq := &vault.SaveBinaryDataRequest{
		Data: testData,
	}
	expectedResp := &vault.SaveBinaryDataResponse{
		Id: "binary-id-101",
	}

	testClient.mockVaultClient.On("SaveBinaryData", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	id, err := testClient.SaveBinaryData(context.Background(), testData)

	assert.NoError(t, err)
	assert.Equal(t, "binary-id-101", id)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_SaveBinaryData_EmptyData(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	id, err := testClient.SaveBinaryData(context.Background(), []byte{})

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Contains(t, err.Error(), "data cannot be empty")
}

func TestClient_SaveMeta_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	testMeta := []*vault.Meta{
		{Key: "category", Value: "personal"},
		{Key: "tags", Value: "important,secret"},
	}
	expectedReq := &vault.SaveMetaRequest{
		Meta: testMeta,
	}
	expectedResp := &vault.SaveMetaResponse{}

	testClient.mockVaultClient.On("SaveMeta", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	err := testClient.SaveMeta(context.Background(), testMeta)

	assert.NoError(t, err)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_SaveMeta_EmptyMeta(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	err := testClient.SaveMeta(context.Background(), []*vault.Meta{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "meta cannot be empty")
}

func TestClient_DeleteVaultItem_Success(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	expectedReq := &vault.DeleteVaultItemRequest{
		Id:   "item-123",
		Type: "text",
	}
	expectedResp := &vault.DeleteVaultItemResponse{}

	testClient.mockVaultClient.On("DeleteVaultItem", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	err := testClient.DeleteVaultItem(context.Background(), "item-123", "text")

	assert.NoError(t, err)
	testClient.mockVaultClient.AssertExpectations(t)
}

func TestClient_DeleteVaultItem_EmptyParameters(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	tests := []struct {
		name     string
		id       string
		itemType string
	}{
		{"empty id", "", "text"},
		{"empty type", "item-123", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testClient.DeleteVaultItem(context.Background(), tt.id, tt.itemType)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be empty")
		})
	}
}

// Test Authentication Context

func TestClient_GetAuthContext(t *testing.T) {
	testClient := newTestClient()

	t.Run("without token", func(t *testing.T) {
		ctx := context.Background()
		authCtx := testClient.GetAuthContext(ctx)
		assert.Equal(t, ctx, authCtx)
	})

	t.Run("with token", func(t *testing.T) {
		testClient.token = "test-token"
		ctx := context.Background()
		authCtx := testClient.GetAuthContext(ctx)
		assert.NotEqual(t, ctx, authCtx)
		// Note: In a real test, you'd extract metadata and verify the authorization header
	})
}

// Test Connection Management

func TestClient_Close(t *testing.T) {
	testClient := newTestClient()

	// Close should not error even with nil connection
	err := testClient.Close()
	assert.NoError(t, err)
}

// Test Authentication Helpers

func TestClient_GetToken(t *testing.T) {
	testClient := newTestClient()

	t.Run("no token", func(t *testing.T) {
		token := testClient.GetToken()
		assert.Empty(t, token)
	})

	t.Run("with token", func(t *testing.T) {
		testClient.token = "test-token-123"
		token := testClient.GetToken()
		assert.Equal(t, "test-token-123", token)
	})
}

func TestClient_IsAuthenticated(t *testing.T) {
	testClient := newTestClient()

	t.Run("not authenticated", func(t *testing.T) {
		assert.False(t, testClient.IsAuthenticated())
	})

	t.Run("authenticated", func(t *testing.T) {
		testClient.token = "test-token"
		assert.True(t, testClient.IsAuthenticated())
	})
}

// Test Error Scenarios

func TestClient_VaultOperations_NotAuthenticated(t *testing.T) {
	testClient := newTestClient()
	// Don't set token

	tests := []struct {
		name string
		fn   func() error
	}{
		{"GetVaultItems", func() error {
			_, err := testClient.GetVaultItems(context.Background())
			return err
		}},
		{"SaveLoginPassword", func() error {
			_, err := testClient.SaveLoginPassword(context.Background(), "login", "pass")
			return err
		}},
		{"SaveTextData", func() error {
			_, err := testClient.SaveTextData(context.Background(), "text")
			return err
		}},
		{"SaveCardData", func() error {
			_, err := testClient.SaveCardData(context.Background(), "1111", "John", "12/25", "123")
			return err
		}},
		{"SaveBinaryData", func() error {
			_, err := testClient.SaveBinaryData(context.Background(), []byte("data"))
			return err
		}},
		{"SaveMeta", func() error {
			return testClient.SaveMeta(
				context.Background(), []*vault.Meta{{Key: "test", Value: "data"}},
			)
		}},
		{"DeleteVaultItem", func() error {
			return testClient.DeleteVaultItem(context.Background(), "id", "type")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not authenticated")
		})
	}
}

// Benchmark Tests

func BenchmarkClient_Login(b *testing.B) {
	testClient := newTestClient()
	expectedReq := &user.LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}
	expectedResp := &user.LoginResponse{
		Token: "test-token-123",
	}

	testClient.mockUserClient.On("Login", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testClient.token = "" // Reset token
		err := testClient.Login(context.Background(), "testuser", "testpass")
		require.NoError(b, err)
	}
}

func BenchmarkClient_SaveTextData(b *testing.B) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	expectedReq := &vault.SaveTextDataRequest{
		Text: "benchmark text data",
	}
	expectedResp := &vault.SaveTextDataResponse{
		Id: "text-id-456",
	}

	testClient.mockVaultClient.On("SaveTextData", mock.Anything, expectedReq).
		Return(expectedResp, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testClient.SaveTextData(context.Background(), "benchmark text data")
		require.NoError(b, err)
	}
}

// Test Context Timeout Handling

func TestClient_Login_ContextTimeout(t *testing.T) {
	testClient := newTestClient()

	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure context is canceled

	expectedReq := &user.LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}

	testClient.mockUserClient.On("Login", mock.Anything, expectedReq).
		Return((*user.LoginResponse)(nil), context.DeadlineExceeded)

	err := testClient.Login(ctx, "testuser", "testpass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

// Test Concurrent Access

func TestClient_ConcurrentAccess(t *testing.T) {
	testClient := newTestClient()
	testClient.token = "valid-token"

	expectedResp := &vault.GetVaultItemsResponse{
		TextData: []*vault.TextData{
			{Base: &vault.VaultItem{Id: "text1"}, Text: "sample text"},
		},
	}

	testClient.mockVaultClient.On("GetVaultItems", mock.Anything, mock.Anything).
		Return(expectedResp, nil).Times(10)

	// Run 10 concurrent requests
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := testClient.GetVaultItems(context.Background())
			errCh <- err
		}()
	}

	// Check all requests succeeded
	for i := 0; i < 10; i++ {
		err := <-errCh
		assert.NoError(t, err)
	}
}
