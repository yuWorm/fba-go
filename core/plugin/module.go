package plugin

type Module interface {
	Meta() Meta
	Register(ctx Context) error
}

type Meta struct {
	ID                    string
	Name                  string
	Version               string
	Description           string
	Author                string
	Tags                  []string
	DependsOn             []Dependency
	Provides              []string
	AutoInjectDefault     bool
	PureDependencyDefault bool
}

type Dependency struct {
	ID       string
	Version  string
	Optional bool
}
