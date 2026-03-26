package delivery

import (
	"fmt"
	"net/http"

	"github.com/oscargsdev/undr/internal/modules/identity/service"
)

type Handler struct {
	Service service.IdentityService
}

func NewHandler(svc service.IdentityService) *Handler {
	return &Handler{
		Service: svc,
	}
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := h.Service.RegisterUser()
	if err != nil {

	}
	fmt.Fprintf(w, "Registering the user with ID %d", id)
}
