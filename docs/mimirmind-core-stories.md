# Mímir - Core Epics: Stories (E1-E14)

**BMAD Artifact - Phase 7 (John, PM)**
Date: 2026-07-23 | Inputs: PRD, Architecture, UX spec, Readiness review
Format: Story / Acceptance Criteria (EARS) / Points (fibonacci) / Depends / Tasks

---

## E1 - Modular Foundation

### E1.1 - Go module + single-binary scaffold
**Story:** As a developer, I want a Go monorepo that builds one static binary per OS/arch, so that Mímir installs anywhere from one artifact.
**AC:**
- The system shall provide `cmd/mimir` building to Win/macOS/Linux (amd64+arm64) as a static binary.
- When run with no args, the system shall start the daemon; with a subcommand, run the CLI.
**Points:** 5 | **Depends:** - | **Tasks:** go.mod; cobra root cmd; build matrix (Makefile/GoReleaser); CI build job.

### E1.2 - Config loader + interpolation
**Story:** As a user, I want a `mimir.json` config (opencode-importable) with secret interpolation, so that I can configure providers and reuse existing definitions.
**AC:**
- The system shall load `mimir.json` and import `opencode.json`.
- The system shall resolve `{env:VAR}` and `{file:path}` in string values.
- If a `{file:...}` secret is missing, then the system shall fail naming the file (no silent empty key).
**Points:** 5 | **Depends:** E1.1 | **Tasks:** config schema (Go structs); JSONC parser; interpolation; opencode import; validation.

### E1.3 - Credential store (encrypted)
**Story:** As a user, I want provider keys stored encrypted, so that they aren't plaintext on disk.
**AC:**
- The system shall store keys in `auth.json` (0600) and encrypt them at rest using `MIMIR_ENCRYPTION_KEY`.
- The system shall provide `mimir auth login/logout/list`.
**Points:** 5 | **Depends:** E1.2 | **Tasks:** AES-GCM encrypt/decrypt; auth.json read/write; login flow; SurrealDB key mirror.

### E1.4 - SurrealDB managed sidecar (Spike C1)
**Story:** As a developer, I want the daemon to auto-manage a local SurrealDB, so that the Cortex is durable with zero manual setup.
**AC:**
- When the daemon starts, the system shall ensure SurrealDB is running (Docker or bundled binary) and connected.
- If SurrealDB stops, then the system shall health-check and restart it and reconnect.
**Points:** 8 | **Depends:** E1.1 | **Tasks:** sidecar manager; health check/reconnect; schema bootstrap (run the SurrealQL); Docker + binary modes. **(Spike first.)**

---

## E2 - Conversation & Agent Loop

### E2.1 - Provider abstraction + OpenAI dialect
**Story:** As a user, I want to use any OpenAI-compatible provider, so that I can plug in my key or a local model.
**AC:**
- The system shall expose a `Provider` interface with a `Dialect`.
- The system shall implement the OpenAI-compatible dialect (`/v1/chat/completions`, SSE streaming + usage).
**Points:** 8 | **Depends:** E1.3 | **Tasks:** Provider/Dialect interfaces; openaiCompatDialect; SSE parser; usage extraction; registry routing by model string.

### E2.2 - Anthropic-native dialect
**Story:** As a user, I want to use Anthropic's native API, so that Claude endpoints work without a translation proxy.
**AC:**
- The system shall implement the Anthropic dialect (`/v1/messages`, streaming deltas + usage).
- When a model is `anthropic/...`, the system shall select the Anthropic dialect.
**Points:** 5 | **Depends:** E2.1 | **Tasks:** anthropicDialect (encode/decode); message-format mapping; streaming events.

### E2.3 - Agent loop (stream + tool calling)
**Story:** As a user, I want the agent to stream and call tools in a loop, so that it can act on my requests.
**AC:**
- The system shall stream model tokens to subscribers as they arrive.
- When the model emits a tool call, the system shall run it through the policy gate, execute it, feed the result back, and loop until the model stops.
**Points:** 13 | **Depends:** E2.1, E3.1, E4.1 | **Tasks:** loop state machine (intake/assemble/infer/act/verify); tool dispatch; result feedback; stop detection; event bus.

### E2.4 - Session persistence + resume
**Story:** As a user, I want sessions saved and resumable, so that I can return to a conversation.
**AC:**
- The system shall persist every message/tool-call/result to SurrealDB.
- When the user resumes a session, the system shall restore the full transcript.
**Points:** 5 | **Depends:** E1.4, E2.3 | **Tasks:** session/message repos; resume; list/fork.

### E2.5 - Context compaction
**Story:** As a user, I want long sessions compacted, so that they don't exceed the context window.
**AC:**
- When context nears the model limit, the system shall compact older turns while re-injecting root memory + active skills.
**Points:** 5 | **Depends:** E2.4 | **Tasks:** token counting; compaction prompt; re-injection; reserve budget.

