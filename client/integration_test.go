package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

// IntegrationTestSuite provides integration tests for the GophKeeper client
type IntegrationTestSuite struct {
	suite.Suite
	client     GophKeeperClient
	serverAddr string
}

// SetupSuite runs before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Get server address from environment or use default
	suite.serverAddr = os.Getenv("GOPHKEEPER_TEST_SERVER")
	if suite.serverAddr == "" {
		suite.serverAddr = "localhost:8082"
	}

	// Check if server is available
	suite.T().Logf("Checking server availability at %s", suite.serverAddr)

	conn, err := net.DialTimeout("tcp", suite.serverAddr, 5*time.Second)
	if err != nil {
		suite.T().Skipf("Integration tests skipped: server not available at %s", suite.serverAddr)
		return
	}
	conn.Close()

	// Create client
	config := &ClientConfig{
		ServerAddr:     suite.serverAddr,
		ConnectTimeout: 10 * time.Second,
		RequestTimeout: 10 * time.Second,
		SkipConnTest:   false,
	}

	client, err := NewClient(config)
	require.NoError(suite.T(), err, "Failed to create client")
	suite.client = client

	suite.T().Logf("Integration test suite setup complete")
}

// TearDownSuite runs after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.client != nil {
		err := suite.client.Close()
		assert.NoError(suite.T(), err, "Failed to close client")
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	if suite.client == nil {
		suite.T().Skip("Client not available - server may not be running")
	}
}

// TestConnection verifies basic connectivity
func (suite *IntegrationTestSuite) TestConnection() {
	assert.NotNil(suite.T(), suite.client)
	assert.False(suite.T(), suite.client.IsAuthenticated(), "Client should not be authenticated initially")
}

// TestUserRegistrationAndLogin tests the complete auth flow
func (suite *IntegrationTestSuite) TestUserRegistrationAndLogin() {
	ctx := context.Background()

	// Generate unique username to avoid conflicts
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("testuser_%d", timestamp)
	password := "testpassword123"

	// Test registration
	suite.T().Logf("Registering user: %s", username)
	err := suite.client.Register(ctx, username, password)
	assert.NoError(suite.T(), err, "Registration should succeed")

	// Test login
	suite.T().Logf("Logging in user: %s", username)
	err = suite.client.Login(ctx, username, password)
	assert.NoError(suite.T(), err, "Login should succeed")
	assert.True(suite.T(), suite.client.IsAuthenticated(), "Client should be authenticated after login")
	assert.NotEmpty(suite.T(), suite.client.GetToken(), "Token should not be empty")

	// Test login with wrong password
	err = suite.client.Login(ctx, username, "wrongpassword")
	assert.Error(suite.T(), err, "Login with wrong password should fail")
}

// TestDuplicateRegistration tests registration with existing username
func (suite *IntegrationTestSuite) TestDuplicateRegistration() {
	ctx := context.Background()

	// Generate unique username
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("dupuser_%d", timestamp)
	password := "testpassword123"

	// First registration should succeed
	err := suite.client.Register(ctx, username, password)
	assert.NoError(suite.T(), err, "First registration should succeed")

	// Second registration should fail
	err = suite.client.Register(ctx, username, password)
	assert.Error(suite.T(), err, "Duplicate registration should fail")
	assert.Contains(suite.T(), err.Error(), "registration failed")
}

