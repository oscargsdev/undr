package identity

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	delivery "github.com/oscargsdev/undr/internal/modules/identity/delivery/http"
	"github.com/oscargsdev/undr/internal/modules/identity/repository"
	"github.com/oscargsdev/undr/internal/modules/identity/service"
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

func New(cfg Config) *Module {
	module := &Module{}

	repo := repository.NewRepository(cfg.DB, cfg.DBTimeout, cfg.Logger)

	svcConfig := service.Config{
		Repository:           repo,
		Logger:               cfg.Logger,
		Issuer:               cfg.Issuer,
		JWTExpiration:        cfg.JWTExpiration,
		RefreshExpiration:    cfg.RefreshExpiration,
		ActivationExpiration: cfg.ActivationExpiration,
	}
	svc := service.New(svcConfig)

	handler := delivery.NewHandler(svc, cfg.Logger)
	router := delivery.NewRouter(*handler)

	module.Router = router
	return module
}
