package postgres

import (
	"errors"

	"github.com/lib/pq"
	"github.com/oscargsdev/undr/internal/identity/store"
)

func mapUniqueViolation(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}

	if pqErr.Code != "23505" {
		return err
	}

	switch pqErr.Constraint {
	case "users_email_key":
		return store.ErrDuplicateEmail
	case "users_username_key":
		return store.ErrDuplicateUsername
	default:
		return err
	}
}
