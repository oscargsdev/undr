package delivery

import (
	"fmt"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /test", test)
	return mux
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Testing the identity routing")
}
