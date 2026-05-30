package observability

import "context"

type Probe func(context.Context) error

type Readiness struct {
	probes map[string]Probe
}

type Result struct {
	Ready  bool                   `json:"ready"`
	Checks map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func NewReadiness() *Readiness {
	return &Readiness{probes: make(map[string]Probe)}
}

func (r *Readiness) Add(name string, probe Probe) {
	r.probes[name] = probe
}

func (r *Readiness) Check(ctx context.Context) Result {
	result := Result{
		Ready:  true,
		Checks: make(map[string]CheckResult, len(r.probes)),
	}
	for name, probe := range r.probes {
		if err := probe(ctx); err != nil {
			result.Ready = false
			result.Checks[name] = CheckResult{OK: false, Error: err.Error()}
			continue
		}
		result.Checks[name] = CheckResult{OK: true}
	}
	return result
}
