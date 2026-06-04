package contract

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TestOptions struct {
	BaseURL   string
	Contracts Contracts
	Client    *http.Client
	AuthToken string
}

type TestResult struct {
	Passed   bool      `json:"passed"`
	Failures []Failure `json:"failures"`
}

type authBootstrap struct {
	AccessToken   string
	RefreshCookie *http.Cookie
}

type Failure struct {
	Method       string `json:"method"`
	Path         string `json:"path"`
	SamplePath   string `json:"sample_path,omitempty"`
	StatusCode   int    `json:"status_code,omitempty"`
	Error        string `json:"error"`
	ResponseBody string `json:"response_body,omitempty"`
}

func Test(opts TestOptions) (TestResult, error) {
	if opts.BaseURL == "" {
		return TestResult{}, fmt.Errorf("base url is required")
	}
	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	authToken := opts.AuthToken
	var refreshCookie *http.Cookie
	adminBootstrapNeeded := opts.Contracts.API.BasePath != "" &&
		authToken == "" &&
		needsAdminToken(opts.Contracts.API.PriorityRoutes, opts.Contracts.API.NegativeRoutes)
	refreshBootstrapNeeded := opts.Contracts.API.BasePath != "" &&
		needsRefreshCookie(opts.Contracts.API.PriorityRoutes)
	if adminBootstrapNeeded {
		auth, err := bootstrapAuthSession(client, opts.BaseURL, opts.Contracts.API.BasePath)
		if err != nil {
			return TestResult{}, err
		}
		authToken = auth.AccessToken
		if !refreshBootstrapNeeded {
			refreshCookie = auth.RefreshCookie
		}
	}
	if refreshBootstrapNeeded {
		// Python create_new_token invalidates the access token tied to the refresh
		// cookie's session. Keep contract probing from invalidating the admin
		// bootstrap token used by the remaining authenticated route checks.
		auth, err := bootstrapAuthSession(client, opts.BaseURL, opts.Contracts.API.BasePath)
		if err != nil {
			return TestResult{}, err
		}
		refreshCookie = auth.RefreshCookie
	}
	limitedToken := ""
	if authToken != "" && hasLimitedNegativeRoute(opts.Contracts.API.NegativeRoutes) {
		token, err := bootstrapLimitedAuthToken(client, opts.BaseURL, opts.Contracts.API.BasePath, authToken)
		if err != nil {
			return TestResult{}, err
		}
		limitedToken = token
	}

	result := TestResult{Passed: true}
	for _, route := range opts.Contracts.API.PriorityRoutes {
		if failure := probeRoute(client, opts.BaseURL, route, opts.Contracts.Response, authToken, refreshCookie); failure != nil {
			result.Passed = false
			result.Failures = append(result.Failures, *failure)
		}
	}
	for _, route := range opts.Contracts.API.NegativeRoutes {
		routeAuthToken, applyAuth, err := negativeRouteAuthToken(client, opts.BaseURL, opts.Contracts.API.BasePath, route, authToken, &limitedToken)
		if err != nil {
			return TestResult{}, err
		}
		if failure := probeNegativeRoute(client, opts.BaseURL, route, opts.Contracts.Response, routeAuthToken, applyAuth); failure != nil {
			result.Passed = false
			result.Failures = append(result.Failures, *failure)
		}
	}
	return result, nil
}

func probeRoute(client *http.Client, baseURL string, route Route, response ResponseContract, authToken string, refreshCookie *http.Cookie) *Failure {
	probePath := route.Path
	if route.SamplePath != "" {
		probePath = route.SamplePath
	}
	var requestBody io.Reader
	if route.Request != nil && route.Request.Body != "" {
		requestBody = strings.NewReader(route.Request.Body)
	}
	req, err := http.NewRequest(route.Method, strings.TrimRight(baseURL, "/")+probePath, requestBody)
	if err != nil {
		return routeFailure(route, probePath, 0, err.Error(), "")
	}
	applyRequestSample(req, route.Request)
	if routeNeedsAuth(route) && authToken != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	if routeNeedsRefreshCookie(route) && refreshCookie != nil && req.Header.Get("Cookie") == "" {
		req.AddCookie(refreshCookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return routeFailure(route, probePath, 0, err.Error(), "")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return routeFailure(route, probePath, resp.StatusCode, err.Error(), "")
		}
		bodyPreview := previewResponseBody(body)
		return routeFailure(route, probePath, resp.StatusCode, fmt.Sprintf("route returned %d", resp.StatusCode), bodyPreview)
	}
	if !responseEnvelopeEnabled(route, response) {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return routeFailure(route, probePath, resp.StatusCode, err.Error(), "")
	}
	bodyPreview := previewResponseBody(body)
	if err := validateResponseEnvelope(body, route, response); err != nil {
		return routeFailure(route, probePath, resp.StatusCode, err.Error(), bodyPreview)
	}
	return nil
}

