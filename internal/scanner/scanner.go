package scanner

import (
	"context"
	"strings"

	"setupper/internal/manifest"
	"setupper/internal/runner"
)

type Scanner struct {
	runner runner.Runner
}

func New(r runner.Runner) *Scanner {
	return &Scanner{runner: r}
}

func (s *Scanner) Scan(ctx context.Context) (*manifest.ObservedManifest, error) {
	obs := &manifest.ObservedManifest{
		SchemaVersion: "1.0",
		Resources:     make(map[string]manifest.Resource),
	}

	// 1. Brew Formulas (skip errors gracefully if not installed)
	_ = s.scanBrewLeaves(ctx, obs)

	// 2. Brew Casks
	_ = s.scanBrewCasks(ctx, obs)

	// 3. Mac App Store (mas)
	_ = s.scanMas(ctx, obs)

	return obs, nil
}

func (s *Scanner) scanBrewLeaves(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "brew", "leaves")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		
		key := manifest.FormatKey("brew", name)
		obs.Resources[key] = manifest.Resource{
			Type: "brew",
			Name: name,
		}
	}
	return nil
}

func (s *Scanner) scanBrewCasks(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "brew", "list", "--cask", "-1")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		
		key := manifest.FormatKey("cask", name)
		obs.Resources[key] = manifest.Resource{
			Type: "cask",
			Name: name,
		}
	}
	return nil
}

func (s *Scanner) scanMas(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "mas", "list")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			id := parts[0]
			name := strings.Join(parts[1:], " ")
			
			key := manifest.FormatKey("mas", id)
			obs.Resources[key] = manifest.Resource{
				Type: "mas",
				Name: name,
				Options: map[string]string{"id": id},
			}
		}
	}
	return nil
}
