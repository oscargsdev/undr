package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/service"
)

func TestNewRefreshTokenCookie(t *testing.T) {
	expires := time.Now().Add(time.Hour).UTC()
	cookie := newRefreshTokenCookie("token-value", expires)

	if cookie.Name != "refresh_token" {
		t.Fatalf("expected cookie name refresh_token, got %q", cookie.Name)
	}
	if cookie.Value != "token-value" {
		t.Fatalf("expected cookie value token-value, got %q", cookie.Value)
	}
	if cookie.Path != "/v1/identity/refresh" {
		t.Fatalf("expected cookie path /v1/identity/refresh, got %q", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected cookie HttpOnly to be true")
	}
	if !cookie.Secure {
		t.Fatal("expected cookie Secure to be true")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite strict mode, got %v", cookie.SameSite)
	}
	if !cookie.Expires.Equal(expires) {
		t.Fatalf("expected cookie expiry %v, got %v", expires, cookie.Expires)
	}
}

func TestHandler_RegisterUserHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*mockIdentityService)
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder, *mockIdentityService)
	}{
		{
			name:       "malformed json returns bad request",
			body:       `{"username":`,
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorContains(t, rr, "body contains badly-formed JSON")
				if svc.registerUserCalls != 0 {
					t.Fatalf("expected RegisterUser not to be called, got %d calls", svc.registerUserCalls)
				}
			},
		},
		{
			name:       "password longer than bcrypt limit returns validation error",
			body:       `{"username":"alice","email":"alice@example.com","password":"` + strings.Repeat("a", 73) + `"}`,
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "password", "must not be more than 72 bytes long")
				if svc.registerUserCalls != 0 {
					t.Fatalf("expected RegisterUser not to be called, got %d calls", svc.registerUserCalls)
				}
			},
		},
		{
			name:       "validation errors return unprocessable entity",
			body:       `{"username":"","email":"invalid","password":"short"}`,
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "username", "must be provided")
				assertErrorField(t, rr, "email", "must be a valid email address")
				assertErrorField(t, rr, "password", "must be at least 8 bytes long")
				if svc.registerUserCalls != 0 {
					t.Fatalf("expected RegisterUser not to be called, got %d calls", svc.registerUserCalls)
				}
			},
		},
		{
			name: "duplicate email maps to validation error",
			body: `{"username":"alice","email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.registerUserFn = func(ctx context.Context, user *domain.User) (*domain.OpaqueToken, error) {
					return nil, service.ErrDuplicateEmail
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "email", "a user with this email address already exists")
				if svc.registerUserCalls != 1 {
					t.Fatalf("expected RegisterUser to be called once, got %d calls", svc.registerUserCalls)
				}
			},
		},
		{
			name: "duplicate username maps to validation error",
			body: `{"username":"alice","email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.registerUserFn = func(ctx context.Context, user *domain.User) (*domain.OpaqueToken, error) {
					return nil, service.ErrDuplicateUsername
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "username", "a user with this username already exists")
				if svc.registerUserCalls != 1 {
					t.Fatalf("expected RegisterUser to be called once, got %d calls", svc.registerUserCalls)
				}
			},
		},
		{
			name: "unexpected service error returns internal server error",
			body: `{"username":"alice","email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.registerUserFn = func(ctx context.Context, user *domain.User) (*domain.OpaqueToken, error) {
					return nil, errors.New("db unavailable")
				}
			},
			wantStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the server encountered an error and could not process your request")
				if svc.registerUserCalls != 1 {
					t.Fatalf("expected RegisterUser to be called once, got %d calls", svc.registerUserCalls)
				}
			},
		},
		{
			name: "success returns accepted user payload",
			body: `{"username":"alice","email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.registerUserFn = func(ctx context.Context, user *domain.User) (*domain.OpaqueToken, error) {
					return &domain.OpaqueToken{Plaintext: "activation-token"}, nil
				}
			},
			wantStatus: http.StatusAccepted,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				if svc.registerUserCalls != 1 {
					t.Fatalf("expected RegisterUser to be called once, got %d calls", svc.registerUserCalls)
				}

				response := decodeJSONResponse(t, rr)
				activationToken := mustString(t, response["activation_token"], "activation_token")
				if activationToken != "activation-token" {
					t.Fatalf("expected activation_token to be activation-token, got %q", activationToken)
				}

				userPayload, ok := response["user"].(map[string]any)
				if !ok {
					t.Fatalf("expected user payload object, got %T (%v)", response["user"], response["user"])
				}

				if mustString(t, userPayload["username"], "user.username") != "alice" {
					t.Fatalf("unexpected user.username value: %v", userPayload["username"])
				}
				if mustString(t, userPayload["email"], "user.email") != "alice@example.com" {
					t.Fatalf("unexpected user.email value: %v", userPayload["email"])
				}
				activated, ok := userPayload["activated"].(bool)
				if !ok {
					t.Fatalf("expected user.activated to be bool, got %T (%v)", userPayload["activated"], userPayload["activated"])
				}
				if activated {
					t.Fatal("expected new user to be non-activated")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{}
			if tt.setup != nil {
				tt.setup(svc)
			}

			handler := newTestHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.registerUserHandler(rr, req)

			assertStatus(t, rr, tt.wantStatus)
			tt.assert(t, rr, svc)
		})
	}
}

