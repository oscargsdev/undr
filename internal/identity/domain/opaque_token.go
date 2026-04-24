package domain

import (
	"errors"
	"time"

	"github.com/oscargsdev/undr/internal/validator"
)

type TokenScope string

const (
	ScopeActivation TokenScope = "activation"
	ScopeRefresh    TokenScope = "refresh"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type OpaqueToken struct {
	Plaintext string     `json:"token"`
	Hash      []byte     `json:"-"`
	UserID    int64      `json:"-"`
	Expiry    time.Time  `json:"expiry"`
	Scope     TokenScope `json:"-"`
}

func ValidateOpaqueTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}
