package errors

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/oscargsdev/undr/internal/common"
	jsonUtils "github.com/oscargsdev/undr/internal/common/json"
)

func logError(r *http.Request, err error, logger *slog.Logger) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	logger.Error(err.Error(), "method", method, "uri", uri)
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, message any, logger *slog.Logger) {
	env := common.Envelope{"error": message}

	err := jsonUtils.WriteJSON(w, status, env, nil)
	if err != nil {
		logError(r, err, logger)
		w.WriteHeader(500)
	}
}

func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	logError(r, err, logger)

	message := "the server encountered an error and could not process your request"
	errorResponse(w, r, http.StatusInternalServerError, message, logger)
}

func notFoundResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "the requested resource could not be found"
	errorResponse(w, r, http.StatusNotFound, message, logger)
}

func methodNotallowedResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	errorResponse(w, r, http.StatusMethodNotAllowed, message, logger)
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	errorResponse(w, r, http.StatusBadRequest, err.Error(), logger)
}

func FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string, logger *slog.Logger) {
	errorResponse(w, r, http.StatusUnprocessableEntity, errors, logger)
}

func EditConflictResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "unable to update the record due to an edit conflict, please try again"
	errorResponse(w, r, http.StatusConflict, message, logger)
}

func rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "rate limit exceeded"
	errorResponse(w, r, http.StatusTooManyRequests, message, logger)
}

func invalidCredentialsResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "invalid authentication credentials"
	errorResponse(w, r, http.StatusUnauthorized, message, logger)
}

func invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	errorResponse(w, r, http.StatusUnauthorized, message, logger)
}

func authenticationRequiredResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "you must be authenticated to access this resource"
	errorResponse(w, r, http.StatusUnauthorized, message, logger)
}

func inactiveAccountResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "your user account must be activated to access this resource"
	errorResponse(w, r, http.StatusForbidden, message, logger)
}

func notPermittedResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "your user account does not have the necessary permissions to access this resource"
	errorResponse(w, r, http.StatusForbidden, message, logger)
}
