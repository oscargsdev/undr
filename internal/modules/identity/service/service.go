package service

import (
	"errors"
	"log/slog"
	"time"

	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotActivated   = errors.New("user not activated")
)

type IdentityService interface {
	RegisterUser(user *domain.User) (*domain.OpaqueToken, error)
	ActivateUser(tokenPlainText string) (refreshTokenString string, accessTokenString string, err error)
	AuthenticateUser(email, password string) (refreshTokenString string, accessTokenString string, err error)
	RefreshToken(oldRefreshToken string) (refreshTokenString string, accessTokenString string, err error)
	Logout(userId int64) error
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

func (s *identityService) RegisterUser(user *domain.User) (*domain.OpaqueToken, error) {
	err := s.repository.InsertUser(user)
	if err != nil {
		return nil, err
	}

	// TODO: domain call to insert permission

	token, err := s.repository.NewOpaqueToken(user.ID, 3*24*time.Hour, domain.ScopeActivation)
	if err != nil {
		return nil, err
	}

	// EVENT: userRegistered
	// TODO: generate email to send to user

	return token, nil
}

func (s *identityService) ActivateUser(tokenPlainText string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.repository.GetForOpaqueToken(domain.ScopeActivation, tokenPlainText)
	if err != nil {
		return "", "", err
	}

	user.Activated = true

	err = s.repository.UpdateUser(user)
	if err != nil {
		return "", "", err
	}

	err = s.repository.DeleteAllFromUser(domain.ScopeActivation, user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.repository.NewOpaqueToken(user.ID, 24*time.Hour, domain.ScopeRefresh)
	if err != nil {
		return "", "", err
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = newAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// EVENT: userActivated
	return
}

func (s *identityService) AuthenticateUser(email, password string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.repository.GetUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	match, err := user.Password.Matches(password)
	if err != nil {
		return "", "", err
	}

	if !match {
		return "", "", ErrInvalidCredentials
	}

	if !user.Activated {
		return "", "", ErrUserNotActivated
	}

	err = s.repository.DeleteAllFromUser(domain.ScopeRefresh, user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.repository.NewOpaqueToken(user.ID, 24*time.Hour, domain.ScopeRefresh)
	if err != nil {
		return "", "", err
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = newAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// EVENT: userAuthenticated
	return
}

func (s *identityService) RefreshToken(oldRefreshToken string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.repository.GetForOpaqueToken(domain.ScopeRefresh, oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	err = s.repository.DeleteAllFromUser(domain.ScopeRefresh, user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.repository.NewOpaqueToken(user.ID, 24*time.Hour, domain.ScopeRefresh)
	if err != nil {
		return "", "", err
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = newAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return
}

func (s *identityService) Logout(userId int64) error {
	err := s.repository.DeleteAllFromUser(domain.ScopeRefresh, userId)
	return err
}
