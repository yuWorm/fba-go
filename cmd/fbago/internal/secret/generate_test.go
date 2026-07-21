package secret_test

import (
	"encoding/base64"
	"testing"

	fbsecret "github.com/yuWorm/fba-go/cmd/fbago/internal/secret"
)

func TestGenerateReturnsRequestedEntropyAsBase64URL(t *testing.T) {
	first, err := fbsecret.Generate(fbsecret.DefaultBytes)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := fbsecret.Generate(fbsecret.DefaultBytes)
	if err != nil {
		t.Fatalf("Generate() second error = %v", err)
	}
	if first == second {
		t.Fatal("Generate() returned the same secret twice")
	}
	raw, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if len(raw) != fbsecret.DefaultBytes {
		t.Fatalf("decoded bytes = %d, want %d", len(raw), fbsecret.DefaultBytes)
	}
}

func TestGenerateRejectsUnsafeSizes(t *testing.T) {
	for _, size := range []int{fbsecret.MinimumBytes - 1, fbsecret.MaximumBytes + 1} {
		if _, err := fbsecret.Generate(size); err == nil {
			t.Fatalf("Generate(%d) succeeded, want size error", size)
		}
	}
}
