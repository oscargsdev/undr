package service

import (
	"log/slog"
	"time"

	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/repository"
)

type IdentityService interface {
	RegisterUser(user *domain.User) (*domain.Token, error)
	ActivateUser(tokenPlainText string) (*domain.User, error)
}

type identityService struct {
	repository repository.IdentityRepository
	logger     *slog.Logger
}

func New(repository repository.IdentityRepository, logger *slog.Logger) *identityService {
	return &identityService{
		repository: repository,
		logger:     logger,
	}
}

func (s *identityService) RegisterUser(user *domain.User) (*domain.Token, error) {
	err := s.repository.InsertUser(user)
	if err != nil {
		return nil, err
	}

	// TODO: domain call to insert permission

	token, err := s.repository.NewToken(user.ID, 3*24*time.Hour, domain.ScopeActivation)
	if err != nil {
		return nil, err
	}

	// TODO: generate email to send to user

	return token, nil
}

func (s *identityService) ActivateUser(tokenPlainText string) (*domain.User, error) {
	user, err := s.repository.GetForToken(domain.ScopeActivation, tokenPlainText)
	if err != nil {
		return nil, err
	}

	user.Activated = true

	err = s.repository.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	err = s.repository.DeleteAllFromUser(domain.ScopeActivation, user.ID)
	if err != nil {
		return nil, err
	}

	// TODO: Generate refresh and auth token, return them

	return user, nil
}
