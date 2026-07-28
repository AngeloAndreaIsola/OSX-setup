# Setupper — Design Document

## Goal

A CLI that scans a developer workstation, recommends improvements, and
reproduces the environment on another machine.

## Core Principles

- Scan first, never modify — extends to the desired manifest too (see
  Manifest Schema Versioning below).
- The manifest is the source of truth.
- Every recommendation, adoption, and profile activation goes through
  the same explicit accept/reject review flow — nothing is applied
  silently.
- Detect capability, use it if present, degrade gracefully otherwise
  (applies to `mas`, secret backends, and any optional integration).
- Treat installation, configuration, authentication, and verification
  as separate lifecycle stages.

## Platform

macOS-only for v1. The package-manager and auth-provider seams are
built as thin interfaces so a Linux backend (apt/dnf) could be added
later without a rewrite — but only one implementation ships in v1.

## Workflow

```text
Init -> Scan -> Inventory -> Recommendations -> Plan -> Apply -> Verify
```

## Manifest Model

Two manifests, diffed like Terraform:

- **Desired manifest** — git-trackable, portable, hand-edited or
  built via `setupper init` / the planner. This is what the user
  intends their machine(s) to look like.
- **Observed manifest** — ephemeral, produced fresh by every `scan`.
  Never committed; represents current reality.

`plan` operates on the diff between the two.

### Resource Identity

Each resource's identity key is `type:name` (e.g. `brew:go`,
`vscode-extension:golang.go`). Version is a separate, optional
constraint field (e.g. `>=1.22`), not part of identity — so a routine
`brew upgrade` doesn't register as drift. Version/health checks live
in `verify`, not identity.

### Schema Versioning

The manifest carries an explicit `schema_version` field. If a binary
encounters a mismatched version, it **fails loudly** rather than
auto-migrating — auto-migration on load would mean an unrelated
command (e.g. `scan`) silently rewrites a git-tracked file. Upgrading
schema versions is a deliberate, reviewable action via `setupper
migrate`.

## Drift & Adoption

Resources found during `scan` that aren't in the desired manifest are
**unmanaged**, not errors. They're surfaced during
`inventory`/`recommend` as "adoptable," and the user chooses whether
to add them to desired state — interactively, during planning.

## Bootstrap: `setupper init`

On a new machine with no desired manifest yet, `init` runs a scan,
auto-suggests likely profiles based on scan evidence, and proposes a
starter desired manifest — through the same accept/reject flow as
every other recommendation. Not a blank file, not a bulk-promotion of
everything observed.

## Architecture

### Modules

- scanner
- inventory
- recommender
- planner
- installer
- configurator
- verifier
- exporters

### Resource Model

Every resource supports: Scan, Recommend, Plan, Install, Configure,
Authenticate (optional), Verify.

Resource types:

- Homebrew formula
- Homebrew cask
- App Store app
- Application *(manifest-only / manually declared — see Scanner)*
- CLI tool
- VS Code extension
- Cursor extension
- Font
- Shell configuration
- Git configuration
- Cloud credential
- Secret
- Service

## Scanner

Discovers: brew packages, App Store apps (via `mas`, if present),
VS Code/Cursor extensions, npm/pnpm/cargo/pipx/uv/go installs, fonts,
shells, git config.

- **`Application` type**: not auto-scanned. Scanning all of
  `/Applications` surfaces mostly noise (system apps, junk) with no
  reliable reinstall source for non-brew/non-App-Store apps. Users
  manually declare these for documentation/inventory purposes only.
- **App Store scanning**: opportunistic. If `mas` is installed, use
  it; if not, skip silently (or note that installing `mas` would
  enable this). No hard MVP dependency on a third-party tool with a
  history of breaking against Apple's API.

Output is a normalized observed manifest.

## Recommendation Engine

Profiles: general, go, node, python, devops, kubernetes, aws,
security, mobile, frontend, backend.

Profiles are shipped as **embedded YAML data** (via Go's `embed`
package), not hardcoded Go logic — easy to edit, diff, and eventually
let users extend with local profile files. No remote/updatable
profile fetching in v1 (avoids a network dependency and supply-chain
surface with no proven need yet).

Profile *activation* is auto-suggested from scan evidence but always
goes through the same review step as other recommendations — nothing
turns on silently.

## Planner

A **Bubble Tea TUI** built from day one (not deferred past MVP) — a
checklist/multi-select interface for accepting or rejecting
recommendations. `--yes` / non-interactive flag supported from the
start for scripted/CI use.

**Dependency ordering**: stage ordering (install → configure → auth
→ verify) plus a small set of hardcoded type-level parent-child rules
(extension requires its app; cloud credential requires its CLI tool).
No general-purpose dependency graph in v1 — the actual dependency
shapes here are a small fixed set, not a general problem.