func TestHandler_RegisterUserHandlerPassesRequestContext(t *testing.T) {
	svc := &mockIdentityService{
		registerUserFn: func(ctx context.Context, user *domain.User) (*domain.OpaqueToken, error) {
			return &domain.OpaqueToken{Plaintext: "activation-token"}, nil
		},
	}
	handler := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(
		`{"username":"alice","email":"alice@example.com","password":"super-secure"}`,
	))
	rr := httptest.NewRecorder()

	handler.registerUserHandler(rr, req)

	assertStatus(t, rr, http.StatusAccepted)
	if svc.lastContext != req.Context() {
		t.Fatal("expected RegisterUser to receive the request context")
	}
}

func TestHandler_ActivateUserHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*mockIdentityService)
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder, *mockIdentityService)
	}{
		{
			name:       "invalid json returns bad request",
			body:       `{"activationToken":`,
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorContains(t, rr, "body contains badly-formed JSON")
				if svc.activateUserCalls != 0 {
					t.Fatalf("expected ActivateUser not to be called, got %d calls", svc.activateUserCalls)
				}
			},
		},
		{
			name:       "invalid token format returns validation error",
			body:       `{"activationToken":"short"}`,
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "token", "must be 26 bytes long")
				if svc.activateUserCalls != 0 {
					t.Fatalf("expected ActivateUser not to be called, got %d calls", svc.activateUserCalls)
				}
			},
		},
		{
			name: "record not found maps to token validation error",
			body: `{"activationToken":"` + strings.Repeat("a", 26) + `"}`,
			setup: func(svc *mockIdentityService) {
				svc.activateUserFn = func(ctx context.Context, tokenPlainText string) (string, string, error) {
					return "", "", service.ErrRecordNotFound
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "token", "invalid or expired activation token")
				if svc.activateUserCalls != 1 {
					t.Fatalf("expected ActivateUser to be called once, got %d calls", svc.activateUserCalls)
				}
			},
		},
		{
			name: "edit conflict maps to conflict status",
			body: `{"activationToken":"` + strings.Repeat("a", 26) + `"}`,
			setup: func(svc *mockIdentityService) {
				svc.activateUserFn = func(ctx context.Context, tokenPlainText string) (string, string, error) {
					return "", "", service.ErrEditConflict
				}
			},
			wantStatus: http.StatusConflict,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "unable to update the record due to an edit conflict, please try again")
				if svc.activateUserCalls != 1 {
					t.Fatalf("expected ActivateUser to be called once, got %d calls", svc.activateUserCalls)
				}
			},
		},
		{
			name: "unexpected service error returns internal server error",
			body: `{"activationToken":"` + strings.Repeat("a", 26) + `"}`,
			setup: func(svc *mockIdentityService) {
				svc.activateUserFn = func(ctx context.Context, tokenPlainText string) (string, string, error) {
					return "", "", errors.New("boom")
				}
			},
			wantStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the server encountered an error and could not process your request")
				if svc.activateUserCalls != 1 {
					t.Fatalf("expected ActivateUser to be called once, got %d calls", svc.activateUserCalls)
				}
			},
		},
		{
			name: "success returns access token and sets refresh cookie",
			body: `{"activationToken":"` + strings.Repeat("a", 26) + `"}`,
			setup: func(svc *mockIdentityService) {
				svc.activateUserFn = func(ctx context.Context, tokenPlainText string) (string, string, error) {
					return "new-refresh-token", "new-access-token", nil
				}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				if svc.activateUserCalls != 1 {
					t.Fatalf("expected ActivateUser to be called once, got %d calls", svc.activateUserCalls)
				}

				response := decodeJSONResponse(t, rr)
				if got := mustString(t, response["access_token"], "access_token"); got != "new-access-token" {
					t.Fatalf("expected access_token new-access-token, got %q", got)
				}

				cookie := cookieByName(t, rr, "refresh_token")
				if cookie.Value != "new-refresh-token" {
					t.Fatalf("expected refresh token cookie value new-refresh-token, got %q", cookie.Value)
				}
				if cookie.Path != "/v1/identity/refresh" {
					t.Fatalf("expected refresh token cookie path /v1/identity/refresh, got %q", cookie.Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{}
			if tt.setup != nil {
				tt.setup(svc)
			}

			handler := newTestHandler(svc)
			req := httptest.NewRequest(http.MethodPut, "/activate", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.activateUserHandler(rr, req)

			assertStatus(t, rr, tt.wantStatus)
			tt.assert(t, rr, svc)
		})
	}
}

func TestHandler_TestSecuredEndpoint(t *testing.T) {
	t.Run("missing auth context returns unauthorized", func(t *testing.T) {
		handler := newTestHandler(&mockIdentityService{})
		req := httptest.NewRequest(http.MethodGet, "/secured", nil)
		rr := httptest.NewRecorder()

		handler.testSecuredEndpoint(rr, req)

		assertStatus(t, rr, http.StatusUnauthorized)
		assertExactError(t, rr, "you must be authenticated to access this resource")
	})

	t.Run("returns user id and roles from context", func(t *testing.T) {
		handler := newTestHandler(&mockIdentityService{})
		req := httptest.NewRequest(http.MethodGet, "/secured", nil)
		req = withAuthContext(t, req, 44, []string{"user", "admin"})
		rr := httptest.NewRecorder()

		handler.testSecuredEndpoint(rr, req)

		assertStatus(t, rr, http.StatusOK)
		response := decodeJSONResponse(t, rr)
		assertFloat64EqualsInt64(t, response["userId"], 44, "userId")

		rawRoles, ok := response["roles"].([]any)
		if !ok {
			t.Fatalf("expected roles to be array, got %T (%v)", response["roles"], response["roles"])
		}
		if len(rawRoles) != 2 || mustString(t, rawRoles[0], "roles[0]") != "user" || mustString(t, rawRoles[1], "roles[1]") != "admin" {
			t.Fatalf("unexpected roles value: %v", rawRoles)
		}
	})
}

func TestHandler_AuthenticateUserHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*mockIdentityService)
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder, *mockIdentityService)
	}{
		{
			name:       "invalid json returns bad request",
			body:       `{"email":`,
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorContains(t, rr, "body contains badly-formed JSON")
				if svc.authenticateUserCalls != 0 {
					t.Fatalf("expected AuthenticateUser not to be called, got %d calls", svc.authenticateUserCalls)
				}
			},
		},
		{
			name:       "validation errors return unprocessable entity",
			body:       `{"email":"not-an-email","password":"short"}`,
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "email", "must be a valid email address")
				assertErrorField(t, rr, "password", "must be at least 8 bytes long")
				if svc.authenticateUserCalls != 0 {
					t.Fatalf("expected AuthenticateUser not to be called, got %d calls", svc.authenticateUserCalls)
				}
			},
		},
		{
			name: "record not found returns invalid credentials",
			body: `{"email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.authenticateUserFn = func(ctx context.Context, email, password string) (string, string, error) {
					return "", "", service.ErrRecordNotFound
				}
			},
			wantStatus: http.StatusUnauthorized,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "invalid authentication credentials")
				if svc.authenticateUserCalls != 1 {
					t.Fatalf("expected AuthenticateUser to be called once, got %d calls", svc.authenticateUserCalls)
				}
			},
		},
		{
			name: "invalid credentials returns invalid credentials",
			body: `{"email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.authenticateUserFn = func(ctx context.Context, email, password string) (string, string, error) {
					return "", "", service.ErrInvalidCredentials
				}
			},
			wantStatus: http.StatusUnauthorized,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "invalid authentication credentials")
				if svc.authenticateUserCalls != 1 {
					t.Fatalf("expected AuthenticateUser to be called once, got %d calls", svc.authenticateUserCalls)
				}
			},
		},
		{
			name: "user not activated returns forbidden",
			body: `{"email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.authenticateUserFn = func(ctx context.Context, email, password string) (string, string, error) {
					return "", "", service.ErrUserNotActivated
				}
			},
			wantStatus: http.StatusForbidden,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "your user account must be activated to access this resource")
				if svc.authenticateUserCalls != 1 {
					t.Fatalf("expected AuthenticateUser to be called once, got %d calls", svc.authenticateUserCalls)
				}
			},
		},
		{
			name: "unexpected service error returns internal server error",
			body: `{"email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.authenticateUserFn = func(ctx context.Context, email, password string) (string, string, error) {
					return "", "", errors.New("db timeout")
				}
			},
			wantStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the server encountered an error and could not process your request")
				if svc.authenticateUserCalls != 1 {
					t.Fatalf("expected AuthenticateUser to be called once, got %d calls", svc.authenticateUserCalls)
				}
			},
		},
		{
			name: "success sets refresh cookie and returns access token",
			body: `{"email":"alice@example.com","password":"super-secure"}`,
			setup: func(svc *mockIdentityService) {
				svc.authenticateUserFn = func(ctx context.Context, email, password string) (string, string, error) {
					return "new-refresh-token", "new-access-token", nil
				}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				if svc.authenticateUserCalls != 1 {
					t.Fatalf("expected AuthenticateUser to be called once, got %d calls", svc.authenticateUserCalls)
				}
				response := decodeJSONResponse(t, rr)
				if got := mustString(t, response["access_token"], "access_token"); got != "new-access-token" {
					t.Fatalf("expected access_token new-access-token, got %q", got)
				}

				cookie := cookieByName(t, rr, "refresh_token")
				if cookie.Value != "new-refresh-token" {
					t.Fatalf("expected refresh token cookie value new-refresh-token, got %q", cookie.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{}
			if tt.setup != nil {
				tt.setup(svc)
			}
			handler := newTestHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/authenticate", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.authenticateUserHandler(rr, req)

			assertStatus(t, rr, tt.wantStatus)
			tt.assert(t, rr, svc)
		})
	}
}

