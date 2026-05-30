package contract

import (
	"fmt"
	"strings"
)

func FormatFailures(result TestResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "contract test failed: %d failure(s)", len(result.Failures))
	for i, failure := range result.Failures {
		fmt.Fprintf(&builder, "\n%d. %s %s", i+1, failure.Method, failure.Path)
		if failure.SamplePath != "" && failure.SamplePath != failure.Path {
			fmt.Fprintf(&builder, "\n   sample: %s", failure.SamplePath)
		}
		if failure.StatusCode != 0 {
			fmt.Fprintf(&builder, "\n   status: %d", failure.StatusCode)
		}
		if failure.Error != "" {
			fmt.Fprintf(&builder, "\n   error: %s", failure.Error)
		}
		if failure.ResponseBody != "" {
			fmt.Fprintf(&builder, "\n   body: %s", failure.ResponseBody)
		}
	}
	return builder.String()
}
