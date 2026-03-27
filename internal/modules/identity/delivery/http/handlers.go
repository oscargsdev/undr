package delivery

import (
	"log/slog"
	"net/http"

	"github.com/oscargsdev/undr/internal/common"
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

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := common.ReadJSON(w, r, &input)
	if err != nil {
		common.BadRequestResponse(w, r, err, h.logger)
		return
	}

	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		common.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	// Validate User struct
	// Call service to register User, passing the User struct
	// Response ->  created + user struct (json)
	err = h.Service.RegisterUser(user)
	if err != nil {

	}

	err = common.WriteJSON(w, http.StatusAccepted, common.Envelope{"user": user}, nil)
	if err != nil {
		common.ServerErrorResponse(w, r, err, h.logger)
	}
}
