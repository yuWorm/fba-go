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
	Method string `json:"method"`
	Path   string `json:"path"`
	Error  string `json:"error"`
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
		if err := probeRoute(client, opts.BaseURL, route, opts.Contracts.Response); err != nil {
			result.Passed = false
			result.Failures = append(result.Failures, Failure{
				Method: route.Method,
				Path:   route.Path,
				Error:  err.Error(),
			})
		}
	}
	return result, nil
}

func probeRoute(client *http.Client, baseURL string, route Route, response ResponseContract) error {
	probePath := route.Path
	if route.SamplePath != "" {
		probePath = route.SamplePath
	}
	req, err := http.NewRequest(route.Method, strings.TrimRight(baseURL, "/")+probePath, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		return fmt.Errorf("route returned %d", resp.StatusCode)
	}
	return validateResponseEnvelope(resp, route, response)
}

func validateResponseEnvelope(resp *http.Response, route Route, response ResponseContract) error {
	envelope := response.Success.Envelope
	if route.ResponseEnvelope != nil {
		envelope = *route.ResponseEnvelope
	}
	if !envelope {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
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

func jsonNumberEquals(value any, want int) bool {
	number, ok := value.(float64)
	if !ok {
		return false
	}
	return number == float64(want)
}
