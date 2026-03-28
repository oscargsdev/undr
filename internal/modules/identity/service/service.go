package service

import (
	"log/slog"
	"time"

	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/repository"
)

type IdentityService interface {
	RegisterUser(user *domain.User) (*domain.Token, error)
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

	// domain call to insert permission
	// generate activation token
	token, err := s.repository.NewToken(user.ID, 3*24*time.Hour, domain.ScopeActivation)
	if err != nil {
		return nil, err
	}

	// TODO: generate email to send to user
	// All good, do not return nothing
	return token, nil
}