---

## E3 - Built-in System Tools

### E3.1 - Shell + process tools
**Story:** As a user, I want the agent to run terminal commands and manage processes, so that it can build/test code.
**AC:**
- The system shall provide start_process, interact_with_process, read_process_output (paginated), force_terminate, list_sessions, list_processes, kill_process.
**Points:** 8 | **Depends:** E1.1 | **Tasks:** session manager for processes; streaming output; pagination; kill by PID.

### E3.2 - Filesystem tools
**Story:** As a user, I want the agent to read/write/search files, so that it can edit my codebase.
**AC:**
- The system shall provide read_file (offset/negative-offset), write_file, read_multiple_files, create_directory, list_directory (recursive/depth), move_file, start_search/get_more/stop_search, get_file_info.
**Points:** 8 | **Depends:** E1.1 | **Tasks:** file ops; ripgrep-backed search; pagination; metadata.

### E3.3 - edit_block (search/replace)
**Story:** As a user, I want surgical edits with fuzzy fallback, so that the agent edits precisely.
**AC:**
- The system shall apply search/replace edits with fuzzy fallback and multiple-occurrence support.
**Points:** 5 | **Depends:** E3.2 | **Tasks:** exact match; fuzzy fallback (similarity); expected_replacements; diff feedback.

### E3.4 - In-memory code execution
**Story:** As a user, I want the agent to run code in memory, so that it can analyze data without littering files.
**AC:**
- The system shall execute Python/Node/R snippets in memory and return output.
**Points:** 3 | **Depends:** E3.1 | **Tasks:** temp-file execution; capture stdout/stderr; cleanup.

### E3.5 - To-do tool (small-model working memory)
**Story:** As a user (and as a small model), I want a persistent to-do list, so that work is planned, tracked, and never lost.
**AC:**
- The system shall provide todowrite/todoread with status (pending/in_progress/completed/blocked), subtasks, dependencies, tags, persisted in SurrealDB.
- The system shall provide task_search (cross-session) and todo_carry (forward to new sessions).
**Points:** 5 | **Depends:** E1.4 | **Tasks:** todo repo (the `todo` table); CRUD; ordering; the GUI/WS live feed.

---

## E4 - Permission Engine & Guardrails

### E4.1 - Policy gate (fail-closed)
**Story:** As a user, I want per-tool/per-command permissions, so that I control what the agent may do.
**AC:**
- The system shall enforce allow/ask/deny per tool and per command pattern (glob), default-deny.
- When a rule is `require_approval`, the system shall pause for user approval.
**Points:** 8 | **Depends:** E1.2 | **Tasks:** rule engine; glob matching; decision (allow/deny/ask); approval channel.

### E4.2 - Shell-AST parsing + circuit breaker
**Story:** As a user, I want compound commands parsed and catastrophic commands blocked, so that I'm protected even in bypass mode.
**AC:**
- The system shall parse `&&`/`|`/`;`, strip wrappers (timeout/nice/xargs), and canonicalize.
- If a command is `rm -rf /` or `~`, then the system shall block it even in bypass mode.
**Points:** 5 | **Depends:** E4.1 | **Tasks:** shell parser; wrapper stripping; circuit-breaker list.

### E4.3 - Audit log (tamper-evident)
**Story:** As a user, I want every tool call logged tamper-evidently, so that I can review what happened.
**AC:**
- The system shall log tool, args, decision, result, timestamp as a hash-chained (Merkle) record.
**Points:** 5 | **Depends:** E4.1, E1.4 | **Tasks:** hash chain; audit repo; Decision BOM fields.

### E4.4 - Doom-loop detection
**Story:** As a user, I want a stuck agent to stop itself, so that it doesn't spin forever.
**AC:**
- If the agent repeats an identical tool call beyond a threshold, then the system shall halt and surface the loop.
**Points:** 3 | **Depends:** E2.3 | **Tasks:** repeat detector; threshold; halt + notify.

---

## E5 - Checkpoints / Undo / Rewind

### E5.1 - File snapshots per edit
**Story:** As a user, I want files snapshotted before each edit, so that I can revert.
**AC:**
- The system shall snapshot affected files before every edit (one restore point per prompt).
**Points:** 5 | **Depends:** E3.2 | **Tasks:** snapshot store; per-prompt grouping; metadata in SurrealDB.

### E5.2 - Undo / rewind
**Story:** As a user, I want to undo code and/or conversation, so that I can roll back mistakes.
**AC:**
- When the user invokes undo/rewind, the system shall restore code, conversation, or both to a chosen point (independent of git).
**Points:** 5 | **Depends:** E5.1, E2.4 | **Tasks:** restore files; truncate/branch transcript; UI hook.

---

