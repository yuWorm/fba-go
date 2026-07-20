package migration

import (
	"context"
	"errors"
	"time"
)

type Store interface {
	IsApplied(ctx context.Context, scope string, version string) (bool, error)
	Record(ctx context.Context, record Record) error
}

type Record struct {
	Scope       string
	Version     string
	Name        string
	Checksum    string
	AppliedAt   time.Time
	ExecutionMS int64
	Success     bool
	Error       string
}

type Runner struct {
	store Store
	lock  Lock
	Now   func() time.Time
}

func NewRunner(store Store, lock Lock) *Runner {
	if lock == nil {
		lock = NoopLock{}
	}
	return &Runner{
		store: store,
		lock:  lock,
		Now:   time.Now,
	}
}

func (r *Runner) Run(ctx context.Context, migrations []Migration) (runErr error) {
	release, err := r.lock.Acquire(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Acquisition contexts are commonly cancelled during shutdown. Use a
		// bounded detached context so session-scoped database locks still release.
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		runErr = errors.Join(runErr, release(releaseCtx))
	}()

	for _, m := range migrations {
		applied, err := r.store.IsApplied(ctx, m.Scope, m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := r.runOne(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runOne(ctx context.Context, m Migration) error {
	start := time.Now()
	record := Record{
		Scope:     m.Scope,
		Version:   m.Version,
		Name:      m.Name,
		Checksum:  m.Checksum,
		AppliedAt: r.Now(),
	}

	err := m.Up(ctx)
	record.ExecutionMS = time.Since(start).Milliseconds()
	record.Success = err == nil
	if err != nil {
		record.Error = err.Error()
	}

	if recordErr := r.store.Record(ctx, record); recordErr != nil {
		return recordErr
	}
	return err
}
