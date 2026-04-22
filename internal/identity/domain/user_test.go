package domain

import (
	"strings"
	"testing"

	"github.com/oscargsdev/undr/internal/validator"
)

func TestUserIsAnonymous(t *testing.T) {
	t.Run("anonymous user singleton is anonymous", func(t *testing.T) {
		if !AnonymousUser.IsAnonymous() {
			t.Fatal("expected AnonymousUser to be anonymous")
		}
	})

	t.Run("different user pointer is not anonymous", func(t *testing.T) {
		user := &User{}
		if user.IsAnonymous() {
			t.Fatal("expected non-singleton user pointer to be non-anonymous")
		}
	})
}

func TestPasswordSetAndMatches(t *testing.T) {
	var password Password

	err := password.Set("correct horse battery staple")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if password.Plaintext == nil {
		t.Fatal("expected plaintext password to be set")
	}

	if got, want := *password.Plaintext, "correct horse battery staple"; got != want {
		t.Fatalf("expected plaintext %q, got %q", want, got)
	}

	if len(password.Hash) == 0 {
		t.Fatal("expected non-empty hash")
	}

	match, err := password.Matches("correct horse battery staple")
	if err != nil {
		t.Fatalf("expected nil error for matching password, got %v", err)
	}
	if !match {
		t.Fatal("expected matching password to return true")
	}

	match, err = password.Matches("wrong-password")
	if err != nil {
		t.Fatalf("expected nil error for mismatched password, got %v", err)
	}
	if match {
		t.Fatal("expected mismatched password to return false")
	}
}

func TestPasswordMatches_InvalidHashReturnsError(t *testing.T) {
	password := Password{Hash: []byte("not-a-bcrypt-hash")}

	match, err := password.Matches("any-password")
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
	if match {
		t.Fatal("expected false match for invalid hash")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		wantErrMsg string
	}{
		{name: "empty email", email: "", wantErrMsg: "must be provided"},
		{name: "invalid format", email: "not-an-email", wantErrMsg: "must be a valid email address"},
		{name: "valid email", email: "user@example.com", wantErrMsg: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateEmail(v, tt.email)

			gotErrMsg, hasErr := v.Errors["email"]
			if tt.wantErrMsg == "" {
				if hasErr {
					t.Fatalf("expected no email validation error, got %q", gotErrMsg)
				}
				return
			}

			if !hasErr {
				t.Fatalf("expected email validation error %q, got none", tt.wantErrMsg)
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Fatalf("expected email validation error %q, got %q", tt.wantErrMsg, gotErrMsg)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		wantErrMsg string
	}{
		{name: "empty password", password: "", wantErrMsg: "must be provided"},
		{name: "too short", password: "1234567", wantErrMsg: "must be at least 8 bytes long"},
		{name: "too long", password: strings.Repeat("a", 73), wantErrMsg: "must not be more than 72 bytes long"},
		{name: "valid length lower boundary", password: "12345678", wantErrMsg: ""},
		{name: "valid length upper boundary", password: strings.Repeat("a", 72), wantErrMsg: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidatePassword(v, tt.password)

			gotErrMsg, hasErr := v.Errors["password"]
			if tt.wantErrMsg == "" {
				if hasErr {
					t.Fatalf("expected no password validation error, got %q", gotErrMsg)
				}
				return
			}

			if !hasErr {
				t.Fatalf("expected password validation error %q, got none", tt.wantErrMsg)
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Fatalf("expected password validation error %q, got %q", tt.wantErrMsg, gotErrMsg)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	validPassword := Password{}
	if err := validPassword.Set("super-secure-password"); err != nil {
		t.Fatalf("setup failed: could not set valid password: %v", err)
	}

	tests := []struct {
		name       string
		user       *User
		field      string
		wantErrMsg string
	}{
		{
			name: "valid user",
			user: &User{
				Username: "alice",
				Email:    "alice@example.com",
				Password: validPassword,
			},
			field:      "",
			wantErrMsg: "",
		},
		{
			name: "allows nil plaintext when hash exists",
			user: &User{
				Username: "alice",
				Email:    "alice@example.com",
				Password: Password{Hash: []byte("hash-only")},
			},
			field:      "",
			wantErrMsg: "",
		},
		{
			name: "username must be provided",
			user: &User{
				Username: "",
				Email:    "alice@example.com",
				Password: Password{Hash: []byte("hash-only")},
			},
			field:      "username",
			wantErrMsg: "must be provided",
		},
		{
			name: "username max length",
			user: &User{
				Username: strings.Repeat("a", 501),
				Email:    "alice@example.com",
				Password: Password{Hash: []byte("hash-only")},
			},
			field:      "username",
			wantErrMsg: "must not be more than 500 bytes long",
		},
		{
			name: "email must be valid",
			user: &User{
				Username: "alice",
				Email:    "invalid-email",
				Password: Password{Hash: []byte("hash-only")},
			},
			field:      "email",
			wantErrMsg: "must be a valid email address",
		},
		{
			name: "missing password hash",
			user: &User{
				Username: "alice",
				Email:    "alice@example.com",
				Password: Password{},
			},
			field:      "password",
			wantErrMsg: "missing password hash for user",
		},
		{
			name: "plaintext password validated when present",
			user: &User{
				Username: "alice",
				Email:    "alice@example.com",
				Password: Password{Plaintext: strPtr("short"), Hash: []byte("hash")},
			},
			field:      "password",
			wantErrMsg: "must be at least 8 bytes long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateUser(v, tt.user)

			if tt.wantErrMsg == "" {
				if !v.Valid() {
					t.Fatalf("expected user to be valid, got errors: %v", v.Errors)
				}
				return
			}

			gotErrMsg, hasErr := v.Errors[tt.field]
			if !hasErr {
				t.Fatalf("expected %s validation error %q, got none (all errors: %v)", tt.field, tt.wantErrMsg, v.Errors)
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Fatalf("expected %s validation error %q, got %q", tt.field, tt.wantErrMsg, gotErrMsg)
			}
		})
	}
}

func strPtr(s string) *string {
	return new(s)
}