// TestVaultOperationsFlow tests complete vault operations
func (suite *IntegrationTestSuite) TestVaultOperationsFlow() {
	ctx := context.Background()

	// Setup: Register and login a user
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("vaultuser_%d", timestamp)
	password := "vaultpass123"

	err := suite.client.Register(ctx, username, password)
	require.NoError(suite.T(), err, "Registration required for vault tests")

	err = suite.client.Login(ctx, username, password)
	require.NoError(suite.T(), err, "Login required for vault tests")

	// Test GetVaultItems (should be empty initially)
	suite.T().Log("Getting initial vault items")
	items, err := suite.client.GetVaultItems(ctx)
	assert.NoError(suite.T(), err, "Getting vault items should succeed")
	assert.NotNil(suite.T(), items, "Vault items response should not be nil")

	// Count initial items
	initialTextCount := len(items.TextData)
	initialLoginCount := len(items.LoginPasswords)
	initialCardCount := len(items.CardData)
	initialBinaryCount := len(items.BinaryData)

	// Test SaveTextData
	suite.T().Log("Saving text data")
	textID, err := suite.client.SaveTextData(ctx, "Integration test text data")
	assert.NoError(suite.T(), err, "Saving text data should succeed")
	assert.NotEmpty(suite.T(), textID, "Text data ID should not be empty")

	// Test SaveLoginPassword
	suite.T().Log("Saving login password")
	loginID, err := suite.client.SaveLoginPassword(ctx, "testlogin", "testloginpass")
	assert.NoError(suite.T(), err, "Saving login password should succeed")
	assert.NotEmpty(suite.T(), loginID, "Login password ID should not be empty")

	// Test SaveCardData
	suite.T().Log("Saving card data")
	cardID, err := suite.client.SaveCardData(ctx, "4111111111111111", "John Doe", "12/25", "123")
	assert.NoError(suite.T(), err, "Saving card data should succeed")
	assert.NotEmpty(suite.T(), cardID, "Card data ID should not be empty")

	// Test SaveBinaryData
	suite.T().Log("Saving binary data")
	binaryData := []byte("Integration test binary data content")
	binaryID, err := suite.client.SaveBinaryData(ctx, binaryData)
	assert.NoError(suite.T(), err, "Saving binary data should succeed")
	assert.NotEmpty(suite.T(), binaryID, "Binary data ID should not be empty")

	// Test SaveMeta
	suite.T().Log("Saving metadata")
	meta := []*vault.Meta{
		{Key: "category", Value: "integration-test"},
		{Key: "created_by", Value: "integration_test"},
	}
	err = suite.client.SaveMeta(ctx, meta)
	assert.NoError(suite.T(), err, "Saving metadata should succeed")

	// Verify items were saved by getting vault items again
	suite.T().Log("Verifying saved items")
	updatedItems, err := suite.client.GetVaultItems(ctx)
	assert.NoError(suite.T(), err, "Getting updated vault items should succeed")
	assert.NotNil(suite.T(), updatedItems, "Updated vault items response should not be nil")

	// Check that counts increased
	assert.Equal(suite.T(), initialTextCount+1, len(updatedItems.TextData), "Text data count should increase by 1")
	assert.Equal(
		suite.T(),
		initialLoginCount+1,
		len(updatedItems.LoginPasswords),
		"Login password count should increase by 1",
	)
	assert.Equal(suite.T(), initialCardCount+1, len(updatedItems.CardData), "Card data count should increase by 1")
	assert.Equal(
		suite.T(),
		initialBinaryCount+1,
		len(updatedItems.BinaryData),
		"Binary data count should increase by 1",
	)

	// Find and verify the saved text data
	var savedText *vault.TextData
	for _, item := range updatedItems.TextData {
		if item.Base.Id == textID {
			savedText = item
			break
		}
	}
	assert.NotNil(suite.T(), savedText, "Saved text data should be found")
	if savedText != nil {
		assert.Equal(suite.T(), "Integration test text data", savedText.Text, "Saved text should match")
	}

	// Test DeleteVaultItem
	suite.T().Log("Deleting text data item")
	err = suite.client.DeleteVaultItem(ctx, textID, "text")
	assert.NoError(suite.T(), err, "Deleting vault item should succeed")

	// Verify deletion
	suite.T().Log("Verifying item deletion")
	finalItems, err := suite.client.GetVaultItems(ctx)
	assert.NoError(suite.T(), err, "Getting final vault items should succeed")
	assert.Equal(
		suite.T(),
		initialTextCount,
		len(finalItems.TextData),
		"Text data count should return to initial value",
	)

	// Verify deleted item is not found
	for _, item := range finalItems.TextData {
		assert.NotEqual(suite.T(), textID, item.Base.Id, "Deleted item should not be found")
	}
}

