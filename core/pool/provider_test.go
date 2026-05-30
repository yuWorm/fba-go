package pool_test

import (
	"testing"

	"github.com/yuWorm/fba-go/core/pool"
)

func TestProviderCreatesNamedPools(t *testing.T) {
	provider := pool.NewProvider(map[string]pool.Options{
		"operation_log": {MaxWorkers: 2, QueueSize: 10},
	})
	defer provider.Stop()

	workerPool, ok := provider.Pool("operation_log")
	if !ok {
		t.Fatal("Pool(operation_log) ok = false, want true")
	}
	if workerPool == nil {
		t.Fatal("Pool(operation_log) = nil")
	}
}

func TestProviderReportsMissingPool(t *testing.T) {
	provider := pool.NewProvider(nil)
	defer provider.Stop()

	if _, ok := provider.Pool("missing"); ok {
		t.Fatal("Pool(missing) ok = true, want false")
	}
}
