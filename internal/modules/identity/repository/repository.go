package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/oscargsdev/undr/internal/common/validator"
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
)

type IdentityRepository interface {
	InsertUser(*domain.User) error
	UpdateUser(*domain.User) error
	NewToken(userID int64, ttl time.Duration, scope string) (*domain.Token, error)
}

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

func (r *Repository) InsertUser(user *domain.User) error {
	query := `
        INSERT INTO users (name, email, password_hash, activated) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []any{user.Username, user.Email, user.Password.Hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key" (23505)`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (r *Repository) UpdateUser(user *domain.User) error {
	return nil
}

func generateToken(userID int64, ttl time.Duration, scope string) *domain.Token {
	token := &domain.Token{
		Plaintext: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func (r *Repository) InsertToken(token *domain.Token) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) NewToken(userID int64, ttl time.Duration, scope string) (*domain.Token, error) {
	token := generateToken(userID, ttl, scope)

	err := r.InsertToken(token)
	return token, err
}
