package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"

	"github.com/MicahParks/jwkset"
)

func (s *identityService) initJWKS() (*rsa.PrivateKey, jwkset.Storage, error) {
	ctx := context.Background()
	jwkStore := jwkset.NewMemoryStorage()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		s.cfg.Logger.Error("failed to generate RSA key")
		return nil, nil, err
	}

	metadata := jwkset.JWKMetadataOptions{
		KID: "my-key-id",
	}
	options := jwkset.JWKOptions{
		Metadata: metadata,
	}

	jwk, err := jwkset.NewJWKFromKey(privateKey, options)
	if err != nil {
		s.cfg.Logger.Error("failed to create JWK from key")
		return nil, nil, err
	}

	err = jwkStore.KeyWrite(ctx, jwk)
	if err != nil {
		s.cfg.Logger.Error("failed to store RSA key")
		return nil, nil, err
	}

	return privateKey, jwkStore, nil
}
