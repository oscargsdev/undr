package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/service"
)

type mockIdentityService struct {
	registerUserFn     func(user *domain.User) (*domain.OpaqueToken, error)
	activateUserFn     func(tokenPlainText string) (string, string, error)
	authenticateUserFn func(email, password string) (string, string, error)
	getUserByIDFn      func(userID int64) (*domain.UserDetails, error)
	refreshTokenFn     func(oldRefreshToken string) (string, string, error)
	logoutFn           func(userID int64) error
	getIssuerFn        func() string
	getJWKSFn          func(r *http.Request) (json.RawMessage, error)
	validateJWTTokenFn func(tokenString string, issuer string) (*jwt.Token, error)

	registerUserCalls     int
	activateUserCalls     int
	authenticateUserCalls int
	getUserByIDCalls      int
	refreshTokenCalls     int
	logoutCalls           int
	getIssuerCalls        int
	getJWKSCalls          int
	validateJWTTokenCalls int
}

func (m *mockIdentityService) RegisterUser(user *domain.User) (*domain.OpaqueToken, error) {
	m.registerUserCalls++
	if m.registerUserFn == nil {
		panic("unexpected RegisterUser call")
	}
	return m.registerUserFn(user)
}

func (m *mockIdentityService) ActivateUser(tokenPlainText string) (string, string, error) {
	m.activateUserCalls++
	if m.activateUserFn == nil {
		panic("unexpected ActivateUser call")
	}
	return m.activateUserFn(tokenPlainText)
}

func (m *mockIdentityService) AuthenticateUser(email, password string) (string, string, error) {
	m.authenticateUserCalls++
	if m.authenticateUserFn == nil {
		panic("unexpected AuthenticateUser call")
	}
	return m.authenticateUserFn(email, password)
}

func (m *mockIdentityService) GetUserById(userID int64) (*domain.UserDetails, error) {
	m.getUserByIDCalls++
	if m.getUserByIDFn == nil {
		panic("unexpected GetUserById call")
	}
	return m.getUserByIDFn(userID)
}

func (m *mockIdentityService) RefreshToken(oldRefreshToken string) (string, string, error) {
	m.refreshTokenCalls++
	if m.refreshTokenFn == nil {
		panic("unexpected RefreshToken call")
	}
	return m.refreshTokenFn(oldRefreshToken)
}

func (m *mockIdentityService) Logout(userID int64) error {
	m.logoutCalls++
	if m.logoutFn == nil {
		panic("unexpected Logout call")
	}
	return m.logoutFn(userID)
}

func (m *mockIdentityService) GetIssuer() string {
	m.getIssuerCalls++
	if m.getIssuerFn == nil {
		panic("unexpected GetIssuer call")
	}
	return m.getIssuerFn()
}

func (m *mockIdentityService) GetJWKS(r *http.Request) (json.RawMessage, error) {
	m.getJWKSCalls++
	if m.getJWKSFn == nil {
		panic("unexpected GetJWKS call")
	}
	return m.getJWKSFn(r)
}

func (m *mockIdentityService) ValidateJWTToken(tokenString string, issuer string) (*jwt.Token, error) {
	m.validateJWTTokenCalls++
	if m.validateJWTTokenFn == nil {
		panic("unexpected ValidateJWTToken call")
	}
	return m.validateJWTTokenFn(tokenString, issuer)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestHandler(svc IdentityService) *Handler {
	return NewHandler(svc, newTestLogger())
}

func decodeJSONResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("expected JSON response, got error: %v; body: %q", err, rr.Body.String())
	}
	return got
}

func errorString(t *testing.T, response map[string]any) string {
	t.Helper()

	raw, ok := response["error"]
	if !ok {
		t.Fatalf("response missing error key: %v", response)
	}

	msg, ok := raw.(string)
	if !ok {
		t.Fatalf("expected error message to be string, got %T (%v)", raw, raw)
	}
	return msg
}

func errorMap(t *testing.T, response map[string]any) map[string]any {
	t.Helper()

	raw, ok := response["error"]
	if !ok {
		t.Fatalf("response missing error key: %v", response)
	}

	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("expected error message to be object, got %T (%v)", raw, raw)
	}
	return m
}

func mustString(t *testing.T, v any, field string) string {
	t.Helper()

	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected %s to be string, got %T (%v)", field, v, v)
	}
	return s
}

func cookieByName(t *testing.T, rr *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}

	t.Fatalf("expected cookie %q to be set, got %v", name, rr.Result().Cookies())
	return nil
}

type authFixtureUsersRepo struct {
	user *domain.User
}

