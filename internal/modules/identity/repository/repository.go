package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/oscargsdev/undr/internal/modules/identity/domain"
)

type UserRepository interface {
	InsertUser(*domain.User) error
	UpdateUser(*domain.User) error
}

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	logger.Info("Entering New Repository Identity")
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		errInspect := err.Error()
		fmt.Printf("%v", errInspect)
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
