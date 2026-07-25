# Research: AI Coder Feature Landscape (2025-2026)

Companion to `PROJECT-MASTER-PLAN.md`. Synthesizes research across: Agent Zero,
Claude Code, Google Antigravity + Gemini CLI, opencode, Cursor, Windsurf,
Cline/Roo Code, Devin, GitHub Copilot coding agent, OpenAI Codex, Kiro, Factory.

Purpose: capture every good idea so Mímir's plan incorporates the cutting edge.

---

## A. Cross-Cutting Trends (table stakes for a 2026 entrant)

These now appear across nearly every serious tool. A new app needs them to be credible.

| Trend | What it means | Seen in |
|---|---|---|
| **Multi-provider models** | One harness, many providers + local; auto-routing by cost/quality | all |
| **Agent loop w/ streaming + tool calling** | gather context -> act -> verify, repeat | all |
| **Built-in system tools** | terminal, filesystem, search, code-exec, web | all |
| **MCP as universal extension protocol** | stdio + remote (SSE/HTTP), OAuth, marketplaces | all |
| **Subagents w/ isolated context** | specialist workers, own context, return only a summary | all |
| **Plan / Act separation** | read-only planning mode, reviewable before execution | all |
| **Checkpoints / undo / rewind** | reversible agent work (files + conversation) | opencode, Claude Code, Cursor, Windsurf, Cline, Kiro, Gemini |
| **AGENTS.md as cross-tool memory** | canonical project instructions file | opencode (canonical), Cursor, Windsurf, Codex |
| **Agent Skills (SKILL.md)** | progressive-disclosure capability bundles, cross-tool standard | Claude Code, opencode, Cursor, Windsurf, Cline, Gemini, Agent Zero |
| **Hooks (lifecycle enforcement)** | deterministic pre/post tool-use handlers | Claude Code, Cursor, Kiro, Factory, Gemini, opencode (plugin hooks) |
| **Granular permissions + sandboxing** | allow/ask/deny per tool, OS/Docker sandbox | all |
| **Background / async agents** | fire-and-forget tasks, status polling, off-foreground | Cursor, Windsurf, Cline, opencode, Copilot, Codex, Devin |
| **Multi-surface from one runtime** | TUI + IDE + desktop + web + headless server + SDK | opencode, Cursor, Cline, Codex, Kiro, Antigravity |
| **Git worktrees for isolation** | parallel agents get isolated worktrees | Cursor, Claude Code, Factory |
| **Context engineering** | compaction, RAG/semantic index, lazy loading, doom-loop detection | all |
| **Embeddable SDK / headless API** | runtime as a library + HTTP/ACP server for CI | Cline, opencode, Codex, Gemini, Antigravity |

---

## B. Standout Features Worth Borrowing (the differentiators)

Grouped by theme. Each notes the originator(s) and why it matters.

### B1. Skills system (SKILL.md) - Claude Code, now cross-tool standard
- Markdown + YAML frontmatter; directory name = `/command`.
- **Progressive disclosure**: only `description` loads at session start; full body
  loads on invocation (cheap context).
- **Invocation control**: `disable-model-invocation` (user-only, e.g. `/deploy`),
  `user-invocable: false` (model-only background knowledge).
- **`context: fork`**: run a skill in an isolated subagent.
- **Dynamic context injection**: `` !`git diff HEAD` `` runs a shell command at load
  and inlines the output into the prompt.
- **String substitution**: `$ARGUMENTS`, `$1`, `${SKILL_DIR}`, `${SESSION_ID}`.
- Supporting files (templates/scripts/reference docs) referenced from SKILL.md.
- WHY: the cleanest "package reusable knowledge/workflows" primitive. Portable
  across tools - skills you write work in Claude Code/Cursor/etc. too.

### B2. Hooks as deterministic enforcement - Claude Code (best impl)
- ~30 lifecycle events: SessionStart, UserPromptSubmit, PreToolUse, PostToolUse,
  PermissionRequest, SubagentStart/Stop, PreCompact/PostCompact, Stop, FileChanged...
- Handler types: `command` (shell), `http`, `mcp_tool`, `prompt` (single LLM call),
  `agent` (spawn a subagent to verify).
- **`PreToolUse` can deny even in bypass-permissions mode** - unbypassable policy.
- Matchers filter by tool name / agent / reason; regex + `Tool(param:value)`.
- KEY INSIGHT: "guardrails in hooks, not prompts." A CLAUDE.md rule is a request; a
  hook is enforcement. Use for auto-lint, auto-test, secret-blocking, validation.

