package errors

import (
	"fmt"
	"log/slog"
	"net/http"

	jsonUtils "github.com/oscargsdev/undr/internal/common/json"
)

type ErrorResponseHelper struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *ErrorResponseHelper {
	return &ErrorResponseHelper{
		logger: logger,
	}
}

func (h *ErrorResponseHelper) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	h.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (h *ErrorResponseHelper) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := jsonUtils.Envelope{"error": message}

	err := jsonUtils.WriteJSON(w, status, env, nil)
	if err != nil {
		h.logError(r, err)
		w.WriteHeader(500)
	}
}

func (h *ErrorResponseHelper) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	h.logError(r, err)

	message := "the server encountered an error and could not process your request"
	h.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (h *ErrorResponseHelper) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	h.errorResponse(w, r, http.StatusNotFound, message)
}

func (h *ErrorResponseHelper) methodNotallowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	h.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (h *ErrorResponseHelper) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	h.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (h *ErrorResponseHelper) InvalidRefresTokenResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	h.errorResponse(w, r, http.StatusBadRequest, errors)
}

func (h *ErrorResponseHelper) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	h.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (h *ErrorResponseHelper) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	h.errorResponse(w, r, http.StatusConflict, message)
}

func (h *ErrorResponseHelper) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	h.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (h *ErrorResponseHelper) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	h.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (h *ErrorResponseHelper) InvalidAccessTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing access token"
	h.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (h *ErrorResponseHelper) MalformedTokenResponse(w http.ResponseWriter, r *http.Request) {
	message := "malformed access token"
	h.errorResponse(w, r, http.StatusBadRequest, message)
}

func (h *ErrorResponseHelper) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	h.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (h *ErrorResponseHelper) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	h.errorResponse(w, r, http.StatusForbidden, message)
}

func (h *ErrorResponseHelper) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account does not have the necessary permissions to access this resource"
	h.errorResponse(w, r, http.StatusForbidden, message)
}
