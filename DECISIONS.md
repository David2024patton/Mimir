<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the settled-decision log for this project. IDEA.md is
the raw capture of what David said; this file is only the calls that are
FINAL, so no one re-asks a settled question or quietly reverses a call.

EXACT STRUCTURE: one entry per decision, newest first:
## <YYYY-MM-DD>: <decision in one line>
- **Decided:** <what was picked>
- **Why:** <the reason, in plain words>
- **Supersedes:** <the earlier call it replaces, or "none">
- **By:** David | <who, with David's approval noted>

RULES:
- A decision enters here the day it is made, in the same pass as any
  PLAN or TODO change it causes.
- Reversing a decision means a NEW entry that names what it supersedes;
  never edit an old entry into a different decision.
- No secrets, passwords, or tokens in this file, ever.
-->
# DECISIONS: Mimir

## 2026-10-19: Engine signing stays inactive until a release key exists
- **Decided:** Ed25519 manifest signature verification is implemented and shipped, but signing is NOT active until David generates a release key offline and MIMIR_ENGINE_PUBKEY is pinned for installs.
- **Why:** A signing system with no real key held by the owner proves nothing; activation is a key custody act, not a code change.
- **Supersedes:** none
- **By:** David (activation step held for him; recorded in TODO Phase 3)

## 2026-10-19: Approval fails closed, everywhere
- **Decided:** RequireApproval with no approver denies the action, on both the gate branch and the ModeEdit branch. Bypass modes waive approval requests, never explicit denies. The interactive shell (F1 posture) stays behind real approval; moving the interactive path to an argv-only jail is a product decision, parked, not a fix.
- **Why:** An approval gate that passes when nobody is there to approve is worse than no gate, because it looks like protection.
- **Supersedes:** the earlier inert approval behavior found in the 2026-10-18 audit
- **By:** David (fix round ordered 2026-10-19; posture note in the fix report)

## 2026-10-18: F22 history purge is held for David's explicit go
- **Decided:** The mail credential in this public repo's git history is treated as live until rotated at Titan Mail. The history purge and force push do NOT happen until David gives an explicit go, because the rewrite breaks his Windows clone and both branches (main and engine/gjallarhorn-ui) need re-syncing after it.
- **Why:** Rotation closes the exposure; the purge is destructive to every clone and is the owner's call to time.
- **Supersedes:** none
- **By:** David (held across four audits, 2026-10-18 to 2026-10-19; recorded in TODO Phase 1)

## 2026-10-15: Build 20 years ahead is the design law
- **Decided:** Every module is designed for the bottleneck that is coming: interface seams at change-prone boundaries (Executor interface, pluggable verify-check registry with targeted and full modes, the full mode stays the hard gate), operational limits as config in run-spec frontmatter with defaults, never compiled-in constants, and data formats carrying version fields.
- **Why:** Models, tools, and hardware will be replaced many times over the life of the harness; the seams are what let a module be swapped without rebuilding the system.
- **Supersedes:** none
- **By:** David (recorded in ~/AGENTS.md and the Mimir blueprint design-law section the same day)

## 2026-10-05: Mimir is standalone, and coding runs go unwatched
- **Decided:** Mimir is David's standalone autonomous coding harness, NOT part of the AlphaFlux SaaS, and stays out of AlphaFlux status updates. It was chosen over MiMo Desktop as the harness direction. Coding runs go start to finish without David watching; the GitHub-integrated harness plus the orchestrator agent keep the build moving and verify it is building correctly.
- **Why:** The harness is a product of its own with its own end goal (a plan-to-ship harness for everyone); folding it into a SaaS status stream would blur both.
- **Supersedes:** the earlier pattern of David coding with MiMo Desktop on his PC
- **By:** David

## 2026-09-30: The product model is 10 stages, unleashed at plan approval, with hard human gates at verify and ship
- **Decided:** A run moves through discover, contract, repo, research, architecture, build, verify, review, ship, handover. The human-guided part ends when the plan is approved; execution is unleashed from there, and it stops for a human at exactly two places: verify and ship.
- **Why:** Planning is where human judgment changes the outcome; execution is where watching adds nothing. The two gates keep shipping a human act.
- **Supersedes:** none
- **By:** David (the end goal model in the goal file, 2026-09-30)

## 2026-09-30: Knowledge is never silently overwritten
- **Decided:** Material the harness ingests lands in a typed knowledge mind map (edges: implements, depends_on, constrains, decided_because, contradicts, supersedes, part_of, learned_from). When two facts contradict, the map flags it and the user resolves it at the plan gate; a merge never silently overwrites.
- **Why:** A memory that quietly rewrites itself cannot be trusted as the record a later run plans against.
- **Supersedes:** none
- **By:** David (knowledge mind map requirement in the goal file, 2026-09-30)

## 2026-09-30: Open-core license model
- **Decided:** The Go harness control plane in this repo is provided openly so developers and teams can run, inspect, and integrate it. The precompiled engines in engines/bin are proprietary to Alpha Flux LLC under the attribution license in LICENSE-ENGINES.md and the NOTICE file.
- **Why:** The harness spreads by being open and inspectable; the engines are the commercial core and keep their attribution terms. (Entry recorded in this log on 2026-10-20 from the repo record: NOTICE and LICENSE-ENGINES.md.)
- **Supersedes:** none
- **By:** David (per NOTICE and LICENSE-ENGINES.md on main)

## 2026-09-29: Go toolchain is pinned
- **Decided:** The repo pins the Go toolchain (go.mod toolchain line, go1.26.9 at the 2026-10-19 audit). Tool versions are part of the gate; a stale linter or scanner is a finding, not a footnote.
- **Why:** A gate run with old tools proves less than it claims; the pin makes every machine prove the same thing. (Pin state verified in audit_code report 2, 2026-10-19.)
- **Supersedes:** none
- **By:** David (recorded in docs/RULES.md, Mimir specifics)
