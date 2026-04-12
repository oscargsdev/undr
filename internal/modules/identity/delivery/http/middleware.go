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
		// This header should only be set for content that is influenced by auth
		// Be aware of it's usage for public content, as it will reduce cache performance
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			h.errorResponses.InvalidAccessTokenResponse(w, r)
			return
		}

		tokenString := headerParts[1]

		token, err := service.ValidateJWTToken(tokenString, h.service.GetIssuer())

		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrTokenMalformed):
				h.errorResponses.MalformedTokenResponse(w, r)
			case errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
				errors.Is(err, jwt.ErrTokenExpired) ||
				errors.Is(err, jwt.ErrTokenNotValidYet) ||
				errors.Is(err, jwt.ErrTokenInvalidIssuer) ||
				errors.Is(err, jwt.ErrInvalidKeyType) ||
				errors.Is(err, service.ErrUnknownClaims):
				h.errorResponses.InvalidAccessTokenResponse(w, r)
			default:
				h.errorResponses.ServerErrorResponse(w, r, err)

			}
			return
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
