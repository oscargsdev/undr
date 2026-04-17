package api

import (
	"net/http"
)

func NewRouter(handler Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handler.registerUserHandler)
	mux.HandleFunc("PUT /activate", handler.activateUserHandler)
	mux.HandleFunc("POST /authenticate", handler.authenticateUserHandler)
	mux.HandleFunc("POST /refresh", handler.refreshTokenHandler)
	mux.Handle("POST /logout", handler.AuthorizationMiddleware(http.HandlerFunc(handler.logoutHandler)))

	mux.Handle("GET /secured", handler.AuthorizationMiddleware(http.HandlerFunc(handler.testSecuredEndpoint)))
	mux.Handle("GET /admin-portal", handler.RequireRoleMiddleware("admin", http.HandlerFunc(handler.OnlyAdminsHandler)))
	mux.Handle("GET /users/{userId}", handler.AuthorizationMiddleware(http.HandlerFunc(handler.MyInfoHandler)))

	mux.HandleFunc("GET /jwks.json", handler.JWKS)

	return mux
}
