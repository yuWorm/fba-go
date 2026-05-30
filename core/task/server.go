package task

import "github.com/hibiken/asynq"

func BuildServeMux(registry *Registry) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	for _, definition := range registry.All() {
		if definition.Handler != nil {
			mux.Handle(definition.Type, definition.Handler)
		}
	}
	return mux
}
