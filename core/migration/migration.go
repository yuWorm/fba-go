package migration

import "context"

type Migration struct {
	Scope    string
	Version  string
	Name     string
	Checksum string
	Up       func(context.Context) error
}
