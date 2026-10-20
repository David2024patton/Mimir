<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the whole plan for this project, step by step. ENDGOAL.md
is the destination; this file is the route; TODO.md is the work list. When
the plan changes, this file changes FIRST, then TODO.md follows.

EXACT STRUCTURE:
- "## S1: What it is" (the product in plain words, who it serves).
- "## S2: Architecture" (the grid: modules, what each owns, the seams).
- Build steps as numbered sections "## S<n>: <step name> (TODO Phase <n>)",
  one section per build step, each naming the shelf repos it uses (from
  ~/workspace/google-org-harvest/USE-IT.md) and the hosting or delivery
  shape where that matters.
- A final "## S<last>: Open questions" section (owner named per question).

RULES:
- Section numbers are stable once published; TODO items point at them with
  (PLAN S<n>). Renumber only in a deliberate pass that updates both files.
- Facts only: what is decided is stated, what is open is an open question.
- Before designing a step, check the shelf: take or retrofit what already
  exists under a usable license instead of rebuilding it.
- No secrets, passwords, or tokens in this file, ever.
-->
# PLAN: Mimir

The whole plan for Mimir, step by step. Sections are numbered (S1,
S2, ...) so TODO.md items can point at them: every TODO item carries a
stable ID and ends with (PLAN S<n>) naming the section it serves, and
each build step below names its TODO phase. When the plan changes,
this file changes first, then TODO.md follows. Facts only: what is
decided is stated, what is open is named as an open question. The
deep engineering plan lives in docs/PROJECT-MASTER-PLAN.md and the
granular engineering checklist in docs/MASTER-TODO.md;
docs/SKEPTICS-AUDIT.md separates working code from blueprint, and
this file follows the skeptics audit where they differ. Shelf repos
from the harvest (~/workspace/google-org-harvest/USE-IT.md) are
worked into the steps that use them, named on the step.

## S1: What Mimir is
Mimir is David's standalone autonomous coding harness: a local-first
agent that takes a project from plan to shipped code without him
watching. It is GitHub-integrated, runs coding runs start to finish
unattended, and keeps a typed knowledge mind map so uploaded material
integrates into the system instead of sitting static. The end goal
(ENDGOAL.md) is a plan-to-ship harness for everyone, not a private
tool for David only. Mimir is its own project; it is NOT part of
AlphaFlux.

The product model is David's 10 stages: discover, contract, repo,
research, architecture, build, verify, review, ship, handover. The
human-guided portion ends when the plan is approved; execution is
unleashed from there, with hard human gates at verify and ship.
Inside a run today, the shipped engine executes a 7-phase SDLC; the
10-stage run graph is designed and lands in S9 below.

## S2: Architecture
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
| internal/runengine | The unattended run engine (on the scaffold branch until S5) |
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

Shelf at this seam: adk-go (Google's Agent Development Kit for Go)
is the dependency candidate for agent loops, tool wiring, and eval
at the seams this grid already owns, adapted only there, never as a
replacement for the policy gate or the SDLC engine (harvest verdict:
depend, Apache-2.0).

Planned stacks, not in the main tree today (docs/SKEPTICS-AUDIT.md):
Rust (Tyr-Core), Mojo (Mjolnir-Compute), and Flutter (Gjallarhorn-UI)
are blueprint; the engines ship as precompiled binaries in
engines/bin under LICENSE-ENGINES.md.

## S3: Step 1, ship and harden the Go core (TODO Phase 0)
Ship and harden the Go core: agent loop, policy gate, SDLC engine,
Decider, Mind Map, Critic, provider catalog, then the 2026-10-18 and
2026-10-19 audit fix rounds to a SHIP verdict. Done: main at
76e1be1, UI branch at 9d1cae0, gates green (build, vet, gofmt,
go test -race, golangci-lint, gitleaks gate). The run-engine scaffold
(T-0.14) was built and tested locally as part of this step; landing
it is S5.

## S4: Step 2, close the F22 secret exposure (TODO Phase 1)
Rotate the SMTP credential at Titan Mail, purge it from git history
and force push, then re-run gitleaks over the full history and
record a clean report. The rotation and the purge go-ahead are
David's; the purge breaks his Windows clone and both branches (main
and engine/gjallarhorn-ui) need re-syncing after it.

