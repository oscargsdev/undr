package identity

import (
	"net/http"

	"github.com/oscargsdev/undr/internal/modules/identity/delivery"
	"github.com/oscargsdev/undr/internal/modules/identity/service"
)

type Module struct {
	Router http.Handler
}

func New() *Module {
	module := &Module{}

	svc := service.New()

	handler := delivery.NewHandler(svc)

	router := delivery.NewRouter(*handler)

	module.Router = router

	return module
}
