package delivery

import (
	"net/http"
)

func NewRouter(handler Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handler.registerUserHandler)
	mux.HandleFunc("PUT /activate", handler.ActivateUserHandler)

	return mux
}
