package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO:  Get the signing key from env
var mySigningKey = []byte("AllYourBase")

var ErrUnknownClaims = errors.New("unknown claims")

type claims struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func newAccessToken(userID int64) (string, error) {
	claims := claims{
		Roles: []string{
			"projectAdmin",
		},
		Permissions: []string{
			"readProjects",
		},
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

func (s *identityService) ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	// TODO: What dows this validate exactly?
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return []byte("AllYourBase"), nil
	})

	if !token.Valid {
		return nil, err
	}

	//TODO: What else needs to be validated?

	//TODO: Validate issuer? Issuer name in env var?

	// TODO: What to do here? Return claims to put in request context?
	if _, ok := token.Claims.(*claims); ok {
		return token, nil
	} else {
		return nil, ErrUnknownClaims
	}
}
