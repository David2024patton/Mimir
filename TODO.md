<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the super-detailed task list for this project. It is the
only place work items live. PLAN.md says why and in what order; ENDGOAL.md
says what finished looks like; this file says exactly what remains to do.

EXACT STRUCTURE:
- Phases in order: "## Phase <n>: <name> (PLAN S<m>)".
- Parked work in its own phase at the end: "## Parked".
- Every item is one checkbox line: "- [ ] T-<phase>.<n> <task> (proof: <named
  proof>) (PLAN S<m>)". Parked items use T-P.<n> and name what blocks them.

RULES:
- A box is checked ONLY when its named proof exists (a test, a log line, a
  bench note, a live check). Never check a box on intention or partial work.
- Emulation-proven and bench-proven are different claims; write which one.
- Item IDs and (PLAN S<n>) tags are stable once published; never renumber
  casually. A deliberate renumber updates PLAN.md in the same pass.
- New work enters as new items at the end of their phase. Completed items
  stay, checked, as the record.
- No secrets, passwords, or tokens in this file, ever.
-->
# TODO: Mimir

The detailed task list for Mimir. Rules: a box is checked only when its
proof exists and is named (a test, a log line, a bench note, a live
check). Plan steps live in PLAN.md; each phase here names its PLAN
section. Deeper engineering checklists live in docs/MASTER-TODO.md; this
file tracks the project state at the top level and never contradicts
the audit reports.

## Phase 0: Shipped Go core, audited (PLAN Step 1)
- [x] T-0.1 Agent tool loop with a fail-closed policy gate, Norse personas
  (odin, thor, loki, heimdall), file memory plus SurrealDB Cortex
  memory, OTP auth, web UI on :8420 (proof: repo review 2026-09-29;
  code on main) (PLAN S3)
- [x] T-0.2 7-phase SDLC engine inside runs, with phase tool enforcement at
  execution (proof: internal/sdlc tests; 2026-10-19 fix round, F5) (PLAN S3)
- [x] T-0.3 Huginn-Decider for routing decisions (proof: internal/decision
  tests; pushed by David in 0c9d034, 2026-10-17) (PLAN S3)
- [x] T-0.4 GitIngest plus Yggdrasil Mind Map into the Cortex (proof:
  internal/ingest tests; pushed by David in 0c9d034, 2026-10-17) (PLAN S3)
- [x] T-0.5 Epistemic Critic and the mimir audit CLI (proof: internal/critic
  tests; pushed by David in 0c9d034, 2026-10-17) (PLAN S3)
- [x] T-0.6 Provider catalog of 12+ providers behind one client interface
  (proof: internal/config and internal/llm tests) (PLAN S3)
- [x] T-0.7 First audit round closed: 2026-10-18 audit found 29 findings, 3
  critical; David's fixes, the 2026-10-19 fix round, and two audit_code
  passes closed the code findings (proof: goal workspace reports
  mimir-audit-2026-10-18.md, mimir-fixes-2026-10-19.md,
  mimir-audit-code-2026-10-19.md, mimir-audit-code-2-2026-10-19.md) (PLAN S3)
- [x] T-0.8 Gates green on main at 76e1be1: go build, go vet, gofmt clean,
  go test -race green on all packages, golangci-lint clean, gitleaks
  gate run (proof: audit_code report #2, 2026-10-19, verdict SHIP for
  the code) (PLAN S3)
- [x] T-0.9 Approval is real and fails closed: RequireApproval with no
  approver denies, on both the gate branch and the ModeEdit branch
  (proof: fix report 2026-10-19, F2/N1, with tests) (PLAN S3)