func TestHandler_RefreshTokenHandler(t *testing.T) {
	tests := []struct {
		name       string
		cookie     *http.Cookie
		setup      func(*mockIdentityService)
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder, *mockIdentityService)
	}{
		{
			name:       "missing cookie returns bad request",
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorContains(t, rr, "named cookie not present")
				if svc.refreshTokenCalls != 0 {
					t.Fatalf("expected RefreshToken not to be called, got %d calls", svc.refreshTokenCalls)
				}
			},
		},
		{
			name:       "invalid cookie token format returns bad request",
			cookie:     &http.Cookie{Name: "refresh_token", Value: "short"},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorField(t, rr, "token", "must be 26 bytes long")
				if svc.refreshTokenCalls != 0 {
					t.Fatalf("expected RefreshToken not to be called, got %d calls", svc.refreshTokenCalls)
				}
			},
		},
		{
			name:       "service not found maps to invalid refresh token",
			cookie:     &http.Cookie{Name: "refresh_token", Value: strings.Repeat("r", 26)},
			wantStatus: http.StatusBadRequest,
			setup: func(svc *mockIdentityService) {
				svc.refreshTokenFn = func(ctx context.Context, oldRefreshToken string) (string, string, error) {
					return "", "", service.ErrRecordNotFound
				}
			},
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "invalid or expired refresh token")
				if svc.refreshTokenCalls != 1 {
					t.Fatalf("expected RefreshToken to be called once, got %d calls", svc.refreshTokenCalls)
				}
			},
		},
		{
			name:       "unexpected service error returns internal server error",
			cookie:     &http.Cookie{Name: "refresh_token", Value: strings.Repeat("r", 26)},
			wantStatus: http.StatusInternalServerError,
			setup: func(svc *mockIdentityService) {
				svc.refreshTokenFn = func(ctx context.Context, oldRefreshToken string) (string, string, error) {
					return "", "", errors.New("unexpected error")
				}
			},
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the server encountered an error and could not process your request")
				if svc.refreshTokenCalls != 1 {
					t.Fatalf("expected RefreshToken to be called once, got %d calls", svc.refreshTokenCalls)
				}
			},
		},
		{
			name:       "success rotates refresh token and returns access token",
			cookie:     &http.Cookie{Name: "refresh_token", Value: strings.Repeat("r", 26)},
			wantStatus: http.StatusOK,
			setup: func(svc *mockIdentityService) {
				svc.refreshTokenFn = func(ctx context.Context, oldRefreshToken string) (string, string, error) {
					return "rotated-refresh", "new-access", nil
				}
			},
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				if svc.refreshTokenCalls != 1 {
					t.Fatalf("expected RefreshToken to be called once, got %d calls", svc.refreshTokenCalls)
				}
				response := decodeJSONResponse(t, rr)
				if got := mustString(t, response["access_token"], "access_token"); got != "new-access" {
					t.Fatalf("expected access_token new-access, got %q", got)
				}

				cookie := cookieByName(t, rr, "refresh_token")
				if cookie.Value != "rotated-refresh" {
					t.Fatalf("expected rotated refresh token cookie, got %q", cookie.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{}
			if tt.setup != nil {
				tt.setup(svc)
			}
			handler := newTestHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rr := httptest.NewRecorder()

			handler.refreshTokenHandler(rr, req)

			assertStatus(t, rr, tt.wantStatus)
			tt.assert(t, rr, svc)
		})
	}
}

