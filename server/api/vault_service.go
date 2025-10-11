package api

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
	"github.com/cmrd-a/GophKeeper/server/auth"
	"github.com/cmrd-a/GophKeeper/server/models"
	"github.com/cmrd-a/GophKeeper/server/service"
)

// VaultServer implements VaultService.
type VaultServer struct {
	vault.UnimplementedVaultServiceServer

	service *service.VaultService
}

func NewVaultServer(svc *service.VaultService) *VaultServer {
	return &VaultServer{service: svc}
}

func vaultItemToProto(item models.VaultItem) *vault.VaultItem {
	return &vault.VaultItem{
		Id:        item.ID.String(),
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
		UserId:    item.UserID.String(),
	}
}

// GetVaultItems returns all vault items for the authenticated user.
func (s *VaultServer) GetVaultItems(
	ctx context.Context,
	req *vault.GetVaultItemsRequest,
) (*vault.GetVaultItemsResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.service.GetVaultItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert items to proto messages
	loginPasswords := make([]*vault.LoginPassword, len(items.LoginPasswords))
	for i, lp := range items.LoginPasswords {
		loginPasswords[i] = &vault.LoginPassword{
			Base:     vaultItemToProto(lp.VaultItem),
			Login:    lp.Login,
			Password: string(lp.Password),
		}
	}

	textData := make([]*vault.TextData, len(items.TextData))
	for i, td := range items.TextData {
		textData[i] = &vault.TextData{
			Base: vaultItemToProto(td.VaultItem),
			Text: td.Text,
		}
	}

	binaryData := make([]*vault.BinaryData, len(items.BinaryData))
	for i, bd := range items.BinaryData {
		binaryData[i] = &vault.BinaryData{
			Base: vaultItemToProto(bd.VaultItem),
			Data: bd.Data,
		}
	}

	cardData := make([]*vault.CardData, len(items.CardData))
	for i, cd := range items.CardData {
		cardData[i] = &vault.CardData{
			Base:   vaultItemToProto(cd.VaultItem),
			Number: string(cd.Number),
			Holder: cd.Holder,
			Expire: cd.Expires.Format("2006-01"),
			Cvv:    string(cd.CVV),
		}
	}

	meta := make(map[string]*vault.Meta)
	for _, metas := range items.Meta {
		for _, m := range metas {
			meta[m.ID.String()] = &vault.Meta{
				Base: vaultItemToProto(models.VaultItem{
					ID:        m.ID,
					UserID:    m.Relation,
					CreatedAt: m.CreatedAt,
					UpdatedAt: m.UpdatedAt,
				}),
				Key:    m.Name,
				Value:  m.Data,
				ItemId: m.Relation.String(),
			}
		}
	}

	return &vault.GetVaultItemsResponse{
		LoginPasswords: loginPasswords,
		TextData:       textData,
		BinaryData:     binaryData,
		CardData:       cardData,
		Meta:           meta,
	}, nil
}

func (s *VaultServer) SaveLoginPassword(
	ctx context.Context,
	req *vault.SaveLoginPasswordRequest,
) (*vault.SaveLoginPasswordResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	itemID := uuid.New()
	lp := models.LoginPassword{
		VaultItem: models.VaultItem{
			ID:        itemID,
			UserID:    parsedUserID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	}

	if err := s.service.SaveLoginPassword(ctx, lp); err != nil {
		return nil, err
	}

	return &vault.SaveLoginPasswordResponse{Id: itemID.String()}, nil
}

func (s *VaultServer) SaveTextData(
	ctx context.Context,
	req *vault.SaveTextDataRequest,
) (*vault.SaveTextDataResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	itemID := uuid.New()
	td := models.TextData{
		VaultItem: models.VaultItem{
			ID:        itemID,
			UserID:    parsedUserID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Text: req.GetText(),
	}

	if err := s.service.SaveTextData(ctx, td); err != nil {
		return nil, err
	}

	return &vault.SaveTextDataResponse{Id: itemID.String()}, nil
}

func (s *VaultServer) SaveBinaryData(
	ctx context.Context,
	req *vault.SaveBinaryDataRequest,
) (*vault.SaveBinaryDataResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	itemID := uuid.New()
	bd := models.BinaryData{
		VaultItem: models.VaultItem{
			ID:        itemID,
			UserID:    parsedUserID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Data: req.GetData(),
	}

	if err := s.service.SaveBinaryData(ctx, bd); err != nil {
		return nil, err
	}

	return &vault.SaveBinaryDataResponse{Id: itemID.String()}, nil
}

func (s *VaultServer) SaveCardData(
	ctx context.Context,
	req *vault.SaveCardDataRequest,
) (*vault.SaveCardDataResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	expires, err := time.Parse("2006-01", req.GetExpire())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	itemID := uuid.New()
	cd := models.CardData{
		VaultItem: models.VaultItem{
			ID:        itemID,
			UserID:    uid,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Number:  []byte(req.GetNumber()),
		Holder:  req.GetHolder(),
		Expires: expires,
		CVV:     []byte(req.GetCvv()),
	}

	if err := s.service.SaveCardData(ctx, cd); err != nil {
		return nil, err
	}

	return &vault.SaveCardDataResponse{Id: itemID.String()}, nil
}

func (s *VaultServer) SaveMeta(ctx context.Context, req *vault.SaveMetaRequest) (*vault.SaveMetaResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	meta := make([]models.Meta, len(req.GetMeta()))
	for i, m := range req.GetMeta() {
		itemID, err := uuid.Parse(m.GetItemId())
		if err != nil {
			return nil, err
		}

		meta[i] = models.Meta{
			ID:        uuid.New(),
			Relation:  itemID,
			Name:      m.GetKey(),
			Data:      m.GetValue(),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	if err := s.service.SaveMeta(ctx, meta); err != nil {
		return nil, err
	}

	return &vault.SaveMetaResponse{}, nil
}

func (s *VaultServer) DeleteVaultItem(
	ctx context.Context,
	req *vault.DeleteVaultItemRequest,
) (*vault.DeleteVaultItemResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.service.DeleteVaultItem(ctx, req.GetId(), userID, req.GetType()); err != nil {
		return nil, err
	}

	return &vault.DeleteVaultItemResponse{}, nil
}