### B3. Checkpoints / Rewind - opencode + Claude Code
- Snapshot files before every edit; one restore point per prompt.
- Rewind code, conversation, or both (`Esc Esc` / `/rewind` / `/undo`+`/redo`).
- opencode: git-backed file reversion + full session snapshots (roll back to any
  earlier message including files).
- Independent of git; huge UX safety net for autonomous edits.

### B4. Plan mode + Auto mode - Claude Code, Cursor, Gemini, Cline, Roo
- **Plan mode**: read-only research + propose a plan for approval before editing;
  delegates research to a Plan subagent so exploration stays out of main context.
- **Auto mode**: a background safety classifier auto-approves safe actions and blocks
  risky ones - the middle ground between approve-all and skip-permissions.
- Cursor: plan with one model, build with another; parallel plans to review.
- Gemini: collaborative external-editor plan editing + inline comments.

### B5. Subagents + orchestration - everyone; Antigravity + Devin + Factory lead
- **Context isolation + summary handoff**: subagent runs in own context, returns only
  a summary (keeps parent lean). The core pattern.
- **Dynamic subagents in 3 flavors** (Antigravity): built-in roles, generic clones
  (inherit parent prompt+env), and on-the-fly goal-defined.
- **Async background subagents** (Antigravity): intercepting client streams child
  output into parent's progress log; a poller concludes when all finish.
- **Coordinator / worker / validator** role split (Factory Missions, Devin managed
  Devins, Kiro sub-agents) with **per-role model selection**.
- **Dependency-graph scheduling in concurrent waves** (Kiro): run independent tasks
  in parallel automatically.
- **Auto worktree-per-subagent + auto cleanup** (Antigravity, Cursor).
- **Nested spawning** capped at N levels (Claude Code: 5).
- **`/best-of-n`** (Cursor): run one task across N models in isolated worktrees, compare.
- **Boomerang/Orchestrator** (Roo): decompose, delegate to specialized modes, summary
  boomerangs back; recursive subtask tree.

### B6. Memory & self-improvement - Agent Zero + Claude Code + Windsurf + Devin
- **AGENTS.md / CLAUDE.md**: persistent instructions, layered scopes (managed > user >
  project > local), directory-tree walk, `@imports`.
- **Path-scoped rules**: load only when matching files are touched (`.claude/rules`,
  `.cursor/rules/*.mdc` with `globs`).
- **Auto memory**: agent writes its own per-repo notes (MEMORY.md index + topic files)
  from your corrections, zero manual effort (Claude Code).
- **Flow awareness** (Windsurf): passively watch real-time user actions (edits,
  terminal runs, navigation, clipboard) and fold into context.
- **Self-improvement** (Agent Zero, Devin): vector memory with areas
  (main/fragments/solutions/skills), AI consolidation on save, staleness-aware
  retrieval, `memory_forget`, human curation dashboard. Agent memorizes solutions and
  creates skills/tools after solving hard problems.
- **Steering files** (Kiro): product/tech/structure context, global vs workspace scope.

### B7. Verification & artifacts (proof, not "trust me") - Cursor, Devin, Copilot, Kiro
- **Run the app + capture proof**: screenshots, screen recordings, logs attached to the
  result/PR. "The artifact is how you trust the result."
- **Self-review pass** (Copilot, Codex): run a second model over the diff before
  presenting to the user.
- **Property-based testing** (Kiro): assert invariants across all inputs (fuzz-style),
  catching edge cases unit tests miss.
- **Requirements analysis** (Kiro): automated reasoning to catch contradictions/gaps
  in requirements BEFORE coding.
- **Computer use / browser-in-loop** (Devin, Cursor, Cline, Antigravity): agent drives
  a real browser/desktop to verify its own UI work; remote-desktop takeover by human.
- **Artifacts as first-class deliverables** (Antigravity): task lists, plans,
  walkthroughs, screenshots, recordings, diffs - with inline feedback (doc comments,
  screenshot annotations) folded in WITHOUT stopping the agent.

### B8. Sandbox + approval tiers - Codex (cleanest model), Gemini, Factory, Kiro
- **Two-layer trust model** (Codex): sandbox = technical boundary (which files/network);
  approval = when to stop and ask. Keep them separate.
- **OS-native enforcement**: macOS Seatbelt (`sandbox-exec`), Linux `bwrap`+seccomp,
  Windows native sandbox / WSL2. Applies to spawned subprocesses too.
- **Sandboxing matrix** (Gemini): Seatbelt, Docker/Podman, Windows, gVisor, LXC + 
  tool-level sandboxing + **sandbox expansion requests** (just-in-time elevation).
- **Two-phase runtime** (Codex cloud): setup phase (network on, install deps) then
  agent phase (network off, secrets stripped). Great default-secure pattern.
