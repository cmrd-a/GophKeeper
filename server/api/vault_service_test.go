package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
	"github.com/cmrd-a/GophKeeper/server/repository"
	"github.com/cmrd-a/GophKeeper/server/service"
)

func TestVaultServer_Creation(t *testing.T) {
	repo := &repository.Repository{}
	svc := service.NewService(repo)
	server := NewVaultServer(svc)

	assert.NotNil(t, server)
}

func TestVaultServer_RequestStructures(t *testing.T) {
	t.Run("GetVaultItemsRequest", func(t *testing.T) {
		req := &vault.GetVaultItemsRequest{}
		assert.NotNil(t, req)
	})

	t.Run("SaveLoginPasswordRequest", func(t *testing.T) {
		req := &vault.SaveLoginPasswordRequest{
			Login:    "test@example.com",
			Password: "password123",
		}
		assert.Equal(t, "test@example.com", req.GetLogin())
		assert.Equal(t, "password123", req.GetPassword())
	})

	t.Run("SaveTextDataRequest", func(t *testing.T) {
		req := &vault.SaveTextDataRequest{
			Text: "secret note",
		}
		assert.Equal(t, "secret note", req.GetText())
	})

	t.Run("SaveBinaryDataRequest", func(t *testing.T) {
		req := &vault.SaveBinaryDataRequest{
			Data: []byte("binary data"),
		}
		assert.Equal(t, []byte("binary data"), req.GetData())
	})

	t.Run("SaveCardDataRequest", func(t *testing.T) {
		req := &vault.SaveCardDataRequest{
			Number: "4111111111111111",
			Cvv:    "123",
			Holder: "Test User",
			Expire: "2025-01",
		}
		assert.Equal(t, "4111111111111111", req.GetNumber())
		assert.Equal(t, "123", req.GetCvv())
		assert.Equal(t, "Test User", req.GetHolder())
		assert.Equal(t, "2025-01", req.GetExpire())
	})

	t.Run("DeleteVaultItemRequest", func(t *testing.T) {
		req := &vault.DeleteVaultItemRequest{
			Id:   "test-id",
			Type: "login_password",
		}
		assert.Equal(t, "test-id", req.GetId())
		assert.Equal(t, "login_password", req.GetType())
	})
}

func TestVaultServer_ResponseStructures(t *testing.T) {
	t.Run("GetVaultItemsResponse", func(t *testing.T) {
		resp := &vault.GetVaultItemsResponse{
			LoginPasswords: []*vault.LoginPassword{},
			TextData:       []*vault.TextData{},
			BinaryData:     []*vault.BinaryData{},
			CardData:       []*vault.CardData{},
		}
		assert.NotNil(t, resp.GetLoginPasswords())
		assert.NotNil(t, resp.GetTextData())
		assert.NotNil(t, resp.GetBinaryData())
		assert.NotNil(t, resp.GetCardData())
	})

	t.Run("SaveLoginPasswordResponse", func(t *testing.T) {
		resp := &vault.SaveLoginPasswordResponse{
			Id: "test-id",
		}
		assert.Equal(t, "test-id", resp.GetId())
	})

	t.Run("SaveTextDataResponse", func(t *testing.T) {
		resp := &vault.SaveTextDataResponse{
			Id: "test-id",
		}
		assert.Equal(t, "test-id", resp.GetId())
	})

	t.Run("SaveBinaryDataResponse", func(t *testing.T) {
		resp := &vault.SaveBinaryDataResponse{
			Id: "test-id",
		}
		assert.Equal(t, "test-id", resp.GetId())
	})

	t.Run("SaveCardDataResponse", func(t *testing.T) {
		resp := &vault.SaveCardDataResponse{
			Id: "test-id",
		}
		assert.Equal(t, "test-id", resp.GetId())
	})

	t.Run("DeleteVaultItemResponse", func(t *testing.T) {
		resp := &vault.DeleteVaultItemResponse{}
		assert.NotNil(t, resp)
	})
}

func TestVaultServer_ProtoStructures(t *testing.T) {
	t.Run("VaultItem", func(t *testing.T) {
		item := &vault.VaultItem{
			Id: "test-id",
		}
		assert.Equal(t, "test-id", item.GetId())
	})

	t.Run("LoginPassword", func(t *testing.T) {
		lp := &vault.LoginPassword{
			Base:     &vault.VaultItem{Id: "test-id"},
			Login:    "test@example.com",
			Password: "password123",
		}
		assert.Equal(t, "test-id", lp.GetBase().GetId())
		assert.Equal(t, "test@example.com", lp.GetLogin())
		assert.Equal(t, "password123", lp.GetPassword())
	})

	t.Run("TextData", func(t *testing.T) {
		td := &vault.TextData{
			Base: &vault.VaultItem{Id: "test-id"},
			Text: "secret note",
		}
		assert.Equal(t, "test-id", td.GetBase().GetId())
		assert.Equal(t, "secret note", td.GetText())
	})

	t.Run("BinaryData", func(t *testing.T) {
		bd := &vault.BinaryData{
			Base: &vault.VaultItem{Id: "test-id"},
			Data: []byte("binary data"),
		}
		assert.Equal(t, "test-id", bd.GetBase().GetId())
		assert.Equal(t, []byte("binary data"), bd.GetData())
	})

	t.Run("CardData", func(t *testing.T) {
		cd := &vault.CardData{
			Base:   &vault.VaultItem{Id: "test-id"},
			Number: "4111111111111111",
			Cvv:    "123",
			Holder: "Test User",
			Expire: "2025-01",
		}
		assert.Equal(t, "test-id", cd.GetBase().GetId())
		assert.Equal(t, "4111111111111111", cd.GetNumber())
		assert.Equal(t, "123", cd.GetCvv())
		assert.Equal(t, "Test User", cd.GetHolder())
		assert.Equal(t, "2025-01", cd.GetExpire())
	})
}
