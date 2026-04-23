package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/postgres"
)

type usersRepoMock struct {
	insertFn      func(*domain.User) error
	updateFn      func(*domain.User) error
	getForTokenFn func(string, string) (*domain.User, error)
	getByEmailFn  func(string) (*domain.User, error)
	getByIDFn     func(int64) (*domain.User, error)

	insertCalls      int
	updateCalls      int
	getForTokenCalls int
	getByEmailCalls  int
	getByIDCalls     int

	lastInsertedUser   *domain.User
	lastUpdatedUser    *domain.User
	lastTokenScope     string
	lastTokenPlaintext string
	lastEmailLookup    string
	lastIDLookup       int64
}

func (m *usersRepoMock) InsertUser(user *domain.User) error {
	m.insertCalls++
	m.lastInsertedUser = user
	if m.insertFn == nil {
		panic("unexpected InsertUser call")
	}
	return m.insertFn(user)
}

func (m *usersRepoMock) UpdateUser(user *domain.User) error {
	m.updateCalls++
	m.lastUpdatedUser = user
	if m.updateFn == nil {
		panic("unexpected UpdateUser call")
	}
	return m.updateFn(user)
}

func (m *usersRepoMock) GetForOpaqueToken(scope, plaintext string) (*domain.User, error) {
	m.getForTokenCalls++
	m.lastTokenScope = scope
	m.lastTokenPlaintext = plaintext
	if m.getForTokenFn == nil {
		panic("unexpected GetForOpaqueToken call")
	}
	return m.getForTokenFn(scope, plaintext)
}

func (m *usersRepoMock) GetUserByEmail(email string) (*domain.User, error) {
	m.getByEmailCalls++
	m.lastEmailLookup = email
	if m.getByEmailFn == nil {
		panic("unexpected GetUserByEmail call")
	}
	return m.getByEmailFn(email)
}

func (m *usersRepoMock) GetUserById(userID int64) (*domain.User, error) {
	m.getByIDCalls++
	m.lastIDLookup = userID
	if m.getByIDFn == nil {
		panic("unexpected GetUserById call")
	}
	return m.getByIDFn(userID)
}

type tokenCall struct {
	userID int64
	ttl    time.Duration
	scope  string
}

type tokensRepoMock struct {
	newOpaqueTokenFn    func(int64, time.Duration, string) (*domain.OpaqueToken, error)
	deleteAllFromUserFn func(string, int64) error

	newOpaqueTokenCalls []tokenCall
	deleteCalls         int
	lastDeleteScope     string
	lastDeleteUserID    int64
}

func (m *tokensRepoMock) NewOpaqueToken(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
	m.newOpaqueTokenCalls = append(m.newOpaqueTokenCalls, tokenCall{userID: userID, ttl: ttl, scope: scope})
	if m.newOpaqueTokenFn == nil {
		panic("unexpected NewOpaqueToken call")
	}
	return m.newOpaqueTokenFn(userID, ttl, scope)
}

func (m *tokensRepoMock) DeleteAllFromUser(scope string, userID int64) error {
	m.deleteCalls++
	m.lastDeleteScope = scope
	m.lastDeleteUserID = userID
	if m.deleteAllFromUserFn == nil {
		panic("unexpected DeleteAllFromUser call")
	}
	return m.deleteAllFromUserFn(scope, userID)
}

type rolesRepoMock struct {
	getAllForUserFn  func(int64) (domain.Roles, error)
	addRoleForUserFn func(int64, ...string) error

	getAllCalls  int
	addRoleCalls int

	lastGetAllUserID int64
	lastAddUserID    int64
	lastAddCodes     []string
}

func (m *rolesRepoMock) GetAllRolesForUser(userID int64) (domain.Roles, error) {
	m.getAllCalls++
	m.lastGetAllUserID = userID
	if m.getAllForUserFn == nil {
		panic("unexpected GetAllRolesForUser call")
	}
	return m.getAllForUserFn(userID)
}

