package delivery

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oscargsdev/undr/internal/modules/identity/service"
)

func (h *Handler) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: What does this do?
		// w.Header().Add("Vary", "Authorization")

		// Extract Auth header
		authorizationHeader := r.Header.Get("Authorization")

		// TODO: Handle anonymous user, do I want them for this app?
		// if authorizationHeader == "" {
		// 	r = app.contextSetUser(r, domain.AnonymousUser)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			h.errorResponses.InvalidAccessTokenResponse(w, r)
			return
		}

		tokenString := headerParts[1]

		token, err := service.ValidateJWTToken(tokenString)

		if err != nil {
			// TODO: Handle extra custom errors introduced in token validation
			switch {
			case errors.Is(err, jwt.ErrTokenMalformed):
				h.errorResponses.MalformedTokenResponse(w, r)
				return
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				h.errorResponses.InvalidAccessTokenResponse(w, r)
				return
			case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
				h.errorResponses.InvalidAccessTokenResponse(w, r)
				return
			case errors.Is(err, service.ErrUnknownClaims):
				h.errorResponses.InvalidAccessTokenResponse(w, r)
			default:
				h.errorResponses.ServerErrorResponse(w, r, err)
				return
			}
		}

		r = service.ContextSetClaims(r, token)
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RequireRoleMiddleware(code string, next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles := service.ContextGetRoles(r)

		if !slices.Contains(roles, code) {
			h.errorResponses.NotPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return h.AuthorizationMiddleware(http.HandlerFunc(fn))
}
