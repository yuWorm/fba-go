package swagger

type DocumentInfo struct {
	Title   string
	Version string
}

type Fragment struct {
	PluginID string              `json:"plugin_id"`
	Paths    map[string]PathItem `json:"paths"`
	Schemas  map[string]any      `json:"schemas"`
}

type PathItem map[string]Operation

type Operation struct {
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Responses   map[string]any `json:"responses,omitempty"`
}

type Document struct {
	OpenAPI    string     `json:"openapi"`
	Info       Info       `json:"info"`
	Paths      Paths      `json:"paths"`
	Components Components `json:"components"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Paths map[string]PathItem

type Components struct {
	Schemas map[string]any `json:"schemas"`
}