func TestHandler_LogoutHandler(t *testing.T) {
	t.Run("missing auth context returns unauthorized", func(t *testing.T) {
		handler := newTestHandler(&mockIdentityService{})
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rr := httptest.NewRecorder()

		handler.logoutHandler(rr, req)

		assertStatus(t, rr, http.StatusUnauthorized)
		assertExactError(t, rr, "you must be authenticated to access this resource")
	})

	t.Run("service error returns internal server error", func(t *testing.T) {
		svc := &mockIdentityService{
			logoutFn: func(ctx context.Context, userID int64) error {
				if userID != 77 {
					t.Fatalf("expected userID 77, got %d", userID)
				}
				return errors.New("cannot revoke")
			},
		}
		handler := newTestHandler(svc)
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req = withAuthContext(t, req, 77, []string{"user"})
		rr := httptest.NewRecorder()

		handler.logoutHandler(rr, req)

		assertStatus(t, rr, http.StatusInternalServerError)
		assertExactError(t, rr, "the server encountered an error and could not process your request")
		if svc.logoutCalls != 1 {
			t.Fatalf("expected Logout to be called once, got %d calls", svc.logoutCalls)
		}
	})

	t.Run("success clears refresh cookie and returns no content", func(t *testing.T) {
		svc := &mockIdentityService{
			logoutFn: func(ctx context.Context, userID int64) error {
				if userID != 77 {
					t.Fatalf("expected userID 77, got %d", userID)
				}
				return nil
			},
		}
		handler := newTestHandler(svc)
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req = withAuthContext(t, req, 77, []string{"user"})
		rr := httptest.NewRecorder()

		handler.logoutHandler(rr, req)

		assertStatus(t, rr, http.StatusNoContent)
		if svc.logoutCalls != 1 {
			t.Fatalf("expected Logout to be called once, got %d calls", svc.logoutCalls)
		}
		cookie := cookieByName(t, rr, "refresh_token")
		if cookie.Value != "" {
			t.Fatalf("expected refresh token cookie to be cleared, got %q", cookie.Value)
		}
		if !cookie.Expires.Equal(time.Unix(0, 0)) {
			t.Fatalf("expected cleared cookie expiry %v, got %v", time.Unix(0, 0), cookie.Expires)
		}
	})
}

