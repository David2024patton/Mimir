# Mímir - Product Requirements Document (PRD)

**BMAD Artifact - Phase 3 (John, Product Manager)**
Date: 2026-07-23 | Status: Draft for review | Feeds: Architecture (Winston)
Source: `mimirmind-product-brief.md`, `PROJECT-MASTER-PLAN.md` (F1-F30, epics E1-E47)
Scope of this PRD: **Core (Tier 1) epics E1-E14.** Tier 2/3 are summarized in section 9.

---

## 1. Overview

Mímir is a local-first, self-learning agentic coding framework with a neural knowledge
brain (the Cortex), built in Go, that runs on the user's own hardware. This PRD defines
the requirements for the Core release (Tier 1): a usable, safe, modular agent with
built-in system tools, permissions, memory, skills, plan mode, subagents, hooks, a
plugin SDK, a GUI + TUI, packaging, and MCP integration.

**Acceptance criteria use EARS** (Easy Approach to Requirements Syntax):
- *Ubiquitous:* "The system shall ..."
- *Event-driven:* "When <trigger>, the system shall ..."
- *State-driven:* "While <state>, the system shall ..."
- *Optional:* "Where <feature>, the system shall ..."
- *Unwanted:* "If <condition>, then the system shall ..."

---

## 2. Personas (from the brief)

- **P1 - The Sovereign Builder** (primary): technical, privacy-conscious, self-hosts,
  wants full ownership + BYOK + local-first data.
- **P2 - The AGI Tinkerer**: building toward self-learning/AGI; wants memory/learning/
  governance primitives exposed and hackable.
- **P3 - The Mobile Delegator**: wants to steer/approve work from phone or chat apps.
- **P4 - The Team** (later): shared knowledge, spend controls, SSO, governance.

---

## 3. Key User Journeys

**J1 - First run to first task (P1):** Install the single binary -> `mimir auth login`
(paste a provider key or point at a local model) -> open a project -> ask the agent to
make a change -> review the diff -> approve -> done.

**J2 - The "it remembers" moment (P1/P2):** In session 1 the user corrects the agent
("always use tabs, not spaces"). In session 2 (days later) the agent applies tabs
without being told - the correction was captured as an engram and recalled.

**J3 - Grounded research (P1/P2):** The user adds a repo + a YouTube talk + a web page
to the Cortex. The agent answers a question citing those sources (RAG with citations).

**J4 - Safe autonomy (P2):** The user sets a goal. The agent plans, executes in a
sandbox, hits a risky command -> the policy gate prompts for approval -> the user
approves -> the agent continues until the goal's success criteria are met.

---

## 4. Functional Requirements by Epic (Core / Tier 1)

### E1 - Modular Foundation
**US1.1** As a developer, I want a single Go binary that runs the daemon + CLI, so that
I can install Mímir on any OS/arch from one artifact.
- AC1.1.1: The system shall cross-compile to Windows, macOS, and Linux (amd64 + arm64)
  as a single static binary with no external runtime dependency.
- AC1.1.2: When invoked with no subcommand, the system shall start the daemon and serve
  the GUI; when invoked with a subcommand, it shall run the CLI.

**US1.2** As a user, I want to configure providers in an opencode-compatible JSON file,
so that I can reuse my existing provider/MCP definitions.
- AC1.2.1: The system shall load `mimir.json` (and import `opencode.json`) with
  `{env:VAR}` and `{file:path}` secret interpolation.
- AC1.2.2: If a referenced `{file:...}` secret is missing, then the system shall fail
  with a clear error naming the file (not a silent empty key).

**US1.3** As a user, I want my provider keys stored securely, so that they are not
plaintext in config.
- AC1.3.1: The system shall store keys in an `auth.json`-style file with 0600 perms and
  encrypt keys at rest in SurrealDB.

### E2 - Conversation & Agent Loop
**US2.1** As a user, I want streaming responses with tool calling, so that I see the
agent think and act in real time.
- AC2.1.1: The system shall stream model tokens to the UI as they arrive.
- AC2.1.2: When the model emits a tool call, the system shall execute it (via the tool
  registry), feed the result back, and loop until the model stops.

**US2.2** As a user, I want sessions persisted and resumable, so that I can return to a
conversation later.
- AC2.2.1: The system shall persist every message/tool-call/result to SurrealDB.
- AC2.2.2: When the user resumes a session, the system shall restore the full transcript.

**US2.3** As a user, I want context compacted when the window fills, so that long
sessions don't fail.
- AC2.3.1: When the context approaches the model limit, the system shall compact older
  turns while re-injecting root memory + active skills.

