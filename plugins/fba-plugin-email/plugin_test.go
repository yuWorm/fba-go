package email_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	emailplugin "github.com/yuWorm/fba-plugin-email"
	"github.com/yuWorm/fba-plugin-email/service"
)

func TestEmailPluginRegistersPythonCompatibleRoute(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := emailplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	routes := ctx.Routes()
	if len(routes) != 1 {
		t.Fatalf("registered route count = %d, want 1", len(routes))
	}
	route := routes[0]
	if route.Method != "POST" || route.Path != "/emails/captcha" {
		t.Fatalf("route = %s %s, want POST /emails/captcha", route.Method, route.Path)
	}
	if !route.AuthRequired || route.Permission != "" {
		t.Fatalf("route auth/perms = auth:%v perm:%q, want auth true and empty permission", route.AuthRequired, route.Permission)
	}
}

func TestEmailCaptchaEndpointMatchesPythonRedisAndSenderBehavior(t *testing.T) {
	redisClient := newFakeEmailRedis()
	sender := &fakeCaptchaSender{}
	app := newEmailApp(t, redisClient, sender)

	resp, body := requestJSON(t, app, "POST", "/api/v1/emails/captcha", `{"recipients":"admin@example.com"}`, "192.0.2.10:1234")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	captchaKey := "fba:email:captcha:0.0.0.0"
	value, ok := redisClient.value(captchaKey)
	if !ok {
		t.Fatalf("email captcha redis key missing; values = %v", redisClient.values)
	}
	if len(value) != 6 {
		t.Fatalf("captcha code length = %d, want 6", len(value))
	}
	if redisClient.ttl(captchaKey) != 180*time.Second {
		t.Fatalf("captcha ttl = %s, want 180s", redisClient.ttl(captchaKey))
	}
	if len(sender.calls) != 1 {
		t.Fatalf("sender calls = %d, want 1", len(sender.calls))
	}
	if sender.calls[0].recipients[0] != "admin@example.com" || sender.calls[0].code != value || sender.calls[0].expireMinutes != 3 {
		t.Fatalf("sender call = %+v, redis code = %q", sender.calls[0], value)
	}

	resp, body = requestJSON(t, app, "POST", "/api/v1/emails/captcha", `{"recipients":["one@example.com","two@example.com"]}`, "198.51.100.20:1234")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	if len(sender.calls) != 2 || len(sender.calls[1].recipients) != 2 {
		t.Fatalf("sender calls = %+v, want second call with two recipients", sender.calls)
	}
}

func newEmailApp(t *testing.T, redisClient service.RedisClient, sender service.CaptchaSender) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler, ProxyHeader: fiber.HeaderXForwardedFor})
	container := di.New()
	if redisClient != nil {
		if err := container.Provide(func() service.RedisClient {
			return redisClient
		}); err != nil {
			t.Fatalf("Provide(redis) error = %v", err)
		}
	}
	if sender != nil {
		if err := container.Provide(func() service.CaptchaSender {
			return sender
		}); err != nil {
			t.Fatalf("Provide(sender) error = %v", err)
		}
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container, APIGroup: app.Group("/api/v1")})
	if err := emailplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	for _, route := range ctx.Routes() {
		ctx.APIGroup().Add([]string{route.Method}, route.Path, route.Handler)
	}
	return app
}

func requestJSON(t *testing.T, app *fiber.App, method string, path string, body string, remoteAddr string) (*http.Response, map[string]any) {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.RemoteAddr = remoteAddr
	req.Header.Set(fiber.HeaderXForwardedFor, strings.Split(remoteAddr, ":")[0])
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s error = %v", method, path, err)
	}
	defer resp.Body.Close()
	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp, decoded
}

func assertStatusOK(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func assertEnvelopeNil(t *testing.T, body map[string]any) {
	t.Helper()
	if body["code"] != float64(200) || body["msg"] != "请求成功" || body["data"] != nil {
		t.Fatalf("envelope = %v, want success null", body)
	}
}

type fakeCaptchaSender struct {
	calls []fakeCaptchaCall
}

type fakeCaptchaCall struct {
	recipients    []string
	code          string
	expireMinutes int
}

func (s *fakeCaptchaSender) SendCaptcha(_ context.Context, recipients []string, code string, expireMinutes int) error {
	s.calls = append(s.calls, fakeCaptchaCall{
		recipients:    append([]string(nil), recipients...),
		code:          code,
		expireMinutes: expireMinutes,
	})
	return nil
}

type fakeEmailRedis struct {
	values map[string]string
	ttls   map[string]time.Duration
}

func newFakeEmailRedis() *fakeEmailRedis {
	return &fakeEmailRedis{
		values: map[string]string{},
		ttls:   map[string]time.Duration{},
	}
}

func (r *fakeEmailRedis) Set(_ context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	r.values[key] = StringValue(value)
	r.ttls[key] = expiration
	return redis.NewStatusResult("OK", nil)
}

func (r *fakeEmailRedis) value(key string) (string, bool) {
	value, ok := r.values[key]
	return value, ok
}

func (r *fakeEmailRedis) ttl(key string) time.Duration {
	return r.ttls[key]
}

func StringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}
