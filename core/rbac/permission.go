package rbac

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrNoEnabledRole    = errors.New("no enabled role")
	ErrStaffRequired    = errors.New("staff required")
	ErrPermissionDenied = errors.New("permission denied")
)

type RouteAccess struct {
	Method      string
	Permission  string
	Whitelisted bool
}

func Authorize(user *CurrentUser, route RouteAccess) error {
	if route.Whitelisted {
		return nil
	}
	if user == nil {
		return ErrUnauthenticated
	}
	if user.IsSuperAdmin {
		return nil
	}

	roles := enabledRoles(user.Roles)
	if len(roles) == 0 {
		return ErrNoEnabledRole
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

func isWriteMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return false
	default:
		return true
	}
}