### E3 - Built-in System Tools
**US3.1** As a user, I want the agent to run terminal commands, so that it can build and
test my code.
- AC3.1.1: The system shall provide shell tools: start_process, interact_with_process,
  read_process_output (paginated), force_terminate, list_sessions.
- AC3.1.2: The system shall provide process tools: list_processes, kill_process.

**US3.2** As a user, I want the agent to read/write/search files, so that it can edit my
codebase.
- AC3.2.1: The system shall provide filesystem tools: read_file (offset/negative-offset),
  write_file, read_multiple_files, create_directory, list_directory (recursive, depth),
  move_file, start_search/get_more_search_results/stop_search, get_file_info.
- AC3.2.2: The system shall provide edit_block with search/replace, fuzzy fallback, and
  multiple-occurrence support.

**US3.3** As a user, I want the agent to run code in memory, so that it can analyze data
without littering files.
- AC3.3.1: The system shall execute Python/Node/R snippets in memory and return output.

### E4 - Permission Engine & Core Guardrails
**US4.1** As a user, I want per-tool/per-command permissions, so that I control what the
agent may do.
- AC4.1.1: The system shall enforce allow/ask/deny per tool and per command pattern (glob).
- AC4.1.2: The system shall parse compound shell commands (`&&`, `|`, `;`), strip
  wrappers, and circuit-break `rm -rf /`/`~` even in bypass mode.

**US4.2** As a user, I want an audit log of all tool calls, so that I can review what
happened.
- AC4.2.1: The system shall log every tool call (tool, args, result code, timestamp) to a
  rotating audit log.

**US4.3** As a user, I want doom-loop detection, so that a stuck agent stops itself.
- AC4.3.1: If the agent repeats an identical tool call beyond a threshold, then the system
  shall halt and surface the loop to the user.

