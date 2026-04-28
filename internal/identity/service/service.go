package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/store"
)

const userIdContextKey = contextKey("userId")
const rolesContextKey = contextKey("roles")

var (
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrUserNotActivated          = errors.New("user not activated")
	ErrUserWithoutRoles          = errors.New("the user has no roles")
	ErrDuplicateEmail            = errors.New("duplicate email")
	ErrDuplicateUsername         = errors.New("duplicate username")
	ErrRecordNotFound            = errors.New("record not found")
	ErrEditConflict              = errors.New("edit conflict")
	ErrUnknownClaims             = errors.New("unknown claims")
	ErrMissingUserIDInContext    = errors.New("missing user id in context")
	ErrMissingUserRolesInContext = errors.New("missing user roles in context")
)

type usersRepository interface {
	InsertUser(context.Context, *domain.User) error
	UpdateUser(context.Context, *domain.User) error
	GetUserForOpaqueToken(context.Context, domain.TokenScope, string) (*domain.User, error)
	GetUserByEmail(context.Context, string) (*domain.User, error)
	GetUserById(context.Context, int64) (*domain.User, error)
}

type opaqueTokensRepository interface {
	NewOpaqueToken(context.Context, int64, time.Duration, domain.TokenScope) (*domain.OpaqueToken, error)
	DeleteAllFromUser(context.Context, domain.TokenScope, int64) error
}

type rolesRepository interface {
	GetAllRolesForUser(context.Context, int64) (domain.Roles, error)
	AddRoleForUser(context.Context, int64, ...string) error
}

type RepositorySet interface {
	usersRepository
	opaqueTokensRepository
	rolesRepository
}

type transactor interface {
	WithinTx(context.Context, func(RepositorySet) error) error
}

type claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type contextKey string

type Config struct {
	UsersRepository        usersRepository
	OpaqueTokensRepository opaqueTokensRepository
	RolesRepository        rolesRepository
	Transactor             transactor
	Logger                 *slog.Logger
	Issuer                 string
	JWTExpiration          time.Duration
	RefreshExpiration      time.Duration
	ActivationExpiration   time.Duration
}

type identityService struct {
	cfg        Config
	jwkStore   jwkset.Storage
	privateKey *rsa.PrivateKey
}

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, store.ErrDuplicateEmail):
		return ErrDuplicateEmail
	case errors.Is(err, store.ErrDuplicateUsername):
		return ErrDuplicateUsername
	case errors.Is(err, store.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, store.ErrEditConflict):
		return ErrEditConflict
	default:
		return err
	}
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

func (s *identityService) RegisterUser(ctx context.Context, user *domain.User) (*domain.OpaqueToken, error) {
	err := s.cfg.UsersRepository.InsertUser(ctx, user)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	err = s.cfg.RolesRepository.AddRoleForUser(ctx, user.ID, "user")
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	activationToken, err := s.cfg.OpaqueTokensRepository.NewOpaqueToken(ctx, user.ID, s.cfg.ActivationExpiration, domain.ScopeActivation)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	// EVENT: userRegistered
	return activationToken, nil
}