func TestHandler_onlyAdminsHandler(t *testing.T) {
	handler := newTestHandler(&mockIdentityService{})
	req := httptest.NewRequest(http.MethodGet, "/admin-portal", nil)
	rr := httptest.NewRecorder()

	handler.onlyAdminsHandler(rr, req)

	assertStatus(t, rr, http.StatusOK)
	response := decodeJSONResponse(t, rr)
	if got := mustString(t, response["howdy"], "howdy"); got != "you are an admin!" {
		t.Fatalf("expected howdy message for admin endpoint, got %q", got)
	}
}

func TestHandler_myInfoHandler(t *testing.T) {
	tests := []struct {
		name       string
		pathUserID string
		withAuth   bool
		authUserID int64
		setup      func(*mockIdentityService)
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder, *mockIdentityService)
	}{
		{
			name:       "invalid path user id returns bad request",
			pathUserID: "abc",
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertErrorContains(t, rr, "invalid syntax")
				if svc.getUserByIDCalls != 0 {
					t.Fatalf("expected GetUserById not to be called, got %d calls", svc.getUserByIDCalls)
				}
			},
		},
		{
			name:       "missing context user returns unauthorized",
			pathUserID: "12",
			wantStatus: http.StatusUnauthorized,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "you must be authenticated to access this resource")
				if svc.getUserByIDCalls != 0 {
					t.Fatalf("expected GetUserById not to be called, got %d calls", svc.getUserByIDCalls)
				}
			},
		},
		{
			name:       "mismatched path and context ids return invalid credentials",
			pathUserID: "12",
			withAuth:   true,
			authUserID: 77,
			wantStatus: http.StatusUnauthorized,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "invalid authentication credentials")
				if svc.getUserByIDCalls != 0 {
					t.Fatalf("expected GetUserById not to be called, got %d calls", svc.getUserByIDCalls)
				}
			},
		},
		{
			name:       "record not found returns not found",
			pathUserID: "12",
			withAuth:   true,
			authUserID: 12,
			setup: func(svc *mockIdentityService) {
				svc.getUserByIDFn = func(ctx context.Context, userID int64) (*domain.UserDetails, error) {
					return nil, service.ErrRecordNotFound
				}
			},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the requested resource could not be found")
				if svc.getUserByIDCalls != 1 {
					t.Fatalf("expected GetUserById to be called once, got %d calls", svc.getUserByIDCalls)
				}
			},
		},
		{
			name:       "user without roles returns internal server error",
			pathUserID: "12",
			withAuth:   true,
			authUserID: 12,
			setup: func(svc *mockIdentityService) {
				svc.getUserByIDFn = func(ctx context.Context, userID int64) (*domain.UserDetails, error) {
					return nil, service.ErrUserWithoutRoles
				}
			},
			wantStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the server encountered an error and could not process your request")
				if svc.getUserByIDCalls != 1 {
					t.Fatalf("expected GetUserById to be called once, got %d calls", svc.getUserByIDCalls)
				}
			},
		},
		{
			name:       "unexpected service error returns internal server error",
			pathUserID: "12",
			withAuth:   true,
			authUserID: 12,
			setup: func(svc *mockIdentityService) {
				svc.getUserByIDFn = func(ctx context.Context, userID int64) (*domain.UserDetails, error) {
					return nil, errors.New("boom")
				}
			},
			wantStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				assertExactError(t, rr, "the server encountered an error and could not process your request")
				if svc.getUserByIDCalls != 1 {
					t.Fatalf("expected GetUserById to be called once, got %d calls", svc.getUserByIDCalls)
				}
			},
		},
		{
			name:       "success returns user details",
			pathUserID: "12",
			withAuth:   true,
			authUserID: 12,
			setup: func(svc *mockIdentityService) {
				svc.getUserByIDFn = func(ctx context.Context, userID int64) (*domain.UserDetails, error) {
					return &domain.UserDetails{
						User: domain.User{
							ID:        12,
							Username:  "alice",
							Email:     "alice@example.com",
							Activated: true,
						},
						Roles: domain.Roles{"user", "admin"},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, rr *httptest.ResponseRecorder, svc *mockIdentityService) {
				if svc.getUserByIDCalls != 1 {
					t.Fatalf("expected GetUserById to be called once, got %d calls", svc.getUserByIDCalls)
				}

				response := decodeJSONResponse(t, rr)
				userDetails, ok := response["user_details"].(map[string]any)
				if !ok {
					t.Fatalf("expected user_details payload object, got %T (%v)", response["user_details"], response["user_details"])
				}

				assertFloat64EqualsInt64(t, userDetails["id"], 12, "user_details.id")
				if got := mustString(t, userDetails["username"], "user_details.username"); got != "alice" {
					t.Fatalf("expected username alice, got %q", got)
				}

				rawRoles, ok := userDetails["roles"].([]any)
				if !ok {
					t.Fatalf("expected user_details.roles to be array, got %T (%v)", userDetails["roles"], userDetails["roles"])
				}
				if len(rawRoles) != 2 || mustString(t, rawRoles[0], "roles[0]") != "user" || mustString(t, rawRoles[1], "roles[1]") != "admin" {
					t.Fatalf("unexpected roles value: %v", rawRoles)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{}
			if tt.setup != nil {
				tt.setup(svc)
			}
			handler := newTestHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.pathUserID, nil)
			req.SetPathValue("userId", tt.pathUserID)
			if tt.withAuth {
				req = withAuthContext(t, req, tt.authUserID, []string{"user"})
			}

			rr := httptest.NewRecorder()

			handler.myInfoHandler(rr, req)

			assertStatus(t, rr, tt.wantStatus)
			tt.assert(t, rr, svc)
		})
	}
}

