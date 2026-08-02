package scanner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

func (s *Scanner) scanApplications(ctx context.Context, obs *manifest.ObservedManifest) error {
	appsDir := "/Applications"
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), ".app") {
			name := entry.Name()
			// Check dedup against cask and mas resources
			if _, exists := obs.Resources[manifest.FormatKey("cask", name)]; exists {
				continue
			}
			// Read bundle identifier and version
			bundleID := ""
			version := ""
			// Use defaults read to get CFBundleIdentifier
			outID, errID := s.runner.Run(ctx, "defaults", "read", filepath.Join(appsDir, name, "Contents", "Info"), "CFBundleIdentifier")
			if errID == nil {
				bundleID = strings.TrimSpace(string(outID))
			}
			outVer, errVer := s.runner.Run(ctx, "defaults", "read", filepath.Join(appsDir, name, "Contents", "Info"), "CFBundleShortVersionString")
			if errVer == nil {
				version = strings.TrimSpace(string(outVer))
			}
			key := manifest.FormatKey("application", name)
			// If bundleID matches existing mas or cask, skip
			if bundleID != "" {
				if _, exists := obs.Resources[manifest.FormatKey("mas", bundleID)]; exists {
					continue
				}
				// Also check cask by bundleID if needed (not implemented)
			}
			opts := map[string]string{}
			if bundleID != "" {
				opts["bundle_id"] = bundleID
			}
			if version != "" {
				opts["version"] = version
			}
			obs.Resources[key] = manifest.Resource{Type: "application", Name: name, Options: opts}
		}
	}
	return nil
}

func (s *Scanner) Scan(ctx context.Context) (*manifest.ObservedManifest, error) {
	obs := &manifest.ObservedManifest{
		SchemaVersion: "1.0",
		Resources:     make(map[string]manifest.Resource),
	}

	// 1. Brew Formulas
	_ = s.scanBrewLeaves(ctx, obs)
	_ = s.scanBrewTaps(ctx, obs)

	// New scanning methods for macOS defaults, Dock items, and default apps
	_ = s.scanMacOSDefaults(ctx, obs)
	_ = s.scanDockItems(ctx, obs)
	_ = s.scanDefaultApps(ctx, obs)

	// 2. Brew Casks
	_ = s.scanBrewCasks(ctx, obs)


	// 3. Mac App Store (mas)
	_ = s.scanMas(ctx, obs)

	// 4. VS Code & Cursor Extensions
	_ = s.scanVSCodeExtensions(ctx, obs)
	_ = s.scanCursorExtensions(ctx, obs)

	// 5. Language package ecosystems
	_ = s.scanNpmPackages(ctx, obs)
	_ = s.scanPnpmPackages(ctx, obs)
	_ = s.scanCargoPackages(ctx, obs)
	_ = s.scanPipxPackages(ctx, obs)
	_ = s.scanUvPackages(ctx, obs)
	_ = s.scanGoPackages(ctx, obs)

	// 6. Fonts
	_ = s.scanFonts(ctx, obs)

	// 7. Git Configurations
	_ = s.scanGitConfigs(ctx, obs)

	// 8. Applications
	_ = s.scanApplications(ctx, obs)

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

func (s *Scanner) scanBrewTaps(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "brew", "tap")
	if err != nil {
		return err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		key := manifest.FormatKey("brew-tap", name)
		obs.Resources[key] = manifest.Resource{Type: "brew-tap", Name: name}
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

func parseExtensionDirName(dirName string) string {
	idx := strings.LastIndex(dirName, "-")
	if idx == -1 {
		return dirName
	}
	suffix := dirName[idx+1:]
	if len(suffix) > 0 && (suffix[0] >= '0' && suffix[0] <= '9') {
		return dirName[:idx]
	}
	return dirName
}

func (s *Scanner) scanVSCodeExtensions(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "code", "--list-extensions")
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			name := strings.TrimSpace(line)
			if name == "" {
				continue
			}
			key := manifest.FormatKey("vscode-extension", name)
			obs.Resources[key] = manifest.Resource{
				Type: "vscode-extension",
				Name: name,
			}
		}
		return nil
	}

	// Fallback to directory scraping
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dirPath := filepath.Join(homeDir, ".vscode", "extensions")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			name := parseExtensionDirName(f.Name())
			key := manifest.FormatKey("vscode-extension", name)
			obs.Resources[key] = manifest.Resource{
				Type: "vscode-extension",
				Name: name,
			}
		}
	}
	return nil
}

