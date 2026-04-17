package service

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/identity/domain"
)

var (
	ErrUnknownClaims             = errors.New("unknown claims")
	ErrMissingUserIDInContext    = errors.New("missing user id in context")
	ErrMissingUserRolesInContext = errors.New("missing user roles in context")
)

type claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type contextKey string

const userIdContextKey = contextKey("userId")
const rolesContextKey = contextKey("roles")

func newAccessToken(userID int64, roles domain.Roles, expiration time.Duration, issuer string) (string, error) {
	claims := claims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			Issuer:    issuer,
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWTToken(tokenString string, issuer string) (*jwt.Token, error) {
	fn := func(token *jwt.Token) (any, error) {
		return privateKey.Public(), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims{}, fn, jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

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
