package delivery

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/service"
)

type Handler struct {
	Service service.IdentityService
	logger  *slog.Logger
}

func NewHandler(svc service.IdentityService, logger *slog.Logger) *Handler {
	logger.Info("Entering NewHandler Identity")

	return &Handler{
		Service: svc,
		logger:  logger,
	}
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Receiving request to register new User")
	// Create input struct
	// Read JSON from request and load to input struct
	// Load input data to User struct
	user := &domain.User{}
	// Validate User struct
	// Call service to register User, passing the User struct
	// Response ->  created + user struct (json)
	err := h.Service.RegisterUser(user)
	if err != nil {

	}
	fmt.Fprintf(w, "New User: %+v", user)
}
