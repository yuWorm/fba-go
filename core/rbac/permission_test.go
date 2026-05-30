package rbac_test

import (
	"errors"
	"testing"

	"github.com/yuWorm/fba-go/core/rbac"
)

func TestAuthorizeAllowsWhitelistWithoutUser(t *testing.T) {
	err := rbac.Authorize(nil, rbac.RouteAccess{Whitelisted: true, Method: "GET"})
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
}

func TestAuthorizeRejectsUnauthenticatedUser(t *testing.T) {
	err := rbac.Authorize(nil, rbac.RouteAccess{Method: "GET"})
	if !errors.Is(err, rbac.ErrUnauthenticated) {
		t.Fatalf("Authorize() error = %v, want unauthenticated", err)
	}
}

func TestAuthorizeAllowsSuperAdmin(t *testing.T) {
	user := &rbac.CurrentUser{ID: 1, IsSuperAdmin: true}
	err := rbac.Authorize(user, rbac.RouteAccess{Method: "DELETE", Permission: "sys:user:del"})
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
}

func TestAuthorizeRejectsNonStaffWrite(t *testing.T) {
	user := &rbac.CurrentUser{
		ID: 2,
		Roles: []rbac.Role{
			{Enabled: true, Permissions: []string{"sys:user:del"}},
		},
	}

	err := rbac.Authorize(user, rbac.RouteAccess{Method: "DELETE", Permission: "sys:user:del"})
	if !errors.Is(err, rbac.ErrStaffRequired) {
		t.Fatalf("Authorize() error = %v, want staff required", err)
	}
}

func TestAuthorizeRejectsUsersWithoutEnabledRoles(t *testing.T) {
	user := &rbac.CurrentUser{ID: 2, IsStaff: true}

	err := rbac.Authorize(user, rbac.RouteAccess{Method: "GET"})
	if !errors.Is(err, rbac.ErrNoEnabledRole) {
		t.Fatalf("Authorize() error = %v, want no enabled role", err)
	}
}

func TestAuthorizeRejectsPermissionMismatch(t *testing.T) {
	user := &rbac.CurrentUser{
		ID:      2,
		IsStaff: true,
		Roles: []rbac.Role{
			{Enabled: true, Permissions: []string{"sys:user:view"}},
		},
	}

	err := rbac.Authorize(user, rbac.RouteAccess{Method: "DELETE", Permission: "sys:user:del"})
	if !errors.Is(err, rbac.ErrPermissionDenied) {
		t.Fatalf("Authorize() error = %v, want permission denied", err)
	}
}