- **Network allowlists** (Copilot/Kiro/Cursor): default-off or registry-only.
- **Autonomy levels** (Factory, Codex): read-only default, opt-in mutations, fail-fast
  on violation. Secure-by-default headless mode for CI.
- **Tiered policy engine** (Gemini): TOML rules, priority tiers (Default < Extension <
  Workspace < User < Admin), `argsPattern` regex, per-MCP and per-subagent rules.

### B9. Model routing - Devin, Kiro, Factory, Copilot, Cursor
- **Auto model selection** by task complexity/latency/cost + manual override.
- **Per-role selection**: cheap/fast model for workers/routine code, strong model for
  planning/validation/architecture.
- **Phase routing** (Gemini): high-reasoning Pro during planning, fast Flash on
  implementation.
- Avoids vendor lock-in - a core local-first value.

### B10. Custom modes / role-based personas - Roo Code
- **Custom Modes**: each bundles a prompt + tool allowlist + model. Built-ins: Code,
  Architect (read-only planning), Ask, Debug, **Orchestrator**, plus user-defined.
- "Architect" literally cannot write code; "Debug" gets a reasoning model.
- Per-mode model assignment ("sticky models"). Community Mode Gallery.
- WHY: clean, user-extensible way to specialize the agent without separate apps.

### B11. Multi-surface runtime + SDK - Cline, opencode, Codex, Antigravity
- **One runtime, many surfaces**: TUI + desktop + web + headless server + SDK + ACP
  (editor integration). Sessions persist and MOVE across surfaces.
- **SDK-first** (Cline `@cline/sdk`, Claude Agent SDK): same harness as a TS/Python
  library. Directly relevant since Mímir is TS/Bun.
- **Headless JSONL streaming** (Gemini, Codex): typed events (init/message/tool_use/
  tool_result/result) + meaningful exit codes - great for CI/automation.
- **ACP (Agent Client Protocol)**: standard for editor integration.
- **`--bare`** flag (Claude Code): skip all auto-discovery for reproducible CI.

### B12. Background tasks + Await + scheduling - Cursor, Antigravity, Cline, Copilot
- **Background-by-default** subagents; main agent keeps working.
- **`Await` tool** (Cursor): first-class primitive to block on a background job/subagent
  or wait for specific stdout ("Ready"/"Error").
- **Scheduled tasks** (Antigravity, Copilot, Cursor, Devin, Cline): cron + prompt +
  project; results stay interactive (async -> human handoff).
- **Event-driven automations**: triggers on file save, issue created, PR updated,
  webhook, Slack/PagerDuty.

### B13. Permission engine details - Claude Code, opencode, Gemini
- **Shell-AST-aware parsing**: parse compound commands (`&&`, `|`, `;`), strip wrappers
  (`timeout`, `nice`, `xargs`), canonicalize aliases; circuit-breaker for `rm -rf /`
  even in bypass mode.
- **Rule syntax**: `Tool` or `Tool(specifier)` allow/ask/deny; glob wildcards;
  `WebFetch(domain:...)`; gitignore-style path patterns.
- **Doom-loop detection** (opencode): auto-detect repeated identical tool calls.
- **Capability-based `permissions.yaml`** (Kiro): one rule allows/denies a whole
  category across all tools.

### B14. Plugins / marketplace - Claude Code, opencode, Cline, Agent Zero
- **Plugin** = bundle of skills + agents + hooks + MCP servers + LSP + settings + `bin/`.
- **Marketplaces**: curated official + reviewed community; private repos for teams.
- **MCP Marketplace** (Cline): in-app discovery/install (paste a GitHub URL).
- **AI security scanning** before plugin/skill install (Agent Zero).

### B15. Spec-driven development - Kiro (best), BMAD
- Three artifacts: `requirements.md` (user stories + **EARS notation** acceptance
  criteria), `design.md` (architecture, sequence/data-flow diagrams, TS interfaces, DB
  schemas, API endpoints), `tasks.md` (dependency-ordered, trackable).
- Feature vs Bugfix vs Quick Spec variants. Specs stay synced with the codebase.
- Approval gates between phases (Requirements -> Design -> Tasks).
- (This is exactly what BMAD does - Mímir can integrate BMAD-style workflows.)

### B16. Stretch / ambitious (later)
- **Full Linux desktop as agent workspace** (Agent Zero): real GUI the agent operates.
- **Browser DOM annotation mode** (Agent Zero): click any element to generate
  inspect/change/lift/comment directives.
