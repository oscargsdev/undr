package domain

import (
	"strings"
	"testing"

	"github.com/oscargsdev/undr/internal/validator"
)

func TestValidateOpaqueTokenPlaintext(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		wantErrMsg string
	}{
		{name: "empty token", token: "", wantErrMsg: "must be provided"},
		{name: "wrong length", token: "too-short", wantErrMsg: "must be 26 bytes long"},
		{name: "valid length", token: strings.Repeat("a", 26), wantErrMsg: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateOpaqueTokenPlaintext(v, tt.token)

			gotErrMsg, hasErr := v.Errors["token"]
			if tt.wantErrMsg == "" {
				if hasErr {
					t.Fatalf("expected no token validation error, got %q", gotErrMsg)
				}
				return
			}

			if !hasErr {
				t.Fatalf("expected token validation error %q, got none", tt.wantErrMsg)
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Fatalf("expected token validation error %q, got %q", tt.wantErrMsg, gotErrMsg)
			}
		})
	}
}
