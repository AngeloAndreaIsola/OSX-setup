package installer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"setupper/internal/manifest"
	"setupper/internal/planner"
	"setupper/internal/runner"
)

type Result struct {
	Step  planner.Step
	Error error
}

type ApplyOptions struct {
	FailFast bool
}

func ExecuteStep(ctx context.Context, r runner.Runner, step planner.Step) error {
	switch step.Action {
	case planner.ActionInstall:
		return runInstall(ctx, r, step.Resource)
	case planner.ActionRemove:
		return runRemove(ctx, r, step.Resource)
	}
	return fmt.Errorf("unknown action: %s", step.Action)
}

func runInstall(ctx context.Context, r runner.Runner, res manifest.Resource) error {
	switch res.Type {
	case "brew":
		_, err := r.Run(ctx, "brew", "install", res.Name)
		return err
	case "cask":
		_, err := r.Run(ctx, "brew", "install", "--cask", res.Name)
		return err
	case "mas":
		id := res.Options["id"]
		if id == "" {
			return fmt.Errorf("missing mas id for %s", res.Name)
		}
		_, err := r.Run(ctx, "mas", "install", id)
		return err
	case "vscode-extension":
		_, err := r.Run(ctx, "code", "--install-extension", res.Name)
		return err
	case "cursor-extension":
		_, err := r.Run(ctx, "cursor", "--install-extension", res.Name)
		return err
	case "npm":
		_, err := r.Run(ctx, "npm", "install", "-g", res.Name)
		return err
	case "pnpm":
		_, err := r.Run(ctx, "pnpm", "add", "-g", res.Name)
		return err
	case "cargo":
		_, err := r.Run(ctx, "cargo", "install", res.Name)
		return err
	case "pipx":
		_, err := r.Run(ctx, "pipx", "install", res.Name)
		return err
	case "uv":
		_, err := r.Run(ctx, "uv", "tool", "install", res.Name)
		return err
	case "go":
		_, err := r.Run(ctx, "go", "install", res.Name+"@latest")
		return err
	case "font":
		url := res.Options["url"]
		path := res.Options["path"]
		if url != "" {
			_, err := r.Run(ctx, "curl", "-fsSL", url, "-o", filepath.Join(res.Options["fontDir"], res.Name))
			if err != nil {
				// Fallback shell commands if fontDir option not present or needs home expansion
				_, err = r.Run(ctx, "/bin/sh", "-c", fmt.Sprintf("mkdir -p \"$HOME/Library/Fonts\" && curl -fsSL %q -o \"$HOME/Library/Fonts/%s\"", url, res.Name))
			}
			return err
		} else if path != "" {
			_, err := r.Run(ctx, "/bin/sh", "-c", fmt.Sprintf("mkdir -p \"$HOME/Library/Fonts\" && cp %q \"$HOME/Library/Fonts/%s\"", path, res.Name))
			return err
		}
		return fmt.Errorf("missing installation url or path for font %s", res.Name)
	case "git-config":
		val := res.Options["value"]
		_, err := r.Run(ctx, "git", "config", "--global", res.Name, val)
		return err
	case "macos-default":
		domain := res.Options["domain"]
		key := res.Options["key"]
		val := res.Options["value"]
		vtype := res.Options["value_type"]
		if domain == "" || key == "" || val == "" || vtype == "" {
			return fmt.Errorf("missing options for macos-default %s", res.Name)
		}
		// Build defaults write command
		// e.g., defaults write <domain> <key> -<type> <value>
		_, err := r.Run(ctx, "defaults", "write", domain, key, "-"+vtype, val)
		if err != nil {
			return err
		}
		// Restart affected process if needed
		if domain == "NSGlobalDomain" || strings.HasPrefix(domain, "com.apple.") {
			// Simplified: kill Finder, Dock, SystemUIServer
			proc := ""
			if strings.Contains(domain, "finder") {
				proc = "Finder"
			} else if strings.Contains(domain, "dock") {
				proc = "Dock"
			} else {
				proc = "SystemUIServer"
			}
			if proc != "" {
				_, _ = r.Run(ctx, "killall", proc)
			}
		}
		return nil
	case "dock-item":
		name := res.Name
		action := res.Options["action"]
		if action == "add" {
			_, err := r.Run(ctx, "dockutil", "--add", name)
			if err != nil {
				return err
			}
		} else if action == "remove" {
			_, err := r.Run(ctx, "dockutil", "--remove", name)
			if err != nil {
				return err
			}
		}
		// Restart Dock
		_, _ = r.Run(ctx, "killall", "Dock")
		return nil
	case "default-app":
		scheme := res.Name
		handler := res.Options["handler"]
		if scheme == "" || handler == "" {
			return fmt.Errorf("missing options for default-app %s", scheme)
		}
		_, err := r.Run(ctx, "duti", "-s", handler, scheme, "all")
		return err
	default:
		return fmt.Errorf("unsupported resource type: %s", res.Type)
	}
}

func runRemove(ctx context.Context, r runner.Runner, res manifest.Resource) error {
	switch res.Type {
	case "brew":
		_, err := r.Run(ctx, "brew", "uninstall", res.Name)
		return err
	case "cask":
		_, err := r.Run(ctx, "brew", "uninstall", "--cask", res.Name)
		return err
	case "mas":
		return fmt.Errorf("automatic removal of Mac App Store apps is not supported")
	case "vscode-extension":
		_, err := r.Run(ctx, "code", "--uninstall-extension", res.Name)
		return err
	case "cursor-extension":
		_, err := r.Run(ctx, "cursor", "--uninstall-extension", res.Name)
		return err
	case "npm":
		_, err := r.Run(ctx, "npm", "uninstall", "-g", res.Name)
		return err
	case "pnpm":
		_, err := r.Run(ctx, "pnpm", "remove", "-g", res.Name)
		return err
	case "cargo":
		_, err := r.Run(ctx, "cargo", "uninstall", res.Name)
		return err
	case "pipx":
		_, err := r.Run(ctx, "pipx", "uninstall", res.Name)
		return err
	case "uv":
		_, err := r.Run(ctx, "uv", "tool", "uninstall", res.Name)
		return err
	case "go":
		binaryName := res.Name
		if idx := strings.LastIndex(res.Name, "/"); idx != -1 {
			binaryName = res.Name[idx+1:]
		}
		_, err := r.Run(ctx, "/bin/sh", "-c", fmt.Sprintf("rm -f \"$HOME/go/bin/%s\"", binaryName))
		return err
	case "font":
		_, err := r.Run(ctx, "/bin/sh", "-c", fmt.Sprintf("rm -f \"$HOME/Library/Fonts/%s\"", res.Name))
		return err
	case "git-config":
		_, err := r.Run(ctx, "git", "config", "--global", "--unset", res.Name)
		return err
	default:
		return fmt.Errorf("unsupported resource type: %s", res.Type)
	}
}
