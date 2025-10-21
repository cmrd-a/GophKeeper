package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/cmrd-a/GophKeeper/server/models"
	"github.com/cmrd-a/GophKeeper/server/repository"
)

type VaultItems struct {
	LoginPasswords []models.LoginPassword
	TextData       []models.TextData
	BinaryData     []models.BinaryData
	CardData       []models.CardData
	Meta           map[uuid.UUID][]models.Meta
}

type VaultService struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *VaultService {
	return &VaultService{repo: repo}
}

func (s *VaultService) GetVaultItems(ctx context.Context, userID string) (*VaultItems, error) {
	items := &VaultItems{
		Meta: make(map[uuid.UUID][]models.Meta),
	}
	var err error

	// Get all items
	items.LoginPasswords, err = s.repo.GetLoginPasswords(ctx, userID)
	if err != nil {
		return nil, err
	}

	items.TextData, err = s.repo.GetTextData(ctx, userID)
	if err != nil {
		return nil, err
	}

	items.BinaryData, err = s.repo.GetBinaryData(ctx, userID)
	if err != nil {
		return nil, err
	}

	items.CardData, err = s.repo.GetCardData(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect all item IDs
	itemIDs := make([]uuid.UUID, 0)
	for _, lp := range items.LoginPasswords {
		itemIDs = append(itemIDs, lp.ID)
	}
	for _, td := range items.TextData {
		itemIDs = append(itemIDs, td.ID)
	}
	for _, bd := range items.BinaryData {
		itemIDs = append(itemIDs, bd.ID)
	}
	for _, cd := range items.CardData {
		itemIDs = append(itemIDs, cd.ID)
	}

	// Get meta for all items
	for _, id := range itemIDs {
		meta, err := s.repo.GetMetaForItem(ctx, id.String())
		if err != nil {
			return nil, err
		}
		if len(meta) > 0 {
			items.Meta[id] = meta
		}
	}

	return items, nil
}

func (s *VaultService) SaveLoginPassword(ctx context.Context, lp models.LoginPassword) error {
	return s.repo.InsertLoginPassword(ctx, lp)
}

func (s *VaultService) SaveTextData(ctx context.Context, td models.TextData) error {
	return s.repo.InsertTextData(ctx, td)
}

func (s *VaultService) SaveBinaryData(ctx context.Context, bd models.BinaryData) error {
	return s.repo.InsertBinaryData(ctx, bd)
}

func (s *VaultService) SaveCardData(ctx context.Context, cd models.CardData) error {
	return s.repo.InsertCardData(ctx, cd)
}

func (s *VaultService) SaveMeta(ctx context.Context, meta []models.Meta) error {
	for _, m := range meta {
		if err := s.repo.InsertMeta(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (s *VaultService) DeleteVaultItem(ctx context.Context, id string, userID string, itemType string) error {
	switch itemType {
	case "login_password":
		return s.repo.DeleteLoginPassword(ctx, id, userID)
	case "text":
		return s.repo.DeleteTextData(ctx, id, userID)
	case "binary":
		return s.repo.DeleteBinaryData(ctx, id, userID)
	case "card":
		return s.repo.DeleteCardData(ctx, id, userID)
	default:
		return nil
	}
}
