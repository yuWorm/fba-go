package model

type Plugin struct {
	ID          string
	Summary     string
	Version     string
	Description string
	Author      string
	Tags        []string
	Database    []string
	DependsOn   []string
	Enabled     bool
	BuiltIn     bool
}
