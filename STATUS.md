<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the living right-now snapshot of this project. Anyone
opening the repo cold reads this first and is current in two minutes.
Updated at the end of every work session on the project; a stale STATUS
is worse than none, so when you finish work here, update it in the same
pass.

EXACT STRUCTURE:
- "## As of <YYYY-MM-DD>": the date this snapshot was written.
- "## Proven state": what is done, with its proof named.
- "## In flight": work currently underway and where it stands.
- "## Blocked": each blocker and exactly who or what it waits on.
- "## Next moves": the next three moves, in order, referencing TODO ids.

RULES:
- Claims here carry the same proof standard as TODO checkboxes.
- Keep it short: a snapshot, not a history (history is CHANGELOG.md).
- No secrets, passwords, or tokens in this file, ever.
-->
# STATUS: Mimir

## As of 2026-10-20

## Proven state
- Main is at 76e1be1 and the UI branch (engine/gjallarhorn-ui) is at 9d1cae0, with a SHIP verdict for the code from the second audit_code pass on 2026-10-19 (proof: goal workspace reports mimir-fixes-2026-10-19.md, mimir-audit-code-2026-10-19.md, mimir-audit-code-2-2026-10-19.md) (TODO T-0.8)
- Gates green on main at that pass: go build, go vet, gofmt clean, go test -race green on all packages, golangci-lint clean, gitleaks gate run (proof: audit_code report 2, 2026-10-19) (TODO T-0.8)
- The Go core is shipped: agent tool loop behind a fail-closed policy gate, Norse personas, 7-phase SDLC engine inside runs, Huginn-Decider, GitIngest plus the Yggdrasil Mind Map in the Cortex, Epistemic Critic and the mimir audit CLI, 12+ provider catalog behind one client interface, web UI on :8420 (proof: TODO Phase 0 items T-0.1 to T-0.6)
- Approval fails closed, steer is session-owned, relay pairing is hardened, engine update verification is implemented, and serve auth refuses an unauthenticated non-loopback bind (proof: TODO items T-0.9 to T-0.13)
- The run-engine scaffold (v4) is built and tested locally on branch run-engine-scaffold at 9e0b5ad: RunContext, run-spec frontmatter, allowlisted argv jail through the policy gate, atomic patches with dry run first, run lock, checklist done-gate, sandbox prep, mimir review CLI (proof: TODO T-0.14; verified independently, not on main)

## In flight
- Repo standard backfill: the jump-back-in file set (this STATUS.md, DECISIONS.md, docs/RUNBOOK.md, CHANGELOG.md, docs/MAP.md) is being added on 2026-10-20, on top of the TODO, IDEA, PLAN, ENDGOAL, and docs/RULES.md set already on main
- No feature code is in flight at this snapshot; the next code work is the run engine landing in TODO Phase 2, which waits on the decisions listed under Blocked

## Blocked
- F22 secret exposure: a mail credential sits in this public repo's git history; rotation at Titan Mail and the history purge are both held for David's explicit go (waiting on: David) (TODO T-1.1, T-1.2)
- Run-engine scaffold merge into main (waiting on: David's merge decision) (TODO T-2.1)
- First live end-to-end unattended run (waiting on: David's model choice, 8B to 14B class) (TODO T-2.3, T-P.2)
- Engine signing activation: verification code is shipped, signing is not active until a release key exists (waiting on: David generating the ed25519 release key offline) (TODO T-3.1)
- Forge OAuth app registration for GitHub, GitLab, Bitbucket, and Gitea/Forgejo (waiting on: David submitting current passwords via secure card; stalled since 2026-09-29) (TODO T-6.1)
- Persistent SurrealDB installs must set MIMIR_SURREAL_USER and MIMIR_SURREAL_PASS before the next restart or the database locks (waiting on: the deployment environment being set) (TODO T-4.1)

## Next moves
1. Close F22: rotate the credential at Titan Mail, then purge history on David's explicit go and record a clean gitleaks report (TODO T-1.1, T-1.2, T-1.3)
2. Land the run engine: merge the scaffold branch and give RunUnattended a production caller (TODO T-2.1, T-2.2)
3. Activate engine signing: generate the release key, pin MIMIR_ENGINE_PUBKEY, and prove a badly signed manifest aborts an update (TODO T-3.1, T-3.2, T-3.3)
