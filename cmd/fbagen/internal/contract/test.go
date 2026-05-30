package contract

import (
	"fmt"
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
		if err := probeRoute(client, opts.BaseURL, route); err != nil {
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

func probeRoute(client *http.Client, baseURL string, route Route) error {
	req, err := http.NewRequest(route.Method, strings.TrimRight(baseURL, "/")+route.Path, nil)
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
	return nil
}
