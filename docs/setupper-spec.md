# Setupper Vision

## Goal

Create a CLI that captures a developer workstation, recommends
improvements, and reproduces it on another machine.

## Core Principles

-   Scan first, never modify.
-   Build an inventory as the source of truth.
-   Let users choose what to install.
-   Recommend industry-standard tools based on profiles.
-   Treat installation, configuration, authentication, and verification
    as separate lifecycle stages.

## Workflow

``` text
Scan -> Inventory -> Recommendations -> Plan -> Apply -> Verify
```

## Architecture

### Modules

-   scanner
-   inventory
-   recommender
-   planner
-   installer
-   configurator
-   verifier
-   exporters

### Resource Model

Every resource supports:

-   Scan
-   Recommend
-   Plan
-   Install
-   Configure
-   Authenticate (optional)
-   Verify

Resource types include:

-   Homebrew formula
-   Homebrew cask
-   App Store app
-   Application
-   CLI tool
-   VS Code extension
-   Cursor extension
-   Font
-   Shell configuration
-   Git configuration
-   Cloud credential
-   Secret
-   Service

## Scanner

Discover: - Brew packages - Applications - App Store apps - VS
Code/Cursor extensions - npm, pnpm, cargo, pipx, uv, go installs -
Fonts - Shells - Git config

Output a normalized manifest.

## Manifest

The manifest is the canonical representation.

``` yaml
installed:
  - brew: go
  - cask: docker
selected: []
recommendations: []
```

## Recommendation Engine

Profiles:

-   general
-   go
-   node
-   python
-   devops
-   kubernetes
-   aws
-   security
-   mobile
-   frontend
-   backend

Profiles suggest tools without forcing installation.

## Planner

Users interactively accept or reject recommendations and create an
execution plan.

## Apply

Stages: 1. Install packages 2. Configure tools 3. Authenticate 4. Verify

Examples: - gh auth login - aws configure sso / aws sso login - Claude
Code login/configuration

## Verification

Check: - Installed - Configured - Authenticated - Healthy

## CLI

``` text
setupper scan
setupper inventory
setupper recommend
setupper plan
setupper apply
setupper verify
setupper export
```

## Project Structure

``` text
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
```

## Roadmap

### MVP

-   Brew scan
-   App scan
-   Manifest
-   Shell exporter

### v0.2

-   Recommendation profiles
-   Interactive planner

### v0.3

-   Apply engine
-   GitHub/AWS authentication tasks

### v0.4

-   VS Code/Cursor extensions
-   Language ecosystems

### v1.0

-   Full verification
-   Multi-export support
-   Plugin system

## Long-term Vision

Become a developer environment manager that can discover, improve,
reproduce, and maintain complete workstations---not just generate setup
scripts.
