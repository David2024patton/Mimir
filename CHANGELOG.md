<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the human-readable milestone trail of this project:
what shipped, dated, newest first. Git history is the raw record; this
is the curated one a person can read in a minute.

EXACT STRUCTURE:
- "## <YYYY-MM-DD>: <milestone in one line>" entries, newest first.
- Under each: what changed, and the proof or commit it landed in.

RULES:
- An entry is added when something SHIPS (lands on the main branch or
  goes live), not when it is started.
- One entry per milestone, not per commit.
- No secrets, passwords, or tokens in this file, ever.
-->
# CHANGELOG: Mimir

## 2026-10-20: Repo standard jump-back-in set added
- STATUS.md, DECISIONS.md, docs/RUNBOOK.md, CHANGELOG.md, and
  docs/MAP.md created, completing the standard file set on main
  (TODO, IDEA, PLAN, ENDGOAL, and docs/RULES.md landed earlier the
  same day, with the PLAN, TODO cross-reference, and IDEA record
  format upgrades).

## 2026-10-19: Audit fix rounds closed, code verdict SHIP
- David's fix round plus two audit_code passes closed the code
  findings from the 2026-10-18 audit: approval fails closed, steer is
  session-owned, relay pairing hardened with expiry and replay
  rejection, engine updates are https-only with ed25519 manifest
  signature verification implemented, serve auth refuses an
  unauthenticated non-loopback bind, and the compiled-in SurrealDB
  defaults are gone.
- Final state: main at 76e1be1, UI branch engine/gjallarhorn-ui at
  9d1cae0; go build, go vet, gofmt clean, go test -race green on all
  packages, golangci-lint clean, gitleaks gate run (proof: goal
  workspace mimir-fixes-2026-10-19.md, mimir-audit-code-2026-10-19.md,
  mimir-audit-code-2-2026-10-19.md).
- Left open on purpose, tracked in TODO Phases 1 to 4: F22 rotation
  and history purge (David's go), engine signing activation (release
  key), and the operations hardening checks.

## 2026-10-18: First full audit of the pushed core
- Full-tree audit of main at 0c9d034 found 29 findings, 3 critical
  (proof: goal workspace mimir-audit-2026-10-18.md). The findings
  drove the 2026-10-19 fix round above.

## 2026-10-17: Core systems pushed to main by David
- The 7-phase SDLC engine, Huginn-Decider, GitIngest plus the
  Yggdrasil Mind Map into the Cortex, the Epistemic Critic and the
  mimir audit CLI, and the 12+ provider catalog landed on main in
  commit 0c9d034 (proof: TODO items T-0.2 to T-0.6).

## 2026-09-29: Go core verified by repo review
- Repo review verified the shipped core: agent tool loop with a
  fail-closed policy gate, Norse personas (odin, thor, loki,
  heimdall), file memory plus SurrealDB Cortex memory, OTP auth, and
  the web UI on :8420, with build clean and tests green except one
  stale test (proof: TODO T-0.1).
