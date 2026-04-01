package delivery

import (
	"net/http"
)

func NewRouter(handler Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handler.registerUserHandler)
	mux.HandleFunc("PUT /activate", handler.ActivateUserHandler)
	mux.HandleFunc("GET /validate", handler.TestTokenValidationHandler)

	return mux
}