## E6 - Memory & Persona (Cortex foundation)

### E6.1 - AGENTS.md loading
**Story:** As a user, I want project instructions auto-loaded, so that the agent knows my conventions.
**AC:**
- The system shall load AGENTS.md (CLAUDE.md fallback) with layered scopes (managed>user>project>local) and `@imports`.
**Points:** 3 | **Depends:** E1.2 | **Tasks:** file discovery; layering; @import resolution; tree walk.

### E6.2 - Cortex store (neuron/synapse/engram)
**Story:** As a developer, I want the neural graph store, so that knowledge and memory have a home.
**AC:**
- The system shall implement neuron/synapse/engram CRUD over SurrealDB, with vector + full-text indexes.
**Points:** 8 | **Depends:** E1.4 | **Tasks:** repos; embedding field; MTREE + BM25 indexes; RELATE edges.

### E6.3 - Outcome-driven auto-capture
**Story:** As a user, I want corrections/preferences/failures remembered automatically, so that I don't repeat myself.
**AC:**
- When the user issues a correction/preference, the system shall capture an engram (no extra LLM call).
- When a tool fails, the system shall store an activity record + an "avoid repeating" experience lesson.
- Where an item is identity/preference/critical, the system shall store it in `__global__` scope.
**Points:** 5 | **Depends:** E6.2 | **Tasks:** regex detectors; engram writer; global vs project scope.

### E6.4 - Persona files
**Story:** As a user, I want per-agent personas, so that different agents behave differently.
**AC:**
- The system shall load per-agent SOUL.md/USER.md/IDENTITY.md into the system prompt.
**Points:** 2 | **Depends:** E6.1 | **Tasks:** persona discovery; prompt injection.

---

## E7 - Skills System

### E7.1 - SKILL.md loading + progressive disclosure
**Story:** As a user, I want reusable skills, so that I can package workflows.
**AC:**
- The system shall load SKILL.md (frontmatter + markdown), exposing only descriptions at start and the body on invocation.
**Points:** 5 | **Depends:** E1.2 | **Tasks:** skill discovery; frontmatter parse; lazy body load; tiered paths.

### E7.2 - Invocation control + fork
**Story:** As a user, I want to control how skills run, so that some are user-only and some run isolated.
**AC:**
- The system shall support user-only vs model-only skills and `context: fork` (run in an isolated subagent).
**Points:** 3 | **Depends:** E7.1, E9.1 | **Tasks:** invocation flags; fork into subagent; substitution ($ARGUMENTS, ${SKILL_DIR}).

---

## E8 - Plan Mode & Permission Modes

### E8.1 - Permission-mode cycling
**Story:** As a user, I want to cycle permission modes, so that I can choose how autonomous the agent is.
**AC:**
- The system shall cycle default/acceptEdits/plan/auto/bypass.
**Points:** 3 | **Depends:** E4.1 | **Tasks:** mode state; gate behavior per mode; UI toggle.

### E8.2 - Plan mode (read-only + plan subagent)
**Story:** As a user, I want a read-only plan mode, so that the agent researches before editing.
**AC:**
- While in plan mode, the system shall forbid edits/bash, delegate research to a Plan subagent, and propose a plan for approval.
**Points:** 5 | **Depends:** E8.1, E9.1 | **Tasks:** plan-mode gate; plan subagent; plan presentation + approval.

---

## E9 - Subagents (core)

### E9.1 - Subagent spawning (isolated context)
**Story:** As a user, I want the agent to spawn subagents, so that it can parallelize work.
**AC:**
- The system shall spawn subagents with isolated context that return only a summary.
- The system shall cap nesting depth.
**Points:** 8 | **Depends:** E2.3 | **Tasks:** subagent runner (goroutine); isolated session; summary return; depth cap.

### E9.2 - Built-in subagent types
**Story:** As a user, I want ready-made subagents, so that common tasks are one call away.
**AC:**
- The system shall provide explore (read-only search), general (multi-step), and plan (research) subagents.
**Points:** 3 | **Depends:** E9.1 | **Tasks:** type definitions; tool restrictions per type.

---

## E10 - Two-Tier Hooks

### E10.1 - Gateway / internal hooks
**Story:** As a user, I want command + lifecycle hooks, so that I can automate around the agent.
**AC:**
- The system shall run gateway/internal hooks on command + lifecycle events (session start/stop, etc.).
**Points:** 5 | **Depends:** E2.3 | **Tasks:** hook registry; event dispatch; command hooks.

### E10.2 - Plugin lifecycle hooks
**Story:** As a developer, I want plugin lifecycle hooks, so that plugins can intercept the loop.
**AC:**
- The system shall support before/after tool call, agent_end, compaction, and session hooks.
- Where a PreToolUse hook denies, the system shall block the call even in bypass mode.
**Points:** 5 | **Depends:** E10.1, E11.1 | **Tasks:** lifecycle hook points; deny-override; matcher (tool/agent/reason).