- **Time Travel** (Agent Zero): workspace snapshot history with diff + revert.
- **Persistent stateful environments / BYOM** (Factory Droid Computers): environment +
  process memory persist across sessions. (For local-first, the user's machine IS this.)
- **Tab next-action prediction** (Cursor): RL-trained inline completion + cross-file
  cursor jumps. (Needs an editor surface.)
- **Chat connectors** (Cline): run/steer agents from Telegram/WhatsApp/Slack.
- **Orchestrator meta-plugin** (Agent Zero): command OTHER agent CLIs (Claude Code,
  Codex, Gemini CLI, opencode) as subordinates.
- **Image generation inside the agent** (Cursor).

### B17. Modular plugin-SDK architecture - OpenClaw (the headline structural idea)
- **Reusable agent-core package**: the agent loop, harness types, messages, compaction,
  prompts, skills, session contracts - pure and framework-agnostic.
- **Typed plugin SDK + central registry**: plugins call `api.register*(...)`
  (registerProvider, registerAgentHarness, registerCliBackend, registerChannel,
  registerTool, registerHook, registerCompactionProvider, registerHttpRoute, ...).
  One-way loading: plugin -> registry registration; core -> registry consumption.
- **Narrow SDK barrels**: plugins import `plugin-sdk/<area>` subpaths, NEVER `src/**`
  internals. The SDK path is the external contract only.
- **Harness registry (pluggable runtimes)**: built-in runtime + plugin-registered
  harnesses (e.g. a native Codex executor); `agentRuntime.id` policy selects per route.
- **5-Piece Kit** module decomposition: execution state machine, context engineering +
  recovery, tool safety, model fallback + error normalization, subagents + skills snapshots.
- **Multi-agent routing**: isolated agents (own workspace + state dir + SQLite session
  store) with bindings routing inbound surfaces to agents; per-agent sandbox + tools.
- **Two-tier hooks**: gateway/internal event hooks + plugin lifecycle hooks
  (before_model_resolve, before_prompt_build, before/after_tool_call, agent_end, ...).
- **Contract tests** assert which plugin owns which capability.
- WHY: this is what keeps a large agent codebase maintainable and extensible - the
  difference between a demo and a long-lived platform. Mímir adopts it as the backbone.

### B18. Repo-as-knowledge via gitmcp - the "Knowledge layer" idea
- **gitmcp.io** turns any GitHub repo into a remote MCP server just by swapping the
  domain: `github.com/o/r` -> `gitmcp.io/o/r`. It serves the repo's llms.txt,
  llms-full.txt, README, and docs as MCP tools/resources.
- **On-the-fly spawning is trivial**: the framework constructs the URL and connects
  via remote MCP (HTTP/SSE) - no install, no config.
- **The idea for Mímir**: a first-class "Knowledge" subsystem (its own
  `knowledge.*` namespace, NOT generic MCP) that dynamically attaches repo-docs as
  knowledge sources, caches/indexes them locally (RAG), and ships curated packs.
- WHY: gives the agent accurate, current library/framework context on demand; a
  clean differentiator and a paid hosted-knowledge angle.

### B19. Monetization models for open-source coding tools
- **Open-core + hosted AI gateway** (opencode Zen/Go): free CLI/BYOK; paid
  OpenAI-compatible gateway with curated models, unified billing, pay-as-you-go
  credits + cheap subscription. The gateway/console is the proprietary backend.
- **Cloud agent execution** (Cursor/Devin/Copilot/Kiro): local free, hosted
  sandboxes/agents paid.
- **Seat subscriptions** (Cursor/Windsurf/Kiro): per-user monthly + usage credits.
- **Team/Enterprise** (all): workspaces, SSO/RBAC, spend limits, analytics,
  self-hosted/air-gapped, compliance, support.
- **Marketplace cut**: plugins/skills marketplace revenue share.
- **Hosted knowledge** (pairs with B18): indexed/private repo knowledge as a paid tier.
- KEY: the open-source tool is the funnel; the hosted gateway + cloud + knowledge +
  enterprise are the revenue. No lock-in (works with any provider/agent) is the trust hook.

### B20. Notebook & memory system - Open Notebook (open-source NotebookLM)
- **Notebooks**: scoped knowledge workspaces grouping sources + notes + chats around a
  project/topic. Multi-notebook organization.
- **Sources (multi-modal)**: PDFs, web pages, video, audio, Office docs, text. Each is
  processed (extract/transcribe/chunk/embed) and indexed.
- **Notes**: AI-generated insights or manual notes; persistent, searchable derived knowledge.
- **RAG retrieval**: hybrid full-text + vector search across all content; context-aware
  chat grounded in the notebook; citations with source references.
- **Transformations**: reusable content-processing actions (summarize, extract, custom).
- **Podcasts**: multi-speaker (1-4) audio overviews (NotebookLM's killer feature).
- **Fine-grained context control**: choose exactly what feeds the model.
- **Tech**: Python/FastAPI + Next.js + SurrealDB + LangChain; 18+ providers via the
  Esperanto unified library (LLM + embedding + STT + TTS); full REST API; MCP integration.
- WHY for Mímir: this is the agent's long-term brain. Unifies repo-docs (gitmcp,
  B18) + behavioral memory (auto-memory) + arbitrary multi-modal knowledge into one
  searchable, RAG-grounded store. Also a paid hosted-knowledge angle.

### B21. The brain as a neural graph on SurrealDB (the "Cortex")
- **SurrealDB** (Open Notebook's DB) is multi-model: graph + vector + document +
  key-value in one database - ideal for an agent brain.
- Model knowledge as a neural graph: **neurons** = nodes (sources, notes, concepts,
  memories), **synapses** = typed edges (`RELATE`), **engrams** = durable memory traces.
- Vector fields on neurons power RAG; graph traversal adds relational context; documents
  hold content. One store for knowledge + memory + sessions.
- WHY: a genuine long-term, queryable, self-organizing brain - with a distinctive
  vocabulary (not "notebook", not copying Gemini/NotebookLM).

### B22. Goal-driven autonomous mode ("end goal" / grind-until-done)
- **ChatGPT agent mode / Devin / Cursor "grind until done" / `/goal`**: state the end
  goal; the agent builds a game plan (spec + tasks), then executes autonomously,
  verifying as it goes, and does NOT stop until the success criteria are met.
- Key pieces: define "done" up front; self-verification (tests/preview/property checks);
  optional human checkpoints; resumable state; budget/safety guardrails still apply.
- WHY: the headline "tell it what to build and walk away" experience.

### B23. Interactive preview + human annotation feedback loop
- The agent pops up a **live browser preview** of what it's building; the user **draws/
  writes/annotates** on it (markup canvas), then **sends the annotated screenshot back**
  as multimodal feedback; the agent applies the changes; iterate.
- Seen as: Antigravity screenshot annotation, Cursor Design Mode, browser-in-loop /
  computer use - but with the human redlining the preview directly.
- WHY: closes the "is this what I wanted?" loop visually and precisely.

### B24. Mobile remote access + tunneling (phone controls the home agent)
- iOS/Android apps (or PWA) connect to the home server over a **secure tunnel**
  (Tailscale / Cloudflare Tunnel / hosted relay); chat, steer, approve, view progress
  from anywhere. Same sessions as desktop (portable sessions).
- Seen as: OpenClaw channels + remote, opencode remote control, Cursor iOS, Factory
  mobile. The hosted relay/tunnel is a natural paid service; local P2P stays free.
- WHY: the agent keeps working for you while you're away.

### B25. Go for the core runtime (server-first, cross-platform)
- **Go** compiles to a single static binary for any OS/arch (`GOOS`/`GOARCH`), no bundled
  runtime; best-in-class concurrency (goroutines) for parallel agents + channels + server;
  ideal for a long-running home-server daemon that mobile clients connect to.
- Tradeoff vs TypeScript: thinner AI ecosystem - but LLM access is just OpenAI-compatible
  HTTP and an official Go MCP SDK exists, so nothing blocks it. GUI is a web frontend
  (React/Solid) served by the Go daemon (also Wails desktop + mobile/PWA).
- WHY: fits a cross-platform single-binary daemon + remote-mobile architecture better
  than a bundled JS runtime; smaller binaries, robust concurrency.

### B26. Self-learning memory engine - recovered from the user's `agence`
- `agence` = the user's prior opencode fork with a genuine self-learning subsystem.
  The ideas are sound; the vehicle broke under fork debt (sprawling monorepo, vendored
  platforms, fragile bundling). Rebuild the IDEAS as clean modules in Mímir.
- **3 memory tiers**: Memory (short ranked facts) / Knowledge (long-form wiki with
  `[[wikilinks]]`) / Heartbeat (scheduled maintenance).
- **5 layers** (LobeHub): activity, context, experience, identity, preference +
  importance rank (low/medium/high/critical).
- **Forgetting curve**: importance-scaled half-life decay + access reinforcement
  (spaced repetition) + expiry (`computeDecayScore`). Makes memory adaptive, not a dump.
- **Outcome-driven auto-capture** (regex, no LLM): preferences, corrections, and
  tool-failure -> "avoid repeating" lesson. Learning from failure.
- **Consolidation "sleep" pass**: merge dupes, prune stale. **Cross-layer associative
  links + concept map** (a semantic graph - SurrealDB's natural home).
- **Procedural memory**: `reflect` (work -> reusable skill) + `quality_gate` (failure ->
  prevention skill). **Global vs project scope** for cross-project identity.
- WHY: this IS the AGI/self-learning core. SurrealDB unifies all three tiers in one
  graph+vector store - the structural upgrade over agence's flat SQLite.

### B27. Deterministic governance - Microsoft agent-governance-toolkit
- Core thesis: prompt-level safety is "a polite request to a stochastic system."
  Intercept every tool call in deterministic middleware BEFORE it hits the wire, so
  denied actions are structurally impossible.
- **Fail-closed policy gate** (YAML/OPA/Cedar) + **privilege rings** (4 tiers) for
  tools/plugins + **require-approval for irreversible actions**.
- **Tamper-evident audit + Decision BOM** (Merkle hash-chained; record policy + request
  + rationale) - provable replay.
- **Agent SRE**: kill switch, SLOs/error budgets, circuit breakers, chaos, reversibility
  verification before acting.
- **MCP security gateway**: tool-poisoning/drift/hidden-instruction scanning.
- WHY: the mature safety/reliability layer an autonomous, self-learning agent needs.

### B28. Native sandbox control plane - vercel/sandbox
- vercel/sandbox is only a CLIENT SDK over a hosted **Firecracker microVM** control
  plane. To own it, build the control plane + runtime in the Go daemon.
- Mechanism: ephemeral microVMs; exec + log streaming; node:fs-compatible file API;
  port -> URL routing; **snapshot/fork** (pause/resume/branch); egress network policy;
  **per-user Linux isolation** (one user per agent, setgid shared dirs); warm pool
  (Firecracker boots ~125ms).
- Native build: tiered runtime (containers fast/trusted vs microVM strong/untrusted);
  in-guest agent (exec/files/PTY); local reverse proxy for port routing; snapshot
  metadata in SurrealDB.
- WHY: the two most transferable ideas are **per-user isolation** (one Linux user per
  agent) and **snapshot/fork** (branch an environment, try a change, roll back).

### B29. Built-in research & ingestion tools
- **Metasearch (SearXNG-style)**: fire one query to many engines in parallel, aggregate/
  dedupe/rank - no single-engine lock-in, privacy-respecting. Built in, not a separate app.
- **YouTube transcripts**: pull subtitles/transcripts (yt-dlp style), chunk, ingest.
- **Web scraping**: fetch + extract clean markdown from sites.
- All auto-ingest into the Cortex as neurons with provenance, available to RAG.
- WHY: broad, current, multi-source knowledge acquisition out of the box.

### B30. Build/task engine + GUI data + components + structured output (Vercel + lift)
- **Turborepo**: declarative task graph + parallel topo execution + **content-addressed
  caching** (hash inputs -> restore outputs). Fast/cheap agent-driven rebuilds.
- **SWR**: stale-while-revalidate GUI data layer (cache + dedup + background revalidate
  + optimistic updates) over the daemon's REST + push stream.
- **components.build**: accessible/composable component standard + copy-in registry +
  agent "building-components" skill.
- **AI SDK**: confirms the Go agent-runtime design (provider abstraction + tool loop +
  structured output + streaming UI protocol) - call providers directly, no gateway.
- **lift**: schema-constrained guaranteed-valid output + per-field confidence/citations.
- WHY: the supporting cast - fast builds, a responsive GUI, accessible components, and
  validated structured output.

### B31. Small-model-friendly structured workflow (<=30B)
- Small models derail in agent loops (too many tools, open-ended tasks, weak planning).
  Patterns that keep them on track:
- **Plan-first + to-do list**: force a game plan, convert to an ordered to-do list, work
  one item at a time (implement -> debug -> test -> done). The to-do list is working
  memory. (Claude Code TodoWrite; agence's extended todo with subtasks/dependencies/carry;
  Kiro tasks.)
- **Lean tool surface**: restrict to read/write/edit/bash/todowrite for small models
  (chimera "shrew" tunes Qwen3.6-35B-A3B this way; smaller step budgets).
- **Focused context + re-orientation**: load only the current task; re-inject the plan;
  doom-loop detection; step budgets.
- **Verify each step**: run tests after each item; debug before advancing.
- WHY: makes <=30B local models productive - central to the local-first/sovereign mission.

### B32. Cross-platform computer use / OS automation
- The agent controls the computer and phone: read the UI tree, click/type, inject input,
  query the OS, trigger app actions. Per-platform APIs:
- **UI automation**: Windows UIA; macOS Accessibility (AXUIElement); Linux AT-SPI2
  (D-Bus); Android AccessibilityService (AccessibilityNodeInfo + dispatchGesture); iOS
  UIAccessibility / XCUITest.
- **Low-level input hooks**: Windows SetWindowsHookEx; macOS Quartz Event Taps
  (CGEventTapCreate); Linux uinput/libevdev; Android restricted (needs root); iOS blocked
  (jailbreak only).
- **System instrumentation**: Windows WMI/COM; macOS sysctl + OSA (AppleScript/JXA);
  Linux D-Bus + /sys + /proc; Android system services / DevicePolicyManager / adb; iOS MDM.
- **AI app actions**: Windows App Actions; macOS App Intents (AppIntent/perform());
  Linux D-Bus; Android App Actions / App Functions API; iOS App Intents (AppIntent/AppEntity).
- Feasibility: desktop fully feasible; Android needs the accessibility permission (root
  for low-level); iOS is sandbox-restricted. This is what powers the visual auditor.
- WHY: full computer use across all platforms is a huge differentiator; most agents are
  desktop-only or cloud-only.

### B33. WASM modules (build in any language)
- Modules as WebAssembly: users write modules (tools, skills, MCP servers, personas,
  triggers) in any language (Rust, C, Go, AssemblyScript), compile to .wasm, and the
  framework loads them via a WASM runtime.
- For a Go framework, **wazero** (pure-Go WASM runtime, no CGO) is the natural choice.
- The **WASI Component Model** is the emerging standard interface so any-language
  modules run identically.
- WASM gives a strong sandbox for marketplace code (a real security boundary).
- Precedent: Envoy/Istio WASM filters, Fastly Compute, Cloudflare Workers, Extism,
  Figma plugins.
- WHY: a polyglot module system makes the marketplace truly open - anyone builds a
  module in their language and it just works, safely.

### B34. Forkable templates & experiments (CodePen-style)
- CodePen's core idea: a "Pen" = a small, shareable, forkable front-end experiment with
  a live preview. Plus templates, embeds, fork & remix, trending discovery.
- Transferable to Mímir: forkable/shareable templates + experiments with live previews in
  the marketplace, embeddable previews, fork & remix, trending discovery.
- CodePen's pricing (free + Pro ~$8/mo + teams) aligns with the freemium model.
- What NOT to take: the whole social front-end playground (different category). Take the
  forkable/template/embed mechanics, not the social playground.
- WHY: forkable experiments + templates make the marketplace sticky and communal.

---

## C. Per-Tool Cheat Sheet (one-liner differentiators)

| Tool | Signature idea to steal |
|---|---|
| **OpenClaw** | THE modular architecture: reusable `agent-core` package + typed plugin SDK (`api.register*()` into a central registry; narrow barrels, never internals) + harness registry (pluggable runtimes) + 5-Piece Kit decomposition + multi-agent routing + two-tier hooks + persona workspace |
| **Agent Zero** | Self-improving memory + agent creates its own tools + full Linux desktop + Time Travel snapshots + orchestrates other CLIs |
| **Open Notebook** | Open-source NotebookLM: Notebooks (scoped knowledge workspaces) + multi-modal Sources (PDF/web/video/audio/docs) + Notes + RAG (full-text + vector search) + Transformations + citations + multi-speaker podcasts; the agent's long-term brain |
| **Claude Code** | Skills + Hooks-as-enforcement + Checkpoints/Rewind + Auto mode + Auto memory + the permission engine + Agent SDK |
| **Antigravity** | Dynamic subagents (3 flavors) + async background subagents + artifacts-with-inline-feedback + scheduled tasks + projects decoupled from repos |
| **Gemini CLI** | Tiered TOML policy engine + 5-backend sandboxing matrix + Agent Skills standard + headless JSONL + task-DAG tracker |
| **opencode** | Git-backed undo/redo + permission-as-core-primitive + doom-loop detection + multi-surface from one core + AGENTS.md canonical |
| **Cursor** | Tab prediction + `/best-of-n` + `/worktree` + `Await` tool + Design Mode + Cloud agent hooks + Cursor Blame |
| **Windsurf** | Flow awareness (passive context) + local-plan->cloud-execute handoff + Agent Command Center (Kanban) + auto-Memories + Arena mode |
| **Cline** | SDK-first (one runtime: IDE+CLI+SDK) + chat connectors + Zen background + cron + MCP Marketplace + Computer Use + audit logging |
| **Roo Code** | Custom Modes (role personas) + Boomerang/Orchestrator orchestration + per-mode models + Mode Gallery |
| **Devin** | Managed parallel sub-agent fleets + Playbooks + self-improving knowledge base + Computer Use video proof + model routing |
| **Copilot agent** | Async PR-based workflow + self-review loop + built-in security scanning + custom agents as repo files + automation triggers |
| **Codex** | Sandbox+approval two-layer model + OS-native enforcement + two-phase runtime + auto-review + subagent fanout + config import |
| **Kiro** | Spec-driven (EARS) + property-based testing + requirements analysis + dependency-wave scheduling + event hooks + steering files + capability permissions |
| **Factory** | Persistent Droid Computers / BYOM + tiered autonomy + Missions (orchestrator/worker/validator) + headless secure-by-default + OTEL observability |

---

## D. Recommended Feature Set for Mímir (tiered)

### Tier 1 - Core v1 (table stakes; must ship)
- Multi-provider via AI SDK (F1)
- Agent loop: streaming + tool calling + sessions (F2)
- Built-in system tools: terminal, fs, process, code-exec, edit_block (F3)
- Permission engine: allow/ask/deny + glob + doom-loop detection (F7)
- Plan/Act separation (runtime plan mode) (new)
- Checkpoints / undo / rewind (git-backed) (new)
- AGENTS.md memory + layered config (new)
- Subagents with context isolation + summary handoff (F5)
- Skills (SKILL.md) with progressive disclosure (new)
- Hooks (lifecycle enforcement) (new)
- MCP client (optional extensibility) (F4)
- TUI (F9) + single binary (F10)
- Modular package architecture + plugin SDK + central registry (OpenClaw) (new)
- Pluggable runtimes (harness registry) + multi-agent routing (OpenClaw) (new)

### Tier 2 - Differentiators (v1.x; what makes it great)
- Custom Modes / role personas (Architect/Debug/Code/Orchestrator)
- Auto mode (LLM permission classifier)
- Auto memory (agent writes its own notes)
- Background task queue + `Await` primitive + scheduled (cron) tasks
- Git worktrees for parallel-agent isolation
- Verification + artifacts (run app, capture screenshots/logs; self-review pass)
- Model routing (auto by complexity + per-role selection)
- Spec-driven workflow with EARS + requirements analysis (F6, BMAD-aligned)
- Sandbox tiers (OS-native where possible + optional Docker)
- Headless mode + JSONL streaming + SDK + ACP (multi-surface)
- Plugins (bundle skills/hooks/agents/MCP)

### Tier 3 - Stretch / future
- Property-based testing integration
- Computer use / browser-in-loop verification
- Flow awareness (passive context from user actions)
- Best-of-n / arena mode (run across N models)
- Self-improving memory + skill/tool synthesis (Agent Zero-style)
- Tauri desktop GUI (E11)
- Chat connectors (Telegram/Slack)
- Orchestrator meta-plugin (command other agent CLIs)
- Persistent stateful environment / BYOM
- Voice input

---

## E. Sources
- Agent Zero: github.com/agent0ai/agent-zero
- Claude Code: docs.claude.com / docs.anthropic.com
- Antigravity: antigravity.google ; Gemini CLI: github.com/google-gemini/gemini-cli
- opencode: opencode.ai ; Cursor: cursor.com/changelog ; Windsurf: docs.windsurf.com
- Cline: docs.cline.bot ; Roo Code: docs.roocode.com
- Devin: devin.ai ; Copilot agent: docs.github.com ; Codex: openai.com/codex
- Kiro: kiro.dev ; Factory: factory.ai
- Agent Skills standard: agentskills.io ; MCP: modelcontextprotocol.io ; ACP: agentclientprotocol.io
- GitMCP (repo -> MCP knowledge server): gitmcp.io (source: github.com/idosal/git-mcp)
- opencode Zen/Go (monetization reference): opencode.ai/zen , opencode.ai/go
- Open Notebook (notebook + memory system reference): open-notebook.ai (source: github.com/lfnovo/open-notebook)
- SurrealDB (the brain's database): surrealdb.com
- Go (core language): go.dev ; Go MCP SDK: github.com/modelcontextprotocol/go-sdk ; Wails (Go+web desktop): wails.io ; Bubble Tea (Go TUI): github.com/charmbracelet/bubbletea
- agence (user's self-learning system, primary recovery source): github.com/David2024patton/agence
- agent-governance-toolkit (governance reference): github.com/microsoft/agent-governance-toolkit
- vercel/sandbox (sandbox mechanism): github.com/vercel/sandbox ; also vercel/swr, vercel/turborepo, vercel/ai, vercel/components.build
- SearXNG (metasearch reference): github.com/searxng/searxng ; lift (structured extraction): github.com/datalab-to/lift ; agentic-inbox (confirm-before-send): github.com/cloudflare/agentic-inbox
