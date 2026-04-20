package service

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/oscargsdev/undr/internal/identity/domain"
)

type identityRepository interface {
	InsertUser(*domain.User) error
	UpdateUser(*domain.User) error
	NewOpaqueToken(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error)
	GetForOpaqueToken(tokenScope, tokenPlaintext string) (*domain.User, error)
	DeleteAllFromUser(scope string, userID int64) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserById(userId int64) (*domain.User, error)
	GetAllRolesForUser(userID int64) (domain.Roles, error)
	AddRoleForUser(userID int64, codes ...string) error
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotActivated   = errors.New("user not activated")
	ErrUserWithoutRoles   = errors.New("the user has no roles")
)

type Config struct {
	Repository           identityRepository
	Logger               *slog.Logger
	Issuer               string
	JWTExpiration        time.Duration
	RefreshExpiration    time.Duration
	ActivationExpiration time.Duration
}

type identityService struct {
	cfg        Config
	jwkStore   jwkset.Storage
	privateKey *rsa.PrivateKey
}

func New(cfg Config) (*identityService, error) {
	identityService := &identityService{
		cfg: cfg,
	}

	privateKey, jwkStore, err := identityService.initJWKS()
	if err != nil {
		return nil, err
	}

	identityService.privateKey = privateKey
	identityService.jwkStore = jwkStore

	return identityService, nil
}

func (s *identityService) RegisterUser(user *domain.User) (*domain.OpaqueToken, error) {
	err := s.cfg.Repository.InsertUser(user)
	if err != nil {
		return nil, err
	}

	err = s.cfg.Repository.AddRoleForUser(user.ID, "user")
	if err != nil {
		return nil, err
	}

	activationToken, err := s.cfg.Repository.NewOpaqueToken(user.ID, s.cfg.ActivationExpiration, domain.ScopeActivation)
	if err != nil {
		return nil, err
	}

	// EVENT: userRegistered
	return activationToken, nil
}

func (s *identityService) ActivateUser(tokenPlainText string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.cfg.Repository.GetForOpaqueToken(domain.ScopeActivation, tokenPlainText)
	if err != nil {
		return "", "", err
	}

	roles, err := s.cfg.Repository.GetAllRolesForUser(user.ID)
	if err != nil {
		return "", "", err
	}

	user.Activated = true

	err = s.cfg.Repository.UpdateUser(user)
	if err != nil {
		return "", "", err
	}

	err = s.cfg.Repository.DeleteAllFromUser(domain.ScopeActivation, user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.cfg.Repository.NewOpaqueToken(user.ID, s.cfg.RefreshExpiration, domain.ScopeRefresh)
	if err != nil {
		return "", "", err
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = s.newAccessToken(user.ID, roles, s.cfg.JWTExpiration, s.cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	// EVENT: userActivated
	return
}

func (s *identityService) AuthenticateUser(email, password string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.cfg.Repository.GetUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	roles, err := s.cfg.Repository.GetAllRolesForUser(user.ID)
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

	err = s.cfg.Repository.DeleteAllFromUser(domain.ScopeRefresh, user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.cfg.Repository.NewOpaqueToken(user.ID, s.cfg.RefreshExpiration, domain.ScopeRefresh)
	if err != nil {
		return "", "", err
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = s.newAccessToken(user.ID, roles, s.cfg.JWTExpiration, s.cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	// EVENT: userAuthenticated
	return
}

func (s *identityService) RefreshToken(oldRefreshToken string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.cfg.Repository.GetForOpaqueToken(domain.ScopeRefresh, oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	roles, err := s.cfg.Repository.GetAllRolesForUser(user.ID)
	if err != nil {
		return "", "", err
	}

	err = s.cfg.Repository.DeleteAllFromUser(domain.ScopeRefresh, user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.cfg.Repository.NewOpaqueToken(user.ID, s.cfg.RefreshExpiration, domain.ScopeRefresh)
	if err != nil {
		return "", "", err
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = s.newAccessToken(user.ID, roles, s.cfg.JWTExpiration, s.cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	return
}

func (s *identityService) Logout(userId int64) error {
	err := s.cfg.Repository.DeleteAllFromUser(domain.ScopeRefresh, userId)
	return err
}

func (s *identityService) GetUserById(userId int64) (*domain.UserDetails, error) {
	user, err := s.cfg.Repository.GetUserById(userId)
	if err != nil {
		return nil, err
	}

	roles, err := s.cfg.Repository.GetAllRolesForUser(user.ID)
	if err != nil {
		return nil, ErrUserWithoutRoles
	}

	userDetails := domain.UserDetails{
		User:  *user,
		Roles: roles,
	}

	return &userDetails, nil
}

func (s *identityService) GetIssuer() string {
	return s.cfg.Issuer
}

var ErrJWKJSON = errors.New("failed to get JWK Set JSON")

func (s *identityService) GetJWKS(r *http.Request) (json.RawMessage, error) {
	response, err := s.jwkStore.JSONPublic(r.Context())
	if err != nil {
		return nil, ErrJWKJSON
	}

	return response, nil
}