- [x] T-0.10 Steer is session-owned: per-session steer queues, context-bound
  delivery, cross-user leak test red before and green after (proof:
  audit_code report #2, C1) (PLAN S3)
- [x] T-0.11 Relay pairing hardened: pairing token required, unpaired sessions
  expire and are torn down with keys wiped, counter nonces with replay
  rejection (proof: fix report 2026-10-19, N3/N5; audit_code #2, C3) (PLAN S3)
- [x] T-0.12 Engines update trust: https-only update base, ed25519 manifest
  signature verification implemented, downgrades refused by default
  (proof: fix report 2026-10-19, N7/N8; activation is Phase 3 below) (PLAN S3)
- [x] T-0.13 Serve auth hardened: non-loopback bind refused with no auth
  backend unless MIMIR_ALLOW_UNAUTHENTICATED=1; root/root SurrealDB
  default removed, credentials resolve from config, then
  MIMIR_SURREAL_USER and MIMIR_SURREAL_PASS, then a per-run random
  sidecar credential (proof: fix report 2026-10-19, F16/F19) (PLAN S3)
- [x] T-0.14 Run-engine scaffold built and tested locally, v4: RunContext,
  run-spec frontmatter, allowlisted argv jail routed through the
  policy gate, atomic patches with dry run first, run lock, checklist
  done-gate on task_complete, per-command timeout and output caps,
  sandbox prep, mimir review CLI for the review stage (proof:
  ~/workspace/mimir-build/mimir, branch run-engine-scaffold, commit
  9e0b5ad; go build and go vet clean, go test green on runengine, llm,
  and cmd, verified independently; merge is Phase 2 below) (PLAN S3)

## Phase 1: Close the F22 secret exposure (PLAN Step 2)
- [ ] T-1.1 David rotates the SMTP password at Titan Mail (blocked on:
  David; the password sitting in public git history is treated as
  live until rotated; finding F22, open across four audits since
  2026-10-18) (PLAN S4)
- [ ] T-1.2 Purge the credential from git history and force push (blocked
  on: David's explicit go; the rewrite breaks his Windows clone and
  both branches, main and engine/gjallarhorn-ui, need re-syncing) (PLAN S4)
- [ ] T-1.3 Re-run gitleaks over the full history after the purge and record
  a clean report (proof: gitleaks report with zero findings for the
  credential) (PLAN S4)

## Phase 2: Land the run engine (PLAN Step 3)
- [ ] T-2.1 Merge the run-engine-scaffold branch into main (blocked on:
  David's merge decision; branch state in Phase 0 above) (PLAN S5)
- [ ] T-2.2 Give RunUnattended a production caller (today it has none;
  proof: a caller wired in cmd/mimir plus a dry run log) (PLAN S5)
- [ ] T-2.3 First live end-to-end unattended run against a real model
  endpoint, 8B to 14B class per the blueprint target (blocked on:
  David's model choice; proof: the run report artifact the run
  itself writes) (PLAN S5)
- [ ] T-2.4 Close the v4 gap items: secret redaction in state dumps,
  completion signal, network-namespace isolation, multi-language
  verify gate (source: GOAL.md gap list, 2026-10-05) (PLAN S5)
- [ ] T-2.5 Session-scoped steering for /api/mobile/steer (today it steers
  the shared agent's live run because the agent exposes no
  per-session steer API; source: fix report 2026-10-19, open item 5) (PLAN S5)

## Phase 3: Activate engine signing (PLAN Step 4)
- [ ] T-3.1 David generates the ed25519 release key offline (blocked on:
  David; verification code shipped 2026-10-19, signing is supported
  but NOT active until a key exists) (PLAN S6)
- [ ] T-3.2 Pin MIMIR_ENGINE_PUBKEY for installs and sign
  engines/bin/manifest.json with engines.SignManifest, committing
  the signature (PLAN S6)
- [ ] T-3.3 Prove an unsigned or badly signed manifest aborts an engine
  update (proof: test run or a recorded update attempt) (PLAN S6)

## Phase 4: Operations hardening (PLAN Step 5)
- [ ] T-4.1 Set MIMIR_SURREAL_USER and MIMIR_SURREAL_PASS in the persistent
  SurrealDB deployment environment before the next restart (without
  them a restart generates a new sidecar password and locks the
  existing database; source: fix report 2026-10-19, open item 3) (PLAN S7)
- [ ] T-4.2 Verify the GitHub CI workflow green on GitHub itself (the
  workflow was added 2026-10-19 and rehearsed locally; a server-side
  run is still unverified; source: audit_code report #2, residual
  risk) (PLAN S7)
- [ ] T-4.3 Verify SurrealDB env-credential first boot at runtime (the
  audits had no surreal binary available, so this stayed unverified;
  source: audit_code report #2, residual risk) (PLAN S7)

## Phase 5: Blueprint stacks (PLAN Step 6)
Separated from working code by docs/SKEPTICS-AUDIT.md; none of these
are built in the main tree today.
- [ ] T-5.1 Rust stack (Tyr-Core) source in the main tree (today the
  engines ship as precompiled binaries in engines/bin under
  LICENSE-ENGINES.md, built from engine branches) (PLAN S8)
- [ ] T-5.2 Mojo stack (Mjolnir-Compute) source in the main tree (PLAN S8)
- [ ] T-5.3 Flutter stack (Gjallarhorn-UI): the branch engine/gjallarhorn-ui
  at 9d1cae0 carries the mobile companion work (pairing, remote diff
  approval cards, mid-turn steering); merge decision is open (PLAN S8)
- [ ] T-5.4 Swarm factory HTTP execution wiring (the factory stops tasks at
  AwaitingMerge and merges only through Factory.ApproveMerge; the
  server /swarm handler still only submits and lists; source: fix
  report 2026-10-19, N6) (PLAN S8)
- [ ] T-5.5 Wasm plugin host (internal/plugins is an unwired draft,
  documented in code as superseded by the planned Wasm host; source:
  fix report 2026-10-19, F26c) (PLAN S8)

## Phase 6: The harness for everyone (PLAN Step 7)
- [ ] T-6.1 OAuth app registration for the git forges (GitHub, GitLab,
  Bitbucket, Gitea/Forgejo) (blocked on: David submitting current
  passwords via secure card; stalled since 2026-09-29) (PLAN S9)
- [ ] T-6.2 Decide the two green unpushed patches from the 2026-09-29 repo
  review: the server test fix via the AuthBackend seam, and the
  forge-link requirement patch (every chat session links a forge
  repo; clone-into-workspace and push-back are its open phase 2)
  (blocked on: David's call) (PLAN S9)
- [ ] T-6.3 Mind map ingestion pipeline to the GOAL.md spec: parse, entity
  and relation extraction, deterministic dedupe, merge with
  contradiction flagging, integration receipt; typed edges
  (implements, depends_on, constrains, decided_because, contradicts,
  supersedes, part_of, learned_from); contradictions resolved by the
  user at the plan gate, never silently overwritten (PLAN S9)
- [ ] T-6.4 Run-graph execution model: the 10 stages (discover, contract,
  repo, research, architecture, build, verify, review, ship,
  handover) as an inspectable graph with typed done gates (design
  delivered 2026-10-07; open decisions in PLAN.md) (PLAN S9)

## Parked (waiting on a decision, hardware, or hands)
- T-P.1 F1 posture: the interactive bash tool runs through a shell behind
  real approval; the argv jail applies only on the unattended path.
  An argv-only interactive shell is a product decision for David,
  not a fix (source: fix report 2026-10-19, posture note) (PLAN S2)
- T-P.2 First live model for the unattended dry run (blocked on: David) (PLAN S5)
- T-P.3 Generator-Critic debate loop as a review-stage mechanism
  (proposal shared by David 2026-10-05; decision: spec it now or
  park it until after the first dry run) (PLAN S9)
