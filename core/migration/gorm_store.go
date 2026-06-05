package migration

import (
	"context"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMStore struct {
	database  *gorm.DB
	ensure    sync.Once
	ensureErr error
}

type GORMRecord struct {
	ID          uint      `gorm:"column:id;primaryKey"`
	Scope       string    `gorm:"column:scope;size:128;not null;uniqueIndex:idx_fba_migration_scope_version"`
	Version     string    `gorm:"column:version;size:64;not null;uniqueIndex:idx_fba_migration_scope_version"`
	Name        string    `gorm:"column:name;size:256"`
	Checksum    string    `gorm:"column:checksum;size:128"`
	AppliedAt   time.Time `gorm:"column:applied_at"`
	ExecutionMS int64     `gorm:"column:execution_ms"`
	Success     bool      `gorm:"column:success;index"`
	Error       string    `gorm:"column:error;type:text"`
}

func (GORMRecord) TableName() string {
	return "fba_migration_records"
}

func NewGORMStore(database *gorm.DB) *GORMStore {
	return &GORMStore{database: database}
}

func (s *GORMStore) IsApplied(ctx context.Context, scope string, version string) (bool, error) {
	if err := s.ensureTable(ctx); err != nil {
		return false, err
	}
	var count int64
	err := s.database.WithContext(ctx).
		Model(&GORMRecord{}).
		Where("scope = ? and version = ? and success = ?", scope, version, true).
		Count(&count).Error
	return count > 0, err
}

func (s *GORMStore) Record(ctx context.Context, record Record) error {
	if err := s.ensureTable(ctx); err != nil {
		return err
	}
	model := GORMRecord{
		Scope:       record.Scope,
		Version:     record.Version,
		Name:        record.Name,
		Checksum:    record.Checksum,
		AppliedAt:   record.AppliedAt,
		ExecutionMS: record.ExecutionMS,
		Success:     record.Success,
		Error:       record.Error,
	}
	return s.database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "scope"}, {Name: "version"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"checksum",
			"applied_at",
			"execution_ms",
			"success",
			"error",
		}),
	}).Create(&model).Error
}

func (s *GORMStore) ensureTable(ctx context.Context) error {
	s.ensure.Do(func() {
		if s.database == nil {
			s.ensureErr = gorm.ErrInvalidDB
			return
		}
		// The migration store bootstraps itself before any application migration
		// can be recorded, so it intentionally uses AutoMigrate outside Runner.
		s.ensureErr = s.database.WithContext(ctx).AutoMigrate(&GORMRecord{})
	})
	return s.ensureErr
}
