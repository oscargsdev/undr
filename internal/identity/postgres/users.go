package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/store"
)

func (r *repository) InsertUser(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (username, email, password_hash, activated) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []any{user.Username, user.Email, user.Password.Hash, user.Activated}

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}

	return nil
}

func (r *repository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users 
        SET username = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version`

	args := []any{
		user.Username,
		user.Email,
		user.Password.Hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return store.ErrEditConflict
		default:
			return mapUniqueViolation(err)
		}
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, created_at, username, email, password_hash, activated, version
        FROM users
        WHERE email = $1`

	var user domain.User

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, email).Scan(
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

func (r *repository) GetUserById(ctx context.Context, userId int64) (*domain.User, error) {
	query := `
        SELECT id, created_at, username, email, password_hash, activated, version
        FROM users
        WHERE id = $1`

	var user domain.User

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, userId).Scan(
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
