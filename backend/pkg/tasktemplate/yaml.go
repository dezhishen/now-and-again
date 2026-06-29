package tasktemplate

// ─── YAML wire format shared by all providers ──────────────────────
//
// Every provider (builtin, http, …) produces YAML that conforms to this
// structure.  The provider parses its source, validates it, then converts
// each template entry into a TemplateRecord and calls
// TemplateStorage.UpsertTemplate().

// TemplateYAMLDocument is the top-level YAML structure.
type TemplateYAMLDocument struct {
	Version   int                  `yaml:"version"`
	Provider  TemplateYAMLProvider `yaml:"provider"`
	Templates []TemplateYAMLEntry  `yaml:"templates"`
}

// TemplateYAMLProvider describes the source provider metadata inside the YAML.
type TemplateYAMLProvider struct {
	Code        string `yaml:"code"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// TemplateYAMLEntry is a single template definition inside the YAML.
type TemplateYAMLEntry struct {
	Code         string                  `yaml:"code"`
	Name         string                  `yaml:"name"`
	Description  string                  `yaml:"description,omitempty"`
	Kind         string                  `yaml:"kind"`
	Icon         string                  `yaml:"icon,omitempty"`
	SortOrder    int                     `yaml:"sort_order,omitempty"`
	Enabled      *bool                   `yaml:"enabled,omitempty"`
	Parameters   []TemplateParameterYAML `yaml:"parameters,omitempty"`
	TaskDefaults map[string]any          `yaml:"task_defaults,omitempty"`
	ExtraSchema  map[string]any          `yaml:"extra_schema,omitempty"`
	Version      string                  `yaml:"version,omitempty"`
}

// TemplateParameterYAML defines a single dynamic parameter for the template.
type TemplateParameterYAML struct {
	Key         string         `yaml:"key"`
	Label       string         `yaml:"label"`
	Type        string         `yaml:"type"` // string, int, float, bool, select
	Description string         `yaml:"description,omitempty"`
	Required    bool           `yaml:"required,omitempty"`
	Default     any            `yaml:"default,omitempty"`
	Options     []SelectOption `yaml:"options,omitempty"`
	Placeholder string         `yaml:"placeholder,omitempty"`
}

// SelectOption is an option item for "select" type parameters.
type SelectOption struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}
