# PLAN: Mimir

The whole plan for Mimir, step by step. Each step names the TODO.md
phase it maps to. When the plan changes, this file changes first,
then TODO.md follows. Facts only: what is decided is stated, what is
open is named as an open question. The deep engineering plan lives in
docs/PROJECT-MASTER-PLAN.md and the granular engineering checklist
in docs/MASTER-TODO.md; docs/SKEPTICS-AUDIT.md separates working
code from blueprint, and this file follows the skeptics audit where
they differ.

## What it is
Mimir is David's standalone autonomous coding harness: a local-first
agent that takes a project from plan to shipped code without him
watching. It is GitHub-integrated, runs coding runs start to finish
unattended, and keeps a typed knowledge mind map so uploaded material
integrates into the system instead of sitting static. The end goal
is a plan-to-ship harness for everyone, not a private tool for David
only. Mimir is its own project; it is NOT part of AlphaFlux.

The product model is David's 10 stages: discover, contract, repo,
research, architecture, build, verify, review, ship, handover. The
human-guided portion ends when the plan is approved; execution is
unleashed from there, with hard human gates at verify and ship.
Inside a run today, the shipped engine executes a 7-phase SDLC; the
10-stage run graph is designed and lands in Step 7 below.

## Architecture
The grid: the Go control plane on main, one module per grid slot.

| Module | What it owns |
| :--- | :--- |
| cmd/mimir | CLI entry, subcommands, the mimir audit CLI |
| internal/agent | The agent loop, personas, steering, compaction |
| internal/policy | The fail-closed policy gate every action passes |
| internal/sdlc | The 7-phase SDLC engine inside runs |
| internal/decision | Huginn-Decider, the routing decider |
| internal/ingest | GitIngest and the Yggdrasil Mind Map |
| internal/cortex | Memory: file store and SurrealDB store |
| internal/critic | The Epistemic Critic |
| internal/llm | Provider clients behind one interface (12+ provider catalog in internal/config) |
| internal/runengine | The unattended run engine (on the scaffold branch until Step 3) |
| internal/engines | Engine updates: manifest, signature verification, platform binary resolution |
| internal/relay | Phone pairing and the E2EE relay |
| internal/server | HTTP server, auth, web UI on :8420 |
| internal/swarm | The multi-agent factory (merge only via Factory.ApproveMerge) |
| internal/tools | The tool set the agent can call |
| internal/tui | The terminal UI |
| internal/plugins | Unwired draft, superseded by the planned Wasm host |

Design law (David, 2026-10-15, build 20 years ahead): interface
seams at every change-prone boundary (Executor interface, pluggable
verify-check registry with targeted and full modes, the full mode
stays the hard gate), operational limits as config in run-spec
frontmatter with defaults, never compiled-in constants, and data
formats carrying version fields. Full law in docs/RULES.md.

Planned stacks, not in the main tree today (docs/SKEPTICS-AUDIT.md):
Rust (Tyr-Core), Mojo (Mjolnir-Compute), and Flutter (Gjallarhorn-UI)
are blueprint; the engines ship as precompiled binaries in
engines/bin under LICENSE-ENGINES.md.

## Build order
1. Ship and harden the Go core: agent loop, policy gate, SDLC
   engine, Decider, Mind Map, Critic, provider catalog, then the
   2026-10-18 and 2026-10-19 audit fix rounds to a SHIP verdict
   (TODO Phase 0) (done; main at 76e1be1, UI branch at 9d1cae0)
2. Close the F22 secret exposure: rotate the credential at Titan
   Mail, purge it from git history, prove the history clean
   (TODO Phase 1)
3. Land the run engine: merge the scaffold branch, wire a
   production caller for RunUnattended, run the first live
   end-to-end unattended run (TODO Phase 2)
4. Activate engine signing: release key generated, pubkey pinned,
   manifest signed, unsigned updates refused (TODO Phase 3)
5. Operations hardening: SurrealDB credentials set for persistent
   installs, CI verified green on GitHub, first-boot credential
   path verified at runtime (TODO Phase 4)
6. Build the blueprint stacks: Rust, Mojo, Flutter, swarm HTTP
   execution, the Wasm plugin host (TODO Phase 5)
7. Open it to everyone: forge OAuth registration, the forge-link
   flow, the mind map ingestion pipeline to spec, and the 10-stage
   run graph (TODO Phase 6)

## Open questions
- F22 history purge go or no-go (owner: David; the purge breaks his
  Windows clone and needs a re-sync of both branches)
- First live model for the unattended dry run, 8B to 14B class
  (owner: David)
- Merge order and timing: the run-engine scaffold branch, and later
  the engine/gjallarhorn-ui branch (owner: David)
- Whether persistent installs move to the pinned MIMIR_SURREAL_USER
  and MIMIR_SURREAL_PASS before or after the next restart (owner:
  David)
- Run-graph decisions from the 2026-10-07 design: the verify-loop
  attempt cap (proposed: three), the default per-run budget,
  whether research files to the mind map mid-run or at handover,
  and which model runs first live (owner: David)
- Generator-Critic debate loop: spec it as the review-stage
  mechanism now, or park it until after the first dry run (owner:
  David)
- F1 posture: keep the interactive shell behind approval, or move
  the interactive path to an argv-only jail as a product change
  (owner: David)
