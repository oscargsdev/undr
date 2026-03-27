package delivery

import (
	"log/slog"
	"net/http"

	"github.com/oscargsdev/undr/internal/common"
	"github.com/oscargsdev/undr/internal/common/validator"
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/service"

	errorResponses "github.com/oscargsdev/undr/internal/common/errors"
	jsonUtils "github.com/oscargsdev/undr/internal/common/json"
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

	err := jsonUtils.ReadJSON(w, r, &input)
	if err != nil {
		errorResponses.BadRequestResponse(w, r, err, h.logger)
		return
	}

	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		errorResponses.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	v := validator.New()

	if domain.ValidateUser(v, user); !v.Valid() {
		errorResponses.FailedValidationResponse(w, r, v.Errors, h.logger)
		return
	}

	// Call service to register User, passing the User struct
	// Response ->  created + user struct (json)
	err = h.Service.RegisterUser(user)
	if err != nil {

	}

	err = jsonUtils.WriteJSON(w, http.StatusAccepted, common.Envelope{"user": user}, nil)
	if err != nil {
		errorResponses.ServerErrorResponse(w, r, err, h.logger)
	}
}