---

## E11 - Plugin SDK & Registry

### E11.1 - MimirPluginApi + PluginRegistry (freeze contract - C2)
**Story:** As a developer, I want a typed plugin SDK + central registry, so that I can extend Mímir without touching internals.
**AC:**
- The system shall expose `MimirPluginApi` (RegisterProvider/Tool/Hook/Channel/Harness/Capability/Command) into a central PluginRegistry.
- Plugins shall import only `@mimirmind/plugin-sdk`, never core internals.
**Points:** 8 | **Depends:** E1.1 | **Tasks:** api interface; registry; one-way load; **freeze + document the contract (C2).**

### E11.2 - Plugin loader
**Story:** As a developer, I want plugins loaded from a manifest, so that I can install extensions.
**AC:**
- The system shall load in-process Go plugins via a manifest (id, version, contracts).
**Points:** 5 | **Depends:** E11.1 | **Tasks:** manifest parse; Go plugin load; registration call.

### E11.3 - Contract tests
**Story:** As a developer, I want contract tests, so that plugins register exactly what they declare.
**AC:**
- The system shall assert each plugin registers exactly the capabilities its manifest declares.
**Points:** 3 | **Depends:** E11.2 | **Tasks:** ownership assertions; CI gate.

---

## E12 - Primary GUI + TUI

### E12.1 - Daemon HTTP + WebSocket server
**Story:** As a developer, I want a local REST + WS API, so that any client (GUI/TUI/mobile) can drive the daemon.
**AC:**
- The system shall serve REST (projects/sessions/cortex/config/auth/tasks/sandboxes) and WS (token/tool/lifecycle/status events).
**Points:** 8 | **Depends:** E2.3 | **Tasks:** HTTP router; REST handlers; WS hub; event projection.

### E12.2 - GUI shell (5-region layout)
**Story:** As a user, I want the desktop GUI shell, so that I have a clear workspace.
**AC:**
- The GUI shall render the static top bar, collapsible left nav, tabbed center workspace, right chat panel, and static far-right quick-launch rail (per UX spec).
**Points:** 13 | **Depends:** E12.1 | **Tasks:** Solid app; layout regions; theming/tokens; routing; responsive collapse.

### E12.3 - Chat panel (stream + tools + approvals + to-do)
**Story:** As a user, I want a rich chat panel, so that I can interact and stay oriented.
**AC:**
- The chat shall stream tokens, show expandable tool-call cards, inline approval prompts, and the live to-do list.
**Points:** 8 | **Depends:** E12.2, E2.3, E3.5 | **Tasks:** message rendering; tool cards; approval card; to-do widget; SWR-style data layer.

### E12.4 - TUI (Bubble Tea)
**Story:** As a power user, I want a TUI, so that I can use Mímir over SSH/headless.
**AC:**
- The system shall provide a Bubble Tea TUI with chat, streaming, tool visibility, approvals, and a to-do rail.
**Points:** 8 | **Depends:** E12.1 | **Tasks:** Bubble Tea app; panes; leader-key commands; palette.

---

## E13 - Packaging & Distribution

### E13.1 - Cross-platform build + installer
**Story:** As a user, I want a simple install, so that I can get running fast.
**AC:**
- The system shall ship a single binary per platform with an installer.
**Points:** 5 | **Depends:** E1.1 | **Tasks:** GoReleaser; installers (Win/mac/Linux); first-run wizard.

### E13.2 - Self-update + version check
**Story:** As a user, I want Mímir to update itself, so that I stay current.
**AC:**
- The system shall check for new versions and self-update on request.
**Points:** 3 | **Depends:** E13.1 | **Tasks:** version check; download + replace; rollback.

---

## E14 - MCP Integration

### E14.1 - MCP client (stdio + remote)
**Story:** As a user, I want to connect external MCP servers, so that I can add tools.
**AC:**
- The system shall support MCP clients over stdio + remote HTTP/SSE, registered via config with per-server enable/disable/env/timeout.
**Points:** 8 | **Depends:** E11.1 | **Tasks:** Go MCP SDK client; stdio + SSE transports; config; lifecycle.

### E14.2 - MCP tool surfacing
**Story:** As a user, I want MCP tools alongside native tools, so that the agent can use them transparently.
**AC:**
- The system shall surface external MCP tools to the agent in the tool registry (core system access stays native).
**Points:** 3 | **Depends:** E14.1, E3.1 | **Tasks:** tool adapter; schema mapping; lazy schema load.

---

## Velocity note
Total Core estimate: ~230 story points. At a notional ~40 pts/sprint, the Core is ~6
sprints. Sprint 1 (below) targets the walking skeleton (E1 + thin E2/E3/E12 path).
