package delivery

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/oscargsdev/undr/internal/common/validator"
	"github.com/oscargsdev/undr/internal/modules/identity/domain"
	"github.com/oscargsdev/undr/internal/modules/identity/repository"
	"github.com/oscargsdev/undr/internal/modules/identity/service"

	errorResponses "github.com/oscargsdev/undr/internal/common/errors"
	jsonUtils "github.com/oscargsdev/undr/internal/common/json"
)

type Handler struct {
	Service service.IdentityService
	logger  *slog.Logger
}

func NewHandler(svc service.IdentityService, logger *slog.Logger) *Handler {
	return &Handler{
		Service: svc,
		logger:  logger,
	}
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
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

	activationToken, err := h.Service.RegisterUser(user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			errorResponses.FailedValidationResponse(w, r, v.Errors, h.logger)
		default:
			errorResponses.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	err = jsonUtils.WriteJSON(w, http.StatusAccepted, jsonUtils.Envelope{"user": user, "activation_token": activationToken.Plaintext}, nil)
	if err != nil {
		errorResponses.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"activationToken"`
	}

	err := jsonUtils.ReadJSON(w, r, &input)
	if err != nil {
		errorResponses.BadRequestResponse(w, r, err, h.logger)
		return
	}

	v := validator.New()

	if domain.ValidateTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		errorResponses.FailedValidationResponse(w, r, v.Errors, h.logger)
		return
	}

	user, err := h.Service.ActivateUser(input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			errorResponses.FailedValidationResponse(w, r, v.Errors, h.logger)
		case errors.Is(err, repository.ErrEditConflict):
			errorResponses.EditConflictResponse(w, r, h.logger)
		default:
			errorResponses.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	err = jsonUtils.WriteJSON(w, http.StatusOK, jsonUtils.Envelope{"user": user}, nil)
	if err != nil {
		errorResponses.ServerErrorResponse(w, r, err, h.logger)
	}
}
