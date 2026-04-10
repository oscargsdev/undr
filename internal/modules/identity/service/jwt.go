package service

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
)

// TODO:  Get the signing key from env
var mySigningKey = []byte("AllYourBase")

var ErrUnknownClaims = errors.New("unknown claims")

type claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type contextKey string

const userIdContextKey = contextKey("userId")
const rolesContextKey = contextKey("roles")

func newAccessToken(userID int64, roles domain.Roles) (string, error) {
	claims := claims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "undr-auth",
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	// TODO: Check algo to sign with asymetric keys
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	// TODO: What dows this validate exactly?
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return []byte("AllYourBase"), nil
	})

	if !token.Valid {
		return nil, err
	}

	//TODO: What else needs to be validated? Create custom errors and handle them in handler

	//TODO: Validate issuer? Issuer name in env var?

	// TODO: What to do here? Return claims to put in request context?
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

func ContextGetUserId(r *http.Request) int64 {
	userId, ok := r.Context().Value(userIdContextKey).(string)
	if !ok {
		panic("missing user id value in request")
	}

	id, _ := strconv.ParseInt(userId, 10, 64)
	return id
}

func ContextGetRoles(r *http.Request) []string {
	roles, ok := r.Context().Value(rolesContextKey).([]string)
	if !ok {
		panic("missing roles value in request")
	}

	return roles
}
