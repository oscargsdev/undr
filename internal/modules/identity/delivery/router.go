package delivery

import (
	"net/http"
)

func NewRouter(handler Handler) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/test", handler.TestHandler)

	return mux
}
