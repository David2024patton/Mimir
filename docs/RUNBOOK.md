<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the operations book for this project, at docs/RUNBOOK.md.
How to build it, run it, deploy it, verify it, and where everything lives,
so anyone can operate the project without tribal knowledge.

EXACT STRUCTURE:
- "## Build": the real commands, copied from the working build.
- "## Run locally": how to start it and what you should see.
- "## Deploy": where it is hosted, how a deploy happens, how to roll back.
- "## Verify": the checks that prove it works after a change or deploy.
- "## Where things live": hosts, services, dashboards, and WHERE
  credentials are stored (names of vault entries or connectors only).
- "## Gotchas": the traps already learned the hard way, each with its fix.

RULES:
- NEVER a credential value in this file: no passwords, keys, or tokens.
  Pointers to secure storage only.
- Commands are copied from what actually works; if a command changes,
  this file changes in the same commit.
- No secrets, passwords, or tokens in this file, ever.
-->
# RUNBOOK: Mimir

## Build
Requires Go 1.26+ (the repo pins toolchain go1.26.9 in go.mod).

```powershell
# Build the single binary (Windows form, from README.md)
go build -o bin/mimir.exe ./cmd/mimir

# Build everything, no output binary
go build ./...
```

The gate suite, all green before a push counts as done (docs/RULES.md):

```bash
gofmt -l .            # must print nothing
go vet ./...
go test -race ./...
golangci-lint run     # zero issues
gitleaks detect       # secret scan over the tree
```

CI (.github/workflows/ci.yml) runs gofmt check, go vet, and
go test -race on push and pull request, with a non-blocking
govulncheck job. A green server-side CI run is still unverified
(TODO T-4.2); local gate runs are the current proof.

## Run locally

```powershell
# Bootstrap the local decision model
.\bin\mimir.exe setup

# Ingest a local repo or a GitHub URL into the Mind Map in the Cortex
.\bin\mimir.exe ingest .
.\bin\mimir.exe ingest https://github.com/David2024patton/Mimir

# List built-in providers and subscription tiers
.\bin\mimir.exe providers

# View or switch the 7-phase SDLC board (agile | waterfall | devops)
.\bin\mimir.exe sdlc

# Interactive TUI
.\bin\mimir.exe

# TUI plus the live web UI on :8420
.\bin\mimir.exe --web
```

Docker shapes in the repo root: docker-compose.yml for local use and
docker-compose.prod.yml for a persistent install, with the Cortex on
SurrealDB.

## Deploy
Mimir is local-first; there is no hosted production deploy to run.
Distribution is the repo itself plus the built-in updater:

```powershell
mimir update --check   # show what changed, install nothing
mimir update           # git fast-forward main, rebuild bin/mimir, swap in changed engines
```

- The harness update uses git pull --ff-only, so it never merges or
  overwrites local edits. If it refuses, commit or stash first.
- Engines are compared by SHA-256 against engines/bin/manifest.json
  and verified before anything is replaced; a failed check leaves the
  old binary untouched.
- Maintainers ship an engine change with scripts/publish-engines.ps1
  from main (see docs/ENGINES.md); engine source lives on engine
  branches and never lands on main.
- Rollback: return the checkout to the previous known-good commit
  and rebuild. Engine binaries roll back the same way, since every
  engine swap is manifest-verified.

## Verify
Three tiers (docs/TESTING-POLICY.md); a feature is not done until the
unit tier and a live test have both run.

1. Unit tests, offline, always green: `go test ./...` (the gate form
   adds -race, see Build above).
2. Live test against a real model: `go test ./internal/agent -run TestLiveRun -v`
   with MIMIR_LIVE_BASE_URL set (and MIMIR_LIVE_MODEL, plus
   MIMIR_LIVE_API_KEY for a cloud endpoint). It skips automatically
   when MIMIR_LIVE_BASE_URL is unset, so an offline run that skips it
   has NOT proven the live path.
3. CLI smoke: `go run ./cmd/mimir "prompt"` against a real provider,
   or `mimir trace "prompt"` to watch each tool call.

Cross-run memory checks: `go test ./internal/agent -run TestCrossRunMemory -v`
proves memory survives a restart on the file-backed Cortex.

## Where things live
- Source of truth: github.com/David2024patton/Mimir (public), default
  branch main; UI work on branch engine/gjallarhorn-ui.
- Web UI: served by the binary on port 8420 when run with --web.
- Cortex memory: file store at .mimir/cortex.json by default
  (override with the MIMIR_CORTEX and MIMIR_HOME environment
  variables), or SurrealDB in the docker shapes.
- Project record: TODO.md for work state, STATUS.md for the current
  snapshot, docs/SKEPTICS-AUDIT.md for working code versus blueprint,
  docs/MASTER-TODO.md and docs/PROJECT-MASTER-PLAN.md for the deep
  engineering detail.
- Credential storage pointers, names only, never values:
  - SurrealDB credentials for a persistent install: the
    MIMIR_SURREAL_USER and MIMIR_SURREAL_PASS environment variables in
    that deployment's environment.
  - Engine signing: the ed25519 release key is generated and held
    offline by David; installs pin it by the MIMIR_ENGINE_PUBKEY
    environment variable name.
  - Forge OAuth apps (GitHub, GitLab, Bitbucket, Gitea/Forgejo):
    registration is stalled until David submits current passwords
    through the secure card flow; no app credentials exist in this
    repo.
  - Live test credentials: supplied per run through the
    MIMIR_LIVE_API_KEY environment variable; never logged, never
    committed.
  - The F22 mail credential: rotation happens at Titan Mail by
    David; nothing about its value belongs in this repo.

## Gotchas
- Restart lockout on persistent SurrealDB installs: set
  MIMIR_SURREAL_USER and MIMIR_SURREAL_PASS in the persistent
  environment BEFORE a restart. Without them, a restart generates a
  new per-run sidecar credential and the existing database cannot be
  opened (fix: set both variables to the install's credentials, then
  restart; TODO T-4.1).
- Unauthenticated non-loopback bind is refused on purpose: serve
  will not bind a non-loopback address with no auth backend unless
  MIMIR_ALLOW_UNAUTHENTICATED=1 is set explicitly (fix: configure an
  auth backend, or set the variable only where that exposure is
  intended).
- Engine signing is implemented but NOT active: updates verify
  signatures only once a release key exists and MIMIR_ENGINE_PUBKEY
  is pinned (fix: TODO Phase 3, David generates the key offline).
- RunUnattended has no production caller yet: the unattended path is
  the scaffold on branch run-engine-scaffold, not a main-branch
  command (fix: TODO T-2.1 and T-2.2).
- The F22 history purge, when David approves it, breaks his Windows
  clone and both branches need re-syncing afterward. Do not run it
  without his explicit go (fix: none; it is held, TODO Phase 1).
- A live test that skips proves nothing about the live path: check
  that MIMIR_LIVE_BASE_URL was actually set before counting it
  (docs/TESTING-POLICY.md).
