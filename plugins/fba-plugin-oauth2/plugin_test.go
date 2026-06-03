package oauth2_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	adminplugin "github.com/yuWorm/fba-plugin-admin"
	adminrepo "github.com/yuWorm/fba-plugin-admin/repo"
	oauth2plugin "github.com/yuWorm/fba-plugin-oauth2"
	"github.com/yuWorm/fba-plugin-oauth2/repo"
	"github.com/yuWorm/fba-plugin-oauth2/service"
)

func TestOAuth2PluginRegistersPythonCompatibleRoutes(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := oauth2plugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got := make(map[string]plugin.Route)
	for _, route := range ctx.Routes() {
		got[route.Method+" "+route.Path] = route
	}
	want := map[string]bool{
		"GET /oauth2/github":          false,
		"GET /oauth2/github/callback": false,
		"GET /oauth2/google":          false,
		"GET /oauth2/google/callback": false,
		"GET /oauth2/me/bindings":     true,
		"GET /oauth2/me/binding":      true,
		"DELETE /oauth2/me/unbinding": true,
	}
	if len(got) != len(want) {
		t.Fatalf("registered route count = %d, want %d; routes = %v", len(got), len(want), routeKeys(got))
	}
	for key, authRequired := range want {
		route, ok := got[key]
		if !ok {
			t.Fatalf("route %s not registered; registered routes: %v", key, routeKeys(got))
		}
		if route.AuthRequired != authRequired {
			t.Fatalf("%s AuthRequired = %v, want %v", key, route.AuthRequired, authRequired)
		}
	}
}

func TestOAuth2AuthURLBindingCallbackAndUnbindingMatchPython(t *testing.T) {
	stateStore := service.NewMemoryStateStore()
	app := newOAuth2App(t, stateStore)

	resp, body := requestJSON(t, app, http.MethodGet, "/api/v1/oauth2/github", "", "")
	assertStatus(t, resp, http.StatusOK)
	githubURL := envelopeString(t, body)
	githubState := queryParam(t, githubURL, "state")
	if githubState == "" {
		t.Fatalf("github auth url missing state: %s", githubURL)
	}
	payload, ok := stateStore.Peek(githubState)
	if !ok || payload.Type != service.AuthTypeLogin {
		t.Fatalf("github state payload = %+v, ok=%v; want login", payload, ok)
	}

	token := loginAdmin(t, app)
	resp, body = requestJSON(t, app, http.MethodGet, "/api/v1/oauth2/me/bindings", "", token)
	assertStatus(t, resp, http.StatusOK)
	if items := envelopeSlice(t, body); len(items) != 0 {
		t.Fatalf("initial bindings = %v, want empty", items)
	}

	resp, body = requestJSON(t, app, http.MethodGet, "/api/v1/oauth2/me/binding?source=Google", "", token)
	assertStatus(t, resp, http.StatusOK)
	googleURL := envelopeString(t, body)
	googleState := queryParam(t, googleURL, "state")
	payload, ok = stateStore.Peek(googleState)
	if !ok || payload.Type != service.AuthTypeBinding || payload.UserID != 1 {
		t.Fatalf("google binding state payload = %+v, ok=%v; want binding user 1", payload, ok)
	}

	resp, _ = requestRaw(t, app, http.MethodGet, "/api/v1/oauth2/google/callback?code=fixture-code&state="+googleState, "", "")
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("binding callback status = %d, want 302", resp.StatusCode)
	}
	if location := resp.Header.Get("Location"); location != "http://localhost:5173/profile" {
		t.Fatalf("binding callback location = %q, want profile", location)
	}

	resp, body = requestJSON(t, app, http.MethodGet, "/api/v1/oauth2/me/bindings", "", token)
	assertStatus(t, resp, http.StatusOK)
	bindings := envelopeSlice(t, body)
	if len(bindings) != 1 || bindings[0] != service.SourceGoogle {
		t.Fatalf("bindings = %v, want [Google]", bindings)
	}

	resp, body = requestJSON(t, app, http.MethodDelete, "/api/v1/oauth2/me/unbinding?source=Google", "", token)
	assertStatus(t, resp, http.StatusOK)
	if body["data"] != nil {
		t.Fatalf("unbinding data = %v, want nil", body["data"])
	}
}

