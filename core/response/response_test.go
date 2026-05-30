package response_test

import (
	"encoding/json"
	"testing"

	"github.com/yuWorm/fba-go/core/response"
)

func TestSuccessMarshalsCompatibleEnvelopeWithNullData(t *testing.T) {
	got, err := json.Marshal(response.Success[any](nil))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	const want = `{"code":200,"msg":"成功","data":null}`
	if string(got) != want {
		t.Fatalf("Success() JSON = %s, want %s", got, want)
	}
}

func TestErrorMarshalsTraceIDForFailures(t *testing.T) {
	got, err := json.Marshal(response.Error(400, "请求参数非法", "trace-1"))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	const want = `{"code":400,"msg":"请求参数非法","data":null,"trace_id":"trace-1"}`
	if string(got) != want {
		t.Fatalf("Error() JSON = %s, want %s", got, want)
	}
}
