package delivery

import (
	"net/http"
)

func NewRouter(handler Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", handler.registerUserHandler)

	return mux
}
