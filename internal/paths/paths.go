package paths

import (
	"os"
	"path/filepath"
)

// EnsureDirs sets up the ~/.setupper directory and its subdirectories (config, cache, logs)
func EnsureDirs() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(home, ".setupper")
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "config"),
		filepath.Join(baseDir, "cache"),
		filepath.Join(baseDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetBaseDir returns the base directory for setupper
func GetBaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".setupper"), nil
}
