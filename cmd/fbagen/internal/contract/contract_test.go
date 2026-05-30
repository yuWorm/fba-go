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

func TestLoadIncludesSecondBatchAdminParityPriorityRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct {
		method     string
		path       string
		samplePath string
	}{
		{"GET", "/api/v1/sys/users/{pk}", "/api/v1/sys/users/1"},
		{"GET", "/api/v1/sys/users/{pk}/roles", "/api/v1/sys/users/1/roles"},
		{"GET", "/api/v1/sys/users", ""},
		{"GET", "/api/v1/sys/roles/all", ""},
		{"GET", "/api/v1/sys/roles/{pk}/menus", "/api/v1/sys/roles/1/menus"},
		{"GET", "/api/v1/sys/roles/{pk}/scopes", "/api/v1/sys/roles/1/scopes"},
		{"GET", "/api/v1/sys/roles/{pk}", "/api/v1/sys/roles/1"},
		{"GET", "/api/v1/sys/roles", ""},
		{"GET", "/api/v1/sys/menus/{pk}", "/api/v1/sys/menus/1"},
		{"GET", "/api/v1/sys/menus", ""},
		{"GET", "/api/v1/sys/depts/{pk}", "/api/v1/sys/depts/1"},
		{"GET", "/api/v1/sys/depts", ""},
		{"GET", "/api/v1/logs/login", ""},
		{"GET", "/api/v1/logs/opera", ""},
		{"GET", "/api/v1/monitors/server", ""},
		{"GET", "/api/v1/monitors/redis", ""},
		{"GET", "/api/v1/monitors/sessions", ""},
	} {
		route := findPriorityRoute(loaded.API.PriorityRoutes, tc.method, tc.path)
		if route == nil {
			t.Fatalf("missing priority route %s %s", tc.method, tc.path)
		}
		if route.SamplePath != tc.samplePath {
			t.Fatalf("%s %s sample_path = %q, want %q", tc.method, tc.path, route.SamplePath, tc.samplePath)
		}
	}
}

func TestLoadIncludesThirdBatchAdminParityPriorityRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct {
		method           string
		path             string
		samplePath       string
		responseEnvelope *bool
	}{
		{"GET", "/api/v1/sys/data-rules/models", "", nil},
		{"GET", "/api/v1/sys/data-rules/models/{model}/columns", "/api/v1/sys/data-rules/models/user/columns", nil},
		{"GET", "/api/v1/sys/data-rules/value-template-variables", "", nil},
		{"GET", "/api/v1/sys/data-rules/all", "", nil},
		{"GET", "/api/v1/sys/data-rules/{pk}", "/api/v1/sys/data-rules/1", nil},
		{"GET", "/api/v1/sys/data-rules", "", nil},
		{"GET", "/api/v1/sys/data-scopes/all", "", nil},
		{"GET", "/api/v1/sys/data-scopes/{pk}", "/api/v1/sys/data-scopes/1", nil},
		{"GET", "/api/v1/sys/data-scopes/{pk}/rules", "/api/v1/sys/data-scopes/1/rules", nil},
		{"GET", "/api/v1/sys/data-scopes", "", nil},
		{"GET", "/api/v1/sys/plugins", "", nil},
		{"GET", "/api/v1/sys/plugins/changed", "", nil},
		{"GET", "/api/v1/sys/plugins/{plugin}", "/api/v1/sys/plugins/dict", boolPtr(false)},
	} {
		route := findPriorityRoute(loaded.API.PriorityRoutes, tc.method, tc.path)
		if route == nil {
			t.Fatalf("missing priority route %s %s", tc.method, tc.path)
		}
		if route.SamplePath != tc.samplePath {
			t.Fatalf("%s %s sample_path = %q, want %q", tc.method, tc.path, route.SamplePath, tc.samplePath)
		}
		if tc.responseEnvelope != nil {
			if route.ResponseEnvelope == nil || *route.ResponseEnvelope != *tc.responseEnvelope {
				t.Fatalf("%s %s response_envelope = %v, want %v", tc.method, tc.path, route.ResponseEnvelope, *tc.responseEnvelope)
			}
		}
	}
}

