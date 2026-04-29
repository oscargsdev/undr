package api

import (
	"log/slog"
	"net/http"
)

type RouterConfig struct {
	EnableDemoRoutes bool
}

func NewRouter(svc IdentityService, logger *slog.Logger, cfg RouterConfig) http.Handler {
	handler := newHandler(svc, logger)
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handler.registerUserHandler)
	mux.HandleFunc("PUT /activate", handler.activateUserHandler)
	mux.HandleFunc("POST /authenticate", handler.authenticateUserHandler)
	mux.HandleFunc("POST /refresh", handler.refreshTokenHandler)
	mux.Handle("POST /logout", handler.authorizationMiddleware(http.HandlerFunc(handler.logoutHandler)))
	mux.Handle("GET /users/{userId}", handler.authorizationMiddleware(http.HandlerFunc(handler.myInfoHandler)))

	if cfg.EnableDemoRoutes {
		mux.Handle("GET /secured", handler.authorizationMiddleware(http.HandlerFunc(handler.testSecuredEndpoint)))
		mux.Handle("GET /admin-portal", handler.requireRoleMiddleware("admin", http.HandlerFunc(handler.onlyAdminsHandler)))
	}

	mux.HandleFunc("GET /jwks.json", handler.jwksHandler)

	return mux
}
