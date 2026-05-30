package task

import (
	"fmt"

	"github.com/hibiken/asynq"
)

type Definition struct {
	Type    string
	Name    string
	Queue   string
	Handler asynq.Handler
}

type Registry struct {
	definitions []Definition
	byType      map[string]Definition
}

func NewRegistry() *Registry {
	return &Registry{byType: make(map[string]Definition)}
}

func (r *Registry) Add(definition Definition) error {
	if definition.Type == "" {
		return fmt.Errorf("task type is required")
	}
	if _, exists := r.byType[definition.Type]; exists {
		return fmt.Errorf("duplicate task type %q", definition.Type)
	}
	r.definitions = append(r.definitions, definition)
	r.byType[definition.Type] = definition
	return nil
}

func (r *Registry) All() []Definition {
	return append([]Definition(nil), r.definitions...)
}
