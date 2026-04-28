package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/service"
	"github.com/oscargsdev/undr/internal/identity/store"
)

const postgresIntegrationDSNEnv = "UNDR_TEST_DB_DSN"

func newPostgresIntegrationRepo(t *testing.T) (*sql.DB, *repository) {
	t.Helper()

	dsn := os.Getenv(postgresIntegrationDSNEnv)
	if dsn == "" {
		t.Skipf("%s is not set; skipping Postgres integration test", postgresIntegrationDSNEnv)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres integration database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close postgres integration database: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres integration database: %v", err)
	}

	return db, NewRepository(db, 3*time.Second)
}

func uniqueIntegrationUser(t *testing.T) *domain.User {
	t.Helper()

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	password := domain.Password{}
	if err := password.Set("correct horse battery staple"); err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	return &domain.User{
		Username:  "tx_user_" + suffix,
		Email:     "tx_user_" + suffix + "@example.com",
		Password:  password,
		Activated: false,
	}
}

func TestIntegrationRepositoryWithinTxRollsBackInsertedUser(t *testing.T) {
	_, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := uniqueIntegrationUser(t)
	callbackErr := errors.New("force rollback")

	err := repo.WithinTx(ctx, func(repos service.RepositorySet) error {
		if err := repos.InsertUser(ctx, user); err != nil {
			return err
		}

		return callbackErr
	})
	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected callback error, got %v", err)
	}

	_, err = repo.GetUserByEmail(ctx, user.Email)
	if !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("expected inserted user to be rolled back, got %v", err)
	}
}

func TestIntegrationRepositoryWithinTxCommitsInsertedUser(t *testing.T) {
	db, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := uniqueIntegrationUser(t)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE email = $1`, user.Email)
	})

	err := repo.WithinTx(ctx, func(repos service.RepositorySet) error {
		return repos.InsertUser(ctx, user)
	})
	if err != nil {
		t.Fatalf("expected transaction commit to succeed, got %v", err)
	}

	persisted, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("expected committed user to be readable, got %v", err)
	}
	if persisted.ID == 0 {
		t.Fatal("expected committed user to have an id")
	}
}
