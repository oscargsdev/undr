package service

import (
	"errors"
	"fmt"
	"log"
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

func (s *identityService) ValidateJWTToken(tokenString string) (*jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return []byte("AllYourBase"), nil
	})

	switch {
	case token.Valid:
		fmt.Println("You look nice today")
	case errors.Is(err, jwt.ErrTokenMalformed):
		fmt.Println("That's not even a token")
		return nil, err
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		fmt.Println("Invalid signature")
		return nil, err
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		// Token is either expired or not active yet
		fmt.Println("Timing is everything")
		return nil, err
	default:
		fmt.Println("Couldn't handle this token:", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*claims); ok {
		fmt.Println(claims.Permission, claims.Issuer)
		return &token.Claims, nil
	} else {
		log.Fatal("unknown claims type, cannot proceed")
		return nil, err
	}
}
