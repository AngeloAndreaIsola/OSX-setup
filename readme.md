# Setupper

Setupper is a declarative workstation configuration manager for macOS. It replaces error-prone, sequential setup scripts with an idempotent, scan-and-apply workflow inspired by modern infrastructure-as-code tools (like Terraform).

Setupper allows you to scan your current machine, choose recommendation profiles (e.g., Go, Node, DevOps), review a structured execution plan, safely apply changes, verify your workstation's state, and even export to standard bootstrap shell scripts.

---

## Features

- **Declarative Manifests**: Separates the **Desired state** (`desired.yaml` - what you want) from the **Observed state** (`observed.yaml` - what is currently installed).
- **Interactive Planner**: A Bubble Tea terminal UI checklist allows you to accept or reject recommendations and unmanaged resources before any changes are made.
- **Opportunistic Scanners**: Automatically scans for:
  - Homebrew formulas & casks
  - Mac App Store apps (`mas`)
  - VS Code & Cursor extensions
  - Global language packages (`npm`, `pnpm`, `cargo`, `pipx`, `uv`, `go` installs)
  - Custom macOS User Fonts
  - Git configurations (`git config`)
- **Safe & Sequential Apply Engine**: Runs through a structured lifecycle: `Install` → `Configure` → `Authenticate` → `Verify`.
- **Verify Engine**: Runs quick structural health checks or deep, service-level credential validations on demand.
- **Shell-script Exporter**: Generates a portable, standalone bootstrap bash/zsh script directly from your desired manifest.

---

## Installation

### Via Homebrew (Tap)

You can tap the repository and install Setupper:

```bash
COMING SOON
```

### Compiling from Source

Ensure you have Go installed (version 1.22 or higher), then run:

```bash
# Clone the repository
git clone https://github.com/angeloandreaisola/OSX-setup.git
cd OSX-setup

# Build the binary
go build -o setupper ./cmd/setupper

# Optionally make it globally accessible
mv setupper /usr/local/bin/
```

---

## Command Reference

| Command | Description |
|---|---|
| `setupper init` | Scans the system and initializes a starter desired manifest (`~/.setupper/config/desired.yaml`) |
| `setupper scan` | Scans current workstation packages and state, producing `observed.yaml` |
| `setupper inventory` | Displays the current observed state of the workstation |
| `setupper recommend` | Auto-detects system state and suggests profiles/resources to add |
| `setupper plan` | Generates and reviews an ordered execution plan to bridge the observed and desired state |
| `setupper apply` | Executes the generated plan (runs package installers and shell configurators) |
| `setupper verify` | Runs post-install verification and reports system drift |
| `setupper export` | Exports the desired manifest to a bootstrap shell script or Brewfile |
| `setupper migrate` | Validates and explicitly upgrades manifest schema versions |

---

## Configuration & Manifests

All configuration files, cached manifests, logs, and plans are stored in a single isolated home subdirectory: `~/.setupper/`.

### Directory Layout

```text
~/.setupper/
├── config/
│   └── desired.yaml      <- Your target configuration (commit this to git!)
├── cache/
│   ├── observed.yaml    <- The state detected on the last scan
│   └── plan.yaml        <- The generated plan waiting to be applied
└── logs/
    └── setupper.log     <- Logs of plan applications and errors
```

---

## Development & Testing

### Running Unit Tests

Unit tests use a faked subprocess runner, ensuring they do not modify your computer's actual configuration:

```bash
go test ./...
```

### Running E2E Integration Tests

Integration tests use the **real** subprocess runner (`SubprocessRunner`) inside a fully sandboxed environment (using isolated temporary home directories and Git config contexts). To run them:

```bash
go test -v -tags=integration ./internal/cli
```

---

## Distribution & CI/CD

When tags starting with `v` (e.g. `v1.0.0`) are pushed to GitHub, a GitHub Actions workflow automatically:
1. Compiles binary executables for macOS Intel (`amd64`) and Apple Silicon (`arm64`).
2. Creates `.tar.gz` packages for both architectures.
3. Computes the SHA256 checksums of the packages.
4. Generates a new GitHub Release containing the packages and checksums.