func (s *Scanner) scanCursorExtensions(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "cursor", "--list-extensions")
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			name := strings.TrimSpace(line)
			if name == "" {
				continue
			}
			key := manifest.FormatKey("cursor-extension", name)
			obs.Resources[key] = manifest.Resource{
				Type: "cursor-extension",
				Name: name,
			}
		}
		return nil
	}

	// Fallback to directory scraping
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dirPath := filepath.Join(homeDir, ".cursor", "extensions")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			name := parseExtensionDirName(f.Name())
			key := manifest.FormatKey("cursor-extension", name)
			obs.Resources[key] = manifest.Resource{
				Type: "cursor-extension",
				Name: name,
			}
		}
	}
	return nil
}

func (s *Scanner) scanNpmPackages(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "npm", "list", "-g", "--depth=0", "--json")
	if err != nil {
		return err
	}

	var data struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(out, &data); err != nil {
		return err
	}

	for pkgName := range data.Dependencies {
		key := manifest.FormatKey("npm", pkgName)
		obs.Resources[key] = manifest.Resource{
			Type: "npm",
			Name: pkgName,
		}
	}
	return nil
}

func (s *Scanner) scanPnpmPackages(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "pnpm", "list", "-g", "--depth=0", "--json")
	if err != nil {
		return err
	}

	var data []struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(out, &data); err != nil {
		return err
	}

	for _, item := range data {
		for pkgName := range item.Dependencies {
			key := manifest.FormatKey("pnpm", pkgName)
			obs.Resources[key] = manifest.Resource{
				Type: "pnpm",
				Name: pkgName,
			}
		}
	}
	return nil
}

func (s *Scanner) scanCargoPackages(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "cargo", "install", "--list")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		if strings.Contains(line, " v") && strings.HasSuffix(line, ":") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				name := parts[0]
				key := manifest.FormatKey("cargo", name)
				obs.Resources[key] = manifest.Resource{
					Type: "cargo",
					Name: name,
				}
			}
		}
	}
	return nil
}

func (s *Scanner) scanPipxPackages(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "pipx", "list", "--json")
	if err != nil {
		return err
	}

	var data struct {
		Venvs map[string]interface{} `json:"venvs"`
	}

	if err := json.Unmarshal(out, &data); err != nil {
		return err
	}

	for name := range data.Venvs {
		key := manifest.FormatKey("pipx", name)
		obs.Resources[key] = manifest.Resource{
			Type: "pipx",
			Name: name,
		}
	}
	return nil
}

func (s *Scanner) scanUvPackages(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "uv", "tool", "list")
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
		if len(parts) > 0 {
			name := parts[0]
			key := manifest.FormatKey("uv", name)
			obs.Resources[key] = manifest.Resource{
				Type: "uv",
				Name: name,
			}
		}
	}
	return nil
}

func (s *Scanner) scanGoPackages(ctx context.Context, obs *manifest.ObservedManifest) error {
	// Determine the GOPATH / GOBIN
	goBinDir := ""
	outGopath, err := s.runner.Run(ctx, "go", "env", "GOPATH")
	if err == nil {
		gopath := strings.TrimSpace(string(outGopath))
		if gopath != "" {
			goBinDir = filepath.Join(gopath, "bin")
		}
	}

	if goBinDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		goBinDir = filepath.Join(homeDir, "go", "bin")
	}

	files, err := os.ReadDir(goBinDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		filePath := filepath.Join(goBinDir, f.Name())
		// Run go version -m <file> to find import path
		vOut, vErr := s.runner.Run(ctx, "go", "version", "-m", filePath)
		if vErr == nil {
			lines := strings.Split(string(vOut), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "path ") || strings.HasPrefix(line, "path\t") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						importPath := parts[1]
						key := manifest.FormatKey("go", importPath)
						obs.Resources[key] = manifest.Resource{
							Type: "go",
							Name: importPath,
						}
					}
				}
			}
		}
	}
	return nil
}

