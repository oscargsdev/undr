package identity

import (
	"net/http"

	delivery "github.com/oscargsdev/undr/internal/modules/identity/delivery/http"
	"github.com/oscargsdev/undr/internal/modules/identity/service"
	"github.com/oscargsdev/undr/internal/modules/identity/store"
)

type Module struct {
	Router http.Handler
}

func New() *Module {
	module := &Module{}

	repo := store.NewRepository()
	svc := service.New(repo)
	handler := delivery.NewHandler(svc)
	router := delivery.NewRouter(*handler)
	module.Router = router

	return module
}
