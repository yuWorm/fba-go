package db

import (
	"fmt"
	"strings"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/yuWorm/fba-go/core/config"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(opts config.DatabaseOptions) (Provider, error) {
	write, err := OpenGORM(opts.Driver, opts.WriteDSN, opts)
	if err != nil {
		return nil, err
	}
	var read *gorm.DB
	if strings.TrimSpace(opts.ReadDSN) != "" {
		read, err = OpenGORM(opts.Driver, opts.ReadDSN, opts)
		if err != nil {
			return nil, err
		}
	}
	return NewGORMProvider(write, read), nil
}

func OpenGORM(driver string, dsn string, opts config.DatabaseOptions) (*gorm.DB, error) {
	driver = normalizeDriver(driver)
	if driver == "" {
		return nil, fmt.Errorf("database driver is required")
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("database dsn is required")
	}

	var dialector gorm.Dialector
	switch driver {
	case "mysql":
		dialector = gormmysql.Open(normalizeMySQLDSN(dsn))
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", driver)
	}
	database, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if driver == "mysql" {
		database = database.Set("gorm:table_options", mysqlTableOptions())
	}
	if err := configurePool(database, opts); err != nil {
		return nil, err
	}
	return database, nil
}

func configurePool(database *gorm.DB, opts config.DatabaseOptions) error {
	sqlDB, err := database.DB()
	if err != nil {
		return err
	}
	if opts.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}
	if opts.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
	}
	return nil
}

func normalizeDriver(driver string) string {
	return strings.ToLower(strings.TrimSpace(driver))
}

func normalizeMySQLDSN(dsn string) string {
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return dsn
	}
	if cfg.Params == nil {
		cfg.Params = make(map[string]string)
	}
	// Admin seed data contains 4-byte Unicode such as emoji. utf8/utf8mb3
	// connections can transmit Chinese text but still fail on those values.
	cfg.Params["charset"] = "utf8mb4"
	if cfg.Collation == "" || !strings.HasPrefix(strings.ToLower(cfg.Collation), "utf8mb4_") {
		cfg.Collation = "utf8mb4_unicode_ci"
	}
	return cfg.FormatDSN()
}

func mysqlTableOptions() string {
	return "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
}
