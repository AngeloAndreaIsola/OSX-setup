package manifest

// Resource represents an installed package, application, or configuration
type Resource struct {
	Type          string            `json:"type" yaml:"type"`
	Name          string            `json:"name" yaml:"name"`
	Version       string            `json:"version,omitempty" yaml:"version,omitempty"`
	Options       map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
	Authenticated bool              `json:"authenticated,omitempty" yaml:"authenticated,omitempty"`
	Account       string            `json:"account,omitempty" yaml:"account,omitempty"`
}

// DesiredManifest represents the desired state of the system
type DesiredManifest struct {
	SchemaVersion string              `json:"schema_version" yaml:"schema_version"`
	Resources     map[string]Resource `json:"resources" yaml:"resources"`
}

// ObservedManifest represents the current state of the system
type ObservedManifest struct {
	SchemaVersion string              `json:"schema_version" yaml:"schema_version"`
	Resources     map[string]Resource `json:"resources" yaml:"resources"`
}

// FormatKey creates a type:name identity key
func FormatKey(rtype, name string) string {
	return rtype + ":" + name
}
