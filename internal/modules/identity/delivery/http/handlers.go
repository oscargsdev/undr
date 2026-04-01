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
	service        service.IdentityService
	logger         *slog.Logger
	errorResponses *errorResponses.ErrorResponseHelper
}

func NewHandler(svc service.IdentityService, logger *slog.Logger) *Handler {
	return &Handler{
		service:        svc,
		logger:         logger,
		errorResponses: errorResponses.New(logger),
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
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if domain.ValidateUser(v, user); !v.Valid() {
		h.errorResponses.FailedValidationResponse(w, r, v.Errors)
		return
	}

	activationToken, err := h.service.RegisterUser(user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			h.errorResponses.FailedValidationResponse(w, r, v.Errors)
		case errors.Is(err, repository.ErrDuplicateUsername):
			v.AddError("username", "a user with this username already exists")
			h.errorResponses.FailedValidationResponse(w, r, v.Errors)
		default:
			h.errorResponses.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = jsonUtils.WriteJSON(w, http.StatusAccepted, jsonUtils.Envelope{"user": user, "activation_token": activationToken.Plaintext}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"activationToken"`
	}

	err := jsonUtils.ReadJSON(w, r, &input)
	if err != nil {
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if domain.ValidateTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		h.errorResponses.FailedValidationResponse(w, r, v.Errors)
		return
	}

	refreshToken, accessToken, err := h.service.ActivateUser(input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			h.errorResponses.FailedValidationResponse(w, r, v.Errors)
		case errors.Is(err, repository.ErrEditConflict):
			h.errorResponses.EditConflictResponse(w, r)
		default:
			h.errorResponses.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = jsonUtils.WriteJSON(w, http.StatusOK, jsonUtils.Envelope{"refreshToken": refreshToken, "accessToken": accessToken}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) TestTokenValidationHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenString string `json:"tokenString"`
	}

	err := jsonUtils.ReadJSON(w, r, &input)
	if err != nil {
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	claims, err := h.service.ValidateJWTToken(input.TokenString)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
		return
	}

	err = jsonUtils.WriteJSON(w, http.StatusOK, jsonUtils.Envelope{"claims": claims}, nil)
}
