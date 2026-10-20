# IDEA: Mimir

David's idea capture for Mimir. Every idea lands here as a full
record, newest first, kept in categories. An idea leaves active
status only by moving into TODO.md (status MOVED, with the TODO item
id) or by David killing it (status KILLED, with the date). Nothing is
silently dropped.

Entry format: category, date and time where the record carries one
(the entries below carry dates; times were not recorded for them),
David's words, Lulu's reply or assessment at the time, and a stats
line. The entries below are recorded summaries from David's
direction in chat and his goal file, upgraded to this format on
2026-10-20; both the David and Lulu lines are marked as summaries
because neither is a verbatim transcript. Quotation marks are used
only where exact words are on record.

## Tech | 2026-10-15
- **David:** (summary) Build 20 years ahead, applied to Mimir first:
  design for the bottleneck that is coming, modules independently
  replaceable, interface seams at the change-prone boundaries
  (Executor interface, pluggable verify-check registry with targeted
  and full modes, limits in run-spec frontmatter instead of
  compiled-in constants).
- **Lulu:** (summary) Recorded the rule in ~/AGENTS.md and in the
  Mimir blueprint design-law section the same day, and steered the
  in-flight run-engine refactor to it: Executor interface, verify
  registry, limits moved to run-spec frontmatter.
- **Stats:** would take: holding every refactor and new module to
  the design law; effort: L (a standing discipline, not a task);
  value: modules stay replaceable for the life of the harness;
  depends on: nothing.
- **Status:** MOVED to docs/RULES.md; applied in TODO Phase 2
  (items T-2.1 to T-2.5) and TODO Phase 5 (items T-5.1 to T-5.5)

## Design | 2026-10-07
- **David:** (summary) Model the 10-stage harness as an explicit,
  inspectable run graph: nodes, allowed transitions, durable state,
  budgets, retries that require new evidence, and human verify and
  ship gates. From an article on graph engineering for agents that
  David shared and asked to apply to Mimir.
- **Lulu:** (summary) The graph design was delivered 2026-10-07;
  four decisions were left open for David (verify-loop attempt cap,
  default per-run budget, when research files to the mind map,
  first live model), now carried in PLAN S10.
- **Stats:** would take: implementing the designed graph in the run
  engine; effort: L; value: runs become inspectable and resumable
  instead of opaque; depends on: the run engine landing (TODO
  T-2.1).
- **Status:** MOVED to TODO T-6.4

## Feature | 2026-10-05
- **David:** (summary) The run loop: on wake, load only the raw
  requirements for the run; an action loop against a local model;
  on task_complete, hand off to a verification gate instead of
  pushing directly, and feed gate failures back into the loop so the
  model self-heals. Behind it, David's standing requirement: coding
  runs go start to finish without him watching.
- **Lulu:** (summary) The scaffold was built to v4 and verified
  independently (build and vet clean, tests green). A pasted
  bash-executor variant was rejected as a command-injection hole;
  the shell-less allowlisted jail was kept.
- **Stats:** would take: merging the scaffold and wiring a
  production caller; effort: M (scaffold exists); value: the
  unattended run, the core promise of the harness; depends on:
  David's merge decision and first live model.
- **Status:** MOVED to TODO T-0.14 (scaffold) and TODO T-2.1, T-2.2
  (land it and call it)

## Feature | 2026-10-05
- **David:** (summary) A Generator-Critic debate loop for the
  review stage: separate Builder and skeptical Reviewer personas,
  bounded turns, context compaction, human escalation. A proposal
  David shared from Gemini; a proposal, not approved or built.
- **Lulu:** (summary) Reviewed 2026-10-05: worth doing with
  structured consensus instead of a magic consensus string; parked
  pending David's spec decision.
- **Stats:** would take: a spec decision from David, then the
  review-stage mechanism; effort: M; value: review quality without
  a human reading every diff; depends on: David's decision, and
  ideally the first dry run.
- **Status:** NEW (parked as TODO T-P.3, waiting on David's
  decision)

## Feature | 2026-10-02
- **David:** (summary) Every chat session links a forge repo;
  account-less users create one, so a session always has real code
  to work on.
- **Lulu:** (summary) The patch was built and green (stdlib OAuth2
  for GitHub, GitLab, Bitbucket, Gitea/Forgejo; sealed tokens;
  verified-repo gate) and left unpushed for David's call, with
  clone-into-workspace and push-back as its phase 2.
- **Stats:** would take: David's push decision, then phase 2;
  effort: M; value: no session ever starts without a repo; depends
  on: David's call, and forge OAuth registration (TODO T-6.1).
- **Status:** MOVED to TODO T-6.2

## Product | 2026-09-30
- **David:** (summary) A plan-to-ship harness for everyone, not a
  private tool: guide a user through planning, create or link a
  repository, research competitors and features, guide architecture
  and development, then carry the project autonomously through
  verification and shipping. The human-guided part ends when the
  plan is approved; execution is unleashed from there, with hard
  human gates at verify and ship.
- **Lulu:** (summary) Recorded as the project's end goal (now
  ENDGOAL.md); Mimir stands as David's standalone harness, and the
  combined stage-and-gate spec was assembled for his yes.
- **Stats:** would take: the 10-stage model below, the run engine,
  and the product work in TODO Phase 6; effort: L; value: this is
  the product itself; depends on: TODO Phases 2 and 6.
- **Status:** MOVED to PLAN S1 and S9, ENDGOAL.md, and TODO Phase 6
  (items T-6.1 to T-6.4)

## Product | 2026-09-30
- **David:** (summary) The 10-stage model: discover, contract,
  repo, research, architecture, build, verify, review, ship,
  handover. Unleash point at plan approval; hard gates at verify
  and ship.
- **Lulu:** (summary) Adopted as the product model in the goal file
  the same day; the shipped engine's 7-phase SDLC is the
  inside-a-run subset, and the full graph is the S9 work.
- **Stats:** would take: the run-graph implementation; effort: L
  (shared with the run-graph entry above); value: one model from
  first question to handover; depends on: TODO T-6.4.
- **Status:** MOVED to TODO T-6.4

## Feature | 2026-09-30
- **David:** (summary) A knowledge mind map with relationships:
  uploaded material must not sit static or isolated; it integrates
  into planning, research, building, verification, and future
  learning across the whole system. Typed edges (implements,
  depends_on, constrains, decided_because, contradicts, supersedes,
  part_of, learned_from); contradictions are resolved by the user,
  never silently overwritten.
- **Lulu:** (summary) The ingestion pipeline spec was written into
  the goal file (parse, extraction, deterministic dedupe, merge
  with contradiction flagging, integration receipt), building on
  the GitIngest Yggdrasil Mind Map already in the Cortex.
- **Stats:** would take: the pipeline build in TODO T-6.3; effort:
  L; value: everything the harness reads compounds into knowledge
  it can plan and verify against; depends on: the Cortex mind map
  already shipped (TODO T-0.4).
- **Status:** MOVED to TODO T-6.3
