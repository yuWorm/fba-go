package plugin

import "fmt"

type Entry struct {
	Module Module
	Mode   Mode
}

type Registry struct {
	entries []Entry
	byID    map[string]Entry
}

func NewRegistry() *Registry {
	return &Registry{
		byID: make(map[string]Entry),
	}
}

func (r *Registry) Add(module Module, mode Mode) error {
	meta := module.Meta()
	if meta.ID == "" {
		return fmt.Errorf("plugin id is required")
	}
	if mode == "" {
		mode = ModeAuto
	}
	if _, exists := r.byID[meta.ID]; exists {
		return fmt.Errorf("duplicate plugin %q", meta.ID)
	}

	entry := Entry{Module: module, Mode: mode}
	r.entries = append(r.entries, entry)
	r.byID[meta.ID] = entry
	return nil
}

func (r *Registry) Resolve() ([]Entry, error) {
	resolved := make([]Entry, 0, len(r.entries))
	state := make(map[string]int, len(r.entries))

	var visit func(entry Entry, stack []string) error
	visit = func(entry Entry, stack []string) error {
		id := entry.Module.Meta().ID
		switch state[id] {
		case 1:
			return fmt.Errorf("plugin dependency cycle: %s -> %s", joinPath(stack), id)
		case 2:
			return nil
		}

		state[id] = 1
		stack = append(stack, id)
		for _, dep := range entry.Module.Meta().DependsOn {
			depEntry, ok := r.byID[dep.ID]
			if !ok {
				if dep.Optional {
					continue
				}
				return fmt.Errorf("plugin %s missing dependency %s", id, dep.ID)
			}
			if depEntry.Mode == ModeDisabled {
				if dep.Optional {
					continue
				}
				return fmt.Errorf("plugin %s has disabled dependency %s", id, dep.ID)
			}
			if err := visit(depEntry, stack); err != nil {
				return err
			}
		}

		state[id] = 2
		resolved = append(resolved, entry)
		return nil
	}

	for _, entry := range r.entries {
		if err := visit(entry, nil); err != nil {
			return nil, err
		}
	}
	return resolved, nil
}

func (r *Registry) RegisterAll(ctx Context) error {
	entries, err := r.Resolve()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Mode != ModeAuto {
			continue
		}
		if err := entry.Module.Register(ctx); err != nil {
			return fmt.Errorf("register plugin %s: %w", entry.Module.Meta().ID, err)
		}
	}
	return nil
}

func joinPath(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for _, item := range items[1:] {
		out += " -> " + item
	}
	return out
}