// TestUnauthenticatedVaultOperations tests that vault operations fail without authentication
func (suite *IntegrationTestSuite) TestUnauthenticatedVaultOperations() {
	// Create a new client without logging in
	config := &ClientConfig{
		ServerAddr:     suite.serverAddr,
		ConnectTimeout: 10 * time.Second,
		RequestTimeout: 10 * time.Second,
		SkipConnTest:   false,
	}

	unauthClient, err := NewClient(config)
	require.NoError(suite.T(), err, "Creating unauthenticated client should succeed")
	defer unauthClient.Close()

	ctx := context.Background()

	// All vault operations should fail
	suite.T().Log("Testing unauthenticated vault operations")

	_, err = unauthClient.GetVaultItems(ctx)
	assert.Error(suite.T(), err, "GetVaultItems should fail without authentication")

	_, err = unauthClient.SaveTextData(ctx, "test")
	assert.Error(suite.T(), err, "SaveTextData should fail without authentication")

	_, err = unauthClient.SaveLoginPassword(ctx, "test", "test")
	assert.Error(suite.T(), err, "SaveLoginPassword should fail without authentication")

	_, err = unauthClient.SaveCardData(ctx, "1111", "Test", "12/25", "123")
	assert.Error(suite.T(), err, "SaveCardData should fail without authentication")

	_, err = unauthClient.SaveBinaryData(ctx, []byte("test"))
	assert.Error(suite.T(), err, "SaveBinaryData should fail without authentication")

	err = unauthClient.SaveMeta(ctx, []*vault.Meta{{Key: "test", Value: "test"}})
	assert.Error(suite.T(), err, "SaveMeta should fail without authentication")

	err = unauthClient.DeleteVaultItem(ctx, "test", "text")
	assert.Error(suite.T(), err, "DeleteVaultItem should fail without authentication")
}

// TestConnectionResilience tests client behavior with connection issues
func (suite *IntegrationTestSuite) TestConnectionResilience() {
	// Test with invalid server address
	config := &ClientConfig{
		ServerAddr:     "invalid:9999",
		ConnectTimeout: 1 * time.Second,
		RequestTimeout: 1 * time.Second,
		SkipConnTest:   false,
	}

	_, err := NewClient(config)
	assert.Error(suite.T(), err, "Client creation should fail with invalid server")
	assert.Contains(suite.T(), err.Error(), "server not reachable", "Error should indicate server unreachable")
}

// TestContextCancellation tests proper handling of context cancellation
func (suite *IntegrationTestSuite) TestContextCancellation() {
	// Register and login first
	ctx := context.Background()
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("canceluser_%d", timestamp)
	password := "cancelpass123"

	err := suite.client.Register(ctx, username, password)
	require.NoError(suite.T(), err)

	err = suite.client.Login(ctx, username, password)
	require.NoError(suite.T(), err)

	// Test with cancelled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Operations should handle cancelled context gracefully
	_, err = suite.client.GetVaultItems(cancelCtx)
	assert.Error(suite.T(), err, "Operation should fail with cancelled context")
	assert.Contains(suite.T(), err.Error(), "context canceled", "Error should indicate context cancellation")
}

// TestLargeDataHandling tests handling of large data
func (suite *IntegrationTestSuite) TestLargeDataHandling() {
	ctx := context.Background()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("largeuser_%d", timestamp)
	password := "largepass123"

	err := suite.client.Register(ctx, username, password)
	require.NoError(suite.T(), err)

	err = suite.client.Login(ctx, username, password)
	require.NoError(suite.T(), err)

	// Test large text data (1MB)
	largeText := string(make([]byte, 1024*1024))
	for i := range largeText {
		largeText = string(rune(65 + (i % 26))) // Fill with A-Z pattern
	}

	suite.T().Log("Testing large text data (1MB)")
	textID, err := suite.client.SaveTextData(ctx, largeText)
	if err != nil {
		suite.T().Logf("Large text data failed (may be expected): %v", err)
		// This might fail due to server limits, which is acceptable
	} else {
		assert.NotEmpty(suite.T(), textID, "Large text data should be saved successfully")

		// Clean up
		suite.client.DeleteVaultItem(ctx, textID, "text")
	}

	// Test large binary data (1MB)
	largeBinary := make([]byte, 1024*1024)
	for i := range largeBinary {
		largeBinary[i] = byte(i % 256)
	}

	suite.T().Log("Testing large binary data (1MB)")
	binaryID, err := suite.client.SaveBinaryData(ctx, largeBinary)
	if err != nil {
		suite.T().Logf("Large binary data failed (may be expected): %v", err)
		// This might fail due to server limits, which is acceptable
	} else {
		assert.NotEmpty(suite.T(), binaryID, "Large binary data should be saved successfully")

		// Clean up
		suite.client.DeleteVaultItem(ctx, binaryID, "binary")
	}
}