func TestLoadIncludesWriteMethodParityPriorityRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct {
		method       string
		path         string
		samplePath   string
		bodyRequired bool
	}{
		{"POST", "/api/v1/dict-types", "", true},
		{"PUT", "/api/v1/dict-types/{pk}", "/api/v1/dict-types/1", true},
		{"DELETE", "/api/v1/dict-types", "", true},
		{"POST", "/api/v1/dict-datas", "", true},
		{"PUT", "/api/v1/dict-datas/{pk}", "/api/v1/dict-datas/1", true},
		{"DELETE", "/api/v1/dict-datas", "", true},
		{"DELETE", "/api/v1/tasks/{task_id}/cancel", "/api/v1/tasks/task-1/cancel", false},
		{"DELETE", "/api/v1/task-results", "", true},
		{"POST", "/api/v1/schedulers", "", true},
		{"PUT", "/api/v1/schedulers/{pk}", "/api/v1/schedulers/1", true},
		{"PUT", "/api/v1/schedulers/{pk}/status", "/api/v1/schedulers/1/status", false},
		{"POST", "/api/v1/schedulers/{pk}/execute", "/api/v1/schedulers/1/execute", false},
		{"DELETE", "/api/v1/schedulers/{pk}", "/api/v1/schedulers/1", false},
	} {
		route := findPriorityRoute(loaded.API.PriorityRoutes, tc.method, tc.path)
		if route == nil {
			t.Fatalf("missing priority route %s %s", tc.method, tc.path)
		}
		if route.SamplePath != tc.samplePath {
			t.Fatalf("%s %s sample_path = %q, want %q", tc.method, tc.path, route.SamplePath, tc.samplePath)
		}
		if tc.bodyRequired && (route.Request == nil || route.Request.Body == "") {
			t.Fatalf("%s %s missing request body sample", tc.method, tc.path)
		}
	}
}

func TestLoadIncludesAdminWriteParityPriorityRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct {
		method       string
		path         string
		samplePath   string
		bodyRequired bool
		permission   string
	}{
		{"POST", "/api/v1/sys/users", "", true, ""},
		{"PUT", "/api/v1/sys/users/{pk}", "/api/v1/sys/users/1", true, ""},
		{"PUT", "/api/v1/sys/users/{pk}/permissions", "/api/v1/sys/users/1/permissions?type=multi_login", false, ""},
		{"PUT", "/api/v1/sys/users/me/password", "", true, ""},
		{"PUT", "/api/v1/sys/users/{pk}/password", "/api/v1/sys/users/1/password", true, ""},
		{"PUT", "/api/v1/sys/users/me/nickname", "", true, ""},
		{"PUT", "/api/v1/sys/users/me/avatar", "", true, ""},
		{"PUT", "/api/v1/sys/users/me/email", "", true, ""},
		{"DELETE", "/api/v1/sys/users/{pk}", "/api/v1/sys/users/999999", false, "sys:user:del"},
		{"POST", "/api/v1/sys/roles", "", true, "sys:role:add"},
		{"PUT", "/api/v1/sys/roles/{pk}", "/api/v1/sys/roles/1", true, "sys:role:edit"},
		{"PUT", "/api/v1/sys/roles/{pk}/menus", "/api/v1/sys/roles/1/menus", true, "sys:role:menu:edit"},
		{"PUT", "/api/v1/sys/roles/{pk}/scopes", "/api/v1/sys/roles/1/scopes", true, ""},
		{"DELETE", "/api/v1/sys/roles", "", true, "sys:role:del"},
		{"POST", "/api/v1/sys/menus", "", true, "sys:menu:add"},
		{"PUT", "/api/v1/sys/menus/{pk}", "/api/v1/sys/menus/1", true, "sys:menu:edit"},
		{"DELETE", "/api/v1/sys/menus/{pk}", "/api/v1/sys/menus/1", false, "sys:menu:del"},
		{"POST", "/api/v1/sys/depts", "", true, ""},
		{"PUT", "/api/v1/sys/depts/{pk}", "/api/v1/sys/depts/1", true, ""},
		{"DELETE", "/api/v1/sys/depts/{pk}", "/api/v1/sys/depts/1", false, ""},
		{"POST", "/api/v1/sys/data-rules", "", true, "data:rule:add"},
		{"PUT", "/api/v1/sys/data-rules/{pk}", "/api/v1/sys/data-rules/1", true, "data:rule:edit"},
		{"DELETE", "/api/v1/sys/data-rules", "", true, "data:rule:del"},
		{"POST", "/api/v1/sys/data-scopes", "", true, "data:scope:add"},
		{"PUT", "/api/v1/sys/data-scopes/{pk}", "/api/v1/sys/data-scopes/1", true, "data:scope:edit"},
		{"PUT", "/api/v1/sys/data-scopes/{pk}/rules", "/api/v1/sys/data-scopes/1/rules", true, "data:scope:rule:edit"},
		{"DELETE", "/api/v1/sys/data-scopes", "", true, "data:scope:del"},
		{"POST", "/api/v1/sys/files/upload", "", true, "sys:file:upload"},
		{"POST", "/api/v1/sys/plugins", "/api/v1/sys/plugins?type=git&repo_url=https://example.invalid/plugin.git", false, ""},
		{"DELETE", "/api/v1/sys/plugins/{plugin}", "/api/v1/sys/plugins/dict", false, ""},
		{"PUT", "/api/v1/sys/plugins/{plugin}/status", "/api/v1/sys/plugins/dict/status", false, ""},
		{"DELETE", "/api/v1/logs/login", "", true, "log:login:del"},
		{"DELETE", "/api/v1/logs/login/all", "", false, "log:login:clear"},
		{"DELETE", "/api/v1/logs/opera", "", true, "log:opera:del"},
		{"DELETE", "/api/v1/logs/opera/all", "", false, "log:opera:clear"},
		{"DELETE", "/api/v1/monitors/sessions/{pk}", "/api/v1/monitors/sessions/1?session_uuid=fixture-session", false, ""},
	} {
		route := findPriorityRoute(loaded.API.PriorityRoutes, tc.method, tc.path)
		if route == nil {
			t.Fatalf("missing priority route %s %s", tc.method, tc.path)
		}
		if route.SamplePath != tc.samplePath {
			t.Fatalf("%s %s sample_path = %q, want %q", tc.method, tc.path, route.SamplePath, tc.samplePath)
		}
		if tc.bodyRequired && (route.Request == nil || route.Request.Body == "") {
			t.Fatalf("%s %s missing request body sample", tc.method, tc.path)
		}
		if route.Permission != tc.permission {
			t.Fatalf("%s %s permission = %q, want %q", tc.method, tc.path, route.Permission, tc.permission)
		}
	}

	upload := findPriorityRoute(loaded.API.PriorityRoutes, "POST", "/api/v1/sys/files/upload")
	if upload == nil {
		t.Fatal("missing priority route POST /api/v1/sys/files/upload")
	}
	if upload.Request == nil || upload.Request.ContentType != "multipart/form-data; boundary=fba-contract" {
		t.Fatalf("POST /api/v1/sys/files/upload content_type = %#v, want multipart sample", upload.Request)
	}
}

func TestLoadIncludesNoticePluginPriorityRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct {
		method       string
		path         string
		samplePath   string
		bodyRequired bool
		permission   string
	}{
		{"GET", "/api/v1/sys/notices/{pk}", "/api/v1/sys/notices/1", false, ""},
		{"GET", "/api/v1/sys/notices", "", false, ""},
		{"POST", "/api/v1/sys/notices", "", true, "sys:notice:add"},
		{"PUT", "/api/v1/sys/notices/{pk}", "/api/v1/sys/notices/1", true, "sys:notice:edit"},
		{"DELETE", "/api/v1/sys/notices", "", true, "sys:notice:del"},
	} {
		route := findPriorityRoute(loaded.API.PriorityRoutes, tc.method, tc.path)
		if route == nil {
			t.Fatalf("missing priority route %s %s", tc.method, tc.path)
		}
		if route.SamplePath != tc.samplePath {
			t.Fatalf("%s %s sample_path = %q, want %q", tc.method, tc.path, route.SamplePath, tc.samplePath)
		}
		if tc.bodyRequired && (route.Request == nil || route.Request.Body == "") {
			t.Fatalf("%s %s missing request body sample", tc.method, tc.path)
		}
		if route.Permission != tc.permission {
			t.Fatalf("%s %s permission = %q, want %q", tc.method, tc.path, route.Permission, tc.permission)
		}
	}
}

