# IDEA: Mimir

David's idea capture for Mimir. When David talks or dictates a thought
about this project, it lands here in his words, dated, newest first,
with a one-line note of what it would take. An idea leaves this file
only by moving into TODO.md (marked MOVED with the TODO phase) or by
David killing it (marked KILLED with the date). Nothing is silently
dropped. Entries below are recorded summaries from David's direction
in chat and his goal file; quotation marks are used only where his
exact words are on record.

## 2026-10-15
- Build 20 years ahead, applied to Mimir first: design for the
  bottleneck that is coming, modules independently replaceable,
  interface seams at the change-prone boundaries (Executor interface,
  pluggable verify-check registry with targeted and full modes,
  limits in run-spec frontmatter instead of compiled-in constants).
  - Would take: holding every refactor and new module to the design
    law; the run-engine refactor was steered to it the same day.
  - Status: MOVED to docs/RULES.md and TODO Phase 2 and Phase 5

## 2026-10-07
- Model the 10-stage harness as an explicit, inspectable run graph:
  nodes, allowed transitions, durable state, budgets, retries that
  require new evidence, and human verify and ship gates (from an
  article on graph engineering for agents that David shared and
  asked to apply to Mimir).
  - Would take: the graph design was delivered 2026-10-07; four open
    decisions sit in PLAN.md (verify-loop attempt cap, per-run
    budget, when research files to the mind map, first live model).
  - Status: MOVED to TODO Phase 6

## 2026-10-05
- The run loop: on wake, load only the raw requirements for the run;
  an action loop against a local model; on task_complete, hand off
  to a verification gate instead of pushing directly, and feed gate
  failures back into the loop so the model self-heals. David's
  standing requirement behind it: coding runs go start to finish
  without him watching.
  - Would take: the scaffold (built, v4, local branch); merging it
    and wiring a production caller is the remaining work.
  - Status: MOVED to TODO Phase 0 (scaffold) and TODO Phase 2 (land
    and call it)
- A Generator-Critic debate loop for the review stage: separate
  Builder and skeptical Reviewer personas, bounded turns, context
  compaction, human escalation (proposal David shared from Gemini;
  a proposal, not approved or built).
  - Would take: a spec decision from David; structured consensus
    instead of a magic consensus string, per the 2026-10-05 review.
  - Status: NEW (parked in TODO, waiting on David's decision)

## 2026-10-02
- Every chat session links a forge repo; account-less users create
  one, so a session always has real code to work on.
  - Would take: the patch is built and green (stdlib OAuth2 for
    GitHub, GitLab, Bitbucket, Gitea/Forgejo; sealed tokens;
    verified-repo gate); David's push decision, then clone-into-
    workspace and push-back as phase 2.
  - Status: MOVED to TODO Phase 6

## 2026-09-30
- A plan-to-ship harness for everyone, not a private tool: guide a
  user through planning, create or link a repository, research
  competitors and features, guide architecture and development,
  then carry the project autonomously through verification and
  shipping. The human-guided part ends when the plan is approved;
  execution is unleashed from there, with hard human gates at
  verify and ship.
  - Would take: the 10-stage model below, the run engine (TODO
    Phase 2), and the product work in TODO Phase 6.
  - Status: MOVED to PLAN.md and TODO Phase 6
- The 10-stage model: discover, contract, repo, research,
  architecture, build, verify, review, ship, handover. Unleash
  point at plan approval; hard gates at verify and ship.
  - Would take: the run-graph implementation in TODO Phase 6; the
    shipped engine already runs a 7-phase SDLC inside runs.
  - Status: MOVED to TODO Phase 6
- A knowledge mind map with relationships: uploaded material must
  not sit static or isolated; it integrates into planning,
  research, building, verification, and future learning across the
  whole system. Typed edges (implements, depends_on, constrains,
  decided_because, contradicts, supersedes, part_of, learned_from);
  contradictions are resolved by the user, never silently
  overwritten.
  - Would take: the ingestion pipeline spec in TODO Phase 6,
    building on the GitIngest Yggdrasil Mind Map already in the
    Cortex.
  - Status: MOVED to TODO Phase 6
