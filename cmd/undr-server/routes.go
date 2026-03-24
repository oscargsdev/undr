package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This thing's starting to be built")
	})

	return router
}
