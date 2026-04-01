package service

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO:  Get the signing key from env
var mySigningKey = []byte("AllYourBase")

type claims struct {
	Permission string `json:"permission"`
	jwt.RegisteredClaims
}

// TODO: Return string and not pointer?
func newAuthToken(userID int64) (*string, error) {
	claims := claims{
		"readProjects",
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "undr-auth",
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	// TODO: Check algo to sign with asymetric keys
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)

	if err != nil {
		return nil, err
	}

	return &ss, nil
}
