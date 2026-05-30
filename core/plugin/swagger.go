package plugin

type SwaggerFragment struct {
	PluginID string
	Paths    map[string]any
	Schemas  map[string]any
}