func TestOAuth2LoginCallbackIssuesAdminCompatibleToken(t *testing.T) {
	stateStore := service.NewMemoryStateStore()
	app := newOAuth2App(t, stateStore)

	_, body := requestJSON(t, app, http.MethodGet, "/api/v1/oauth2/github", "", "")
	state := queryParam(t, envelopeString(t, body), "state")
	resp, _ := requestRaw(t, app, http.MethodGet, "/api/v1/oauth2/github/callback?code=fixture-code&state="+state, "", "")
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("login callback status = %d, want 302", resp.StatusCode)
	}
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, "http://localhost:5173/oauth2/callback?") {
		t.Fatalf("login callback location = %q, want frontend callback", location)
	}
	accessToken := queryParam(t, location, "access_token")
	sessionUUID := queryParam(t, location, "session_uuid")
	if accessToken == "" || sessionUUID == "" {
		t.Fatalf("login callback location missing token/session: %s", location)
	}

	resp, body = requestJSON(t, app, http.MethodGet, "/api/v1/sys/users/me", "", accessToken)
	assertStatus(t, resp, http.StatusOK)
	current := envelopeMap(t, body)
	if current["username"] != "fixture_github" {
		t.Fatalf("current user username = %v, want fixture_github", current["username"])
	}
}

func TestAdminSeedMarksOAuth2AndUnsupportedCodeGenerator(t *testing.T) {
	seed := adminrepo.SeedData()
	plugins := make(map[string]bool)
	var codeGeneratorDescription string
	for _, item := range seed.Plugins {
		plugins[item.ID] = item.Enabled
		if item.ID == "code_generator" {
			codeGeneratorDescription = item.Description
		}
	}
	if !plugins["oauth2"] {
		t.Fatal("oauth2 plugin is missing or disabled in admin seed")
	}
	if plugins["code_generator"] {
		t.Fatal("code_generator plugin should be present but disabled")
	}
	if !strings.Contains(codeGeneratorDescription, "Go 版本不实现") {
		t.Fatalf("code_generator description = %q, want unsupported marker", codeGeneratorDescription)
	}
}

func newOAuth2App(t *testing.T, stateStore service.StateStore) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	container := di.New()
	adminRepository := adminrepo.NewMemoryRepository(adminrepo.SeedData())
	socialRepository := repo.NewMemoryRepository(repo.SeedData())
	if err := container.Provide(func() adminrepo.Repository {
		return adminRepository
	}); err != nil {
		t.Fatalf("Provide(admin repo) error = %v", err)
	}
	if err := container.Provide(func() repo.Repository {
		return socialRepository
	}); err != nil {
		t.Fatalf("Provide(oauth2 repo) error = %v", err)
	}
	if stateStore != nil {
		if err := container.Provide(func() service.StateStore {
			return stateStore
		}); err != nil {
			t.Fatalf("Provide(state store) error = %v", err)
		}
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container, APIGroup: app.Group("/api/v1")})
	if err := adminplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("admin Register() error = %v", err)
	}
	if err := oauth2plugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("oauth2 Register() error = %v", err)
	}
	plugin.MountRoutes(ctx.APIGroup(), ctx.Routes(), plugin.WithContainer(ctx.Container()))
	return app
}

func loginAdmin(t *testing.T, app *fiber.App) string {
	t.Helper()
	_, body := requestJSON(t, app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin","uuid":"fixture-captcha","captcha":"1234"}`, "")
	data := envelopeMap(t, body)
	token, ok := data["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("login access_token = %v, want non-empty", data["access_token"])
	}
	return token
}

func requestJSON(t *testing.T, app *fiber.App, method string, path string, body string, token string) (*http.Response, map[string]any) {
	t.Helper()
	resp, raw := requestRaw(t, app, method, path, body, token)
	defer resp.Body.Close()
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode %s %s response: %v; body=%s", method, path, err, string(raw))
	}
	return resp, decoded
}

func requestRaw(t *testing.T, app *fiber.App, method string, path string, body string, token string) (*http.Response, []byte) {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s error = %v", method, path, err)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, path, err)
	}
	return resp, raw
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d", resp.StatusCode, want)
	}
}

func envelopeString(t *testing.T, body map[string]any) string {
	t.Helper()
	value, ok := body["data"].(string)
	if !ok || value == "" {
		t.Fatalf("response data = %v, want string", body["data"])
	}
	return value
}

func envelopeMap(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	value, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("response data = %T %v, want object", body["data"], body["data"])
	}
	return value
}

func envelopeSlice(t *testing.T, body map[string]any) []any {
	t.Helper()
	value, ok := body["data"].([]any)
	if !ok {
		t.Fatalf("response data = %T %v, want slice", body["data"], body["data"])
	}
	return value
}

func queryParam(t *testing.T, rawURL string, key string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url %q: %v", rawURL, err)
	}
	return parsed.Query().Get(key)
}

func routeKeys(routes map[string]plugin.Route) []string {
	keys := make([]string, 0, len(routes))
	for key := range routes {
		keys = append(keys, key)
	}
	return keys
}
