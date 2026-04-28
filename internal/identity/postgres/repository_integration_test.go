package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/service"
	"github.com/oscargsdev/undr/internal/identity/store"
)

const postgresIntegrationDSNEnv = "UNDR_REPOSITORY_TEST_DB_DSN"

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

	var usersTable sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.users')`).Scan(&usersTable); err != nil {
		t.Fatalf("check postgres integration schema: %v", err)
	}
	if !usersTable.Valid {
		t.Skipf("%s is set, but migrations are not applied; skipping Postgres integration test", postgresIntegrationDSNEnv)
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

func insertIntegrationUser(t *testing.T, db *sql.DB, repo *repository, ctx context.Context) *domain.User {
	t.Helper()

	user := uniqueIntegrationUser(t)
	if err := repo.InsertUser(ctx, user); err != nil {
		t.Fatalf("insert integration user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID)
	})

	return user
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

func TestIntegrationRepositoryInsertUserMapsDuplicateEmailAndUsername(t *testing.T) {
	db, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := insertIntegrationUser(t, db, repo, ctx)

	duplicateEmail := uniqueIntegrationUser(t)
	duplicateEmail.Email = user.Email
	if err := repo.InsertUser(ctx, duplicateEmail); !errors.Is(err, store.ErrDuplicateEmail) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}

	duplicateUsername := uniqueIntegrationUser(t)
	duplicateUsername.Username = user.Username
	if err := repo.InsertUser(ctx, duplicateUsername); !errors.Is(err, store.ErrDuplicateUsername) {
		t.Fatalf("expected duplicate username error, got %v", err)
	}
}

func TestIntegrationRepositoryGetUserByEmailAndID(t *testing.T) {
	db, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := insertIntegrationUser(t, db, repo, ctx)

	byEmail, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if byEmail.ID != user.ID || byEmail.Username != user.Username || byEmail.Email != user.Email {
		t.Fatalf("unexpected user by email: %+v", byEmail)
	}

	byID, err := repo.GetUserById(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if byID.ID != user.ID || byID.Username != user.Username || byID.Email != user.Email {
		t.Fatalf("unexpected user by id: %+v", byID)
	}

	_, err = repo.GetUserByEmail(ctx, "missing_"+user.Email)
	if !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("expected missing email to map to record not found, got %v", err)
	}

	_, err = repo.GetUserById(ctx, user.ID+1_000_000_000)
	if !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("expected missing id to map to record not found, got %v", err)
	}
}

func TestIntegrationRepositoryUpdateUserPersistsChangesAndDetectsConflict(t *testing.T) {
	db, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := insertIntegrationUser(t, db, repo, ctx)

	user.Username += "_updated"
	user.Activated = true
	initialVersion := user.Version
	if err := repo.UpdateUser(ctx, user); err != nil {
		t.Fatalf("update user: %v", err)
	}
	if user.Version != initialVersion+1 {
		t.Fatalf("expected version to increment from %d to %d, got %d", initialVersion, initialVersion+1, user.Version)
	}

	updated, err := repo.GetUserById(ctx, user.ID)
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}
	if updated.Username != user.Username || !updated.Activated {
		t.Fatalf("expected updated user values, got %+v", updated)
	}

	stale := *user
	stale.Version = initialVersion
	if err := repo.UpdateUser(ctx, &stale); !errors.Is(err, store.ErrEditConflict) {
		t.Fatalf("expected stale update to map to edit conflict, got %v", err)
	}
}

func TestIntegrationRepositoryOpaqueTokenLookupScopeExpiryAndDeletion(t *testing.T) {
	db, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := insertIntegrationUser(t, db, repo, ctx)

	token, err := repo.NewOpaqueToken(ctx, user.ID, time.Hour, domain.ScopeRefresh)
	if err != nil {
		t.Fatalf("new opaque token: %v", err)
	}

	tokenUser, err := repo.GetUserForOpaqueToken(ctx, domain.ScopeRefresh, token.Plaintext)
	if err != nil {
		t.Fatalf("get user for refresh token: %v", err)
	}
	if tokenUser.ID != user.ID {
		t.Fatalf("expected token user id %d, got %d", user.ID, tokenUser.ID)
	}

	_, err = repo.GetUserForOpaqueToken(ctx, domain.ScopeActivation, token.Plaintext)
	if !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("expected wrong token scope to map to record not found, got %v", err)
	}

	expiredToken, err := repo.NewOpaqueToken(ctx, user.ID, -time.Hour, domain.ScopeRefresh)
	if err != nil {
		t.Fatalf("new expired token: %v", err)
	}
	_, err = repo.GetUserForOpaqueToken(ctx, domain.ScopeRefresh, expiredToken.Plaintext)
	if !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("expected expired token to map to record not found, got %v", err)
	}

	if err := repo.DeleteAllFromUser(ctx, domain.ScopeRefresh, user.ID); err != nil {
		t.Fatalf("delete refresh tokens: %v", err)
	}
	_, err = repo.GetUserForOpaqueToken(ctx, domain.ScopeRefresh, token.Plaintext)
	if !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("expected deleted token to map to record not found, got %v", err)
	}
}

func TestIntegrationRepositoryRoles(t *testing.T) {
	db, repo := newPostgresIntegrationRepo(t)
	ctx := context.Background()
	user := insertIntegrationUser(t, db, repo, ctx)

	if err := repo.AddRoleForUser(ctx, user.ID, "user", "admin"); err != nil {
		t.Fatalf("add roles for user: %v", err)
	}

	roles, err := repo.GetAllRolesForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("get roles for user: %v", err)
	}
	if !reflect.DeepEqual([]string(roles), []string{"user", "admin"}) && !reflect.DeepEqual([]string(roles), []string{"admin", "user"}) {
		t.Fatalf("expected user and admin roles, got %v", roles)
	}

	noRolesUser := insertIntegrationUser(t, db, repo, ctx)
	roles, err = repo.GetAllRolesForUser(ctx, noRolesUser.ID)
	if err != nil {
		t.Fatalf("get roles for user without roles: %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected no roles, got %v", roles)
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
