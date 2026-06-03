package migration

import (
	"context"

	"github.com/yuWorm/fba-go/core/db"
	coremigration "github.com/yuWorm/fba-go/core/migration"
	"github.com/yuWorm/fba-plugin-oauth2/model"
)

func AutoMigrate(provider db.Provider) coremigration.Migration {
	return coremigration.Migration{
		Scope:   "plugin:oauth2",
		Version: "2026060301",
		Up: func(ctx context.Context) error {
			return provider.Write().WithContext(ctx).AutoMigrate(&model.UserSocial{})
		},
	}
}
