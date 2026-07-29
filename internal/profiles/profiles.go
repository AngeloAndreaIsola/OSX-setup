package profiles

import (
	_ "embed"
	"fmt"
	
	"gopkg.in/yaml.v3"
	"setupper/internal/manifest"
)

//go:embed data/default_profiles.yaml
var defaultProfilesYAML []byte

type Profile struct {
	ID          string              `yaml:"id"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Triggers    []string            `yaml:"triggers"`
	Resources   []manifest.Resource `yaml:"resources"`
}

type Collection struct {
	Profiles []Profile `yaml:"profiles"`
}

// LoadDefaults parses the embedded default profiles YAML
func LoadDefaults() ([]Profile, error) {
	var c Collection
	if err := yaml.Unmarshal(defaultProfilesYAML, &c); err != nil {
		return nil, fmt.Errorf("failed to parse default profiles: %w", err)
	}
	return c.Profiles, nil
}
