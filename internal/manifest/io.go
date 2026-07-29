package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const currentSchemaVersion = "1.0"

func SaveObserved(path string, m *ObservedManifest) error {
	m.SchemaVersion = currentSchemaVersion
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadObserved(path string) (*ObservedManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m ObservedManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.SchemaVersion != currentSchemaVersion {
		return nil, fmt.Errorf("schema version mismatch: expected %s, got %s. Run 'setupper migrate' first", currentSchemaVersion, m.SchemaVersion)
	}
	return &m, nil
}

func SaveDesired(path string, m *DesiredManifest) error {
	m.SchemaVersion = currentSchemaVersion
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadDesired(path string) (*DesiredManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m DesiredManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.SchemaVersion != currentSchemaVersion {
		return nil, fmt.Errorf("schema version mismatch: expected %s, got %s. Run 'setupper migrate' first", currentSchemaVersion, m.SchemaVersion)
	}
	return &m, nil
}
