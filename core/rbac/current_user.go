package rbac

type CurrentUser struct {
	ID           int64
	Username     string
	IsSuperAdmin bool
	IsStaff      bool
	Roles        []Role
}

type Role struct {
	ID          int64
	Code        string
	Enabled     bool
	Permissions []string
}
