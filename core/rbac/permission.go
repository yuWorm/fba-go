package rbac

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrNoEnabledRole     = errors.New("no enabled role")
	ErrNoRoleMenus       = errors.New("no role menus")
	ErrStaffRequired     = errors.New("staff required")
	ErrSuperuserRequired = errors.New("superuser required")
	ErrPermissionDenied  = errors.New("permission denied")
)

type RouteAccess struct {
	Method            string
	Permission        string
	SuperuserRequired bool
	Whitelisted       bool
}

func Authorize(user *CurrentUser, route RouteAccess) error {
	if route.Whitelisted {
		return nil
	}
	if user == nil {
		return ErrUnauthenticated
	}
	// Superuser-only routes are stricter than role permissions and do not depend on enabled roles.
	if route.SuperuserRequired {
		if user.IsSuperAdmin {
			return nil
		}
		return ErrSuperuserRequired
	}
	if user.IsSuperAdmin {
		return nil
	}

	roles := enabledRoles(user.Roles)
	if len(roles) == 0 {
		return ErrNoEnabledRole
	}
	// Keep this before the staff/write guard to match the Python admin RBAC contract:
	// an enabled role without any menus is a role assignment problem, not an operation privilege problem.
	if !hasAnyRoleMenu(roles) {
		return ErrNoRoleMenus
	}
	if isWriteMethod(route.Method) && !user.IsStaff {
		return ErrStaffRequired
	}
	if route.Permission == "" {
		return nil
	}
	for _, role := range roles {
		if slices.Contains(role.Permissions, route.Permission) {
			return nil
		}
	}
	return ErrPermissionDenied
}

func enabledRoles(roles []Role) []Role {
	out := make([]Role, 0, len(roles))
	for _, role := range roles {
		if role.Enabled {
			out = append(out, role)
		}
	}
	return out
}

func hasAnyRoleMenu(roles []Role) bool {
	for _, role := range roles {
		if role.MenuCount > 0 {
			return true
		}
	}
	return false
}

func isWriteMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return false
	default:
		return true
	}
}
