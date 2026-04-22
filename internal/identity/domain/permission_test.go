package domain

import "testing"

func TestRolesInclude(t *testing.T) {
	roles := Roles{"user", "admin"}

	if !roles.Include("user") {
		t.Fatal("expected roles to include user")
	}

	if roles.Include("auditor") {
		t.Fatal("expected roles not to include auditor")
	}
}
