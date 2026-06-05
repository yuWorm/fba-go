package db

import (
	"fmt"
	"strings"

	"github.com/yuWorm/fba-go/core/config"
	"gorm.io/driver/mysql"
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
		dialector = mysql.Open(dsn)
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
