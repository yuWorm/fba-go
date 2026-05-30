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

	result := TestResult{Passed: true}
	for _, route := range opts.Contracts.API.PriorityRoutes {
		if failure := probeRoute(client, opts.BaseURL, route, opts.Contracts.Response); failure != nil {
			result.Passed = false
			result.Failures = append(result.Failures, *failure)
		}
	}
	return result, nil
}

func probeRoute(client *http.Client, baseURL string, route Route, response ResponseContract) *Failure {
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
