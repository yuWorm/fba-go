package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	fba "github.com/yuWorm/fba-go"
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/migration"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/examples/compat-host/internal/generated"
	adminrepo "github.com/yuWorm/fba-plugin-admin/repo"
	configmodel "github.com/yuWorm/fba-plugin-config/model"
	dictmodel "github.com/yuWorm/fba-plugin-dict/model"
	noticemodel "github.com/yuWorm/fba-plugin-notice/model"
	oauth2model "github.com/yuWorm/fba-plugin-oauth2/model"
	taskmodel "github.com/yuWorm/fba-plugin-task/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type applicationOptions struct {
	DBMode    string
	SQLiteDSN string
}

func main() {
	app, err := newApplication()
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func newApplication() (fba.Application, error) {
	return newApplicationWithOptions(applicationOptions{
		DBMode:    os.Getenv("FBA_COMPAT_DB"),
		SQLiteDSN: os.Getenv("FBA_COMPAT_SQLITE_DSN"),
	})
}

func newApplicationWithOptions(options applicationOptions) (fba.Application, error) {
	app, err := fba.NewApplication(fba.Options{})
	if err != nil {
		return nil, err
	}
	provider, err := configureCompatDB(app.Container(), options)
	if err != nil {
		return nil, err
	}

	registry := plugin.NewRegistry()
	if err := generated.RegisterPlugins(registry); err != nil {
		return nil, err
	}

	pluginContext := plugin.NewContext(plugin.ContextOptions{
		Container: app.Container(),
		Router:    app.HTTP(),
		APIGroup:  app.HTTP().Group("/api/v1"),
	})
	if err := registry.RegisterAll(pluginContext); err != nil {
		return nil, err
	}
	if provider != nil {
		if err := runCompatMigrations(context.Background(), pluginContext.Migrations()); err != nil {
			return nil, err
		}
		if err := seedCompatDB(context.Background(), provider); err != nil {
			return nil, err
		}
	}
	plugin.MountRoutes(pluginContext.APIGroup(), pluginContext.Routes(), plugin.WithContainer(pluginContext.Container()))

	return app, nil
}

func configureCompatDB(container *di.Container, options applicationOptions) (db.Provider, error) {
	switch strings.ToLower(options.DBMode) {
	case "", "memory":
		return nil, nil
	case "sqlite":
	default:
		return nil, fmt.Errorf("unsupported compat db mode %q", options.DBMode)
	}

	dsn := options.SQLiteDSN
	if dsn == "" {
		dsn = "file:fba_compat?mode=memory&cache=shared"
	}
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	// In-memory SQLite databases are scoped to a connection; a single connection keeps migrated tables visible to all requests.
	sqlDB.SetMaxOpenConns(1)

	provider := db.NewGORMProvider(gormDB, nil)
	if err := container.Provide(func() db.Provider {
		return provider
	}); err != nil {
		return nil, err
	}
	return provider, nil
}

func runCompatMigrations(ctx context.Context, migrations []migration.Migration) error {
	runner := migration.NewRunner(newCompatMigrationStore(), migration.NoopLock{})
	return runner.Run(ctx, migrations)
}

func seedCompatDB(ctx context.Context, provider db.Provider) error {
	if err := adminrepo.NewGORMRepository(provider, adminrepo.SeedData()).Seed(ctx); err != nil {
		return fmt.Errorf("seed admin fixtures: %w", err)
	}
	return provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedCompatIfEmpty(tx, &dictmodel.DictType{}, dictmodel.SeedDictTypes()); err != nil {
			return fmt.Errorf("seed dict types: %w", err)
		}
		if err := seedCompatIfEmpty(tx, &dictmodel.DictData{}, dictmodel.SeedDictData()); err != nil {
			return fmt.Errorf("seed dict data: %w", err)
		}
		if err := seedCompatIfEmpty(tx, &configmodel.Config{}, configmodel.SeedConfigs()); err != nil {
			return fmt.Errorf("seed configs: %w", err)
		}
		if err := seedCompatIfEmpty(tx, &noticemodel.Notice{}, noticemodel.SeedNotices()); err != nil {
			return fmt.Errorf("seed notices: %w", err)
		}
		if err := seedCompatIfEmpty(tx, &oauth2model.UserSocial{}, []oauth2model.UserSocial{}); err != nil {
			return fmt.Errorf("seed oauth2 user socials: %w", err)
		}
		if err := seedCompatIfEmpty(tx, &taskmodel.TaskScheduler{}, taskmodel.SeedSchedulers()); err != nil {
			return fmt.Errorf("seed task schedulers: %w", err)
		}
		if err := seedCompatIfEmpty(tx, &taskmodel.TaskResult{}, taskmodel.SeedTaskResults()); err != nil {
			return fmt.Errorf("seed task results: %w", err)
		}
		return nil
	})
}

func seedCompatIfEmpty[T any](tx *gorm.DB, table any, items []T) error {
	if len(items) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(table).Count(&count).Error; err != nil {
		return err
	}
	// Contract DB mode may use a file-backed DSN, so fixture seeding stays idempotent across restarts.
	if count > 0 {
		return nil
	}
	return tx.Create(&items).Error
}

type compatMigrationStore struct {
	applied map[string]bool
}

func newCompatMigrationStore() *compatMigrationStore {
	return &compatMigrationStore{applied: make(map[string]bool)}
}

func (s *compatMigrationStore) IsApplied(_ context.Context, scope string, version string) (bool, error) {
	return s.applied[scope+":"+version], nil
}

func (s *compatMigrationStore) Record(_ context.Context, record migration.Record) error {
	if record.Success {
		s.applied[record.Scope+":"+record.Version] = true
	}
	return nil
}