### E5 - Checkpoints / Undo / Rewind
**US5.1** As a user, I want to undo the agent's changes, so that I can revert mistakes.
- AC5.1.1: The system shall snapshot files before every edit (one restore point per prompt).
- AC5.1.2: When the user invokes undo/rewind, the system shall restore code, conversation,
  or both, to a chosen point (independent of the user's git history).

### E6 - Memory & Persona (Cortex foundation)
**US6.1** As a user, I want project instructions auto-loaded, so that the agent knows my
conventions.
- AC6.1.1: The system shall load AGENTS.md (CLAUDE.md fallback) with layered scopes
  (managed > user > project > local) and `@imports`.

**US6.2** As a user, I want the agent to remember my corrections/preferences, so that I
don't repeat myself.
- AC6.2.1: When the user issues a correction or preference, the system shall capture it as
  an engram (no extra LLM call) and recall it in future sessions.
- AC6.2.2: Where an item is identity/preference/critical, the system shall store it in the
  `__global__` cortex so it transfers across projects.

**US6.3** As a user, I want per-agent persona files, so that different agents behave
differently.
- AC6.3.1: The system shall support per-agent SOUL.md / USER.md / IDENTITY.md persona files.

### E7 - Skills System
**US7.1** As a user, I want reusable skills (SKILL.md), so that I can package workflows.
- AC7.1.1: The system shall load SKILL.md files (frontmatter + markdown) with progressive
  disclosure (description at start; body on invocation).
- AC7.1.2: The system shall support invocation control (user-only vs model-only) and
  `context: fork` (run in an isolated subagent).

### E8 - Plan Mode & Permission Modes
**US8.1** As a user, I want a read-only plan mode, so that the agent researches before
editing.
- AC8.1.1: While in plan mode, the system shall forbid edits/bash and delegate research to
  a Plan subagent, then propose a plan for approval.
- AC8.1.2: The system shall cycle permission modes (default/acceptEdits/plan/auto/bypass).

### E9 - Subagents (core)
**US9.1** As a user, I want the agent to spawn subagents, so that it can parallelize work.
- AC9.1.1: The system shall spawn subagents with isolated context that return only a summary.
- AC9.1.2: The system shall provide built-in subagent types: explore (read-only search),
  general (multi-step), plan (research).

### E10 - Two-Tier Hooks
**US10.1** As a user, I want lifecycle hooks, so that I can automate around the agent.
- AC10.1.1: The system shall support gateway/internal hooks (command + lifecycle events)
  and plugin lifecycle hooks (before/after tool call, agent_end, compaction, session).
- AC10.1.2: Where a PreToolUse hook denies, the system shall block the call even in bypass
  mode (unbypassable enforcement).

### E11 - Plugin SDK & Registry
**US11.1** As a developer, I want a typed plugin SDK, so that I can extend Mímir without
touching internals.
- AC11.1.1: The system shall expose `api.register*(...)` (provider, tool, hook, channel,
  harness, capability) into a central PluginRegistry.
- AC11.1.2: The system shall load plugins via documented `@mimirmind/plugin-sdk` barrels;
  plugins shall not import core internals.
- AC11.1.3: The system shall include contract tests asserting which plugin registers which
  capability.

### E12 - Primary GUI + TUI
**US12.1** As a user, I want a desktop GUI with a clear layout, so that I can work
comfortably.
- AC12.1.1: The GUI shall provide a static top bar, a collapsible left nav, a center
  workspace, a right chat panel, and a static far-right quick-launch rail.
- AC12.1.2: The GUI shall be a web app served by the Go daemon (also packaged as a Wails
  desktop app).

**US12.2** As a power user, I want a TUI, so that I can use Mímir over SSH/headless.
- AC12.2.1: The system shall provide a Bubble Tea TUI with chat, streaming, tool-call
  visibility, and approval prompts.

### E13 - Packaging & Distribution
**US13.1** As a user, I want a simple install, so that I can get running fast.
- AC13.1.1: The system shall ship a single binary per platform with an installer, version
  check, and self-update.

### E14 - MCP Integration
**US14.1** As a user, I want to connect external MCP servers, so that I can add tools.
- AC14.1.1: The system shall support MCP clients (stdio + remote HTTP/SSE) registered via
  config, with per-server enable/disable/env/timeout.
- AC14.1.2: The system shall surface external MCP tools to the agent alongside native tools
  (core system access remains native, not MCP).

---

## 5. Non-Functional Requirements

- **NFR1 Performance:** tool dispatch < 50ms overhead; GUI streaming latency < 200ms.
- **NFR2 Reliability:** sessions survive daemon restart; crash-safe persistence.
- **NFR3 Security:** keys encrypted at rest; fail-closed permission gate; audit log.
- **NFR4 Portability:** single static binary for Win/macOS/Linux (amd64/arm64).
- **NFR5 Privacy:** local-first; no telemetry by default; user owns all data.
- **NFR6 Extensibility:** every capability reachable via the plugin SDK.
- **NFR7 Concurrency:** parallel agents/subagents via goroutines without races
  (serialized per-session lane).

---

## 6. MoSCoW Prioritization (Core)

- **Must:** E1, E2, E3, E4, E11, E12 (foundation, loop, tools, permissions, plugins, GUI).
- **Should:** E5, E6, E7, E8, E9 (checkpoints, memory, skills, plan mode, subagents).
- **Could:** E10, E14 (hooks, MCP).
- **Won't (this release):** Tier 2/3 (see section 9).

---

## 7. Assumptions & Dependencies

- **A1:** LLM providers expose OpenAI-compatible HTTP APIs (no SDK lock-in).
- **A2:** A Go MCP SDK is available and maintained.
- **A3:** SurrealDB runs as a companion service (embedded or sidecar).
- **A4:** Docker is available for sandboxing (F28, Tier 2); Core uses permission gates only.
- **D1:** Provider keys supplied by the user (BYOK) or a local model (Ollama).

---

## 8. Open Questions

- **Q1:** SurrealDB embedded-in-Go vs sidecar process for the default install?
- **Q2:** Default GUI framework: React vs Solid? (Both viable; decide in architecture.)
- **Q3:** Which local model is the default "zero-config" experience (Ollama bundle?)?
- **Q4:** Telemetry: fully off by default, or opt-in anonymous usage?

---

## 9. Out of Scope (this PRD; deferred)

**Tier 2 (Differentiators, v1.x):** pluggable runtimes/harness registry (E15), governance
& deterministic safety (E44), native sandbox runtime (E45), the full self-learning memory
engine (E43), multi-agent routing (E16), custom modes (E17), advanced orchestration (E19),
background tasks/automation (E20), spec-driven workflow + BMAD (E21), verification/
artifacts + model routing (E22), sandbox tiers (E23), multi-surface runtime (E24), research
& ingestion tools / Gjallarhorn (E46), task/build engine (E47), plugins marketplace (E25).

**Tier 3 (Stretch):** property-based testing + computer use (E26), self-improving memory
synthesis (E27), best-of-n/arena (E28), messaging channels (E29), mobile remote (E42),
Tauri desktop + voice (E30), podcasts/audio overviews (E39), hosted platform (E32-E37).

---

## 10. Next BMAD Step

Hand off to **Winston (Architect)** for `bmad-architecture`: turn this PRD + the master
plan into an architecture document with ADRs (Go module layout, SurrealDB schema for the
Cortex, the plugin-SDK contract, the agent-loop state machine, the GUI<->daemon protocol),
then `bmad-check-implementation-readiness` to gate the build.
