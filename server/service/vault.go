package service

import (
	"context"

	"github.com/cmrd-a/GophKeeper/server/models"
	"github.com/cmrd-a/GophKeeper/server/repository"
)

type VaultService struct {
	repo repository.Repository
}

func NewService() *VaultService {
	return &VaultService{}
}

func (s *VaultService) SaveLoginPassword(ctx context.Context, lp models.LoginPassword) error {
	if lp.ID == nil {
		return s.repo.InsertLoginPassword(ctx, lp)
	}
	return s.repo.UpdateLoginPassword(ctx, lp)
}
