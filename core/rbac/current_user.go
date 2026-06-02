package rbac

type CurrentUser struct {
	ID           int64
	Username     string
	DeptID       *int64
	IsSuperAdmin bool
	IsStaff      bool
	Roles        []Role
}

type Role struct {
	ID             int64
	Code           string
	Enabled        bool
	IsFilterScopes bool
	Permissions    []string
}
