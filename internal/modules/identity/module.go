package identity

import (
	"database/sql"
	"log/slog"
	"net/http"

	delivery "github.com/oscargsdev/undr/internal/modules/identity/delivery/http"
	"github.com/oscargsdev/undr/internal/modules/identity/service"
	"github.com/oscargsdev/undr/internal/modules/identity/store"
)

type Module struct {
	Router http.Handler
	logger *slog.Logger
}

func New(db *sql.DB, logger *slog.Logger) *Module {
	logger.Info("Entering New Identity Module")
	module := &Module{}

	repo := store.NewRepository(db, logger)
	svc := service.New(repo, logger)
	handler := delivery.NewHandler(svc, logger)
	router := delivery.NewRouter(*handler)
	module.Router = router

	return module
}
