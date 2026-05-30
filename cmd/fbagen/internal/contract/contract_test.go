package contract_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/yuWorm/fba-go/cmd/fbagen/internal/contract"
)

func TestLoadParsesContracts(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.API.PriorityRoutes) == 0 {
		t.Fatal("PriorityRoutes is empty")
	}
	if loaded.Response.Success.Code != 200 {
		t.Fatalf("success code = %d, want 200", loaded.Response.Success.Code)
	}
	if loaded.Response.Success.Msg != "请求成功" {
		t.Fatalf("success msg = %q, want 请求成功", loaded.Response.Success.Msg)
	}
	if loaded.Redis.Keys["access_token"].Pattern != "fba:token:{user_id}:{session_uuid}" {
		t.Fatalf("access token key = %+v", loaded.Redis.Keys["access_token"])
	}
}

func TestSnapshotWritesAPIContractSummary(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	snapshot, err := contract.Snapshot(loaded)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snapshot.RouteCount == 0 {
		t.Fatal("RouteCount = 0")
	}
	if snapshot.ResponseEnvelope != true {
		t.Fatal("ResponseEnvelope = false, want true")
	}
}

func TestRunnerReportsMissingPriorityRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/captcha" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":200,"msg":"请求成功","data":null}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	loaded := contract.Contracts{
		API: contract.APIContract{
			PriorityRoutes: []contract.Route{
				{Method: "GET", Path: "/api/v1/auth/captcha"},
				{Method: "GET", Path: "/api/v1/missing"},
			},
		},
	}

	result, err := contract.Test(contract.TestOptions{
		BaseURL:   server.URL,
		Contracts: loaded,
	})
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if len(result.Failures) != 1 {
		t.Fatalf("failures = %d, want 1", len(result.Failures))
	}
}
