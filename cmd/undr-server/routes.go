package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	router.Handle("/v1/identity/", http.StripPrefix("/v1/identity", app.identityModule.Router))

	return router
}