func TestPriorityRoutesCoverDeclaredAPIRoutes(t *testing.T) {
	loaded, err := contract.Load(filepath.Join("..", "..", "..", "..", "contracts"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var missing []string
	for _, route := range loaded.API.Routes {
		if findPriorityRoute(loaded.API.PriorityRoutes, route.Method, route.Path) == nil {
			missing = append(missing, route.Method+" "+route.Path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("priority_routes missing %d declared routes: %s", len(missing), strings.Join(missing, ", "))
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
	failure := result.Failures[0]
	if failure.Method != "GET" {
		t.Fatalf("failure method = %q, want GET", failure.Method)
	}
	if failure.Path != "/api/v1/missing" {
		t.Fatalf("failure path = %q, want /api/v1/missing", failure.Path)
	}
	if failure.SamplePath != "/api/v1/missing" {
		t.Fatalf("failure sample_path = %q, want /api/v1/missing", failure.SamplePath)
	}
	if failure.StatusCode != http.StatusNotFound {
		t.Fatalf("failure status = %d, want 404", failure.StatusCode)
	}
	if !strings.Contains(failure.ResponseBody, "Not Found") {
		t.Fatalf("failure response body = %q, want Not Found", failure.ResponseBody)
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

func TestRunnerSendsRequestSampleBody(t *testing.T) {
	loaded := contract.Contracts{
		API: contract.APIContract{
			PriorityRoutes: []contract.Route{
				{
					Method: "POST",
					Path:   "/api/v1/dict-types",
					Request: &contract.RequestSample{
						ContentType: "application/json",
						Body:        `{"name":"契约类型","code":"contract_status","remark":null}`,
					},
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
		Client: &http.Client{Transport: assertRequestTransport{
			t:            t,
			wantMethod:   "POST",
			wantPath:     "/api/v1/dict-types",
			wantType:     "application/json",
			wantBody:     `{"name":"契约类型","code":"contract_status","remark":null}`,
			status:       http.StatusOK,
			responseBody: `{"code":200,"msg":"请求成功","data":null}`,
		}},
	})
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("Passed = false, failures = %+v", result.Failures)
	}
}

func TestRunnerBootstrapsAuthForProtectedPriorityRoutes(t *testing.T) {
	loaded := contract.Contracts{
		API: contract.APIContract{
			BasePath: "/api/v1",
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
		Client: &http.Client{Transport: authBootstrapTransport{
			t: t,
		}},
	})
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("Passed = false, failures = %+v", result.Failures)
	}
}

func TestFormatFailuresIncludesRouteDetails(t *testing.T) {
	result := contract.TestResult{
		Failures: []contract.Failure{
			{
				Method:       "GET",
				Path:         "/api/v1/sys/users/{pk}",
				SamplePath:   "/api/v1/sys/users/1",
				StatusCode:   http.StatusInternalServerError,
				Error:        "unexpected response code 500, want 200",
				ResponseBody: `{"code":500,"msg":"内部服务器错误","data":null}`,
			},
		},
	}

	formatted := contract.FormatFailures(result)
	for _, want := range []string{
		"contract test failed: 1 failure(s)",
		"GET /api/v1/sys/users/{pk}",
		"sample: /api/v1/sys/users/1",
		"status: 500",
		"unexpected response code 500, want 200",
		`{"code":500`,
	} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("formatted failure missing %q:\n%s", want, formatted)
		}
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

func boolPtr(value bool) *bool {
	return &value
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

type assertRequestTransport struct {
	t            *testing.T
	wantMethod   string
	wantPath     string
	wantType     string
	wantBody     string
	status       int
	responseBody string
}

type authBootstrapTransport struct {
	t *testing.T
}

func (transport authBootstrapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport.t.Helper()
	switch req.URL.Path {
	case "/api/v1/auth/login":
		if req.Header.Get("Authorization") != "" {
			transport.t.Fatalf("login Authorization = %q, want empty", req.Header.Get("Authorization"))
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			transport.t.Fatalf("ReadAll(login body) error = %v", err)
		}
		if !strings.Contains(string(body), `"username":"admin"`) {
			transport.t.Fatalf("login body = %q, want admin credentials", body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"code":200,"msg":"请求成功","data":{"access_token":"contract-token"}}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	case "/api/v1/auth/codes":
		if req.Header.Get("Authorization") != "Bearer contract-token" {
			transport.t.Fatalf("Authorization = %q, want Bearer contract-token", req.Header.Get("Authorization"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"code":200,"msg":"请求成功","data":[]}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	default:
		transport.t.Fatalf("unexpected path %s", req.URL.Path)
		return nil, nil
	}
}

func (transport assertRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport.t.Helper()
	if req.Method != transport.wantMethod {
		transport.t.Fatalf("method = %s, want %s", req.Method, transport.wantMethod)
	}
	if req.URL.Path != transport.wantPath {
		transport.t.Fatalf("path = %s, want %s", req.URL.Path, transport.wantPath)
	}
	if got := req.Header.Get("Content-Type"); got != transport.wantType {
		transport.t.Fatalf("Content-Type = %q, want %q", got, transport.wantType)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		transport.t.Fatalf("ReadAll(request body) error = %v", err)
	}
	if string(body) != transport.wantBody {
		transport.t.Fatalf("body = %q, want %q", body, transport.wantBody)
	}
	return &http.Response{
		StatusCode: transport.status,
		Body:       io.NopCloser(strings.NewReader(transport.responseBody)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
