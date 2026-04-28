package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/store"
)

func generateOpaqueToken(userID int64, ttl time.Duration, scope domain.TokenScope) *domain.OpaqueToken {
	token := &domain.OpaqueToken{
		Plaintext: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

func (r *repository) insertOpaqueToken(ctx context.Context, token *domain.OpaqueToken) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, string(token.Scope)}

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) NewOpaqueToken(ctx context.Context, userID int64, ttl time.Duration, scope domain.TokenScope) (*domain.OpaqueToken, error) {
	token := generateOpaqueToken(userID, ttl, scope)

	err := r.insertOpaqueToken(ctx, token)
	return token, err
}

func (r *repository) GetUserForOpaqueToken(ctx context.Context, tokenScope domain.TokenScope, tokenPlaintext string) (*domain.User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT users.id, users.created_at, users.username, users.email, users.password_hash, users.activated, users.version
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []any{tokenHash[:], string(tokenScope), time.Now()}

	var user domain.User

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, store.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (r *repository) DeleteAllFromUser(ctx context.Context, scope domain.TokenScope, userID int64) error {
	query := `
        DELETE FROM tokens 
        WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, string(scope), userID)
	return err
}
