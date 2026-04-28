package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/oscargsdev/undr/internal/identity/service"
)

const txTestDriverName = "undr-postgres-tx-test"

var txTestStores sync.Map

func init() {
	sql.Register(txTestDriverName, txTestDriver{})
}

type txTestStats struct {
	mu        sync.Mutex
	begins    int
	commits   int
	rollbacks int
}

func (s *txTestStats) recordBegin() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.begins++
}

func (s *txTestStats) recordCommit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commits++
}

func (s *txTestStats) recordRollback() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollbacks++
}

func (s *txTestStats) snapshot() (begins int, commits int, rollbacks int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.begins, s.commits, s.rollbacks
}

type txTestDriver struct{}

func (txTestDriver) Open(name string) (driver.Conn, error) {
	stats, ok := txTestStores.Load(name)
	if !ok {
		return nil, errors.New("unknown tx test store")
	}
	return &txTestConn{stats: stats.(*txTestStats)}, nil
}

type txTestConn struct {
	stats *txTestStats
}

func (c *txTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not implemented")
}

func (c *txTestConn) Close() error {
	return nil
}

func (c *txTestConn) Begin() (driver.Tx, error) {
	c.stats.recordBegin()
	return &txTestTx{stats: c.stats}, nil
}

func (c *txTestConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}

func (c *txTestConn) CheckNamedValue(*driver.NamedValue) error {
	return nil
}

func (c *txTestConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (c *txTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return txTestRows{}, nil
}

type txTestTx struct {
	stats *txTestStats
}

func (tx *txTestTx) Commit() error {
	tx.stats.recordCommit()
	return nil
}

func (tx *txTestTx) Rollback() error {
	tx.stats.recordRollback()
	return nil
}

type txTestRows struct{}

func (txTestRows) Columns() []string {
	return nil
}

func (txTestRows) Close() error {
	return nil
}

func (txTestRows) Next([]driver.Value) error {
	return io.EOF
}

func newTxTestDB(t *testing.T) (*sql.DB, *txTestStats) {
	t.Helper()

	name := t.Name()
	stats := &txTestStats{}
	txTestStores.Store(name, stats)
	t.Cleanup(func() {
		txTestStores.Delete(name)
	})

	db, err := sql.Open(txTestDriverName, name)
	if err != nil {
		t.Fatalf("open tx test db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close tx test db: %v", err)
		}
	})

	return db, stats
}

func TestRepositoryWithinTxCommitsOnSuccess(t *testing.T) {
	db, stats := newTxTestDB(t)
	repo := NewRepository(db, time.Second)

	err := repo.WithinTx(context.Background(), func(repos service.RepositorySet) error {
		txRepo, ok := repos.(*repository)
		if !ok {
			t.Fatalf("expected transaction callback repository to be *repository, got %T", repos)
		}
		if txRepo == repo {
			t.Fatal("expected transaction callback to receive a distinct transaction repository")
		}
		if txRepo.rootDB != repo.rootDB {
			t.Fatal("expected transaction repository to keep root DB for nested transaction creation")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	begins, commits, rollbacks := stats.snapshot()
	if begins != 1 || commits != 1 || rollbacks != 0 {
		t.Fatalf("expected begin=1 commit=1 rollback=0, got begin=%d commit=%d rollback=%d", begins, commits, rollbacks)
	}
}

func TestRepositoryWithinTxRollsBackOnCallbackError(t *testing.T) {
	db, stats := newTxTestDB(t)
	repo := NewRepository(db, time.Second)
	callbackErr := errors.New("callback failed")

	err := repo.WithinTx(context.Background(), func(repos service.RepositorySet) error {
		return callbackErr
	})
	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected callback error, got %v", err)
	}

	begins, commits, rollbacks := stats.snapshot()
	if begins != 1 || commits != 0 || rollbacks != 1 {
		t.Fatalf("expected begin=1 commit=0 rollback=1, got begin=%d commit=%d rollback=%d", begins, commits, rollbacks)
	}
}
