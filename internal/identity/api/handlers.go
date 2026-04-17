package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/repository"
	"github.com/oscargsdev/undr/internal/identity/service"
	"github.com/oscargsdev/undr/internal/validator"

	"github.com/oscargsdev/undr/internal/jsonx"
	"github.com/oscargsdev/undr/internal/responses"
)

type Handler struct {
	service        service.IdentityService
	logger         *slog.Logger
	errorResponses *responses.ErrorResponseHelper
}

func newRefreshTokenCookie(refreshToken string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/v1/identity/refresh",
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

func NewHandler(svc service.IdentityService, logger *slog.Logger) *Handler {
	return &Handler{
		service:        svc,
		logger:         logger,
		errorResponses: responses.New(logger),
	}
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonx.ReadJSON(w, r, &input)
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

	err = jsonx.WriteJSON(w, http.StatusAccepted, jsonx.Envelope{"user": user, "activation_token": activationToken.Plaintext}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"activationToken"`
	}

	err := jsonx.ReadJSON(w, r, &input)
	if err != nil {
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if domain.ValidateOpaqueTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
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

	http.SetCookie(w, newRefreshTokenCookie(refreshToken, time.Now().Add(24*time.Hour)))

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"access_token": accessToken}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) testSecuredEndpoint(w http.ResponseWriter, r *http.Request) {
	userId := service.ContextGetUserId(r)
	roles := service.ContextGetRoles(r)
	jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"userId": userId, "roles": roles}, nil)
}

func (h *Handler) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonx.ReadJSON(w, r, &input)
	if err != nil {
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	domain.ValidateEmail(v, input.Email)
	domain.ValidatePassword(v, input.Password)

	if !v.Valid() {
		h.errorResponses.FailedValidationResponse(w, r, v.Errors)
		return
	}

	refreshToken, accessToken, err := h.service.AuthenticateUser(input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			h.errorResponses.InvalidCredentialsResponse(w, r)
		case errors.Is(err, service.ErrInvalidCredentials):
			h.errorResponses.InvalidCredentialsResponse(w, r)
		case errors.Is(err, service.ErrUserNotActivated):
			h.errorResponses.InactiveAccountResponse(w, r)
		default:
			h.errorResponses.ServerErrorResponse(w, r, err)
		}
		return
	}

	http.SetCookie(w, newRefreshTokenCookie(refreshToken, time.Now().Add(24*time.Hour)))

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"access_token": accessToken}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	oldToken := refreshTokenCookie.Value

	v := validator.New()

	if domain.ValidateOpaqueTokenPlaintext(v, oldToken); !v.Valid() {
		h.errorResponses.InvalidRefresTokenResponse(w, r, v.Errors)
		return
	}

	refreshToken, accessToken, err := h.service.RefreshToken(oldToken)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			h.errorResponses.BadRequestResponse(w, r, domain.ErrInvalidRefreshToken)
		default:
			h.errorResponses.ServerErrorResponse(w, r, err)
		}
		return
	}

	http.SetCookie(w, newRefreshTokenCookie(refreshToken, time.Now().Add(24*time.Hour)))

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"access_token": accessToken}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	userId := service.ContextGetUserId(r)

	err := h.service.Logout(userId)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
		return
	}

	http.SetCookie(w, newRefreshTokenCookie("", time.Unix(0, 0)))

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) OnlyAdminsHandler(w http.ResponseWriter, r *http.Request) {
	jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"howdy": "you are an admin!"}, nil)
}

func (h *Handler) MyInfoHandler(w http.ResponseWriter, r *http.Request) {
	pathUserId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		h.errorResponses.BadRequestResponse(w, r, err)
		return
	}

	contextUserId := service.ContextGetUserId(r)

	if pathUserId != contextUserId {
		h.errorResponses.InvalidCredentialsResponse(w, r)
		return
	}

	user, err := h.service.GetUserById(contextUserId)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			h.errorResponses.NotFoundResponse(w, r)
		case errors.Is(err, service.ErrUserWithoutRoles):
			panic("user has no roles, should have at least 1")
		default:
			h.errorResponses.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"user_details": user}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}

func (h *Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.GetJWKS(r)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
		return
	}

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"jwks": response}, nil)
	if err != nil {
		h.errorResponses.ServerErrorResponse(w, r, err)
	}
}