## S5: Step 3, land the run engine (TODO Phase 2)
Merge the run-engine-scaffold branch into main, give RunUnattended a
production caller, and run the first live end-to-end unattended run
against a real model endpoint (8B to 14B class; the model choice is
David's). Then close the v4 gap items (secret redaction in state
dumps, completion signal, network-namespace isolation, multi-language
verify gate) and land session-scoped steering for /api/mobile/steer.

Shelf at this step:
- substrate (agent-substrate): the zero-trust sandbox runtime whose
  architecture the Executor seam is retrofitted to: no credentials
  inside the sandbox, credentials injected at an egress gateway,
  per-agent identity, snapshot and resume of idle runs (harvest
  verdict: retrofit the architecture; run as-is only if a Kubernetes
  host tier joins).
- gVisor: the application kernel sandbox for running generated and
  actively distrusted code on server-side hosts (harvest verdict:
  take as-is, server side).
- go-git: repo operations inside runs (clone, commit, diff, history)
  become pure-Go library calls instead of shelling out to the git
  binary (harvest verdict: depend, upstream v6 line).
- go-github: the GitHub API client behind Mimir's own GitHub
  interface seam, deleting the hand-written GitHub plumbing
  (harvest verdict: depend, BSD-3-Clause).
- oss-fuzz-gen: the pattern for the verify loop: an LLM writes
  checks and fuzz targets, runs them, and iterates on measured
  results instead of a single pass; the blueprint for how the
  multi-language verify gate grows its check registry (harvest
  verdict: take for the dev pipeline).

## S6: Step 4, activate engine signing (TODO Phase 3)
David generates the ed25519 release key offline; pin
MIMIR_ENGINE_PUBKEY for installs and sign engines/bin/manifest.json
with engines.SignManifest; prove an unsigned or badly signed
manifest aborts an engine update. Verification code shipped
2026-10-19; signing is supported but not active until the key
exists.

## S7: Step 5, operations hardening (TODO Phase 4)
Set MIMIR_SURREAL_USER and MIMIR_SURREAL_PASS in the persistent
SurrealDB deployment before the next restart (without them a restart
generates a new sidecar password and locks the existing database);
verify the GitHub CI workflow green on GitHub itself; verify the
SurrealDB env-credential first boot at runtime.

## S8: Step 6, build the blueprint stacks (TODO Phase 5)
Bring the blueprint stacks into the main tree, one at a time: the
Rust stack (Tyr-Core) and Mojo stack (Mjolnir-Compute) sources, the
Flutter stack (Gjallarhorn-UI, branch at 9d1cae0; merge decision
open), swarm factory HTTP execution wiring (the /swarm handler gains
execute and approve paths on top of submit and list), and the Wasm
plugin host that supersedes the unwired internal/plugins draft.

## S9: Step 7, the harness for everyone (TODO Phase 6)
Open the harness up: OAuth app registration for the git forges
(GitHub, GitLab, Bitbucket, Gitea/Forgejo; stalled on David
submitting passwords via secure card); the forge-link flow decided
and landed (every chat session links a forge repo; clone-into-
workspace and push-back are its phase 2, built on go-github and
go-git library calls per S5); the mind map ingestion pipeline to
the GOAL.md spec (parse, entity and relation extraction,
deterministic dedupe, merge with contradiction flagging,
integration receipt; typed edges: implements, depends_on,
constrains, decided_because, contradicts, supersedes, part_of,
learned_from; contradictions resolved by the user at the plan gate,
never silently overwritten); and the 10-stage run-graph execution
model as an inspectable graph with typed done gates (design
delivered 2026-10-07; open decisions in S10).

Shelf at this step:
- agentmesh: the contract draft for networked Mimir nodes on host
  tiers: verifiable identity, publish and discover, and
  direct-plus-relay transport, so a run can span machines without a
  new protocol being invented here (harvest verdict: retrofit).
- mangle: the embeddable Datalog engine for deductive queries over
  the mind map's facts: dependency reachability, ownership, and call
  graphs answered by recursive rules over what GitIngest extracted
  (harvest verdict: depend, embedded in-process).
- graph-renderer: the mind map's visual surface: the standalone web
  component for interactive node graphs (pan, zoom, dragging,
  optional ELK layout) that the map UI renders with (harvest
  verdict: depend via npm).

## S10: Open questions
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