func (s *identityService) ActivateUser(ctx context.Context, tokenPlainText string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.cfg.UsersRepository.GetUserForOpaqueToken(ctx, domain.ScopeActivation, tokenPlainText)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	roles, err := s.cfg.RolesRepository.GetAllRolesForUser(ctx, user.ID)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	user.Activated = true

	err = s.cfg.UsersRepository.UpdateUser(ctx, user)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	err = s.cfg.OpaqueTokensRepository.DeleteAllFromUser(ctx, domain.ScopeActivation, user.ID)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	refreshToken, err := s.cfg.OpaqueTokensRepository.NewOpaqueToken(ctx, user.ID, s.cfg.RefreshExpiration, domain.ScopeRefresh)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = s.newAccessToken(user.ID, roles, s.cfg.JWTExpiration, s.cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	// EVENT: userActivated
	return
}

func (s *identityService) AuthenticateUser(ctx context.Context, email, password string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.cfg.UsersRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	roles, err := s.cfg.RolesRepository.GetAllRolesForUser(ctx, user.ID)
	if err != nil {
		return "", "", mapRepositoryError(err)
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

	err = s.cfg.OpaqueTokensRepository.DeleteAllFromUser(ctx, domain.ScopeRefresh, user.ID)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	refreshToken, err := s.cfg.OpaqueTokensRepository.NewOpaqueToken(ctx, user.ID, s.cfg.RefreshExpiration, domain.ScopeRefresh)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = s.newAccessToken(user.ID, roles, s.cfg.JWTExpiration, s.cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	// EVENT: userAuthenticated
	return
}

func (s *identityService) RefreshToken(ctx context.Context, oldRefreshToken string) (refreshTokenString string, accessTokenString string, err error) {
	user, err := s.cfg.UsersRepository.GetUserForOpaqueToken(ctx, domain.ScopeRefresh, oldRefreshToken)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	roles, err := s.cfg.RolesRepository.GetAllRolesForUser(ctx, user.ID)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	err = s.cfg.OpaqueTokensRepository.DeleteAllFromUser(ctx, domain.ScopeRefresh, user.ID)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}

	refreshToken, err := s.cfg.OpaqueTokensRepository.NewOpaqueToken(ctx, user.ID, s.cfg.RefreshExpiration, domain.ScopeRefresh)
	if err != nil {
		return "", "", mapRepositoryError(err)
	}
	refreshTokenString = refreshToken.Plaintext

	accessTokenString, err = s.newAccessToken(user.ID, roles, s.cfg.JWTExpiration, s.cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	return
}

func (s *identityService) Logout(ctx context.Context, userId int64) error {
	err := s.cfg.OpaqueTokensRepository.DeleteAllFromUser(ctx, domain.ScopeRefresh, userId)
	return mapRepositoryError(err)
}

func (s *identityService) GetUserById(ctx context.Context, userId int64) (*domain.UserDetails, error) {
	user, err := s.cfg.UsersRepository.GetUserById(ctx, userId)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	roles, err := s.cfg.RolesRepository.GetAllRolesForUser(ctx, user.ID)
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

func (s *identityService) initJWKS() (*rsa.PrivateKey, jwkset.Storage, error) {
	ctx := context.Background()
	jwkStore := jwkset.NewMemoryStorage()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		s.cfg.Logger.Error("failed to generate RSA key")
		return nil, nil, err
	}

	metadata := jwkset.JWKMetadataOptions{
		KID: "my-key-id",
	}
	options := jwkset.JWKOptions{
		Metadata: metadata,
	}

	jwk, err := jwkset.NewJWKFromKey(privateKey, options)
	if err != nil {
		s.cfg.Logger.Error("failed to create JWK from key")
		return nil, nil, err
	}

	err = jwkStore.KeyWrite(ctx, jwk)
	if err != nil {
		s.cfg.Logger.Error("failed to store RSA key")
		return nil, nil, err
	}

	return privateKey, jwkStore, nil
}

func (s *identityService) newAccessToken(userID int64, roles domain.Roles, expiration time.Duration, issuer string) (string, error) {
	claims := claims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			Issuer:    issuer,
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *identityService) ValidateJWTToken(tokenString string, issuer string) (*jwt.Token, error) {
	fn := func(token *jwt.Token) (any, error) {
		return s.privateKey.Public(), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims{}, fn, jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	if token == nil {
		return nil, jwt.ErrTokenMalformed
	}

	if !token.Valid {
		return nil, err
	}

	if _, ok := token.Claims.(*claims); ok {
		return token, nil
	} else {
		return nil, ErrUnknownClaims
	}
}

func ContextSetClaims(r *http.Request, token *jwt.Token) *http.Request {
	userId, _ := token.Claims.GetSubject()
	roles := token.Claims.(*claims).Roles

	ctx := context.WithValue(r.Context(), userIdContextKey, userId)
	ctx = context.WithValue(ctx, rolesContextKey, roles)

	return r.WithContext(ctx)
}

func ContextGetUserId(r *http.Request) (int64, error) {
	userId, ok := r.Context().Value(userIdContextKey).(string)
	if !ok {
		return -1, ErrMissingUserIDInContext
	}

	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func ContextGetRoles(r *http.Request) ([]string, error) {
	roles, ok := r.Context().Value(rolesContextKey).([]string)
	if !ok {
		return nil, ErrMissingUserRolesInContext
	}

	return roles, nil
}
