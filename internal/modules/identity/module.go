package identity

import (
	"net/http"

	"github.com/oscargsdev/undr/internal/modules/identity/delivery"
)

type Module struct {
	Router http.Handler
}

func New() *Module {
	module := &Module{}

	module.Router = delivery.NewRouter()

	return module
}
