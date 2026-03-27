package store

import (
	"database/sql"
	"log/slog"

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

func (r *Repository) InsertUser(user *domain.User) error {
	return nil
}

func (r *Repository) UpdateUser(user *domain.User) error {
	return nil
}
