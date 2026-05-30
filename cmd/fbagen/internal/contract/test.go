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
	if authToken == "" && opts.Contracts.API.BasePath != "" && hasAuthenticatedRoute(opts.Contracts.API.PriorityRoutes) {
		token, err := bootstrapAuthToken(client, opts.BaseURL, opts.Contracts.API.BasePath)
		if err != nil {
			return TestResult{}, err
		}
		authToken = token
	}

	result := TestResult{Passed: true}
	for _, route := range opts.Contracts.API.PriorityRoutes {
		if failure := probeRoute(client, opts.BaseURL, route, opts.Contracts.Response, authToken); failure != nil {
			result.Passed = false
			result.Failures = append(result.Failures, *failure)
		}
	}
	return result, nil
}

func probeRoute(client *http.Client, baseURL string, route Route, response ResponseContract, authToken string) *Failure {
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

func hasAuthenticatedRoute(routes []Route) bool {
	for _, route := range routes {
		if routeNeedsAuth(route) {
			return true
		}
	}
	return false
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

func bootstrapAuthToken(client *http.Client, baseURL string, basePath string) (string, error) {
	if basePath == "" {
		basePath = "/api/v1"
	}
	body := strings.NewReader(`{"username":"admin","password":"admin","uuid":"fixture-captcha","captcha":"1234"}`)
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+strings.TrimRight(basePath, "/")+"/auth/login", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("contract auth bootstrap returned %d: %s", resp.StatusCode, previewResponseBody(payload))
	}
	var decoded struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return "", fmt.Errorf("decode contract auth bootstrap: %w", err)
	}
	if decoded.Data.AccessToken == "" {
		return "", fmt.Errorf("contract auth bootstrap missing access_token")
	}
	return decoded.Data.AccessToken, nil
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
	if response.Success.Code != 0 && !jsonNumberEquals(payload["code"], response.Success.Code) {
		return fmt.Errorf("unexpected response code %v, want %d", payload["code"], response.Success.Code)
	}
	if response.Success.Msg != "" && payload["msg"] != response.Success.Msg {
		return fmt.Errorf("unexpected response msg %v, want %q", payload["msg"], response.Success.Msg)
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