func (s *Scanner) scanFonts(ctx context.Context, obs *manifest.ObservedManifest) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fontDirs := []string{
		filepath.Join(homeDir, "Library", "Fonts"),
		"/Library/Fonts",
		"/System/Library/Fonts",
	}

	for _, dir := range fontDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext == ".ttf" || ext == ".otf" || ext == ".woff" || ext == ".woff2" {
				key := manifest.FormatKey("font", f.Name())
				obs.Resources[key] = manifest.Resource{
					Type: "font",
					Name: f.Name(),
				}
			}
		}
	}
	return nil
}

func (s *Scanner) scanGitConfigs(ctx context.Context, obs *manifest.ObservedManifest) error {
	out, err := s.runner.Run(ctx, "git", "config", "--list", "--global")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, "=")
		if idx != -1 {
			key := line[:idx]
			val := line[idx+1:]
			rKey := manifest.FormatKey("git-config", key)
			obs.Resources[rKey] = manifest.Resource{
				Type: "git-config",
				Name: key,
				Options: map[string]string{
					"value": val,
				},
			}
		}
	}
	return nil
}

func (s *Scanner) scanMacOSDefaults(ctx context.Context, obs *manifest.ObservedManifest) error {
    // Define a curated list of defaults to scan
    defaults := []struct {
        domain string
        key    string
    }{
        {"NSGlobalDomain", "AppleShowAllExtensions"},
        {"NSGlobalDomain", "ApplePressAndHoldEnabled"},
        {"com.apple.dock", "autohide"},
        {"com.apple.finder", "ShowPathbar"},
        {"com.apple.screencapture", "type"},
    }
    for _, d := range defaults {
        out, err := s.runner.Run(ctx, "defaults", "read", d.domain, d.key)
        if err != nil {
            // If the key does not exist, skip it
            continue
        }
        val := strings.TrimSpace(string(out))
        key := manifest.FormatKey("macos-default", d.domain+":"+d.key)
        obs.Resources[key] = manifest.Resource{
            Type: "macos-default",
            Name: d.key,
            Options: map[string]string{
                "domain": d.domain,
                "key": d.key,
                "value": val,
                "value_type": "string",
            },
        }
    }
    return nil
}

func (s *Scanner) scanDockItems(ctx context.Context, obs *manifest.ObservedManifest) error {
    out, err := s.runner.Run(ctx, "dockutil", "--list")
    if err != nil {
        // dockutil may not be installed; skip scanning dock items
        return nil
    }
    lines := strings.Split(string(out), "\n")
    for _, line := range lines {
        name := strings.TrimSpace(line)
        if name == "" {
            continue
        }
        key := manifest.FormatKey("dock-item", name)
        obs.Resources[key] = manifest.Resource{
            Type: "dock-item",
            Name: name,
            Options: map[string]string{
                "action": "present",
            },
        }
    }
    return nil
}

func (s *Scanner) scanDefaultApps(ctx context.Context, obs *manifest.ObservedManifest) error {
    // Scan a set of common URL schemes for their default handlers
    schemes := []string{"http", "https", "mailto", "ftp"}
    for _, scheme := range schemes {
        out, err := s.runner.Run(ctx, "duti", "-x", scheme)
        if err != nil {
            continue
        }
        lines := strings.Split(string(out), "\n")
        if len(lines) == 0 {
            continue
        }
        handler := strings.TrimSpace(lines[0])
        if handler == "" {
            continue
        }
        key := manifest.FormatKey("default-app", scheme)
        obs.Resources[key] = manifest.Resource{
            Type: "default-app",
            Name: scheme,
            Options: map[string]string{
                "handler": handler,
            },
        }
    }
    return nil
}
