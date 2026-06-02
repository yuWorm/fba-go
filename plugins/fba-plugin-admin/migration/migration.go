package migration

import (
	"context"

	"github.com/yuWorm/fba-go/core/db"
	coremigration "github.com/yuWorm/fba-go/core/migration"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

func AutoMigrate(provider db.Provider) coremigration.Migration {
	return coremigration.Migration{
		Scope:    "plugin:admin",
		Version:  "0001",
		Name:     "admin core tables",
		Checksum: "go:auto-migrate:admin:0001",
		Up: func(ctx context.Context) error {
			return provider.Write().WithContext(ctx).AutoMigrate(
				&model.User{},
				&model.Role{},
				&model.Menu{},
				&model.Dept{},
				&model.DataRule{},
				&model.DataScope{},
				&model.Plugin{},
				&model.LoginLog{},
				&model.OperaLog{},
				&model.Session{},
				&repo.UserRole{},
				&repo.RoleMenu{},
				&repo.RoleDataScope{},
				&repo.DataScopeRule{},
				&repo.PluginState{},
			)
		},
	}
}
