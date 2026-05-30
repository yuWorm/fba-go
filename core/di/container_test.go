package di_test

import (
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/core/di"
)

type sampleService struct {
	Name string
}

func TestContainerProvideAndInvoke(t *testing.T) {
	container := di.New()

	if err := container.Provide(func() *sampleService {
		return &sampleService{Name: "service"}
	}); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	if err := container.Invoke(func(service *sampleService) {
		if service.Name != "service" {
			t.Fatalf("service.Name = %q, want service", service.Name)
		}
	}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
}

func TestContainerWrapsProvideErrors(t *testing.T) {
	container := di.New()

	err := container.Provide(42)
	if err == nil {
		t.Fatal("Provide() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "provide dependency") {
		t.Fatalf("Provide() error = %q, want wrapped context", err)
	}
}

func TestContainerWrapsInvokeErrors(t *testing.T) {
	container := di.New()

	err := container.Invoke(func(_ *sampleService) {})
	if err == nil {
		t.Fatal("Invoke() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invoke dependency") {
		t.Fatalf("Invoke() error = %q, want wrapped context", err)
	}
}
