package plugin

type Mode string

const (
	ModeAuto           Mode = "auto"
	ModeDisabled       Mode = "disabled"
	ModePureDependency Mode = "pure_dependency"
)
