package responses

import (
	"fmt"
	"net/http"

	jsonUtils "github.com/oscargsdev/undr/internal/jsonx"
)

type Responder struct{}

func New() *Responder {
	return &Responder{}
}

func (rp *Responder) errorResponse(w http.ResponseWriter, status int, message any) {
	env := jsonUtils.Envelope{"error": message}

	err := jsonUtils.WriteJSON(w, status, env, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (rp *Responder) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "the server encountered an error and could not process your request"
	rp.errorResponse(w, http.StatusInternalServerError, message)
}

func (rp *Responder) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	rp.errorResponse(w, http.StatusNotFound, message)
}

func (rp *Responder) methodNotallowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	rp.errorResponse(w, http.StatusMethodNotAllowed, message)
}

func (rp *Responder) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	rp.errorResponse(w, http.StatusBadRequest, err.Error())
}

func (rp *Responder) InvalidRefreshTokenResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	rp.errorResponse(w, http.StatusBadRequest, errors)
}

func (rp *Responder) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	rp.errorResponse(w, http.StatusUnprocessableEntity, errors)
}

func (rp *Responder) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	rp.errorResponse(w, http.StatusConflict, message)
}

func (rp *Responder) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	rp.errorResponse(w, http.StatusTooManyRequests, message)
}

func (rp *Responder) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	rp.errorResponse(w, http.StatusUnauthorized, message)
}

func (rp *Responder) InvalidAccessTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing access token"
	rp.errorResponse(w, http.StatusUnauthorized, message)
}

func (rp *Responder) MalformedTokenResponse(w http.ResponseWriter, r *http.Request) {
	message := "malformed access token"
	rp.errorResponse(w, http.StatusBadRequest, message)
}

func (rp *Responder) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	rp.errorResponse(w, http.StatusUnauthorized, message)
}

func (rp *Responder) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	rp.errorResponse(w, http.StatusForbidden, message)
}

func (rp *Responder) NotPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account does not have the necessary roles/permissions to access this resource"
	rp.errorResponse(w, http.StatusForbidden, message)
}
