package errors_test

import (
	stderrors "errors"
	"testing"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
)

func TestAppErrorKeepsPublicAndInternalDetails(t *testing.T) {
	cause := stderrors.New("database rejected value")
	err := fbaerrors.New(400, 10001, "请求参数非法", cause)

	if err.HTTPStatus() != 400 {
		t.Fatalf("HTTPStatus() = %d, want 400", err.HTTPStatus())
	}
	if err.Code() != 10001 {
		t.Fatalf("Code() = %d, want 10001", err.Code())
	}
	if err.PublicMessage() != "请求参数非法" {
		t.Fatalf("PublicMessage() = %q", err.PublicMessage())
	}
	if !stderrors.Is(err, cause) {
		t.Fatal("AppError does not unwrap cause")
	}
}

func TestAppErrorDefaultsCompatibleCodeToHTTPStatus(t *testing.T) {
	err := fbaerrors.New(404, 0, "未找到", nil)

	if err.Code() != 404 {
		t.Fatalf("Code() = %d, want 404", err.Code())
	}
}
