package store

import (
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
)

type UserRepository interface {
	InsertUser(*domain.User) error
	UpdateUser(*domain.User) error
}

type Repository struct {
	// db *sql.DB
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) InsertUser(user *domain.User) error {
	user.Username = "New User"
	user.Email = "user@mail.com"
	user.Password = "pa55word"
	return nil
}

func (r *Repository) UpdateUser(user *domain.User) error {
	return nil
}