func TestHandler_JWKS(t *testing.T) {
	t.Run("service error returns internal server error", func(t *testing.T) {
		svc := &mockIdentityService{
			getJWKSFn: func(r *http.Request) (json.RawMessage, error) {
				return nil, errors.New("jwks unavailable")
			},
		}
		handler := newTestHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/jwks.json", nil)
		rr := httptest.NewRecorder()

		handler.jwksHandler(rr, req)

		assertStatus(t, rr, http.StatusInternalServerError)
		assertExactError(t, rr, "the server encountered an error and could not process your request")
		if svc.getJWKSCalls != 1 {
			t.Fatalf("expected GetJWKS to be called once, got %d calls", svc.getJWKSCalls)
		}
	})

	t.Run("success returns jwks payload", func(t *testing.T) {
		svc := &mockIdentityService{
			getJWKSFn: func(r *http.Request) (json.RawMessage, error) {
				return json.RawMessage(`{"keys":[{"kid":"my-key"}]}`), nil
			},
		}
		handler := newTestHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/jwks.json", nil)
		rr := httptest.NewRecorder()

		handler.jwksHandler(rr, req)

		assertStatus(t, rr, http.StatusOK)
		if svc.getJWKSCalls != 1 {
			t.Fatalf("expected GetJWKS to be called once, got %d calls", svc.getJWKSCalls)
		}

		response := decodeJSONResponse(t, rr)
		jwks, ok := response["jwks"].(map[string]any)
		if !ok {
			t.Fatalf("expected jwks payload object, got %T (%v)", response["jwks"], response["jwks"])
		}
		keys, ok := jwks["keys"].([]any)
		if !ok || len(keys) != 1 {
			t.Fatalf("expected jwks.keys to contain one key, got %v", jwks["keys"])
		}
	})
}
