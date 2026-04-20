package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/oscargsdev/undr/internal/identity"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

type application struct {
	config         config
	logger         *slog.Logger
	wg             sync.WaitGroup
	identityModule *identity.Module
}

func main() {
	var cfg config
	var identityFlags identity.FlagConfig

	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("UNDR_DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.StringVar(&identityFlags.Issuer, "identity-issuer", "undr-auth", "identity token issuer name")
	flag.IntVar(&identityFlags.JWTExpiration, "jwt-expiration", 15, "JWT token expiration time in minutes")
	flag.IntVar(&identityFlags.RefreshExpiration, "refresh-token-expiration", 24, "refresh token expiration time in hours")
	flag.IntVar(&identityFlags.ActivationExpiration, "activation-expiration", 3, "activation token expiration time in days")
	flag.IntVar(&identityFlags.DBTimeout, "db-timeout", 3, "db timeout in seconds") // Maybe global flag for other services

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	identityCfg := identity.NewConfig(db, logger, identityFlags)
	identityModule, err := identity.New(identityCfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config:         cfg,
		logger:         logger,
		identityModule: identityModule,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
