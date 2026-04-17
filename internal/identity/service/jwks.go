package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"

	"github.com/MicahParks/jwkset"
)

var jwkStore jwkset.Storage
var privateKey *rsa.PrivateKey

func (s *identityService) initJWKS() error {
	ctx := context.Background()
	jwkStore = jwkset.NewMemoryStorage()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		s.cfg.Logger.Error("failed to generate RSA key")
		return err
	}

	privateKey = key

	metadata := jwkset.JWKMetadataOptions{
		KID: "my-key-id",
	}
	options := jwkset.JWKOptions{
		Metadata: metadata,
	}

	jwk, err := jwkset.NewJWKFromKey(key, options)
	if err != nil {
		s.cfg.Logger.Error("failed to create JWK from key")
		return err
	}

	err = jwkStore.KeyWrite(ctx, jwk)
	if err != nil {
		s.cfg.Logger.Error("failed to store RSA key")
		return err
	}

	return nil
}
