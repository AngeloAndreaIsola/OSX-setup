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
- [ ] `general` profile (wget, htop, tree, jq, curl, coreutils,
      ripgrep, fd, bat — common developer utilities)
- [ ] `devops` profile (docker, terraform, ansible, packer,
      vagrant triggers/resources)
- [ ] `kubernetes` profile (kubectl, helm, k9s, kubectx, stern,
      kustomize triggers/resources)
- [ ] `aws` profile (awscli, aws-vault, session-manager-plugin
      triggers/resources)
- [ ] `security` profile (gnupg, age, sops, 1password-cli
      triggers/resources)
- [ ] `mobile` profile (flutter, cocoapods, fastlane,
      android-studio, xcode-select triggers/resources)
- [ ] `frontend` profile (prettier, eslint, vite, sass,
      tailwindcss triggers/resources)
- [ ] `backend` profile (protobuf, grpcurl, redis, nginx,
      postgresql triggers/resources)

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

- [x] VS Code / Cursor extension scanning (resolve the
      how-do-we-tell-them-apart open question first)
- [x] npm/pnpm/cargo/pipx/uv/go install scanning
- [x] Font scanning
- [x] Git configuration scanning

## Milestone 10 — Polish & Distribution

- [x] GitHub Releases build pipeline with checksums
- [ ] Homebrew tap
- [x] End-to-end integration test suite (real `Runner`, gated
      separately from unit tests)
- [x] Documentation pass

## Milestone 11 - Application Scanner

Goal: be able to also scan the application folder

- [x] `scanner`: `/Applications` directory auto-scan — enumerate
      `.app` bundles and register them as `application` resources
- [x] `/Applications` scan: extract bundle identifier and version
      from `Info.plist` (`CFBundleIdentifier`, `CFBundleShortVersionString`)
- [x] `/Applications` scan: deduplicate against already-discovered
      `cask` and `mas` resources (skip apps that are already tracked
      by brew or the App Store)
- [x] `/Applications` scan: skip system/Apple-provided apps
      (e.g. Safari, Xcode command-line tools stubs) via a
      configurable exclusion list
- [x] Tests for `/Applications` scanner using the fake `Runner`
      and a temporary directory tree
- [x] `scanner`: `brew tap` to list active taps — register each
      as a `brew-tap` resource
- [x] `installer`: `brew tap <name>` for install,
      `brew untap <name>` for remove
- [x] Brew-tap dependency inference: during scan, cross-reference
      each cask/formula against `brew info --json=v2` to record
      which tap it originates from; during plan, automatically
      order `brew-tap` resources before any formula/cask that
      depends on them

## Milestone 12 — macOS System Preferences

Goal: scan, apply, and verify macOS system-level settings so a
fresh machine feels identical to the old one.

### macOS Defaults

- [x] New resource type `macos-default` with fields: `domain`,
   `key`, `value`, `value_type` (bool / int / float / string)
- [x] `scanner`: read a curated set of known-useful defaults
   categories — Finder, Dock, NSGlobalDomain, screencapture,
   Mail, keyboard, trackpad
- [x] `installer`: `defaults write <domain> <key> -<type> <val>`
   for apply; `defaults delete <domain> <key>` for remove;
   `killall` the affected process when needed (Finder, Dock,
   SystemUIServer)
- [x] `verifier`: `defaults read <domain> <key>` and compare
   against desired value

- [x] Curated default set seeded from `osx_setup.sh`:
       AppleShowAllExtensions, Dock autohide + timing, Finder
       status/path bar, screenshot location/format, save-to-disk
       default, keyboard full-access mode, show-recents in Dock

### Dock Layout

- [x] New resource type `dock-item` with fields: `name`,
       `action` (add / remove), `position` (optional index)
- [x] `scanner`: read current Dock contents via
       `dockutil --list` (skip gracefully if absent)
- [x] `installer`: `dockutil --add` / `dockutil --remove` +
       `killall Dock`
- [ ] Dependency: `dock-item` resources depend on `brew:dockutil`

### Default Applications

- [x] New resource type `default-app` with fields:
      `scheme` (http / mailto / public.html / etc.),
      `handler` (bundle ID, e.g. `com.google.Chrome`)
- [x] `scanner`: read current handler for common schemes via
      Launch Services or `duti`
- [x] `installer`: set handler via `duti` or
      `open -a <app> --args --make-default-browser` for http/https

### Keyboard, Language & Input

- [x] `macos-default` entries for keyboard settings:
      KeyRepeat, InitialKeyRepeat,
      ApplePressAndHoldEnabled
- [x] `macos-default` entries for locale/language:
      AppleLanguages, AppleLocale, AppleMeasurementUnits
- [x] `macos-default` entries for trackpad:
      Clicking (tap to click), TrackpadThreeFingerDrag,
      com.apple.trackpad.scaling

## Deferred / Not in Scope for v1

- Linux support (interfaces kept open, no second backend implemented)
- Unified single-session TUI shell
- Parallel apply execution
- Remote/updatable recommendation profiles
- Full dotfile ownership
- Open-source licensing decision

## Still Unresolved

- Exact shell-exporter output format
