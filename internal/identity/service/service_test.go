package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/postgres"
)

type opaqueTokensRepositoryStub struct {
	deleteErr error

	deleteCalls  int
	deletedScope string
	deletedUser  int64
}

func (s *opaqueTokensRepositoryStub) NewOpaqueToken(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
	return nil, nil
}

func (s *opaqueTokensRepositoryStub) DeleteAllFromUser(scope string, userID int64) error {
	s.deleteCalls++
	s.deletedScope = scope
	s.deletedUser = userID
	return s.deleteErr
}

func TestNewInitializedServiceWithJWKSAndPrivateKey(t *testing.T) {
	svc, err := New(Config{Issuer: "https://issuer.example"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	if svc.privateKey == nil {
		t.Fatal("expected privateKey to be initialized")
	}

	if svc.jwkStore == nil {
		t.Fatal("expected jwkStore to be initialized")
	}

	if got, want := svc.cfg.Issuer, "https://issuer.example"; got != want {
		t.Fatalf("expected issuer %q, got %q", want, got)
	}
}

func TestGetIssuerReturnedConfiguredValue(t *testing.T) {
	svc := &identityService{cfg: Config{Issuer: "https://issuer.example"}}

	if got, want := svc.GetIssuer(), "https://issuer.example"; got != want {
		t.Fatalf("expected issuer %q, got %q", want, got)
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
		{name: "unknown error passthrough", in: unknownErr, want: unknownErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapRepositoryError(tt.in)
			if !errors.Is(got, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, got)
			}

			if tt.want == unknownErr && got != unknownErr {
				t.Fatalf("expected unknown error passthrough to preserve error identity")
			}
		})
	}
}

func TestGetJWKSReturnedPublicJWKSJSON(t *testing.T) {
	svc, err := New(Config{Issuer: "https://issuer.example"})
	if err != nil {
		t.Fatalf("setup failed: expected nil error, got %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/jwks.json", nil)
	response, err := svc.GetJWKS(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(response) == 0 {
		t.Fatal("expected non-empty JWKS JSON")
	}

	var payload map[string]any
	if err := json.Unmarshal(response, &payload); err != nil {
		t.Fatalf("expected valid JWKS JSON payload, got error: %v", err)
	}

	if _, ok := payload["keys"]; !ok {
		t.Fatalf("expected JWKS payload to include \"keys\", got %v", payload)
	}
}

func TestLogoutDeletedRefreshTokensAndMappedRepositoryErrors(t *testing.T) {
	unknownErr := errors.New("db down")

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "success", repoErr: nil, wantErr: nil},
		{name: "record not found is mapped", repoErr: postgres.ErrRecordNotFound, wantErr: ErrRecordNotFound},
		{name: "unknown error passthrough", repoErr: unknownErr, wantErr: unknownErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &opaqueTokensRepositoryStub{deleteErr: tt.repoErr}
			svc := &identityService{cfg: Config{OpaqueTokensRepository: repo}}

			err := svc.Logout(42)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if tt.name == "unknown error passthrough" {
				if err != tt.wantErr {
					t.Fatalf("expected passthrough error identity to be preserved")
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if repo.deleteCalls != 1 {
				t.Fatalf("expected DeleteAllFromUser to be called once, got %d", repo.deleteCalls)
			}

			if repo.deletedScope != domain.ScopeRefresh {
				t.Fatalf("expected scope %q, got %q", domain.ScopeRefresh, repo.deletedScope)
			}

			if repo.deletedUser != 42 {
				t.Fatalf("expected user id 42, got %d", repo.deletedUser)
			}
		})
	}
}