func probeNegativeRoute(client *http.Client, baseURL string, route Route, response ResponseContract, authToken string, applyAuth bool) *Failure {
	probePath := route.Path
	if route.SamplePath != "" {
		probePath = route.SamplePath
	}
	if route.ExpectedStatus == 0 {
		return routeFailure(route, probePath, 0, "negative route expected_status is required", "")
	}
	var requestBody io.Reader
	if route.Request != nil && route.Request.Body != "" {
		requestBody = strings.NewReader(route.Request.Body)
	}
	req, err := http.NewRequest(route.Method, strings.TrimRight(baseURL, "/")+probePath, requestBody)
	if err != nil {
		return routeFailure(route, probePath, 0, err.Error(), "")
	}
	applyRequestSample(req, route.Request)
	if applyAuth && authToken != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return routeFailure(route, probePath, 0, err.Error(), "")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return routeFailure(route, probePath, resp.StatusCode, err.Error(), "")
	}
	bodyPreview := previewResponseBody(body)
	if resp.StatusCode != route.ExpectedStatus {
		return routeFailure(route, probePath, resp.StatusCode, fmt.Sprintf("unexpected status %d, want %d", resp.StatusCode, route.ExpectedStatus), bodyPreview)
	}
	if err := validateErrorEnvelope(body, route, response, route.ExpectedStatus); err != nil {
		return routeFailure(route, probePath, resp.StatusCode, err.Error(), bodyPreview)
	}
	return nil
}

func needsAdminToken(priorityRoutes []Route, negativeRoutes []Route) bool {
	if hasAuthenticatedRoute(priorityRoutes) {
		return true
	}
	for _, route := range negativeRoutes {
		switch route.Auth {
		case "admin", "limited":
			return true
		}
	}
	return false
}

func hasAuthenticatedRoute(routes []Route) bool {
	for _, route := range routes {
		if routeNeedsAuth(route) {
			return true
		}
	}
	return false
}

func needsRefreshCookie(routes []Route) bool {
	for _, route := range routes {
		if routeNeedsRefreshCookie(route) {
			return true
		}
	}
	return false
}

func hasLimitedNegativeRoute(routes []Route) bool {
	for _, route := range routes {
		if route.Auth == "limited" {
			return true
		}
	}
	return false
}

func negativeRouteAuthToken(client *http.Client, baseURL string, basePath string, route Route, adminToken string, limitedToken *string) (string, bool, error) {
	switch route.Auth {
	case "", "none":
		return "", false, nil
	case "admin":
		if adminToken == "" {
			return "", false, fmt.Errorf("negative route %s %s requires admin auth token", route.Method, route.Path)
		}
		return adminToken, true, nil
	case "limited":
		if adminToken == "" {
			return "", false, fmt.Errorf("negative route %s %s requires admin auth token", route.Method, route.Path)
		}
		if *limitedToken == "" {
			token, err := bootstrapLimitedAuthToken(client, baseURL, basePath, adminToken)
			if err != nil {
				return "", false, err
			}
			*limitedToken = token
		}
		return *limitedToken, true, nil
	default:
		return "", false, fmt.Errorf("negative route %s %s has unsupported auth mode %q", route.Method, route.Path, route.Auth)
	}
}

func routeNeedsRefreshCookie(route Route) bool {
	return strings.HasSuffix(strings.TrimRight(route.Path, "/"), "/auth/refresh")
}

func routeNeedsAuth(route Route) bool {
	path := strings.TrimRight(route.Path, "/")
	switch {
	case strings.HasSuffix(path, "/auth/captcha"),
		strings.HasSuffix(path, "/auth/login"),
		strings.HasSuffix(path, "/auth/login/swagger"),
		strings.HasSuffix(path, "/auth/refresh"),
		strings.HasSuffix(path, "/auth/logout"):
		return false
	default:
		return true
	}
}

