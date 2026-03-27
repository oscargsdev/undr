package service

import (
	"log/slog"

	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/repository"
)

type IdentityService interface {
	RegisterUser(user *domain.User) error
}

type identityService struct {
	repository repository.IdentityRepository
	logger     *slog.Logger
}

func New(repository repository.IdentityRepository, logger *slog.Logger) *identityService {
	logger.Info("Entering New Service Identity")

	return &identityService{
		repository: repository,
		logger:     logger,
	}
}

func (s *identityService) RegisterUser(user *domain.User) error {
	// Receive User pointer
	// domain repository call to insert user
	err := s.repository.InsertUser(user)
	if err != nil {
		return err
	}
	// domain call to insert permission
	// generate activation token
	// TODO: generate email to send to user
	// All good, do not return nothing
	return nil
}
