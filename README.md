<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the project's front door on GitHub. A stranger decides
in ten seconds whether this project is real. Everything here must be true
today: badges only when true, status only with proof, quickstart commands
copied from the working build.

EXACT STRUCTURE: centered title in exact brand casing, one-line promise,
true badges, the standard-files link row, What it is, Quickstart, the
module map (the grid), Status, License. Full art direction lives in the
github-repo-standard skill (README standard + README-PATTERNS.md).

RULES:
- Brand casing is exact: AlphaVision, AlphaFlux, iTaK, Mimir,
  Patriot Pest Control in full.
- David's taste: black background with a bright color on top so things
  pop is his preferred look; black and red are his favorite scheme, not
  a law. Each project picks its own palette.
- No secrets, passwords, or tokens in this file, ever.
-->
<div align="center">

# Mimir

The agent that remembers. A local-first autonomous coding harness
that takes a project from plan to shipped code, with a knowledge
mind map that keeps what it learns.

![Language: Go](https://img.shields.io/badge/language-Go-00ADD8)
![Version](https://img.shields.io/badge/version-0.2.0--dev-lightgrey)
![Repo](https://img.shields.io/badge/repo-public-brightgreen)

</div>

**Plan:** [PLAN.md](PLAN.md) · **Tasks:** [TODO.md](TODO.md) · **Ideas:** [IDEA.md](IDEA.md) · **Build rules:** [docs/RULES.md](docs/RULES.md)

Deep docs: [docs/PROJECT-MASTER-PLAN.md](docs/PROJECT-MASTER-PLAN.md) ·
[docs/MASTER-TODO.md](docs/MASTER-TODO.md) ·
[docs/SKEPTICS-AUDIT.md](docs/SKEPTICS-AUDIT.md) (working code vs
blueprint, kept honest)

Named for Mimir of Norse mythology, the rememberer who guards the
well of wisdom beneath Yggdrasil. Domain: mimirmind.com.

## What it is

Mimir is a standalone autonomous coding harness written in Go with a
SurrealDB-backed memory (the Cortex). An agent loop works through a
fail-closed policy gate, routes decisions through the Huginn-Decider,
ingests repositories into a GitIngest Yggdrasil Mind Map it can query
across sessions, critiques its own work with an Epistemic Critic, and
runs a 7-phase SDLC inside each run. It serves a web UI on :8420 and
targets 12+ model providers behind one client interface.

The end goal is bigger than the current build: a plan-to-ship
harness for everyone, running David's 10-stage model (discover,
contract, repo, research, architecture, build, verify, review, ship,
handover), unleashed at plan approval with hard human gates at
verify and ship. The unattended run engine that carries that model
is built and tested on a local branch and lands per TODO Phase 2.

What is NOT built yet, stated plainly per docs/SKEPTICS-AUDIT.md:
the Rust (Tyr-Core), Mojo (Mjolnir-Compute), and Flutter
(Gjallarhorn-UI) stacks are blueprint in the main tree; engines ship
as precompiled binaries under LICENSE-ENGINES.md.

## Quickstart

Requires Go 1.26+ (the repo pins toolchain go1.26.9).

```powershell
# Build the single binary
go build -o bin/mimir.exe ./cmd/mimir

# Bootstrap the local decision model
.\bin\mimir.exe setup

# Ingest any local repo or GitHub URL into the Mind Map in the Cortex
.\bin\mimir.exe ingest .
.\bin\mimir.exe ingest https://github.com/David2024patton/Mimir

# List built-in providers and subscription tiers
.\bin\mimir.exe providers

# View or switch the 7-phase SDLC board (agile | waterfall | devops)
.\bin\mimir.exe sdlc

# Launch the interactive TUI
.\bin\mimir.exe

# Launch the TUI plus the live web UI on :8420
.\bin\mimir.exe --web
```

## The grid: module map

| Module | What it owns |
| :--- | :--- |
| cmd/mimir | CLI entry, subcommands, the mimir audit CLI |
| internal/agent | The agent loop, Norse personas, steering, compaction |
| internal/policy | The fail-closed policy gate |
| internal/sdlc | The 7-phase SDLC engine inside runs |
| internal/decision | Huginn-Decider, the routing decider |
| internal/ingest | GitIngest and the Yggdrasil Mind Map |
| internal/cortex | Memory: file store and SurrealDB store |
| internal/critic | The Epistemic Critic |
| internal/llm | Provider clients behind one interface |
| internal/config | The 12+ provider catalog and discovery |
| internal/runengine | The unattended run engine (scaffold branch) |
| internal/engines | Engine updates and signature verification |
| internal/relay | Phone pairing and the E2EE relay |
| internal/server | HTTP server, auth, web UI on :8420 |
| internal/swarm | The multi-agent factory |
| internal/tools | The tool set the agent can call |
| internal/tui | The terminal UI |
| internal/plugins | Unwired draft, superseded by the planned Wasm host |

## Status

Main is at 76e1be1 and the UI branch (engine/gjallarhorn-ui) at
9d1cae0. As of the 2026-10-19 audit_code pass: go build, go vet, and
gofmt clean, go test -race green on all packages, golangci-lint
clean, gitleaks gate run, verdict SHIP for the code.

One standing critical is open and it is not a code defect in the
current tree: finding F22, a mail credential sitting in this public
repo's git history. It closes only with a credential rotation at
Titan Mail and a history purge, both David's call, held for his
explicit go. Detail in TODO.md Phase 1.

Also open, tracked in TODO.md: the run-engine scaffold branch is
unmerged, RunUnattended has no production caller yet, engine update
signing is implemented but not active until a release key exists,
and persistent SurrealDB installs must set MIMIR_SURREAL_USER and
MIMIR_SURREAL_PASS or a restart locks the database.

## License

Open-core, per NOTICE and LICENSE-ENGINES.md. The Go harness control
plane in this repo is provided openly so developers and teams can
run, inspect, and integrate it. The precompiled engines in
engines/bin are proprietary to Alpha Flux LLC under the attribution
license in LICENSE-ENGINES.md.

Copyright (c) 2026 Alpha Flux LLC (https://alphaflux.net).
