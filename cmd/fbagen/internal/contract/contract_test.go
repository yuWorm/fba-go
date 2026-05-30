package contract_test

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
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

func TestLoadIncludesFirstBatchPythonParityPriorityRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/dict-types/all"},
		{"GET", "/api/v1/dict-types"},
		{"GET", "/api/v1/dict-datas/all"},
		{"GET", "/api/v1/dict-datas"},
		{"GET", "/api/v1/tasks/registered"},
		{"GET", "/api/v1/task-results"},
		{"GET", "/api/v1/schedulers/all"},
		{"GET", "/api/v1/schedulers"},
	} {
		if !hasPriorityRoute(loaded.API.PriorityRoutes, tc.method, tc.path) {
			t.Fatalf("missing priority route %s %s", tc.method, tc.path)
		}
	}

	swaggerLogin := findPriorityRoute(loaded.API.PriorityRoutes, "POST", "/api/v1/auth/login/swagger")
	if swaggerLogin == nil {
		t.Fatal("missing priority route POST /api/v1/auth/login/swagger")
	}
	if swaggerLogin.ResponseEnvelope == nil || *swaggerLogin.ResponseEnvelope {
		t.Fatal("swagger login priority route must opt out of response envelope")
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
	loaded := contract.Contracts{
		API: contract.APIContract{
			PriorityRoutes: []contract.Route{
				{Method: "GET", Path: "/api/v1/auth/captcha"},
				{Method: "GET", Path: "/api/v1/missing"},
			},
		},
	}

	result, err := contract.Test(contract.TestOptions{
		BaseURL:   "http://fba.test",
		Contracts: loaded,
		Client:    &http.Client{Transport: fakeTransport{}},
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

func TestRunnerReportsPriorityRouteWithoutResponseEnvelope(t *testing.T) {
	loaded := contract.Contracts{
		API: contract.APIContract{
			PriorityRoutes: []contract.Route{
				{Method: "GET", Path: "/api/v1/auth/codes"},
			},
		},
		Response: contract.ResponseContract{
			Success: contract.ResponseSuccess{
				Envelope:       true,
				RequiredFields: []string{"code", "msg", "data"},
				Code:           200,
				Msg:            "请求成功",
			},
		},
	}

	result, err := contract.Test(contract.TestOptions{
		BaseURL:   "http://fba.test",
		Contracts: loaded,
		Client: &http.Client{Transport: fakeTransport{
			"/api/v1/auth/codes": {
				Status: http.StatusOK,
				Body:   `{}`,
			},
		}},
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
	if !strings.Contains(result.Failures[0].Error, "missing response envelope field") {
		t.Fatalf("failure = %q, want missing response envelope field", result.Failures[0].Error)
	}
}

func TestRunnerReportsUnexpectedSuccessEnvelopeValues(t *testing.T) {
	loaded := contract.Contracts{
		API: contract.APIContract{
			PriorityRoutes: []contract.Route{
				{Method: "GET", Path: "/api/v1/auth/codes"},
			},
		},
		Response: contract.ResponseContract{
			Success: contract.ResponseSuccess{
				Envelope:       true,
				RequiredFields: []string{"code", "msg", "data"},
				Code:           200,
				Msg:            "请求成功",
			},
		},
	}

	result, err := contract.Test(contract.TestOptions{
		BaseURL:   "http://fba.test",
		Contracts: loaded,
		Client: &http.Client{Transport: fakeTransport{
			"/api/v1/auth/codes": {
				Status: http.StatusOK,
				Body:   `{"code":0,"msg":"ok","data":[]}`,
			},
		}},
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
	if !strings.Contains(result.Failures[0].Error, "unexpected response code") {
		t.Fatalf("failure = %q, want unexpected response code", result.Failures[0].Error)
	}
}

func TestRunnerUsesPriorityRouteSamplePath(t *testing.T) {
	loaded := contract.Contracts{
		API: contract.APIContract{
			PriorityRoutes: []contract.Route{
				{
					Method:     "GET",
					Path:       "/api/v1/dict-datas/type-codes/{code}",
					SamplePath: "/api/v1/dict-datas/type-codes/sys_status",
				},
			},
		},
		Response: contract.ResponseContract{
			Success: contract.ResponseSuccess{
				Envelope:       true,
				RequiredFields: []string{"code", "msg", "data"},
				Code:           200,
				Msg:            "请求成功",
			},
		},
	}

	result, err := contract.Test(contract.TestOptions{
		BaseURL:   "http://fba.test",
		Contracts: loaded,
		Client: &http.Client{Transport: fakeTransport{
			"/api/v1/dict-datas/type-codes/{code}": {
				Status: http.StatusInternalServerError,
				Body:   `{"code":500,"msg":"内部服务器错误","data":null}`,
			},
			"/api/v1/dict-datas/type-codes/sys_status": {
				Status: http.StatusOK,
				Body:   `{"code":200,"msg":"请求成功","data":[]}`,
			},
		}},
	})
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("Passed = false, failures = %+v", result.Failures)
	}
}

func findPriorityRoute(routes []contract.Route, method, path string) *contract.Route {
	for i := range routes {
		if routes[i].Method == method && routes[i].Path == path {
			return &routes[i]
		}
	}
	return nil
}

func hasPriorityRoute(routes []contract.Route, method, path string) bool {
	return findPriorityRoute(routes, method, path) != nil
}

type fakeResponse struct {
	Status int
	Body   string
}

type fakeTransport map[string]fakeResponse

func (transport fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	status := http.StatusNotFound
	body := `{"code":404,"msg":"Not Found","data":null}`
	if response, ok := transport[req.URL.Path]; ok {
		status = response.Status
		body = response.Body
	} else if req.URL.Path == "/api/v1/auth/captcha" {
		status = http.StatusOK
		body = `{"code":200,"msg":"请求成功","data":null}`
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