func (m *rolesRepoMock) AddRoleForUser(userID int64, codes ...string) error {
	m.addRoleCalls++
	m.lastAddUserID = userID
	m.lastAddCodes = append([]string(nil), codes...)
	if m.addRoleForUserFn == nil {
		panic("unexpected AddRoleForUser call")
	}
	return m.addRoleForUserFn(userID, codes...)
}

func newTestIdentityService(t *testing.T) (*identityService, *usersRepoMock, *tokensRepoMock, *rolesRepoMock) {
	t.Helper()

	users := &usersRepoMock{}
	tokens := &tokensRepoMock{}
	roles := &rolesRepoMock{}

	svc, err := New(Config{
		UsersRepository:        users,
		OpaqueTokensRepository: tokens,
		RolesRepository:        roles,
		Issuer:                 "https://issuer.example",
		JWTExpiration:          5 * time.Minute,
		RefreshExpiration:      24 * time.Hour,
		ActivationExpiration:   48 * time.Hour,
	})
	if err != nil {
		t.Fatalf("setup failed creating identity service: %v", err)
	}

	return svc, users, tokens, roles
}

func newHashedPassword(t *testing.T, plaintext string) domain.Password {
	t.Helper()

	var password domain.Password
	if err := password.Set(plaintext); err != nil {
		t.Fatalf("could not setup password hash: %v", err)
	}
	return password
}

func assertError(t *testing.T, got error, want error) {
	t.Helper()

	switch {
	case want == nil && got != nil:
		t.Fatalf("expected nil error, got %v", got)
	case want != nil && !errors.Is(got, want):
		t.Fatalf("expected error %v, got %v", want, got)
	}
}

func TestMapRepositoryError(t *testing.T) {
	unknownErr := errors.New("unknown")

	tests := []struct {
		name string
		in   error
		want error
	}{
		{name: "duplicate email", in: postgres.ErrDuplicateEmail, want: ErrDuplicateEmail},
		{name: "wrapped duplicate username", in: errors.Join(errors.New("wrapped"), postgres.ErrDuplicateUsername), want: ErrDuplicateUsername},
		{name: "record not found", in: postgres.ErrRecordNotFound, want: ErrRecordNotFound},
		{name: "edit conflict", in: postgres.ErrEditConflict, want: ErrEditConflict},
		{name: "unknown passthrough", in: unknownErr, want: unknownErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapRepositoryError(tt.in)
			if tt.want == unknownErr {
				if got != unknownErr {
					t.Fatalf("expected unknown error identity to be preserved")
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, got)
			}
		})
	}
}

