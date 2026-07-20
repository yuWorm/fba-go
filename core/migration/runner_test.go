package migration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/migration"
)

func TestRunnerRecordsSuccessfulMigration(t *testing.T) {
	store := &fakeStore{}
	lock := &fakeLock{}
	runner := migration.NewRunner(store, lock)
	runner.Now = func() time.Time {
		return time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	}

	err := runner.Run(context.Background(), []migration.Migration{
		{
			Scope:    "core",
			Version:  "0001",
			Name:     "init",
			Checksum: "sha256:one",
			Up:       func(context.Context) error { return nil },
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !lock.acquired || !lock.released {
		t.Fatalf("lock acquired/released = %v/%v, want true/true", lock.acquired, lock.released)
	}
	if len(store.records) != 1 {
		t.Fatalf("records = %d, want 1", len(store.records))
	}
	record := store.records[0]
	if record.Scope != "core" || record.Version != "0001" || record.Checksum != "sha256:one" {
		t.Fatalf("record = %+v", record)
	}
	if !record.Success {
		t.Fatal("record.Success = false, want true")
	}
	if !record.AppliedAt.Equal(runner.Now()) {
		t.Fatalf("AppliedAt = %v, want %v", record.AppliedAt, runner.Now())
	}
}

func TestRunnerRecordsFailedMigration(t *testing.T) {
	store := &fakeStore{}
	runner := migration.NewRunner(store, &fakeLock{})
	boom := errors.New("boom")

	err := runner.Run(context.Background(), []migration.Migration{
		{
			Scope:   "plugin:admin",
			Version: "0001",
			Name:    "init",
			Up:      func(context.Context) error { return boom },
		},
	})
	if !errors.Is(err, boom) {
		t.Fatalf("Run() error = %v, want boom", err)
	}
	if len(store.records) != 1 {
		t.Fatalf("records = %d, want 1", len(store.records))
	}
	if store.records[0].Success {
		t.Fatal("record.Success = true, want false")
	}
	if store.records[0].Error != "boom" {
		t.Fatalf("record.Error = %q, want boom", store.records[0].Error)
	}
}

func TestRunnerReturnsLockReleaseError(t *testing.T) {
	releaseErr := errors.New("release lock")
	runner := migration.NewRunner(&fakeStore{}, &fakeLock{releaseErr: releaseErr})

	err := runner.Run(context.Background(), nil)

	if !errors.Is(err, releaseErr) {
		t.Fatalf("Run() error = %v, want release error", err)
	}
}

type fakeStore struct {
	applied map[string]bool
	records []migration.Record
}

func (s *fakeStore) IsApplied(_ context.Context, scope string, version string) (bool, error) {
	return s.applied[scope+":"+version], nil
}

func (s *fakeStore) Record(_ context.Context, record migration.Record) error {
	s.records = append(s.records, record)
	return nil
}

type fakeLock struct {
	acquired   bool
	released   bool
	releaseErr error
}

func (l *fakeLock) Acquire(context.Context) (func(context.Context) error, error) {
	l.acquired = true
	return func(context.Context) error {
		l.released = true
		return l.releaseErr
	}, nil
}
