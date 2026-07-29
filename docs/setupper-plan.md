# Setupper — Implementation Plan

This plan sequences the design decisions in `setupper-design.md` into
buildable milestones. Each milestone is meant to produce something
runnable, not just scaffolding.

## Milestone 0 — Foundations

Goal: nothing user-facing yet, but the seams every later milestone
depends on are in place.

- [x] `cmd/setupper` skeleton, CLI framework wired up
- [x] `internal/runner`: thin `Runner` interface + real subprocess
      implementation (shells out to arbitrary commands)
- [x] Fake `Runner` for tests
- [x] Manifest types (desired + observed) with `schema_version` field
- [x] `~/.setupper/` directory setup (config, cache, logs)

## Milestone 1 — Scan & Inventory (MVP core)

Goal: `setupper scan` produces a real observed manifest for brew.

- [x] `scanner`: brew formulas + casks via `Runner`
- [x] `scanner`: App Store apps via `mas`, if present — skip
      gracefully if absent
- [x] Normalized observed manifest output (`type:name` identity keys)
- [x] `setupper inventory` — read-only display of observed state
- [x] Manually-declared `Application` entries supported in the
      desired manifest schema (no auto-scan)

## Milestone 2 — Desired Manifest & Diffing

Goal: the two-manifest model actually works end to end.

- [x] Desired manifest schema + YAML (de)serialization
- [x] Diff engine: observed vs. desired → unmanaged / missing /
      matching
- [x] `setupper init`: scan + profile auto-suggestion (stub profiles
      OK for now) → starter desired manifest, reviewed before write
- [x] `setupper migrate`: explicit schema-version upgrade command
      (fail loudly on mismatch elsewhere)

## Milestone 3 — Recommendation Profiles

Goal: profiles suggest tools based on scan evidence.

- [x] Embedded YAML profile definitions (`go`, `node`, `python`, ...)
- [x] `recommender`: match scan evidence → suggested profiles +
      suggested resources
- [x] `setupper recommend` — plain output first (no TUI required yet)

## Milestone 4 — Planner TUI

Goal: the first real interactive surface.

- [x] Bubble Tea checklist screen: accept/reject recommendations and
      unmanaged/adoptable resources
- [x] `--yes` / non-interactive mode producing the same plan output
- [x] `planner`: hardcoded type-level dependency rules
      (extension→app, credential→cli-tool) applied when ordering the
      plan
- [x] `setupper plan` writes a concrete, ordered execution plan

## Milestone 5 — Apply Engine

Goal: the plan actually does something to the machine.

- [x] `installer`: brew install/upgrade via `Runner`, sequential
- [x] Best-effort/continue semantics + `--fail-fast` flag
- [x] Per-resource stage scoping (a failed install blocks later
      stages for that resource only)
- [x] Bubble Tea progress screen for `apply`
- [x] `configurator`: narrow idempotent shell config blocks (PATH,
      tool init lines), with dotfile-manager detection/deferral

## Milestone 6 — Authentication & Secrets

Goal: auth flows are orchestrated, never stored.

- [x] `Authenticate` stage: trigger `gh auth login`,
      `aws sso login` / `aws configure sso`, Claude Code login
- [x] Manifest records `authenticated: true/false` + account/profile
      only — no token/secret persistence
- [x] Interface designed to allow a future pluggable secret-backend
      reference (1Password CLI, Keychain) without implementing one
      yet

## Milestone 7 — Verification

Goal: standalone, repeatable health checks.

- [x] `verifier`: fast structural checks (binary on PATH, file
      exists, token file present)
- [x] Deep check mode (on demand): exercise the credential/service
      for real
- [x] Verify output feeds into the same drift surface as scan
- [x] Bubble Tea results screen for `verify`

## Milestone 8 — Exporters

Goal: reproduce a machine without Setupper installed.

- [x] Shell-script exporter: desired manifest → bootstrap script
- [x] Exporter interface generalized enough to add Brewfile /
      devcontainer.json later

## Milestone 9 — Extensions & Language Ecosystems

- [ ] VS Code / Cursor extension scanning (resolve the
      how-do-we-tell-them-apart open question first)
- [ ] npm/pnpm/cargo/pipx/uv/go install scanning
- [ ] Font scanning
- [ ] Git configuration scanning

## Milestone 10 — Polish & Distribution

- [ ] GitHub Releases build pipeline with checksums
- [ ] Homebrew tap
- [ ] End-to-end integration test suite (real `Runner`, gated
      separately from unit tests)
- [ ] Documentation pass

## Deferred / Not in Scope for v1

- Linux support (interfaces kept open, no second backend implemented)
- Unified single-session TUI shell
- Parallel apply execution
- Remote/updatable recommendation profiles
- Full dotfile ownership
- `/Applications` auto-scanning
- Open-source licensing decision

## Still Unresolved

- CLI framework choice (Cobra likely, unconfirmed)
- Exact shell-exporter output format
- VS Code vs. Cursor extension disambiguation during scan