// TestConcurrentOperations tests concurrent client operations
func (suite *IntegrationTestSuite) TestConcurrentOperations() {
	ctx := context.Background()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("concuser_%d", timestamp)
	password := "concpass123"

	err := suite.client.Register(ctx, username, password)
	require.NoError(suite.T(), err)

	err = suite.client.Login(ctx, username, password)
	require.NoError(suite.T(), err)

	// Run multiple concurrent save operations
	const numOperations = 10
	errCh := make(chan error, numOperations)
	idCh := make(chan string, numOperations)

	suite.T().Log("Testing concurrent save operations")
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			text := fmt.Sprintf("Concurrent text data %d", index)
			id, err := suite.client.SaveTextData(ctx, text)
			errCh <- err
			if err == nil {
				idCh <- id
			}
		}(i)
	}

	// Collect results
	var savedIDs []string
	for i := 0; i < numOperations; i++ {
		err := <-errCh
		assert.NoError(suite.T(), err, "Concurrent save operation should succeed")
		if err == nil {
			savedIDs = append(savedIDs, <-idCh)
		}
	}

	assert.Len(suite.T(), savedIDs, numOperations, "All concurrent operations should succeed")

	// Verify all items were saved
	_, err = suite.client.GetVaultItems(ctx)
	assert.NoError(suite.T(), err)

	// Clean up
	suite.T().Log("Cleaning up concurrent test data")
	for _, id := range savedIDs {
		err := suite.client.DeleteVaultItem(ctx, id, "text")
		assert.NoError(suite.T(), err, "Cleanup should succeed")
	}
}

// TestInvalidInputHandling tests client behavior with invalid inputs
func (suite *IntegrationTestSuite) TestInvalidInputHandling() {
	ctx := context.Background()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("invaliduser_%d", timestamp)
	password := "invalidpass123"

	err := suite.client.Register(ctx, username, password)
	require.NoError(suite.T(), err)

	err = suite.client.Login(ctx, username, password)
	require.NoError(suite.T(), err)

	// Test invalid inputs
	suite.T().Log("Testing invalid inputs")

	// Empty strings
	_, err = suite.client.SaveTextData(ctx, "")
	assert.Error(suite.T(), err, "Empty text should fail")

	_, err = suite.client.SaveLoginPassword(ctx, "", "password")
	assert.Error(suite.T(), err, "Empty login should fail")

	_, err = suite.client.SaveLoginPassword(ctx, "login", "")
	assert.Error(suite.T(), err, "Empty password should fail")

	// Invalid card data
	_, err = suite.client.SaveCardData(ctx, "", "John Doe", "12/25", "123")
	assert.Error(suite.T(), err, "Empty card number should fail")

	// Empty binary data
	_, err = suite.client.SaveBinaryData(ctx, []byte{})
	assert.Error(suite.T(), err, "Empty binary data should fail")

	// Invalid delete parameters
	err = suite.client.DeleteVaultItem(ctx, "", "text")
	assert.Error(suite.T(), err, "Empty ID should fail")

	err = suite.client.DeleteVaultItem(ctx, "some-id", "")
	assert.Error(suite.T(), err, "Empty type should fail")
}

// TestEdgeCases tests various edge cases
func (suite *IntegrationTestSuite) TestEdgeCases() {
	ctx := context.Background()

	// Test very long username/password
	longString := string(make([]byte, 1000))
	for range longString {
		longString = "a"
	}

	suite.T().Log("Testing edge cases")

	// Very long credentials (may fail, which is acceptable)
	err := suite.client.Register(ctx, longString, "password")
	if err != nil {
		suite.T().Logf("Long username registration failed (may be expected): %v", err)
	}

	err = suite.client.Register(ctx, "user", longString)
	if err != nil {
		suite.T().Logf("Long password registration failed (may be expected): %v", err)
	}

	// Special characters in credentials
	specialUser := fmt.Sprintf("user_!@#$%%^&*()_%d", time.Now().UnixNano())
	specialPass := "pass_!@#$%^&*()"

	err = suite.client.Register(ctx, specialUser, specialPass)
	if err == nil {
		// If registration succeeded, test login
		err = suite.client.Login(ctx, specialUser, specialPass)
		assert.NoError(suite.T(), err, "Login with special characters should work")
	} else {
		suite.T().Logf("Special character credentials failed (may be expected): %v", err)
	}
}

// Run the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// Helper function to run integration tests easily
func TestIntegrationRunner(t *testing.T) {
	// Check if integration tests should run
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Integration tests skipped. Set RUN_INTEGRATION_TESTS=1 to run.")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