func bootstrapAuthSession(client *http.Client, baseURL string, basePath string) (authBootstrap, error) {
	if basePath == "" {
		basePath = "/api/v1"
	}
	body := strings.NewReader(`{"username":"admin","password":"admin","uuid":"fixture-captcha","captcha":"1234"}`)
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+strings.TrimRight(basePath, "/")+"/auth/login", body)
	if err != nil {
		return authBootstrap{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return authBootstrap{}, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return authBootstrap{}, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return authBootstrap{}, fmt.Errorf("contract auth bootstrap returned %d: %s", resp.StatusCode, previewResponseBody(payload))
	}
	var decoded struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return authBootstrap{}, fmt.Errorf("decode contract auth bootstrap: %w", err)
	}
	if decoded.Data.AccessToken == "" {
		return authBootstrap{}, fmt.Errorf("contract auth bootstrap missing access_token")
	}
	return authBootstrap{
		AccessToken:   decoded.Data.AccessToken,
		RefreshCookie: findCookie(resp.Cookies(), "fba_refresh_token"),
	}, nil
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func bootstrapLimitedAuthToken(client *http.Client, baseURL string, basePath string, adminToken string) (string, error) {
	if basePath == "" {
		basePath = "/api/v1"
	}
	root := strings.TrimRight(baseURL, "/") + strings.TrimRight(basePath, "/")
	createBody := `{"username":"contract_limited_user","password":"secret","nickname":"Contract Limited","email":null,"phone":null,"dept_id":1,"roles":[2]}`
	createPayload, err := doContractJSON(client, http.MethodPost, root+"/sys/users", createBody, adminToken)
	if err != nil {
		return "", err
	}
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createPayload, &created); err != nil {
		return "", fmt.Errorf("decode limited user bootstrap: %w", err)
	}
	if created.Data.ID == 0 {
		return "", fmt.Errorf("limited user bootstrap missing user id")
	}
	if _, err := doContractJSON(client, http.MethodPut, fmt.Sprintf("%s/sys/users/%d/permissions?type=staff", root, created.Data.ID), "", adminToken); err != nil {
		return "", err
	}
	loginPayload, err := doContractJSON(client, http.MethodPost, root+"/auth/login", `{"username":"contract_limited_user","password":"secret","uuid":"fixture-captcha","captcha":"1234"}`, "")
	if err != nil {
		return "", err
	}
	var login struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginPayload, &login); err != nil {
		return "", fmt.Errorf("decode limited login bootstrap: %w", err)
	}
	if login.Data.AccessToken == "" {
		return "", fmt.Errorf("limited login bootstrap missing access_token")
	}
	return login.Data.AccessToken, nil
}

func doContractJSON(client *http.Client, method string, url string, body string, bearerToken string) ([]byte, error) {
	var requestBody io.Reader
	if body != "" {
		requestBody = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("contract bootstrap %s %s returned %d: %s", method, url, resp.StatusCode, previewResponseBody(payload))
	}
	return payload, nil
}

func applyRequestSample(req *http.Request, sample *RequestSample) {
	if sample == nil {
		return
	}
	for key, value := range sample.Headers {
		req.Header.Set(key, value)
	}
	if sample.Body == "" {
		return
	}
	contentType := sample.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)
}

func validateResponseEnvelope(body []byte, route Route, response ResponseContract) error {
	if !responseEnvelopeEnabled(route, response) {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("response is not a JSON object envelope: %w", err)
	}
	for _, field := range response.Success.RequiredFields {
		if _, ok := payload[field]; !ok {
			return fmt.Errorf("missing response envelope field %q", field)
		}
	}
	expectedCode := response.Success.Code
	if route.ExpectedCode != 0 {
		expectedCode = route.ExpectedCode
	}
	if expectedCode != 0 && !jsonNumberEquals(payload["code"], expectedCode) {
		return fmt.Errorf("unexpected response code %v, want %d", payload["code"], expectedCode)
	}
	expectedMsg := response.Success.Msg
	if route.ExpectedMsg != "" {
		expectedMsg = route.ExpectedMsg
	}
	if expectedMsg != "" && payload["msg"] != expectedMsg {
		return fmt.Errorf("unexpected response msg %v, want %q", payload["msg"], expectedMsg)
	}
	return nil
}

func validateErrorEnvelope(body []byte, route Route, response ResponseContract, expectedStatus int) error {
	if !response.Error.Envelope {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("response is not a JSON object error envelope: %w", err)
	}
	for _, field := range response.Error.RequiredFields {
		if _, ok := payload[field]; !ok {
			return fmt.Errorf("missing response error envelope field %q", field)
		}
	}
	if expectedStatus != 0 && !jsonNumberEquals(payload["code"], expectedStatus) {
		return fmt.Errorf("unexpected response code %v, want %d", payload["code"], expectedStatus)
	}
	return nil
}

func responseEnvelopeEnabled(route Route, response ResponseContract) bool {
	if route.ResponseEnvelope != nil {
		return *route.ResponseEnvelope
	}
	return response.Success.Envelope
}

func routeFailure(route Route, samplePath string, statusCode int, message string, responseBody string) *Failure {
	return &Failure{
		Method:       route.Method,
		Path:         route.Path,
		SamplePath:   samplePath,
		StatusCode:   statusCode,
		Error:        message,
		ResponseBody: responseBody,
	}
}

func previewResponseBody(body []byte) string {
	const limit = 500
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "..."
}

func jsonNumberEquals(value any, want int) bool {
	number, ok := value.(float64)
	if !ok {
		return false
	}
	return number == float64(want)
}
