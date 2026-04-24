package postgres

import (
	"database/sql"
	"time"
)

type repository struct {
	db        *sql.DB
	dbTimeout time.Duration
}

func NewRepository(db *sql.DB, dbTimeout time.Duration) *repository {
	return &repository{
		db:        db,
		dbTimeout: dbTimeout,
	}
}
