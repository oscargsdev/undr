package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/identity/domain"
	"github.com/oscargsdev/undr/internal/identity/service"
	"github.com/oscargsdev/undr/internal/validator"

	"github.com/oscargsdev/undr/internal/jsonx"
	"github.com/oscargsdev/undr/internal/responses"
)

type IdentityService interface {
	RegisterUser(context.Context, *domain.User) (*domain.OpaqueToken, error)
	ActivateUser(context.Context, string) (refreshTokenString string, accessTokenString string, err error)
	AuthenticateUser(context.Context, string, string) (refreshTokenString string, accessTokenString string, err error)
	GetUserById(context.Context, int64) (*domain.UserDetails, error)
	RefreshToken(context.Context, string) (refreshTokenString string, accessTokenString string, err error)
	Logout(context.Context, int64) error
	GetIssuer() string
	GetRefreshExpiration() time.Duration
	GetJWKS(r *http.Request) (json.RawMessage, error)
	ValidateJWTToken(tokenString string, issuer string) (*jwt.Token, error)
}

type handler struct {
	service   IdentityService
	logger    *slog.Logger
	responder *responses.Responder
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

func newHandler(svc IdentityService, logger *slog.Logger) *handler {
	return &handler{
		service:   svc,
		logger:    logger,
		responder: responses.New(),
	}
}

func (h *handler) logError(r *http.Request, err error) {
	h.logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
}

func (h *handler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonx.ReadJSON(w, r, &input)
	if err != nil {
		h.responder.BadRequestResponse(w, r, err)
		return
	}

	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
	}

	v := validator.New()

	if domain.ValidateNewUser(v, user, input.Password); !v.Valid() {
		h.responder.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
		return
	}

	activationToken, err := h.service.RegisterUser(r.Context(), user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			h.responder.FailedValidationResponse(w, r, v.Errors)
		case errors.Is(err, service.ErrDuplicateUsername):
			v.AddError("username", "a user with this username already exists")
			h.responder.FailedValidationResponse(w, r, v.Errors)
		default:
			h.logError(r, err)
			h.responder.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = jsonx.WriteJSON(w, http.StatusAccepted, jsonx.Envelope{"user": user, "activation_token": activationToken.Plaintext}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"activationToken"`
	}

	err := jsonx.ReadJSON(w, r, &input)
	if err != nil {
		h.responder.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if domain.ValidateOpaqueTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		h.responder.FailedValidationResponse(w, r, v.Errors)
		return
	}

	refreshToken, accessToken, err := h.service.ActivateUser(r.Context(), input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			h.responder.FailedValidationResponse(w, r, v.Errors)
		case errors.Is(err, service.ErrEditConflict):
			h.responder.EditConflictResponse(w, r)
		default:
			h.logError(r, err)
			h.responder.ServerErrorResponse(w, r, err)
		}
		return
	}

	http.SetCookie(w, newRefreshTokenCookie(refreshToken, time.Now().Add(h.service.GetRefreshExpiration())))

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"access_token": accessToken}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) testSecuredEndpoint(w http.ResponseWriter, r *http.Request) {
	userId, err := service.ContextGetUserId(r)
	if err != nil {
		h.responder.AuthenticationRequiredResponse(w, r)
		return
	}

	roles, err := service.ContextGetRoles(r)
	if err != nil {
		h.responder.AuthenticationRequiredResponse(w, r)
		return
	}

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"userId": userId, "roles": roles}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonx.ReadJSON(w, r, &input)
	if err != nil {
		h.responder.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	domain.ValidateEmail(v, input.Email)
	domain.ValidatePassword(v, input.Password)

	if !v.Valid() {
		h.responder.FailedValidationResponse(w, r, v.Errors)
		return
	}

	refreshToken, accessToken, err := h.service.AuthenticateUser(r.Context(), input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			h.responder.InvalidCredentialsResponse(w, r)
		case errors.Is(err, service.ErrInvalidCredentials):
			h.responder.InvalidCredentialsResponse(w, r)
		case errors.Is(err, service.ErrUserNotActivated):
			h.responder.InactiveAccountResponse(w, r)
		default:
			h.logError(r, err)
			h.responder.ServerErrorResponse(w, r, err)
		}
		return
	}

	http.SetCookie(w, newRefreshTokenCookie(refreshToken, time.Now().Add(h.service.GetRefreshExpiration())))

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"access_token": accessToken}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.responder.BadRequestResponse(w, r, err)
		return
	}

	oldToken := refreshTokenCookie.Value

	v := validator.New()

	if domain.ValidateOpaqueTokenPlaintext(v, oldToken); !v.Valid() {
		h.responder.InvalidRefreshTokenResponse(w, r, v.Errors)
		return
	}

	refreshToken, accessToken, err := h.service.RefreshToken(r.Context(), oldToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			h.responder.BadRequestResponse(w, r, domain.ErrInvalidRefreshToken)
		default:
			h.logError(r, err)
			h.responder.ServerErrorResponse(w, r, err)
		}
		return
	}

	http.SetCookie(w, newRefreshTokenCookie(refreshToken, time.Now().Add(h.service.GetRefreshExpiration())))

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"access_token": accessToken}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := service.ContextGetUserId(r)
	if err != nil {
		h.responder.AuthenticationRequiredResponse(w, r)
		return
	}

	err = h.service.Logout(r.Context(), userId)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
		return
	}

	http.SetCookie(w, newRefreshTokenCookie("", time.Unix(0, 0)))

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) onlyAdminsHandler(w http.ResponseWriter, r *http.Request) {
	err := jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"howdy": "you are an admin!"}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) myInfoHandler(w http.ResponseWriter, r *http.Request) {
	pathUserId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		h.responder.BadRequestResponse(w, r, err)
		return
	}

	contextUserId, err := service.ContextGetUserId(r)
	if err != nil {
		h.responder.AuthenticationRequiredResponse(w, r)
		return
	}

	if pathUserId != contextUserId {
		h.responder.InvalidCredentialsResponse(w, r)
		return
	}

	user, err := h.service.GetUserById(r.Context(), contextUserId)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			h.responder.NotFoundResponse(w, r)
		case errors.Is(err, service.ErrUserWithoutRoles):
			h.logError(r, err)
			h.responder.ServerErrorResponse(w, r, err)
		default:
			h.logError(r, err)
			h.responder.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"user_details": user}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}

func (h *handler) jwksHandler(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.GetJWKS(r)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
		return
	}

	err = jsonx.WriteJSON(w, http.StatusOK, jsonx.Envelope{"jwks": response}, nil)
	if err != nil {
		h.logError(r, err)
		h.responder.ServerErrorResponse(w, r, err)
	}
}
