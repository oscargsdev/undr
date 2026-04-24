package api

import (
	"log/slog"
	"net/http"
)

func NewRouter(svc IdentityService, logger *slog.Logger) http.Handler {
	handler := newHandler(svc, logger)
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handler.registerUserHandler)
	mux.HandleFunc("PUT /activate", handler.activateUserHandler)
	mux.HandleFunc("POST /authenticate", handler.authenticateUserHandler)
	mux.HandleFunc("POST /refresh", handler.refreshTokenHandler)
	mux.Handle("POST /logout", handler.authorizationMiddleware(http.HandlerFunc(handler.logoutHandler)))

	mux.Handle("GET /secured", handler.authorizationMiddleware(http.HandlerFunc(handler.testSecuredEndpoint)))
	mux.Handle("GET /admin-portal", handler.requireRoleMiddleware("admin", http.HandlerFunc(handler.onlyAdminsHandler)))
	mux.Handle("GET /users/{userId}", handler.authorizationMiddleware(http.HandlerFunc(handler.myInfoHandler)))

	mux.HandleFunc("GET /jwks.json", handler.jwksHandler)

	return mux
}
