package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
)

type IdentityRepository interface {
	InsertUser(*domain.User) error
	UpdateUser(*domain.User) error
	NewOpaqueToken(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error)
	GetForOpaqueToken(tokenScope, tokenPlaintext string) (*domain.User, error)
	DeleteAllFromUser(scope string, userID int64) error
	GetUserByEmail(email string) (*domain.User, error)
	GetAllRolesForUser(userID int64) (domain.Roles, error)
	AddRoleForUser(userID int64, codes ...string) error
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
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrRecordNotFound    = errors.New("record not found")
	ErrEditConflict      = errors.New("edit conflict")
)

func (r *Repository) InsertUser(user *domain.User) error {
	query := `
        INSERT INTO users (username, email, password_hash, activated) 
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
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key" (23505)`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (r *Repository) UpdateUser(user *domain.User) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (r *Repository) GetUserByEmail(email string) (*domain.User, error) {
	query := `
        SELECT id, created_at, username, email, password_hash, activated, version
        FROM users
        WHERE email = $1`

	var user domain.User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func generateOpaqueToken(userID int64, ttl time.Duration, scope string) *domain.OpaqueToken {
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

func (r *Repository) InsertOpaqueToken(token *domain.OpaqueToken) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) NewOpaqueToken(userID int64, ttl time.Duration, scope string) (*domain.OpaqueToken, error) {
	token := generateOpaqueToken(userID, ttl, scope)

	err := r.InsertOpaqueToken(token)
	return token, err
}

func (r *Repository) GetForOpaqueToken(tokenScope, tokenPlaintext string) (*domain.User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT users.id, users.created_at, users.username, users.email, users.password_hash, users.activated, users.version
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user domain.User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (r *Repository) DeleteAllFromUser(scope string, userID int64) error {
	query := `
        DELETE FROM tokens 
        WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, scope, userID)
	return err
}

func (r *Repository) GetAllRolesForUser(userID int64) (domain.Roles, error) {
	query := `
		SELECT roles.code
		FROM roles
		INNER JOIN users_roles ON users_roles.role_id = roles.id
		INNER JOIN users ON users_roles.user_id = users.id
		WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles domain.Roles

	for rows.Next() {
		var role string

		err := rows.Scan(&role)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *Repository) AddRoleForUser(userID int64, codes ...string) error {
	query := `
        INSERT INTO users_roles
        SELECT $1, roles.id FROM roles WHERE roles.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
