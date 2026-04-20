package identity

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/oscargsdev/undr/internal/identity/api"
	"github.com/oscargsdev/undr/internal/identity/postgres"
	"github.com/oscargsdev/undr/internal/identity/service"
)

type Module struct {
	Router http.Handler
}

type Config struct {
	DB                   *sql.DB
	Logger               *slog.Logger
	Issuer               string
	JWTExpiration        time.Duration
	RefreshExpiration    time.Duration
	ActivationExpiration time.Duration
	DBTimeout            time.Duration
}

func NewConfig(db *sql.DB, logger *slog.Logger, flagConfig FlagConfig) Config {
	return Config{
		DB:                   db,
		Logger:               logger,
		Issuer:               flagConfig.Issuer,
		JWTExpiration:        time.Duration(flagConfig.JWTExpiration) * time.Minute,
		RefreshExpiration:    time.Duration(flagConfig.RefreshExpiration) * time.Hour,
		ActivationExpiration: time.Duration(flagConfig.ActivationExpiration) * 24 * time.Hour,
		DBTimeout:            time.Duration(flagConfig.DBTimeout) * time.Second,
	}
}

type FlagConfig struct {
	Issuer               string
	JWTExpiration        int // in minutes
	RefreshExpiration    int // in hours
	ActivationExpiration int // in days
	DBTimeout            int // in seconds
}

func New(cfg Config) (*Module, error) {
	module := &Module{}

	repo := postgres.NewRepository(cfg.DB, cfg.DBTimeout, cfg.Logger)

	svcConfig := service.Config{
		Repository:           repo,
		Logger:               cfg.Logger,
		Issuer:               cfg.Issuer,
		JWTExpiration:        cfg.JWTExpiration,
		RefreshExpiration:    cfg.RefreshExpiration,
		ActivationExpiration: cfg.ActivationExpiration,
	}
	svc, err := service.New(svcConfig)
	if err != nil {
		return nil, err
	}

	handler := api.NewHandler(svc, cfg.Logger)
	router := api.NewRouter(*handler)

	module.Router = router
	return module, nil
}
