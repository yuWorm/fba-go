package rbac

import "context"

type SQLExpr struct {
	Clause string
	Args   []any
}

type ModelRef struct {
	Name  string
	Alias string
}

type DataPermissionFilter interface {
	Build(ctx context.Context, user *CurrentUser, models ...ModelRef) (SQLExpr, error)
}
