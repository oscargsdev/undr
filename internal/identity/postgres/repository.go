package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/oscargsdev/undr/internal/identity/service"
)

type dbExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type repository struct {
	db        dbExecutor
	rootDB    *sql.DB
	dbTimeout time.Duration
}

func NewRepository(db *sql.DB, dbTimeout time.Duration) *repository {
	return &repository{
		db:        db,
		rootDB:    db,
		dbTimeout: dbTimeout,
	}
}

func (r *repository) WithinTx(ctx context.Context, fn func(service.RepositorySet) error) error {
	tx, err := r.rootDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txRepo := &repository{
		db:        tx,
		rootDB:    r.rootDB,
		dbTimeout: r.dbTimeout,
	}

	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
