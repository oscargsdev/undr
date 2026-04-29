package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter_DemoRoutes(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		path       string
		wantStatus int
	}{
		{
			name:       "secured route enabled",
			enabled:    true,
			path:       "/secured",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "admin route enabled",
			enabled:    true,
			path:       "/admin-portal",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "secured route disabled",
			enabled:    false,
			path:       "/secured",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "admin route disabled",
			enabled:    false,
			path:       "/admin-portal",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(&mockIdentityService{}, newTestLogger(), RouterConfig{EnableDemoRoutes: tt.enabled})
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assertStatus(t, rr, tt.wantStatus)
		})
	}
}
