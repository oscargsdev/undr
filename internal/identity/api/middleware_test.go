package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/identity/service"
)

func TestHandler_AuthorizationMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header", header: ""},
		{name: "missing token part", header: "Bearer"},
		{name: "wrong scheme", header: "Token value"},
		{name: "too many header parts", header: "Bearer value extra"},
		{name: "scheme is case-sensitive", header: "bearer value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{}
			handler := newTestHandler(svc)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})
			mw := handler.AuthorizationMiddleware(next)

			req := httptest.NewRequest(http.MethodGet, "/secured", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			assertStatus(t, rr, http.StatusUnauthorized)
			assertExactError(t, rr, "invalid or missing access token")
			if rr.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("expected WWW-Authenticate header Bearer, got %q", rr.Header().Get("WWW-Authenticate"))
			}
			assertHeaderContains(t, rr.Header().Values("Vary"), "Authorization", "Vary")
			if nextCalled {
				t.Fatal("expected next handler not to be called")
			}
			if svc.validateJWTTokenCalls != 0 {
				t.Fatalf("expected ValidateJWTToken not to be called, got %d calls", svc.validateJWTTokenCalls)
			}
		})
	}
}

func TestHandler_AuthorizationMiddleware_ValidateJWTTokenErrorMapping(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "malformed token returns bad request",
			err:         jwt.ErrTokenMalformed,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "malformed access token",
		},
		{
			name:        "invalid signature returns unauthorized",
			err:         jwt.ErrTokenSignatureInvalid,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "invalid or missing access token",
		},
		{
			name:        "expired token returns unauthorized",
			err:         jwt.ErrTokenExpired,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "invalid or missing access token",
		},
		{
			name:        "token not valid yet returns unauthorized",
			err:         jwt.ErrTokenNotValidYet,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "invalid or missing access token",
		},
		{
			name:        "invalid issuer returns unauthorized",
			err:         jwt.ErrTokenInvalidIssuer,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "invalid or missing access token",
		},
		{
			name:        "invalid key type returns unauthorized",
			err:         jwt.ErrInvalidKeyType,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "invalid or missing access token",
		},
		{
			name:        "unknown claims returns unauthorized",
			err:         service.ErrUnknownClaims,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "invalid or missing access token",
		},
		{
			name:        "unexpected validation error returns server error",
			err:         errors.New("jwt parser crashed"),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "the server encountered an error and could not process your request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockIdentityService{
				getIssuerFn: func() string { return "https://issuer.example" },
				validateJWTTokenFn: func(tokenString string, issuer string) (*jwt.Token, error) {
					if tokenString != "token-value" {
						t.Fatalf("expected token-value, got %q", tokenString)
					}
					if issuer != "https://issuer.example" {
						t.Fatalf("expected issuer https://issuer.example, got %q", issuer)
					}
					return nil, tt.err
				},
			}
			handler := newTestHandler(svc)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})
			mw := handler.AuthorizationMiddleware(next)

			req := httptest.NewRequest(http.MethodGet, "/secured", nil)
			req.Header.Set("Authorization", "Bearer token-value")
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			assertStatus(t, rr, tt.wantStatus)
			assertExactError(t, rr, tt.wantMessage)
			assertHeaderContains(t, rr.Header().Values("Vary"), "Authorization", "Vary")

			if nextCalled {
				t.Fatal("expected next handler not to be called")
			}
			if svc.getIssuerCalls != 1 {
				t.Fatalf("expected GetIssuer to be called once, got %d calls", svc.getIssuerCalls)
			}
			if svc.validateJWTTokenCalls != 1 {
				t.Fatalf("expected ValidateJWTToken to be called once, got %d calls", svc.validateJWTTokenCalls)
			}
		})
	}
}

func TestHandler_AuthorizationMiddleware_SuccessSetsClaimsAndCallsNext(t *testing.T) {
	realService, accessToken := newJWTFixtureServiceAndToken(t, 501, []string{"user", "admin"})
	handler := newTestHandler(realService)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		userID, err := service.ContextGetUserId(r)
		if err != nil {
			t.Fatalf("expected user id in context, got error: %v", err)
		}
		if userID != 501 {
			t.Fatalf("expected user id 501, got %d", userID)
		}

		roles, err := service.ContextGetRoles(r)
		if err != nil {
			t.Fatalf("expected roles in context, got error: %v", err)
		}
		if len(roles) != 2 || roles[0] != "user" || roles[1] != "admin" {
			t.Fatalf("expected roles [user admin], got %v", roles)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	mw := handler.AuthorizationMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/secured", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusNoContent)
	assertHeaderContains(t, rr.Header().Values("Vary"), "Authorization", "Vary")
	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
}

func TestHandler_RequireRoleMiddleware(t *testing.T) {
	t.Run("invalid auth header is rejected by wrapped authorization middleware", func(t *testing.T) {
		handler := newTestHandler(&mockIdentityService{})

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})
		mw := handler.RequireRoleMiddleware("admin", next)

		req := httptest.NewRequest(http.MethodGet, "/admin-portal", nil)
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		assertStatus(t, rr, http.StatusUnauthorized)
		assertExactError(t, rr, "invalid or missing access token")
		if nextCalled {
			t.Fatal("expected next handler not to be called")
		}
	})

	t.Run("authenticated user without required role gets forbidden", func(t *testing.T) {
		realService, accessToken := newJWTFixtureServiceAndToken(t, 901, []string{"user"})
		handler := newTestHandler(realService)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})
		mw := handler.RequireRoleMiddleware("admin", next)

		req := httptest.NewRequest(http.MethodGet, "/admin-portal", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		assertStatus(t, rr, http.StatusForbidden)
		assertExactError(t, rr, "your user account does not have the necessary roles/permissions to access this resource")
		if nextCalled {
			t.Fatal("expected next handler not to be called")
		}
	})

	t.Run("authenticated user with role can access next handler", func(t *testing.T) {
		realService, accessToken := newJWTFixtureServiceAndToken(t, 902, []string{"user", "admin"})
		handler := newTestHandler(realService)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusAccepted)
		})
		mw := handler.RequireRoleMiddleware("admin", next)

		req := httptest.NewRequest(http.MethodGet, "/admin-portal", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		assertStatus(t, rr, http.StatusAccepted)
		if !nextCalled {
			t.Fatal("expected next handler to be called")
		}
	})
}
