package manifest

// DesiredManifest represents the desired state of the system
type DesiredManifest struct {
	SchemaVersion string `json:"schema_version" yaml:"schema_version"`
}

// ObservedManifest represents the current state of the system
type ObservedManifest struct {
	SchemaVersion string `json:"schema_version" yaml:"schema_version"`
}