## Apply

Stages: Install → Configure → Authenticate → Verify.

- **Failure handling**: best-effort/continue by default — a failure
  in one resource shouldn't block unrelated resources. `--fail-fast`
  flag available for stricter runs. Failures are scoped per-resource
  across stages (a failed install blocks later stages *for that
  resource*, not for others).
- **Execution model**: fully sequential per-resource for v1. Homebrew
  itself doesn't support safe concurrent `brew install` calls (global
  lock), so the "obvious" parallelization win isn't even available
  for the most common resource type. Parallelism is deferred to
  resource types later proven independently safe (VS Code
  extensions, npm global installs).
- Each stage gets its own Bubble Tea screen when run interactively
  (see TUI Scope below).

Examples of auth flows orchestrated (never storing credentials):
`gh auth login`, `aws configure sso` / `aws sso login`, Claude Code
login.

## Secrets & Credentials

Setupper is a **pure auth-flow orchestrator**. It never stores,
reads, or transmits secret values — it triggers native auth flows
(`gh`, `aws`, `claude`, etc.) and records only `authenticated:
true/false` plus which account/profile. The `Authenticate` stage
interface is designed so a pluggable secret backend (1Password CLI,
macOS Keychain) could later be *referenced* by a resource without
Setupper ever touching the value itself.

## Verification

Standalone and repeatable at any time (not just right after apply) —
supports the long-term "maintain complete workstations" goal.

- **Fast check** (default): structural — binary on PATH, file
  exists, token file present.
- **Deep check** (on demand): actually exercises the
  credential/service (e.g. pings AWS SSO rather than just checking a
  token file).

Verify output feeds the same drift-detection surface as `scan` — an
expired `gh` token is just another kind of drift.

## Configuration Stage — Dotfiles

Setupper does **not** own dotfile management. It manages a narrow,
clearly-marked, idempotent slice of shell config it directly cares
about — PATH additions, tool init lines (e.g. `eval "$(direnv hook
zsh)"`) — the same pattern nvm/pyenv/direnv already use for their own
install. If an existing dotfile manager (chezmoi, Stow) is detected
owning a file, Setupper defers rather than competing with it.

## TUI Scope

Per-command Bubble Tea screens: `scan`, `plan`, `apply`, and `verify`
each get their own interactive screen when run interactively. Not one
unified continuous TUI session — that would be a much bigger build
(session/navigation state across very different screens) and risks
becoming the whole product before the core engine is proven. Each
command remains independently invocable and scriptable.

## Exporters

MVP ships a **shell-script exporter** — a portable escape hatch that
reproduces the desired manifest on a machine without Setupper
installed at all. The exporter interface is built so other formats
(Brewfile, `.tool-versions`, devcontainer.json) are just additional
implementations added later, feeding the v1.0 plugin system goal.

## Testing Strategy

`installer`/`configurator` shell out to real system tools
(`brew`, `mas`, `gh`). Rather than a heavyweight abstraction, a thin
`Runner`-style seam wraps subprocess calls:

```go
type Runner interface {
    Run(cmd string, args ...string) (Output, error)
}
```

Unit tests fake this seam to test plan/apply logic without mutating a
real machine. Real subprocess execution is exercised only in a
separate, explicitly-tagged integration test suite.

## Config & Data Location

Single flat `~/.setupper/` directory for config, observed manifest
cache, and logs together — not XDG-split. Matches the convention of
peer tools (`aws`, `docker`) rather than Linux-style XDG separation,
which has no meaningful multi-user story to solve here anyway.

## Distribution

GitHub Releases binaries (checksummed, no toolchain assumed —
bootstrap-safe for a genuinely fresh machine) plus a Homebrew tap as
the primary install path for the target audience.

## Licensing

Private/closed for now. Open-source licensing deferred until (if)
the project is published.

## CLI

```text
setupper init
setupper scan
setupper inventory
setupper recommend
setupper plan
setupper apply
setupper verify
setupper migrate
setupper export
```

## Project Structure

```text
cmd/setupper
internal/
  scanner/
  inventory/
  recommender/
  planner/
  installer/
  configurator/
  verifier/
  exporter/
  profiles/
  runner/        # thin subprocess-execution seam
```

## Long-term Vision

Become a developer environment manager that can discover, improve,
reproduce, and maintain complete workstations — not just generate
setup scripts.

## Open Questions (not yet resolved)

- CLI framework choice (Cobra is the likely default, unconfirmed).
- Exact shell-exporter output format/structure.
- How VS Code vs. Cursor extensions are distinguished during scan,
  since they share a similar extensions-folder pattern.