func (r *authFixtureUsersRepo) InsertUser(*domain.User) error { panic("unexpected InsertUser call") }
func (r *authFixtureUsersRepo) UpdateUser(*domain.User) error { panic("unexpected UpdateUser call") }
func (r *authFixtureUsersRepo) GetForOpaqueToken(string, string) (*domain.User, error) {
	panic("unexpected GetForOpaqueToken call")
}
func (r *authFixtureUsersRepo) GetUserByEmail(string) (*domain.User, error) {
	return r.user, nil
}
func (r *authFixtureUsersRepo) GetUserById(int64) (*domain.User, error) {
	panic("unexpected GetUserById call")
}

type authFixtureTokensRepo struct{}

func (r *authFixtureTokensRepo) NewOpaqueToken(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
	return &domain.OpaqueToken{
		UserID:    userID,
		Scope:     scope,
		Plaintext: strings.Repeat("r", 26),
		Expiry:    time.Now().Add(ttl),
	}, nil
}

func (r *authFixtureTokensRepo) DeleteAllFromUser(string, int64) error { return nil }

type authFixtureRolesRepo struct {
	roles domain.Roles
}

func (r *authFixtureRolesRepo) GetAllRolesForUser(int64) (domain.Roles, error) {
	return r.roles, nil
}
func (r *authFixtureRolesRepo) AddRoleForUser(int64, ...string) error {
	panic("unexpected AddRoleForUser call")
}

func newJWTFixtureServiceAndToken(t *testing.T, userID int64, roles []string) (IdentityService, string) {
	t.Helper()

	password := domain.Password{}
	if err := password.Set("correct horse battery staple"); err != nil {
		t.Fatalf("setup failed hashing password: %v", err)
	}

	users := &authFixtureUsersRepo{
		user: &domain.User{
			ID:        userID,
			Username:  "alice",
			Email:     "alice@example.com",
			Password:  password,
			Activated: true,
		},
	}

	svc, err := service.New(service.Config{
		UsersRepository:        users,
		OpaqueTokensRepository: &authFixtureTokensRepo{},
		RolesRepository:        &authFixtureRolesRepo{roles: roles},
		Logger:                 newTestLogger(),
		Issuer:                 "https://issuer.example",
		JWTExpiration:          time.Hour,
		RefreshExpiration:      24 * time.Hour,
		ActivationExpiration:   24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("setup failed creating identity service: %v", err)
	}

	_, accessToken, err := svc.AuthenticateUser("alice@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("setup failed creating access token: %v", err)
	}

	return svc, accessToken
}

func withAuthContext(t *testing.T, req *http.Request, userID int64, roles []string) *http.Request {
	t.Helper()

	svc, tokenString := newJWTFixtureServiceAndToken(t, userID, roles)
	token, err := svc.ValidateJWTToken(tokenString, svc.GetIssuer())
	if err != nil {
		t.Fatalf("setup failed validating access token: %v", err)
	}

	return service.ContextSetClaims(req, token)
}

func assertHeaderContains(t *testing.T, got []string, want string, key string) {
	t.Helper()

	for _, v := range got {
		if v == want {
			return
		}
	}

	t.Fatalf("expected header %s to contain %q, got %v", key, want, got)
}

func assertErrorContains(t *testing.T, rr *httptest.ResponseRecorder, fragment string) {
	t.Helper()

	response := decodeJSONResponse(t, rr)
	msg := errorString(t, response)
	if !strings.Contains(msg, fragment) {
		t.Fatalf("expected error message to contain %q, got %q", fragment, msg)
	}
}

func assertExactError(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	response := decodeJSONResponse(t, rr)
	got := errorString(t, response)
	if got != want {
		t.Fatalf("expected error message %q, got %q", want, got)
	}
}

func assertStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rr.Code != want {
		t.Fatalf("expected status %d, got %d. body: %s", want, rr.Code, rr.Body.String())
	}
}

func assertFloat64EqualsInt64(t *testing.T, got any, want int64, field string) {
	t.Helper()

	n, ok := got.(float64)
	if !ok {
		t.Fatalf("expected %s to be number, got %T (%v)", field, got, got)
	}
	if int64(n) != want {
		t.Fatalf("expected %s to be %d, got %v", field, want, got)
	}
}

func assertErrorField(t *testing.T, rr *httptest.ResponseRecorder, field string, want string) {
	t.Helper()

	response := decodeJSONResponse(t, rr)
	errs := errorMap(t, response)
	got := mustString(t, errs[field], fmt.Sprintf("error.%s", field))
	if got != want {
		t.Fatalf("expected error.%s to be %q, got %q", field, want, got)
	}
}
