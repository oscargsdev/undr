package service

import (
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/store"
)

type IdentityService interface {
	RegisterUser(user *domain.User) error
}

type identityService struct {
	repository store.UserRepository
}

func New(repository store.UserRepository) *identityService {
	return &identityService{
		repository: repository,
	}
}

func (s *identityService) RegisterUser(user *domain.User) error {
	// Receive User pointer
	// domain repository call to insert user
	s.repository.InsertUser(user)
	// domain call to insert permission
	// generate activation token
	// TODO: generate email to send to user
	// All good, do not return nothing
	return nil
}
