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

func (h *Handler) TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, h.Service.InterfaceTest())
}