func TestIdentityService_NewAndGetIssuer(t *testing.T) {
	svc, err := New(Config{Issuer: "https://issuer.example"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if svc.privateKey == nil {
		t.Fatal("expected private key to be initialized")
	}
	if svc.jwkStore == nil {
		t.Fatal("expected JWK storage to be initialized")
	}
	if got, want := svc.GetIssuer(), "https://issuer.example"; got != want {
		t.Fatalf("expected issuer %q, got %q", want, got)
	}
}

func TestIdentityService_GetJWKS(t *testing.T) {
	svc, _, _, _ := newTestIdentityService(t)

	req := httptest.NewRequest(http.MethodGet, "/jwks.json", nil)
	response, err := svc.GetJWKS(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(response) == 0 {
		t.Fatal("expected non-empty JWKS response")
	}

	var payload map[string]any
	if err := json.Unmarshal(response, &payload); err != nil {
		t.Fatalf("expected valid JSON, got %v", err)
	}
	if _, ok := payload["keys"]; !ok {
		t.Fatalf("expected JWKS payload to include keys, got %v", payload)
	}
}

func TestIdentityService_Logout(t *testing.T) {
	unknownErr := errors.New("db down")
	tests := []struct {
		name        string
		repoErr     error
		wantErr     error
		wantSameErr bool
	}{
		{name: "success", repoErr: nil, wantErr: nil},
		{name: "mapped record not found", repoErr: postgres.ErrRecordNotFound, wantErr: ErrRecordNotFound},
		{name: "unknown passthrough", repoErr: unknownErr, wantErr: unknownErr, wantSameErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := &tokensRepoMock{
				deleteAllFromUserFn: func(scope string, userID int64) error {
					return tt.repoErr
				},
			}
			svc := &identityService{cfg: Config{OpaqueTokensRepository: tokens}}

			err := svc.Logout(42)
			if tt.wantSameErr {
				if err != tt.wantErr {
					t.Fatalf("expected passthrough error identity")
				}
			} else {
				assertError(t, err, tt.wantErr)
			}

			if tokens.deleteCalls != 1 {
				t.Fatalf("expected DeleteAllFromUser to be called once, got %d", tokens.deleteCalls)
			}
			if tokens.lastDeleteScope != domain.ScopeRefresh {
				t.Fatalf("expected scope %q, got %q", domain.ScopeRefresh, tokens.lastDeleteScope)
			}
			if tokens.lastDeleteUserID != 42 {
				t.Fatalf("expected user id 42, got %d", tokens.lastDeleteUserID)
			}
		})
	}
}

func TestIdentityService_RegisterUser(t *testing.T) {
	tests := []struct {
		name                string
		insertErr           error
		addRoleErr          error
		newTokenErr         error
		wantErr             error
		wantInsertCalls     int
		wantAddRoleCalls    int
		wantNewTokenCalls   int
		wantActivationToken string
	}{
		{
			name:                "success",
			wantInsertCalls:     1,
			wantAddRoleCalls:    1,
			wantNewTokenCalls:   1,
			wantActivationToken: "activation-token",
		},
		{
			name:            "insert error mapped",
			insertErr:       postgres.ErrDuplicateEmail,
			wantErr:         ErrDuplicateEmail,
			wantInsertCalls: 1,
		},
		{
			name:             "add role error mapped",
			addRoleErr:       postgres.ErrDuplicateUsername,
			wantErr:          ErrDuplicateUsername,
			wantInsertCalls:  1,
			wantAddRoleCalls: 1,
		},
		{
			name:              "new token error mapped",
			newTokenErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantInsertCalls:   1,
			wantAddRoleCalls:  1,
			wantNewTokenCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, users, tokens, roles := newTestIdentityService(t)

			users.insertFn = func(user *domain.User) error {
				if tt.insertErr != nil {
					return tt.insertErr
				}
				user.ID = 99
				return nil
			}
			roles.addRoleForUserFn = func(userID int64, codes ...string) error {
				return tt.addRoleErr
			}
			tokens.newOpaqueTokenFn = func(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
				if tt.newTokenErr != nil {
					return nil, tt.newTokenErr
				}
				return &domain.OpaqueToken{Plaintext: "activation-token"}, nil
			}

			user := &domain.User{Username: "alice", Email: "alice@example.com", Password: domain.Password{Hash: []byte("hash")}}
			activationToken, err := svc.RegisterUser(user)

			assertError(t, err, tt.wantErr)
			if tt.wantErr != nil {
				if activationToken != nil {
					t.Fatal("expected nil activation token on error")
				}
			} else if activationToken == nil || activationToken.Plaintext != tt.wantActivationToken {
				t.Fatalf("expected activation token %q, got %+v", tt.wantActivationToken, activationToken)
			}

			if users.insertCalls != tt.wantInsertCalls {
				t.Fatalf("expected InsertUser calls %d, got %d", tt.wantInsertCalls, users.insertCalls)
			}
			if roles.addRoleCalls != tt.wantAddRoleCalls {
				t.Fatalf("expected AddRoleForUser calls %d, got %d", tt.wantAddRoleCalls, roles.addRoleCalls)
			}
			if len(tokens.newOpaqueTokenCalls) != tt.wantNewTokenCalls {
				t.Fatalf("expected NewOpaqueToken calls %d, got %d", tt.wantNewTokenCalls, len(tokens.newOpaqueTokenCalls))
			}

			if tt.wantErr == nil {
				if roles.lastAddUserID != 99 {
					t.Fatalf("expected role assignment user id 99, got %d", roles.lastAddUserID)
				}
				if !reflect.DeepEqual(roles.lastAddCodes, []string{"user"}) {
					t.Fatalf("expected role codes [user], got %v", roles.lastAddCodes)
				}
				if got, want := tokens.newOpaqueTokenCalls[0].scope, domain.ScopeActivation; got != want {
					t.Fatalf("expected activation scope %q, got %q", want, got)
				}
				if got, want := tokens.newOpaqueTokenCalls[0].userID, int64(99); got != want {
					t.Fatalf("expected token user id %d, got %d", want, got)
				}
				if got, want := tokens.newOpaqueTokenCalls[0].ttl, svc.cfg.ActivationExpiration; got != want {
					t.Fatalf("expected token ttl %s, got %s", want, got)
				}
			}
		})
	}
}

func TestIdentityService_ActivateUser(t *testing.T) {
	tests := []struct {
		name              string
		getUserErr        error
		getRolesErr       error
		updateErr         error
		deleteErr         error
		newTokenErr       error
		wantErr           error
		wantGetUserCalls  int
		wantGetRolesCalls int
		wantUpdateCalls   int
		wantDeleteCalls   int
		wantNewTokenCalls int
		wantRefreshToken  string
		wantAccessToken   bool
	}{
		{
			name:              "success",
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantUpdateCalls:   1,
			wantDeleteCalls:   1,
			wantNewTokenCalls: 1,
			wantRefreshToken:  "refresh-token",
			wantAccessToken:   true,
		},
		{
			name:             "get user error mapped",
			getUserErr:       postgres.ErrRecordNotFound,
			wantErr:          ErrRecordNotFound,
			wantGetUserCalls: 1,
		},
		{
			name:              "get roles error mapped",
			getRolesErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
		{
			name:              "update error mapped",
			updateErr:         postgres.ErrEditConflict,
			wantErr:           ErrEditConflict,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantUpdateCalls:   1,
		},
		{
			name:              "delete activation tokens error mapped",
			deleteErr:         postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantUpdateCalls:   1,
			wantDeleteCalls:   1,
		},
		{
			name:              "new refresh token error mapped",
			newTokenErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantUpdateCalls:   1,
			wantDeleteCalls:   1,
			wantNewTokenCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, users, tokens, roles := newTestIdentityService(t)
			user := &domain.User{ID: 77, Activated: false, Password: domain.Password{Hash: []byte("hash")}}

			users.getForTokenFn = func(scope, plaintext string) (*domain.User, error) {
				if tt.getUserErr != nil {
					return nil, tt.getUserErr
				}
				return user, nil
			}
			roles.getAllForUserFn = func(userID int64) (domain.Roles, error) {
				if tt.getRolesErr != nil {
					return nil, tt.getRolesErr
				}
				return domain.Roles{"user", "admin"}, nil
			}
			users.updateFn = func(updated *domain.User) error {
				return tt.updateErr
			}
			tokens.deleteAllFromUserFn = func(scope string, userID int64) error {
				return tt.deleteErr
			}
			tokens.newOpaqueTokenFn = func(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
				if tt.newTokenErr != nil {
					return nil, tt.newTokenErr
				}
				return &domain.OpaqueToken{Plaintext: "refresh-token"}, nil
			}
			refreshToken, accessToken, err := svc.ActivateUser("activation-plaintext")
			assertError(t, err, tt.wantErr)

			if refreshToken != tt.wantRefreshToken {
				t.Fatalf("expected refresh token %q, got %q", tt.wantRefreshToken, refreshToken)
			}
			if tt.wantAccessToken && accessToken == "" {
				t.Fatal("expected non-empty access token")
			}
			if !tt.wantAccessToken && accessToken != "" {
				t.Fatalf("expected empty access token, got %q", accessToken)
			}

			if users.getForTokenCalls != tt.wantGetUserCalls {
				t.Fatalf("expected GetForOpaqueToken calls %d, got %d", tt.wantGetUserCalls, users.getForTokenCalls)
			}
			if roles.getAllCalls != tt.wantGetRolesCalls {
				t.Fatalf("expected GetAllRolesForUser calls %d, got %d", tt.wantGetRolesCalls, roles.getAllCalls)
			}
			if users.updateCalls != tt.wantUpdateCalls {
				t.Fatalf("expected UpdateUser calls %d, got %d", tt.wantUpdateCalls, users.updateCalls)
			}
			if tokens.deleteCalls != tt.wantDeleteCalls {
				t.Fatalf("expected DeleteAllFromUser calls %d, got %d", tt.wantDeleteCalls, tokens.deleteCalls)
			}
			if len(tokens.newOpaqueTokenCalls) != tt.wantNewTokenCalls {
				t.Fatalf("expected NewOpaqueToken calls %d, got %d", tt.wantNewTokenCalls, len(tokens.newOpaqueTokenCalls))
			}

			if tt.wantErr == nil {
				if users.lastTokenScope != domain.ScopeActivation {
					t.Fatalf("expected activation lookup scope, got %q", users.lastTokenScope)
				}
				if users.lastTokenPlaintext != "activation-plaintext" {
					t.Fatalf("expected activation plaintext, got %q", users.lastTokenPlaintext)
				}
				if users.lastUpdatedUser == nil || !users.lastUpdatedUser.Activated {
					t.Fatal("expected updated user to be activated")
				}
				if tokens.lastDeleteScope != domain.ScopeActivation {
					t.Fatalf("expected activation delete scope, got %q", tokens.lastDeleteScope)
				}
				if got, want := tokens.newOpaqueTokenCalls[0].scope, domain.ScopeRefresh; got != want {
					t.Fatalf("expected refresh scope, got %q", got)
				}
			}
		})
	}
}

func TestIdentityService_AuthenticateUser(t *testing.T) {
	validUser := &domain.User{ID: 55, Email: "alice@example.com", Activated: true, Password: newHashedPassword(t, "correct-password")}
	inactiveUser := &domain.User{ID: 55, Email: "alice@example.com", Activated: false, Password: newHashedPassword(t, "correct-password")}
	invalidHashUser := &domain.User{ID: 55, Email: "alice@example.com", Activated: true, Password: domain.Password{Hash: []byte("invalid-hash")}}

	tests := []struct {
		name              string
		user              *domain.User
		password          string
		getUserErr        error
		getRolesErr       error
		deleteErr         error
		newTokenErr       error
		wantErr           error
		wantGetUserCalls  int
		wantGetRolesCalls int
		wantDeleteCalls   int
		wantNewTokenCalls int
		wantRefreshToken  string
		wantAccessToken   bool
		wantRawHashError  bool
	}{
		{
			name:              "success",
			user:              validUser,
			password:          "correct-password",
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantDeleteCalls:   1,
			wantNewTokenCalls: 1,
			wantRefreshToken:  "new-refresh-token",
			wantAccessToken:   true,
		},
		{
			name:             "get user error mapped",
			user:             validUser,
			password:         "correct-password",
			getUserErr:       postgres.ErrRecordNotFound,
			wantErr:          ErrRecordNotFound,
			wantGetUserCalls: 1,
		},
		{
			name:              "get roles error mapped",
			user:              validUser,
			password:          "correct-password",
			getRolesErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
		{
			name:              "invalid credentials",
			user:              validUser,
			password:          "wrong-password",
			wantErr:           ErrInvalidCredentials,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
		{
			name:              "inactive account",
			user:              inactiveUser,
			password:          "correct-password",
			wantErr:           ErrUserNotActivated,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
		{
			name:              "invalid stored hash returns raw compare error",
			user:              invalidHashUser,
			password:          "any-password",
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantRawHashError:  true,
		},
		{
			name:              "delete refresh tokens error mapped",
			user:              validUser,
			password:          "correct-password",
			deleteErr:         postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantDeleteCalls:   1,
		},
		{
			name:              "new refresh token error mapped",
			user:              validUser,
			password:          "correct-password",
			newTokenErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantDeleteCalls:   1,
			wantNewTokenCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, users, tokens, roles := newTestIdentityService(t)

			users.getByEmailFn = func(email string) (*domain.User, error) {
				if tt.getUserErr != nil {
					return nil, tt.getUserErr
				}
				return tt.user, nil
			}
			roles.getAllForUserFn = func(userID int64) (domain.Roles, error) {
				if tt.getRolesErr != nil {
					return nil, tt.getRolesErr
				}
				return domain.Roles{"user"}, nil
			}
			tokens.deleteAllFromUserFn = func(scope string, userID int64) error {
				return tt.deleteErr
			}
			tokens.newOpaqueTokenFn = func(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
				if tt.newTokenErr != nil {
					return nil, tt.newTokenErr
				}
				return &domain.OpaqueToken{Plaintext: "new-refresh-token"}, nil
			}
			refreshToken, accessToken, err := svc.AuthenticateUser("alice@example.com", tt.password)

			if tt.wantRawHashError {
				if err == nil {
					t.Fatal("expected compare hash error, got nil")
				}
				if errors.Is(err, ErrInvalidCredentials) {
					t.Fatal("expected raw hash compare error, got ErrInvalidCredentials")
				}
			} else {
				assertError(t, err, tt.wantErr)
			}

			if refreshToken != tt.wantRefreshToken {
				t.Fatalf("expected refresh token %q, got %q", tt.wantRefreshToken, refreshToken)
			}
			if tt.wantAccessToken && accessToken == "" {
				t.Fatal("expected non-empty access token")
			}
			if !tt.wantAccessToken && accessToken != "" {
				t.Fatalf("expected empty access token, got %q", accessToken)
			}

			if users.getByEmailCalls != tt.wantGetUserCalls {
				t.Fatalf("expected GetUserByEmail calls %d, got %d", tt.wantGetUserCalls, users.getByEmailCalls)
			}
			if roles.getAllCalls != tt.wantGetRolesCalls {
				t.Fatalf("expected GetAllRolesForUser calls %d, got %d", tt.wantGetRolesCalls, roles.getAllCalls)
			}
			if tokens.deleteCalls != tt.wantDeleteCalls {
				t.Fatalf("expected DeleteAllFromUser calls %d, got %d", tt.wantDeleteCalls, tokens.deleteCalls)
			}
			if len(tokens.newOpaqueTokenCalls) != tt.wantNewTokenCalls {
				t.Fatalf("expected NewOpaqueToken calls %d, got %d", tt.wantNewTokenCalls, len(tokens.newOpaqueTokenCalls))
			}

			if tt.wantErr == nil && !tt.wantRawHashError {
				if users.lastEmailLookup != "alice@example.com" {
					t.Fatalf("expected email lookup alice@example.com, got %q", users.lastEmailLookup)
				}
				if tokens.lastDeleteScope != domain.ScopeRefresh {
					t.Fatalf("expected refresh delete scope, got %q", tokens.lastDeleteScope)
				}
			}
		})
	}
}

func TestIdentityService_RefreshToken(t *testing.T) {
	tests := []struct {
		name              string
		getUserErr        error
		getRolesErr       error
		deleteErr         error
		newTokenErr       error
		wantErr           error
		wantGetUserCalls  int
		wantGetRolesCalls int
		wantDeleteCalls   int
		wantNewTokenCalls int
		wantRefreshToken  string
		wantAccessToken   bool
	}{
		{
			name:              "success",
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantDeleteCalls:   1,
			wantNewTokenCalls: 1,
			wantRefreshToken:  "rotated-refresh-token",
			wantAccessToken:   true,
		},
		{
			name:             "get user error mapped",
			getUserErr:       postgres.ErrRecordNotFound,
			wantErr:          ErrRecordNotFound,
			wantGetUserCalls: 1,
		},
		{
			name:              "get roles error mapped",
			getRolesErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
		{
			name:              "delete error mapped",
			deleteErr:         postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantDeleteCalls:   1,
		},
		{
			name:              "new token error mapped",
			newTokenErr:       postgres.ErrRecordNotFound,
			wantErr:           ErrRecordNotFound,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
			wantDeleteCalls:   1,
			wantNewTokenCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, users, tokens, roles := newTestIdentityService(t)
			users.getForTokenFn = func(scope, plaintext string) (*domain.User, error) {
				if tt.getUserErr != nil {
					return nil, tt.getUserErr
				}
				return &domain.User{ID: 88, Password: domain.Password{Hash: []byte("hash")}}, nil
			}
			roles.getAllForUserFn = func(userID int64) (domain.Roles, error) {
				if tt.getRolesErr != nil {
					return nil, tt.getRolesErr
				}
				return domain.Roles{"user"}, nil
			}
			tokens.deleteAllFromUserFn = func(scope string, userID int64) error {
				return tt.deleteErr
			}
			tokens.newOpaqueTokenFn = func(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
				if tt.newTokenErr != nil {
					return nil, tt.newTokenErr
				}
				return &domain.OpaqueToken{Plaintext: "rotated-refresh-token"}, nil
			}
			refreshToken, accessToken, err := svc.RefreshToken("old-refresh-token")
			assertError(t, err, tt.wantErr)

			if refreshToken != tt.wantRefreshToken {
				t.Fatalf("expected refresh token %q, got %q", tt.wantRefreshToken, refreshToken)
			}
			if tt.wantAccessToken && accessToken == "" {
				t.Fatal("expected non-empty access token")
			}
			if !tt.wantAccessToken && accessToken != "" {
				t.Fatalf("expected empty access token, got %q", accessToken)
			}

			if users.getForTokenCalls != tt.wantGetUserCalls {
				t.Fatalf("expected GetForOpaqueToken calls %d, got %d", tt.wantGetUserCalls, users.getForTokenCalls)
			}
			if roles.getAllCalls != tt.wantGetRolesCalls {
				t.Fatalf("expected GetAllRolesForUser calls %d, got %d", tt.wantGetRolesCalls, roles.getAllCalls)
			}
			if tokens.deleteCalls != tt.wantDeleteCalls {
				t.Fatalf("expected DeleteAllFromUser calls %d, got %d", tt.wantDeleteCalls, tokens.deleteCalls)
			}
			if len(tokens.newOpaqueTokenCalls) != tt.wantNewTokenCalls {
				t.Fatalf("expected NewOpaqueToken calls %d, got %d", tt.wantNewTokenCalls, len(tokens.newOpaqueTokenCalls))
			}

			if tt.wantErr == nil {
				if users.lastTokenScope != domain.ScopeRefresh {
					t.Fatalf("expected refresh lookup scope, got %q", users.lastTokenScope)
				}
				if users.lastTokenPlaintext != "old-refresh-token" {
					t.Fatalf("expected old refresh token lookup, got %q", users.lastTokenPlaintext)
				}
			}
		})
	}
}

func TestIdentityService_GetUserByID(t *testing.T) {
	tests := []struct {
		name              string
		getUserErr        error
		getRolesErr       error
		wantErr           error
		wantGetUserCalls  int
		wantGetRolesCalls int
	}{
		{
			name:              "success",
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
		{
			name:             "get user error mapped",
			getUserErr:       postgres.ErrRecordNotFound,
			wantErr:          ErrRecordNotFound,
			wantGetUserCalls: 1,
		},
		{
			name:              "get roles error returns ErrUserWithoutRoles",
			getRolesErr:       errors.New("roles fail"),
			wantErr:           ErrUserWithoutRoles,
			wantGetUserCalls:  1,
			wantGetRolesCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, users, _, roles := newTestIdentityService(t)

			users.getByIDFn = func(userID int64) (*domain.User, error) {
				if tt.getUserErr != nil {
					return nil, tt.getUserErr
				}
				return &domain.User{ID: userID, Email: "alice@example.com", Username: "alice", Password: domain.Password{Hash: []byte("hash")}}, nil
			}
			roles.getAllForUserFn = func(userID int64) (domain.Roles, error) {
				if tt.getRolesErr != nil {
					return nil, tt.getRolesErr
				}
				return domain.Roles{"user", "admin"}, nil
			}

			userDetails, err := svc.GetUserById(21)
			assertError(t, err, tt.wantErr)
			if tt.wantErr != nil {
				if userDetails != nil {
					t.Fatal("expected nil user details on error")
				}
			} else {
				if userDetails == nil {
					t.Fatal("expected non-nil user details")
				}
				if userDetails.ID != 21 {
					t.Fatalf("expected user id 21, got %d", userDetails.ID)
				}
				if !reflect.DeepEqual(userDetails.Roles, domain.Roles{"user", "admin"}) {
					t.Fatalf("expected roles [user admin], got %v", userDetails.Roles)
				}
			}

			if users.getByIDCalls != tt.wantGetUserCalls {
				t.Fatalf("expected GetUserById calls %d, got %d", tt.wantGetUserCalls, users.getByIDCalls)
			}
			if roles.getAllCalls != tt.wantGetRolesCalls {
				t.Fatalf("expected GetAllRolesForUser calls %d, got %d", tt.wantGetRolesCalls, roles.getAllCalls)
			}
		})
	}
}

func TestIdentityService_ValidateJWTToken(t *testing.T) {
	svc, _, _, _ := newTestIdentityService(t)

	goodToken, err := svc.newAccessToken(123, domain.Roles{"user"}, time.Hour, svc.cfg.Issuer)
	if err != nil {
		t.Fatalf("failed to generate good token: %v", err)
	}

	wrongIssuerToken, err := svc.newAccessToken(123, domain.Roles{"user"}, time.Hour, "https://other-issuer")
	if err != nil {
		t.Fatalf("failed to generate wrong issuer token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		issuer    string
		wantErr   error
		wantValid bool
	}{
		{name: "valid token", token: goodToken, issuer: svc.cfg.Issuer, wantValid: true},
		{name: "malformed token", token: "not-a-token", issuer: svc.cfg.Issuer, wantErr: jwt.ErrTokenMalformed},
		{name: "wrong issuer", token: wrongIssuerToken, issuer: svc.cfg.Issuer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.ValidateJWTToken(tt.token, tt.issuer)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if token != nil {
					t.Fatal("expected nil token on error")
				}
				return
			}

			if tt.wantValid {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if token == nil {
					t.Fatal("expected non-nil token")
				}
				claims, ok := token.Claims.(*claims)
				if !ok {
					t.Fatalf("expected *claims, got %T", token.Claims)
				}
				if claims.Subject != "123" {
					t.Fatalf("expected subject 123, got %q", claims.Subject)
				}
				if !reflect.DeepEqual(claims.Roles, []string{"user"}) {
					t.Fatalf("expected roles [user], got %v", claims.Roles)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if token != nil {
				t.Fatal("expected nil token on invalid token")
			}
		})
	}
}

func TestContextClaimsHelpers(t *testing.T) {
	t.Run("set and get claims", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		token := &jwt.Token{Claims: &claims{Roles: []string{"user", "admin"}, RegisteredClaims: jwt.RegisteredClaims{Subject: "42"}}}

		req = ContextSetClaims(req, token)

		userID, err := ContextGetUserId(req)
		if err != nil {
			t.Fatalf("expected nil error for user id, got %v", err)
		}
		if userID != 42 {
			t.Fatalf("expected user id 42, got %d", userID)
		}

		roles, err := ContextGetRoles(req)
		if err != nil {
			t.Fatalf("expected nil error for roles, got %v", err)
		}
		if !reflect.DeepEqual(roles, []string{"user", "admin"}) {
			t.Fatalf("expected roles [user admin], got %v", roles)
		}
	})

	t.Run("missing user id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := ContextGetUserId(req)
		assertError(t, err, ErrMissingUserIDInContext)
	})

	t.Run("invalid user id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), userIdContextKey, "not-an-int")
		req = req.WithContext(ctx)

		_, err := ContextGetUserId(req)
		if err == nil {
			t.Fatal("expected parse error, got nil")
		}
	})

	t.Run("missing roles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := ContextGetRoles(req)
		assertError(t, err, ErrMissingUserRolesInContext)
	})
}
