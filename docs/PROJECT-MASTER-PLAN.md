# Mímir - Project Master Plan

Product name: **Mímir** (domain mimirmind.com - reserved pending build completion).
Status: Concept complete; entering BMAD workflow (Product Brief next).
Date: 2026-07-23

---

## 0. Identity & Mythology

Named for **Mímir** of Norse mythology - "the rememberer" / "the wise one" - who
guards **Mímisbrunnr**, the Well of Wisdom beneath **Yggdrasil** (the World Tree).
The well grants cosmic knowledge to those who drink from it; Odin sacrificed an eye
for a single draught. After Mímir was beheaded by the Vanir, Odin preserved his head
with magic so it would keep giving counsel - wisdom that outlives its keeper. Mímir
drinks from the well daily using the **Gjallarhorn**.

This is the perfect myth for a self-learning agent: it drinks from a well of knowledge,
remembers everything, and keeps advising long after any single session ends. The
mythology maps directly onto the architecture and gives every major subsystem a name.

### Mythology -> architecture
| Norse myth | Mímir architecture |
|---|---|
| **Mímir** ("the rememberer / wise one") | The framework - the self-learning agent that remembers |
| **Mímisbrunnr** (the Well of Wisdom) | The **Cortex** - the SurrealDB knowledge brain, the well of knowledge |
| **Drinking from the well grants knowledge** | **RAG retrieval + ingestion** - the agent "drinks" to learn |
| **Yggdrasil** (World Tree; roots reach the well) | The **knowledge graph** - neurons/synapses linking all knowledge, rooted in the well |
| **Gjallarhorn** (the horn Mímir drinks from) | The **research & ingestion tools** (F29): metasearch, YouTube transcripts, web scraping - the instruments that draw knowledge in |
| **Mímir's preserved head** (counsel after death) | **Persistent memory / engrams** (F26) - knowledge that survives sessions/restarts and keeps advising |
| **Odin's sacrificed eye** (the price of wisdom) | The **cost of knowledge** - compute/tokens spent to learn; the trust/permission the user grants |
| **Mímir as Odin's advisor** | The agent as the user's counselor |
| **Runes** (Odin won them on Yggdrasil) | **Skills** (F11) - hard-won, reusable procedural knowledge |
| **The Nine Realms** | Different **projects/agents** the knowledge spans (global vs project scope, F26.5) |

### Naming scheme
- **Mímir** - the product / brand name (ASCII: Mimir). Domain: **mimirmind.com**.
  Technical namespace (packages, config, gateway subdomains): **mimirmind**
  (e.g. `@mimirmind/agent-core`, `mimirmind.json`, `gateway.mimirmind.com`).
- **The Well (Mímisbrunnr)** - the Cortex knowledge store (SurrealDB)
- **Yggdrasil** - the knowledge-graph layer
- **Gjallarhorn** - the research/ingestion tool suite (F29)
- **Mímir's Head (Höfuð)** - the persistent memory engine (F26)
- Data units keep the neural vocabulary: **neurons** (nodes), **synapses** (edges),
  **engrams** (durable memories) - the brain's cells, held in the Well, connected by
  the Tree.

### Agent personas (Norse)
Mímir's agent modes and subagents take Norse names. (The BMAD planning personas -
Mary/John/Winston/Sally/Amelia - stay separate: those are the *process*, these are
the *product*.)

| Norse figure | Mímir role |
|---|---|
| **Odin** | Orchestrator / planner - the All-Father; makes the game plan, sees the whole board |
| **Thor** | Builder - the workhorse; writes code, runs tools, executes |
| **Loki** | Debugger / tester - cunning; hunts bugs and edge cases |
| **Heimdall** | Reviewer / watchman - sees everything; code review + the policy gate |
| **Bragi** | Skald - documentation, comments, clean output |
| **Huginn & Muninn** | Scout subagents - "Thought" & "Memory"; explore / research |
| **Ratatoskr** | Messenger - inter-agent communication (runs up and down Yggdrasil) |
| **The Norns** (Urd/Verdandi/Skuld) | Spec / planning - past / present / future |
| **Forseti** | Arbiter - permission/policy decisions, conflict resolution |

> The story writes itself: Mímir drinks from the Well (Mímisbrunnr) through the
> Gjallarhorn, grows its Yggdrasil of connected neurons, lays down engrams, and - like
> Mímir's preserved head - keeps giving counsel long after each session ends.

---

## 1. Executive Summary

Build a standalone, self-hosted agentic coding application that combines the
strongest ideas from the current generation of AI dev tools:

- **opencode** - broad multi-provider model support, MCP integration, agent/subagent
  system, skills, plugins, config-driven design.
- **Desktop Commander** - deep local system access: terminal control, filesystem
  operations, process management, in-memory code execution, audit logging, safety
  guardrails.
- **Kiro (AWS)** - spec-driven development: prompts become requirements, design,
  and sequenced tasks before code; property-based correctness checks.
- **Google Antigravity** - parallel agent orchestration, dynamic subagents,
  scheduled (cron) background tasks, artifacts as deliverables.
- **IBM Bob** - approval modes / guardrails, "literate coding" (natural language to
  code in context), shell/CI integration.
- **Cursor** - fast editor-grade UX, tab completion, agent + MCP.

The app is built on proven open foundations rather than from scratch:
the **Vercel AI SDK** for providers and the **Bun** runtime, with a full
**Desktop-Commander-class tool set built in natively** (terminal, filesystem,
process control, in-memory code execution). The **Model Context Protocol (MCP) SDK**
is included only as an optional extensibility layer for OTHER third-party tools -
system access is first-class native code, not an external dependency. This delivers
"all the providers opencode has" plus deep local system access in a clean, owned codebase.

---

## 2. Competitive Landscape: What to Borrow

| Tool | Borrow this | Skip / differentiate |
|---|---|---|
| **OpenClaw** | **THE ARCHITECTURAL REFERENCE** (see Section 5): modular package structure (`@openclaw/agent-core`), typed plugin SDK (`api.register*()` into a central registry; plugins import narrow barrels, never internals), harness registry (pluggable agent runtimes), the "5-Piece Kit" module decomposition, multi-agent routing (per-agent workspace/session/auth + bindings), two-tier hooks, persona workspace (SOUL.md/USER.md/IDENTITY.md) | Very broad surface (chat channels, voice, nodes); we borrow the architecture, not every surface |
| **opencode** | Provider registry pattern, MCP client, agent loop, skills/plugins, JSON config, credential store (`auth.json` pattern) | TUI-only UX is limiting; we add a richer UI later |
| **Desktop Commander** | Tool set used as the REFERENCE SPEC for our built-in tools: `start_process`, `interact_with_process`, `read_process_output`, `list_processes`, `kill_process`, filesystem read/write/search, `edit_block` (search/replace with fuzzy fallback), in-memory code exec (Python/Node/R), config management, audit logging, symlink + blocklist guardrails | We do NOT depend on it. We reimplement this tool set natively in our own codebase. |
| **Claude Code** | Skills (SKILL.md), hooks-as-enforcement, checkpoints/rewind, plan + auto mode, auto-memory, the permission-rule engine, Agent SDK | Anthropic-only by default |
| **Agent Zero** | Self-improving vector memory, agent creates its own tools, full Linux desktop workspace, Time Travel snapshots, orchestrates other CLIs | Docker-heavy, Python |
| **Gemini CLI** | Tiered TOML policy engine, 5-backend sandboxing matrix, Agent Skills standard, headless JSONL streaming, task-DAG tracker | Gemini-only models |
| **Kiro** | Spec-driven pipeline (requirements -> design -> tasks -> code), property-based testing, parallel subagents, AGENTS.md / Skills / MCP / ACP compatibility | Heavyweight enterprise pricing model |
| **Antigravity** | Agent orchestration UI, dynamic subagents, scheduled cron tasks, artifacts, projects (scoped folders + permissions), voice input | Closed Google ecosystem |
| **IBM Bob** | Approval modes (review before apply), shell/CI integration, analytics on agent contributions | Enterprise/legacy (RPG/COBOL) focus |
| **Cursor** | Editor UX quality, inline diff review, tab completion | Proprietary, closed |

> Full 15-tool feature breakdown (OpenClaw, Agent Zero, Claude Code, Antigravity,
> Gemini CLI, opencode, Cursor, Windsurf, Cline, Roo, Devin, Copilot, Codex, Kiro,
> Factory) is in `RESEARCH-FEATURE-LANDSCAPE.md`. **OpenClaw is the primary
> architectural model** for the modular design in Section 5.

### Competitive parity & edge (gap analysis)
Beyond the differentiators (self-learning memory, sovereignty, small-model mode,
governance), Mímir must match the table-stakes features users expect from Cursor /
Claude Code / Copilot / Kiro / Windsurf. The gap analysis added F34-F40:
- **Code Intelligence / LSP** (F34) - a coding agent that can't see compile errors is
  blind; every competitor has this.
- **Git & PR Automation** (F35) - the "assign an issue, get a PR" autonomous-developer
  story (Copilot agent / Devin / BugBot) is a headline selling point.
- **Deep Codebase Indexing** (F36) - AST-aware understanding beats naive text RAG.
- **Usage & Cost Tracking** (F37) - builds trust and demonstrates the gateway's value.
- **Context Mastery** (F38) - output styles, steering files, flow awareness, prompt
  caching, /context visualization: the context-engineering edge.
- **Inline Tab Completion** (F39) - Cursor's moat; an adoption driver (needs an editor).
- **Sharing & Collaboration** (F40) - share links + team sessions for growth.

**Key decision:** Desktop Commander's tool set is the proven design we mirror, but
we REIMPLEMENT it natively rather than depending on the external MCP server. The
tools (terminal, filesystem, process, code-exec, edit_block) are first-class code in
our app - faster, no subprocess/IPC overhead, fully under our control, and they work
with zero external servers configured. MCP remains available purely as an optional
way to add OTHER third-party tools later.

---

## 3. Product Vision and Scope

### Vision
A local-first, model-agnostic agentic coding companion that can plan (spec-driven),
execute (terminal + filesystem + code), orchestrate (parallel subagents), and
automate (scheduled tasks) - while keeping the human in control via approval modes.

### In scope (v1)
- Multi-provider model access (all major providers via AI SDK).
- Conversational agent loop with streaming + tool calling.
- Built-in system tools (native, no external dependency): shell/terminal, filesystem,
  process management, in-memory code exec, edit_block.
- Optional MCP client for connecting OTHER third-party tool servers (extensibility only).
- Subagents / parallel task execution.
- Spec-driven workflow: brief -> spec -> tasks -> implementation.
- Permission/approval modes + command blocklist + audit log.
- TUI interface (primary), config file, credential store.
- Single-binary distribution.

### Out of scope (v1, deferred to later epics)
- Full desktop GUI (Tauri shell) - Epic 11.
- Editor/IDE-grade inline completion (Cursor-style) - future.
- Cloud sandbox execution (Kiro/Antigravity web) - future.
- Voice input - future.
- Enterprise analytics dashboard (Bobalytics-style) - future.

### Non-goals
- Not a hosted SaaS. Local-first, user owns their keys and data.
- Not a sandbox by default (mirror Desktop Commander's stance: guardrails, not a
  sandbox; offer Docker isolation as an option).

---

## 4. Recommended Tech Stack and Foundation

| Layer | Choice | Rationale |
|---|---|---|
| **Core language / runtime** | **Go** (daemon, agent loop, tools, sandbox, transport, gateway, channels, server) | Compiles to a single static binary for ANY OS/arch (`GOOS`/`GOARCH`) with no bundled runtime; best-in-class concurrency (goroutines) for parallel agents + channels + server; ideal for a long-running home-server daemon that mobile clients connect to. Smaller, more robust binaries than a bundled JS runtime for this server-first design. |
| **GUI / frontend** | **TypeScript + React (or Solid)** web app served by the Go daemon; **Wails** to package as a desktop app | The UI (top bar, nav, chat, quick-launch, preview) is a web app - so it runs locally in a browser, as a desktop app (Wails), AND remotely (mobile/PWA) against the same Go server. TS lives in the frontend, not the core. |
| **TUI (alternative)** | **Bubble Tea** (Charm) | Native Go terminal UI for power users / headless / SSH. |
| **Mobile** | **React Native / Flutter** or responsive **PWA**, over a secure tunnel | iOS + Android apps that talk to the home Go daemon remotely (see F25). |
| **AI / Providers** | **OpenAI-compatible HTTP** called directly from Go (+ `go-openai`-style libs); provider registry pattern | LLM calls are just HTTP to OpenAI-compatible endpoints - no SDK lock-in. One registry routes to OpenAI/Anthropic/Google/OpenRouter/xAI/Mistral/Groq/Alibaba-DashScope/Ollama/local, etc. (same breadth as opencode via the compatible-mode API). |
| **MCP (optional)** | **Go MCP SDK** (`github.com/modelcontextprotocol/go-sdk`) | MCP client for OTHER third-party tool servers. Core system access is native, not via MCP. |
| **System access (BUILT IN)** | Native Go tools (`os/exec`, `os`, process APIs) | Desktop-Commander-class tool set in Go: terminal, filesystem, process mgmt, code-exec, edit_block. No external dependency. |
| **Agent orchestration** | Custom Go (goroutines + channels), opencode-style agent/subagent pattern | Task delegation, parallel goroutines, context isolation per subagent. |
| **Storage / the "brain"** | **SurrealDB** (companion service) | Multi-model: graph + vector + document + key-value in one DB. Powers the Cortex (neurons/synapses/engrams), RAG embeddings, sessions, audit. Open Notebook's choice; Rust-fast, embeddable. |
| **Sandboxing** | **Docker containers spun up on the fly** + OS-native fallback (Seatbelt/bwrap/Windows) | Agent creates/destroys isolated sandboxes dynamically (see F7.12). |
| **Config** | JSON/JSONC (opencode-style) + `{env:VAR}` / `{file:path}` interpolation | Familiar, version-controllable; prefer `{file:...}` for secrets. |
| **Credential store** | `auth.json`-style file (0600) + encrypted keys in SurrealDB | Per-provider keys, decoupled from config. |
| **Packaging** | `GOOS/GOARCH` cross-compile + Wails bundler (desktop) | One static binary per platform; desktop app via Wails; mobile via app stores/PWA. |

### Foundation decision: build fresh in Go (recommended)
Two viable paths:

1. **Fork opencode** - instant features, but you inherit ~2000-file TypeScript complexity
   and a TUI-first design that doesn't match your server + mobile vision.
2. **Build fresh in Go, with native tools** (RECOMMENDED) - clean, owned codebase that
   compiles to a single binary for any system and runs as a home-server daemon. You
   borrow opencode's proven PATTERNS (provider config schema, agent loop, credential
   store, plugin SDK) and reimplement the tool set natively in Go.

Recommendation: **Path 2**. Go fits the direction: a cross-platform single-binary
daemon that parallelizes agents, hosts the web GUI, exposes an API for mobile apps,
and connects chat channels - all from one process. The tradeoff (vs TypeScript) is a
thinner AI ecosystem, but LLM access is just OpenAI-compatible HTTP and a Go MCP SDK
exists, so nothing blocks you. A thin "compatibility layer" imports opencode-style
`opencode.json` provider/MCP configs so existing definitions work on day one.

---

## 5. System Architecture

```
+---------------------------------------------------------------+
|                         UI LAYER                               |
|   TUI (Ink/opentui)   |   (later) Tauri desktop shell         |
+---------------------------------------------------------------+
                              |
+---------------------------------------------------------------+
|                      CORE AGENT RUNTIME                        |
|  Conversation Engine  |  Agent Loop  |  Subagent Orchestrator  |
|  Session/Message Store (SQLite)  |  Spec Workflow Engine        |
+---------------------------------------------------------------+
            |                         |                    |
+----------------------+   +---------------------+   +------------------+
|   PROVIDER LAYER     |   |    TOOL LAYER       |   |  SAFETY LAYER    |
| Vercel AI SDK        |   | Native tools:       |   | Permission modes |
| @ai-sdk/* providers  |   |  - shell/terminal   |   | Command blocklist|
| Credential store     |   |  - filesystem       |   | Symlink guard    |
| (auth.json)          |   |  - process mgmt     |   | Audit log        |
|                      |   |  - code exec        |   | Approval gates   |
|                      |   |  - edit_block       |   |                  |
|                      |   | (all built in,      |   |                  |
|                      |   |  no external dep)   |   |                  |
|                      |   | Optional MCP client |   |                  |
|                      |   |  for 3rd-party tools|   |                  |
+----------------------+   +---------------------+   +------------------+
```

### Modular Package Architecture (OpenClaw-style)

The defining idea we borrow from OpenClaw: a clean, package-based modular
architecture where the agent runtime is a reusable core, everything else is a
swappable module, and third parties extend the app through a documented plugin
SDK - never by reaching into internals. This is what "do what OpenClaw does with
modules" means for Mímir.

**Core principles**
- **Reusable agent core package** (`@mimirmind/agent-core`): the agent loop,
  harness types, messages, compaction helpers, prompt templates, skills, and
  session-storage contracts. Pure, framework-agnostic, unit-testable.
- **One-way plugin loading**: plugin module -> registry registration; core runtime
  -> registry consumption. Plugins never mutate core globals; core reads a central
  `PluginRegistry`.
- **Narrow SDK barrels**: plugins import `@mimirmind/plugin-sdk/<area>` subpaths
  (plugin-entry, provider, channel, runtime, capability, memory...). They NEVER
  import `src/**` internals. The SDK path is the external contract only.
- **Typed registration contract** (`MímirPluginApi`): plugins call
  `api.register*(...)` to contribute capabilities (see F19).
- **Contract tests** assert ownership: which plugin registers which capability.

**The "5-Piece Kit" module decomposition** (the five modules that separate a
demo-grade agent from a long-running system - our core is built around these):
1. **Execution state machine**: `run` / `attempt` / `subscribe` + active-runs
   registry. Serialized per-session lane + optional global lane (no tool/session races).
2. **Context engineering + recovery**: guards, transcript hygiene, compaction
   retries, timeout snapshots.
3. **Tool safety**: layered policy + human approvals for risky execution.
4. **Model fallback + error normalization**: structured attempts, fallback chains,
   abort semantics.
5. **Subagents + skills snapshots**: isolated runs, recovery, hot-update of skills.

**Harness registry (pluggable agent runtimes)**: a built-in runtime (id
`mimirmind`) plus plugin-registered harnesses (e.g. a `codex` or `claude-code`
native executor). Runtime policy is model/provider-scoped (`agentRuntime.id`);
`auto` selects a registered harness that supports the resolved route, else falls
back to the built-in runtime. A harness owns low-level execution of one prepared
turn (native threads, compaction, resume id) while the core keeps the channel,
transcript mirror, tool policy, and approvals.

**Multi-agent routing**: run multiple isolated agents in one process. Each agent
has its own workspace (files + persona), state dir (auth profiles, model registry,
per-agent config), and SQLite session store. Bindings route an inbound surface
(TUI session, chat channel, API) to an agent. Per-agent sandbox + tool allow/deny.

**The Cortex - Knowledge & Memory subsystem** (unifies F15 + F21 + F22; stored in
SurrealDB): the agent's long-term **brain**, modeled as a neural graph. Knowledge
items are **neurons** (nodes), relationships are **synapses** (edges), and durable
memories are **engrams**. Every agent owns a scoped **cortex** that links multi-modal
sources (repo-docs via gitmcp, web pages, PDFs, video, audio, code), derived notes,
and behavioral memory (AGENTS.md, auto-memory) into one graph. A processing pipeline
(extract -> transcribe -> chunk -> embed) feeds hybrid full-text + vector **RAG**, so
chat is grounded in the cortex with citations. SurrealDB's graph + vector + document
model maps exactly: neurons = records, synapses = `RELATE` edges, engrams = memory
records, embeddings = vector fields. Local-first, encrypted at rest.

### Proposed project layout (Go)
```
mimirmind/
  cmd/
    mimirmind/               # main entrypoint (daemon + CLI dispatch)
  internal/
    agent/                   # agent loop (run/attempt/subscribe state machine)
      loop.go                # core loop
      harness.go             # harness registry + selection policy
      subagent.go            # subagent orchestration (goroutines)
      compaction.go          # context compaction
      prompts.go             # prompt templates
    llm/                     # provider registry + OpenAI-compatible transport
      registry.go            # provider registry + routing
      credentials.go         # auth.json + encrypted key store
      models.go              # model catalog + limits + routing
      transport.go           # OpenAI-compatible HTTP client (streaming)
    tools/                   # built-in native system tools
      registry.go            # tool registry (native + optional MCP)
      shell.go               # terminal: start/interact/read/kill
      filesystem.go          # read/write/list/search/move/edit_block
      process.go             # list/kill processes
      codeexec.go            # in-memory python/node/r
      web.go                 # web_search + web_fetch
      sandbox.go             # on-the-fly Docker/OS sandbox lifecycle
    cortex/                  # the brain: knowledge + memory (SurrealDB)
      neuron.go              # neurons (nodes): sources, notes, concepts, memories
      synapse.go             # synapses (edges/relationships)
      engram.go              # engrams (durable memories)
      ingest.go              # processing pipeline (extract/transcribe/chunk/embed)
      rag.go                 # hybrid full-text + vector retrieval
      store.go               # SurrealDB client + queries
    plugins/                 # PluginRegistry + load pipeline + contract tests
    sdk/                     # plugin SDK contracts (register* API)
    agents/                  # multi-agent routing, per-agent workspace/session
    sessions/                # session manager, lanes/queues, persistence
    hooks/                   # two-tier hooks (gateway + plugin lifecycle)
    safety/                  # permissions, blocklist, audit
    spec/                    # spec-driven workflow + goal/autonomous mode
    skills/                  # skills discovery + snapshots
    server/                  # HTTP/WebSocket gateway + REST API + GUI serving
    channels/                # chat channel adapters (Discord/Slack/Telegram...)
    tunnel/                  # secure remote access (Tailscale/Cloudflare/relay)
  gui/                       # TypeScript + React/Solid web frontend
    src/
      shell/                 # top bar, left nav, right chat, quick-launch rail
      preview/               # browser preview + annotation canvas
  mobile/                    # React Native/Flutter app (or PWA in gui/)
  plugins/                   # bundled plugins (Go packages using sdk)
  desktop/                   # Wails packaging (Go + gui -> desktop app)
  mimirmind.json             # default config
  go.mod
```

---

## 6. Feature Requirements (grouped)

### F1. Provider & Model Access [Core]
- F1.1 Configure providers via JSON (opencode-compatible schema).
- F1.2 Support all major providers via native API dialects: **OpenAI-compatible**
  (`/v1/chat/completions` - OpenAI, OpenRouter, xAI, Mistral, Groq, Cerebras, DeepInfra,
  Together, Alibaba/DashScope, Bedrock, Azure, Perplexity, Cohere, Ollama/local, etc.)
  AND **Anthropic-native** (`/v1/messages`). Extensible to other dialects (e.g. Gemini)
  via plugins.
- F1.3 Custom OpenAI-compatible endpoints (baseURL + apiKey) for self-hosted/local
  (Ollama, llama.cpp, vLLM) and token-plan endpoints.
- F1.4 Per-provider credential store (`auth.json`), `auth login`-style flow.
- F1.5 Model metadata: context/output limits, modalities, reasoning/tool-call flags.
- F1.6 `{env:VAR}` and `{file:path}` secret interpolation (empty-string fallback fixed).
- F1.8 Pluggable API dialects/encoders: the provider abstraction supports multiple
  request/response formats (OpenAI-compatible + Anthropic-native built in; more via
  plugins), so both OpenAI and Anthropic endpoints work natively.
- F1.7 Model routing [Differentiator]: auto-select by complexity/latency/cost; per-role
  model assignment (cheap for workers, strong for planning/validation); manual override.

### F2. Conversation & Agent Loop [Core]
- F2.1 Streaming responses with tool calling (gather -> act -> verify loop).
- F2.2 Multi-turn sessions persisted to SQLite (JSONL-style for portability).
- F2.3 Session resume / list / branch / fork.
- F2.4 Context compaction when window fills (re-attach recent skills + root memory).
- F2.5 System prompt assembly (AGENTS.md + skills injection).
- F2.6 Effort levels (low/medium/high/max) + reasoning toggle.

### F3. Built-In System Tools [Core] (Desktop-Commander-class, reimplemented natively)
- F3.1 Shell: `start_process`, `interact_with_process`, `read_process_output`
  (paginated), `force_terminate`, `list_sessions`.
- F3.2 Process: `list_processes`, `kill_process`.
- F3.3 Filesystem: `read_file` (with offset/negative-offset), `write_file`,
  `read_multiple_files`, `create_directory`, `list_directory` (recursive, depth),
  `move_file`, `start_search`/`get_more_search_results`/`stop_search`, `get_file_info`.
- F3.4 Editing: `edit_block` (search/replace, fuzzy fallback, multiple occurrences).
- F3.5 Code exec: run Python/Node/R in memory without saving files.
- F3.6 Optional rich formats: Excel/PDF/DOCX read (defer if heavy).
- F3.7 Web tools: `web_search` + `web_fetch`.

#### Built-in tool inventory (the full set - Desktop Commander parity and beyond)
- **Terminal / process:** `bash`, `start_process`, `interact_with_process`,
  `read_process_output`, `force_terminate`, `list_sessions`, `list_processes`,
  `kill_process`.
- **Files:** `read_file`, `write_file`, `read_multiple_files`, `create_directory`,
  `list_directory`, `move_file`, `start_search` / `get_more_search_results` /
  `stop_search`, `get_file_info`, `edit_block`.
- **Code:** `code_exec` (Python/Node/R in-memory), `sandbox_create` / `sandbox_exec` /
  `sandbox_destroy` (F28 - sandbox any kind of code).
- **Web / research:** `web_search` (metasearch), `web_fetch` (scrape),
  `youtube_transcript`, plus the Gjallarhorn ingestion pipeline (PDF / audio / video /
  Office / images, F29).
- **Agent / workflow:** `todo` (todowrite/todoread), `question` (F45), `task` (spawn
  subagent), `cortex_search` / `cortex_add` (RAG knowledge).
- **Computer use / OS automation (F48):** `screen_read` (UI tree), `ui_click` /
  `ui_type` (click/type UI elements), `screenshot`, `input_inject`, `os_query`
  (WMI/sysctl/D-Bus), `app_action` (App Intents / App Actions) - across Windows, macOS,
  Linux, Android, iOS.

### F4. MCP Integration [Core, optional extensibility only]
NOTE: Core system access (F3) is native and does NOT use MCP. MCP here is purely
for adding OTHER third-party tool servers.
- F4.1 Optional MCP client supporting stdio + remote (HTTP/SSE) servers.
- F4.2 Config-driven server registration (opencode-compatible `mcp` block).
- F4.3 Surface external MCP tools to the agent alongside the built-in native tools.
- F4.4 Per-server enable/disable, env, timeout. Lazy-load schemas to save context.

### F5. Agent Orchestration & Subagents [Core isolation; Differentiator advanced]
- F5.1 Subagents with strict isolated context + summary-only handoff (parent stays lean). [Core]
- F5.2 Parallel task execution (fiber pool). [Core]
- F5.3 Built-in subagent types: explore (read-only search), general (multi-step), plan (research). [Core]
- F5.4 Custom subagents via markdown + frontmatter (tools/model/permission/skills). [Diff]
- F5.5 Dynamic subagents: clones (inherit parent) + on-the-fly goal-defined. [Diff]
- F5.6 Coordinator/worker/validator role split with per-role models. [Diff]
- F5.7 Dependency-graph scheduling in concurrent waves. [Diff]
- F5.8 Git worktree per subagent + auto cleanup (parallel changes never collide). [Diff]
- F5.9 Nested spawning capped at N levels. [Diff]
- F5.10 (Stretch) best-of-n: run one task across N models in isolated worktrees, compare.

### F6. Spec-Driven Workflow [Differentiator] (Kiro/BMAD inspired)
- F6.1 Three artifacts: requirements.md (EARS acceptance criteria), design.md
  (architecture, interfaces, DB schemas, API endpoints), tasks.md (dependency-ordered).
- F6.2 Requirements analysis: model pass to catch contradictions/gaps BEFORE coding.
- F6.3 Approval gates between phases (Requirements -> Design -> Tasks).
- F6.4 Implement tasks from a spec with an agent; track completion.
- F6.5 (Stretch) property-based correctness checks (invariants across all inputs).
- F6.6 BMAD integration: run BMAD workflows (brief/PRD/architecture/stories) in-app.

### F7. Safety, Permissions & Sandboxing [Core engine; Differentiator sandbox]
- F7.1 Permission engine: allow/ask/deny per tool + per command pattern (glob). [Core]
- F7.2 Shell-AST-aware parsing: parse `&&`/`|`/`;`, strip wrappers, canonicalize aliases;
  circuit-breaker for `rm -rf /` even in bypass mode. [Core]
- F7.3 Command blocklist + symlink traversal prevention. [Core]
- F7.4 Audit log of all tool calls (rotating). [Core]
- F7.5 Doom-loop detection (repeated identical tool calls). [Core]
- F7.6 Approval gates before destructive actions. [Core]
- F7.7 Sandbox tiers: OS-native (Windows sandbox/WSL2, macOS Seatbelt, Linux bwrap)
  where possible + optional Docker/devcontainer. [Diff]
- F7.8 Two-phase runtime for high-risk work: setup (network on) then agent (network
  off, secrets stripped). [Diff]
- F7.9 Network allowlists (default-off / registry-only). [Diff]
- F7.10 Autonomy levels (read-only default, opt-in mutations, fail-fast). [Diff]
- F7.11 Capability-based permissions (one rule covers a whole tool category). [Diff]
- F7.12 On-the-fly sandboxing: the agent dynamically creates/destroys isolated sandboxes
  (Docker containers, or OS-native) per risky task - `sandbox_create` / `sandbox_exec` /
  `sandbox_destroy` tools; ephemeral or persistent dev environment the agent returns to
  (Factory Droid / Devin VM style). [Diff]

### F8. Background Tasks & Automation [Differentiator]
- F8.1 Background-by-default subagents; main agent keeps working.
- F8.2 `Await` primitive: block on a background job/subagent or wait for specific stdout.
- F8.3 Scheduled (cron) tasks: cron + prompt + project; results stay interactive.
- F8.4 Event-driven automations: triggers on file save, lifecycle, webhook.
- F8.5 Artifacts: agent-produced deliverables (plans, diffs, screenshots, logs).

### F9. UI & Multi-Surface Runtime [Core GUI; Differentiator multi-surface]
- F9.1 Primary GUI: web app served by the Go daemon (also packaged as desktop via
  Wails). Layout: [Core]
  - **Static top bar**: app title, project/agent switcher, status, search, account.
  - **Left nav menu** (collapsible/closable): projects, cortex/knowledge, sessions,
    agents, skills, settings.
  - **Center workspace**: files, editor, terminal, spec/tasks, live preview.
  - **Right chat panel**: talk to the LLMs (streaming, tool-call visibility, approvals).
  - **Far-right quick-launch rail** (static): favorite tools/agents/commands, one-click.
- F9.2 TUI (Bubble Tea) as an alternative / power-user / headless / SSH interface. [Core]
- F9.3 Headless mode: one-shot + JSONL streaming events + meaningful exit codes (CI). [Diff]
- F9.4 SDK / REST + WebSocket API: the harness as a library + server (the GUI and mobile
  app both consume it). [Diff]
- F9.5 ACP server: editor integration (VS Code/JetBrains/Zed). [Diff]
- F9.6 Portable sessions that move across surfaces (desktop <-> mobile <-> TUI). [Diff]
- F9.7 Messaging channels (OpenClaw-style modules): Discord, Slack, Telegram, etc. - each
  a plugin that routes messages to an agent (see E29). [Diff]
- F9.8 Mobile remote apps (iOS + Android) over a secure tunnel (see F25). [Diff]
- F9.9 GUI data layer (SWR-style, from vercel/swr): key-based cache + request dedup +
  stale-while-revalidate; render cached sandbox/agent state instantly, refresh from the
  daemon in the background; revalidate on tab focus / WebSocket reconnect; optimistic
  updates (stop sandbox, approve tool). Backed by the daemon's push stream. [Core]
- F9.10 Accessible component design system (from components.build): composable,
  WCAG-accessible, themeable components (shadcn/Radix-style), code owned in-repo;
  keyboard-navigable GUI. [Core]

### F10. Packaging & Distribution [Core]
- F10.1 Single-binary build (`bun build --compile`) for Win/macOS/Linux.
- F10.2 Installer + version check + self-update.

### F11. Skills System [Core] (SKILL.md, cross-tool standard)
- F11.1 SKILL.md = YAML frontmatter + markdown; directory name = `/command`.
- F11.2 Progressive disclosure: only description loads at start; body on invocation.
- F11.3 Invocation control: user-only vs model-only skills.
- F11.4 `context: fork`: run a skill in an isolated subagent.
- F11.5 Dynamic context injection: `` !`cmd` `` inlines shell output at load.
- F11.6 String substitution: `$ARGUMENTS`, `$1`, `${SKILL_DIR}`, `${SESSION_ID}`.
- F11.7 Supporting files (templates/scripts/reference) referenced from SKILL.md.
- F11.8 Tiered discovery: user + project + `.agents/skills` interop alias.
- F11.9 `building-components` skill (from components.build): the agent scaffolds
  accessible, standard-conformant UI components into target projects on demand.

### F12. Hooks / Lifecycle Enforcement [Differentiator]
- F12.1 Events: SessionStart, UserPromptSubmit, PreToolUse, PostToolUse,
  PermissionRequest, SubagentStart/Stop, PreCompact/PostCompact, Stop, FileChanged.
- F12.2 Handler types: command (shell), http, mcp_tool, prompt (single LLM call),
  agent (spawn a verifier subagent).
- F12.3 PreToolUse can deny even in bypass mode (unbypassable policy).
- F12.4 Matchers: filter by tool/agent/reason; regex + `Tool(param:value)`.
- F12.5 Hooks definable in settings, plugins, or skill/agent frontmatter.
- F12.6 Use for auto-lint, auto-test, secret-blocking, validation (enforcement, not guidance).
- F12.7 Two tiers (OpenClaw): gateway/internal hooks (event scripts for commands +
  lifecycle) + plugin lifecycle hooks (before_model_resolve, before_prompt_build,
  before_agent_reply, agent_end, before/after_compaction, before/after_tool_call,
  tool_result_persist, message_received/sending/sent, session_start/end).

### F13. Checkpoints / Undo / Rewind [Core]
- F13.1 Snapshot files before every edit; one restore point per prompt.
- F13.2 Git-backed file reversion + conversation rewind (code, conversation, or both).
- F13.3 Full session snapshots: roll back to any earlier message including files.
- F13.4 Independent of the user's git history.

### F14. Plan Mode & Auto Mode [Core plan; Differentiator auto]
- F14.1 Plan mode: read-only research + propose plan for approval before editing;
  research delegated to a Plan subagent. [Core]
- F14.2 Cycle permission modes (default/acceptEdits/plan/auto/bypass). [Core]
- F14.3 Auto mode: background safety classifier auto-approves safe actions, blocks risky. [Diff]
- F14.4 `/goal`: keep working across turns until a completion condition holds. [Diff]
- F14.5 Goal-driven autonomous mode ("end goal"): state the goal -> auto game plan ->
  execute and verify until the goal is met. See F23. [Diff]

### F15. Memory & Context [Core AGENTS.md; Differentiator auto-memory]
- F15.1 AGENTS.md canonical (CLAUDE.md fallback); layered scopes (managed > user >
  project > local); directory-tree walk; `@imports`. [Core]
- F15.2 Path-scoped rules: load only when matching files are touched (globs). [Diff]
- F15.3 Auto memory: agent writes its own per-repo notes (index + topic files). [Diff]
- F15.4 Steering files: product/tech/structure context, global vs workspace scope. [Diff]
- F15.5 (Stretch) Self-improving vector memory: areas (solutions/skills), AI
  consolidation, staleness-aware retrieval, human curation UI.
- F15.6 Persona/identity workspace files (per-agent): SOUL.md (who the agent is),
  USER.md (about the human), IDENTITY.md, GOALS.md (OpenClaw-style).

### F16. Custom Modes / Role Personas [Differentiator] (Roo-style)
- F16.1 Built-in modes: Code, Architect (read-only planning), Ask, Debug, Orchestrator.
- F16.2 Each mode bundles prompt + tool allowlist + model.
- F16.3 User-defined modes; per-mode ("sticky") model assignment.
- F16.4 Mode gallery (share/import community modes).

### F17. Verification & Artifacts [Differentiator]
- F17.1 Verification loop: run the app/tests, capture proof (screenshots, logs,
  recordings) attached to the result.
- F17.2 Self-review pass: run a second model over the diff before presenting to user.
- F17.3 Artifacts as first-class deliverables (plans, task lists, walkthroughs, diffs).
- F17.4 Inline feedback on artifacts (comments) folded in without stopping the agent.
- F17.5 (Stretch) Computer use / browser-in-loop: agent drives a real browser to verify UI.
- F17.6 Schema-constrained output (from lift): guaranteed-valid structured output
  (validate model JSON against a schema before acting) + per-field confidence +
  citations/provenance.

### F18. Plugins & Extensibility Packaging [Differentiator]
- F18.1 Plugin = bundle of skills + agents + hooks + MCP servers + settings + `bin/`.
- F18.2 Namespaced skills (`/plugin:skill`) to prevent conflicts.
- F18.3 Marketplace: curated + community; install from path/URL/git.
- F18.4 (Stretch) AI security scanning before plugin/skill install.

### F19. Modular Architecture & Plugin SDK [Core backbone] (OpenClaw-style)
- F19.1 Reusable `@mimirmind/agent-core` package (loop, harness types, messages,
  compaction, prompts, skills, session contracts).
- F19.2 Monorepo of packages: agent-core, llm, tools, plugin-sdk, core, tui, cli.
- F19.3 Central `PluginRegistry`; one-way loading (plugins register, core consumes).
- F19.4 Typed `MímirPluginApi` with `api.register*(...)`: registerProvider,
  registerModelCatalogProvider, registerAgentHarness, registerCliBackend,
  registerChannel, registerTool, registerHook, registerEmbeddingProvider,
  registerSpeechProvider, registerImageGenerationProvider, registerWebSearchProvider,
  registerWebFetchProvider, registerCompactionProvider, registerHttpRoute,
  registerCommand, registerMemoryBackend.
- F19.5 Narrow SDK barrels (`@mimirmind/plugin-sdk/<area>`); plugins never import `src/**`.
- F19.6 Plugin manifest + load pipeline (in-process native; optional subprocess/HTTP/MCP).
- F19.7 Contract tests asserting plugin ownership of capabilities.
- F19.8 The "5-Piece Kit" module decomposition as the core skeleton.

### F20. Pluggable Runtimes & Multi-Agent Routing [Differentiator] (OpenClaw-style)
- F20.1 Harness registry: built-in runtime (`mimirmind`) + plugin-registered harnesses.
- F20.2 Runtime policy: model/provider-scoped `agentRuntime.id`; `auto` selects a
  supporting harness else built-in.
- F20.3 A harness owns low-level turn execution (native threads, compaction, resume);
  core keeps channel/transcript/policy/approvals.
- F20.4 Multi-agent routing: multiple isolated agents per process.
- F20.5 Per-agent: workspace (files + persona), state dir (auth/model registry/config),
  SQLite session store.
- F20.6 Bindings: route an inbound surface (TUI/chat/API) to an agent; per-agent
  sandbox + tool allow/deny.
- F20.7 Persona workspace files: AGENTS.md, SOUL.md (identity), USER.md (the human),
  IDENTITY.md, GOALS.md - per-agent.
- F20.8 Session concurrency: serialized per-session lane + optional global lane.

### F21. Knowledge Layer (gitmcp-powered) [Differentiator]
A first-class "Knowledge" subsystem - NOT lumped under generic MCP. The framework
turns any GitHub repo into an agent knowledge base on the fly via gitmcp.io.
- F21.1 A knowledge source = a GitHub repo whose docs (llms.txt, llms-full.txt,
  README, docs/) are exposed to the agent.
- F21.2 On-the-fly spawning: the framework constructs the gitmcp URL and connects
  via remote MCP (HTTP/SSE) with zero manual config:
  `github.com/o/r` -> `gitmcp.io/o/r`; `o.github.io/r` -> `o.gitmcp.io/r`;
  any repo -> `gitmcp.io/docs`.
- F21.3 Dedicated `knowledge.*` tool namespace (e.g. `knowledge.fetch`,
  `knowledge.search`, `knowledge.list`, `knowledge.add`) - presented separately
  from generic `mcp.*` tools.
- F21.4 Dynamic loading mid-session: the agent or user adds knowledge on demand
  ("learn modular/mojo" -> auto-spawn `gitmcp.io/modular/mojo`).
- F21.5 Config: a `knowledge:` block listing repos + auto-discovery from
  imports/dependencies + per-source enable/disable.
- F21.6 Local cache/index of fetched docs for fast retrieval (optional RAG over
  the repo docs; embeddings + sqlite-vec).
- F21.7 Curated built-in knowledge packs (popular libraries/frameworks).
- F21.8 Architecture: a registered capability (`registerKnowledgeProvider`) in the
  plugin registry; uses the MCP client internally but abstracts it behind the
  `knowledge.*` namespace.
- F21.9 (Stretch / monetization) Hosted knowledge bases: indexed/cached/private
  repo knowledge on Mímir servers; curated packs; team-shared knowledge.
- F21.10 Naming: functional category "Knowledge" (namespace `knowledge.*`);
  thematic product-name candidates for brainstorming: Grimoire, Lore, Codex,
  Atlas, Library. (OpenClaw uses "Lore".)

### F22. The Cortex - Knowledge & Memory Brain (Open Notebook-inspired) [Differentiator]
The agent's long-term brain, stored in SurrealDB and modeled as a neural graph (NOT a
"notebook" - we use our own brain metaphor so we're not copying NotebookLM/Gemini).
Merges F15 memory + F21 gitmcp knowledge with Open Notebook's content model.
- F22.1 Cortex: each agent owns a scoped knowledge brain (tied to its workspace,
  F20.5). Multiple cortices for multiple projects/topics.
- F22.2 Neurons (nodes): the knowledge items - sources, notes, concepts, memories.
  Multi-modal sources: repo-docs (gitmcp, F21), web pages, PDFs, video, audio,
  Office docs, code, plain text.
- F22.3 Synapses (edges): typed relationships between neurons (references, derives-from,
  relates-to, contradicts) - SurrealDB `RELATE` graph edges, traversable.
- F22.4 Engrams (durable memories): the agent's persistent memories (F15 auto-memory,
  self-improving) stored as neurons flagged as memory traces.
- F22.5 Ingestion pipeline: extract text (PDF/Office), transcribe (audio/video via STT),
  chunk, embed. Pluggable embedder (provider or local). Each source becomes neurons.
- F22.6 Retrieval (RAG): hybrid full-text + vector search + graph traversal across the
  cortex; context-aware chat grounded in the cortex; citations with source references.
- F22.7 Notes: AI-generated insights or manual notes - neurons that are themselves
  searchable and linked via synapses.
- F22.8 Transformations: reusable content-processing actions (summarize, extract,
  custom) - implemented as skills/hooks (F11/F12), create new neurons.
- F22.9 Fine-grained context control: choose exactly which neurons/cortex feed a chat.
- F22.10 Storage: SurrealDB (graph + vector + document) as a companion service;
  encrypted at rest. Neurons = records, synapses = edges, embeddings = vector fields.
- F22.11 (Stretch) Podcast / audio overviews: multi-speaker (1-4) audio summaries of a
  cortex (NotebookLM-style) via TTS providers.
- F22.12 (Monetization) Hosted cortex = paid tier (extends E37): synced, indexed,
  private, team-shared knowledge brains.
- F22.13 Naming: subsystem = **Cortex** (recommended); items = neurons / synapses /
  engrams. Alternatives: Synapse, Engram, Mind, Cerebrum, Hippocampus (brainstorming).
- F22.14 The Cortex is the substrate for the self-learning memory engine (F26): the
  3-tier Memory/Knowledge/Heartbeat model, the 5 memory layers, and the concept graph
  all live as neurons/synapses/engrams in SurrealDB.

### F23. Goal-Driven Autonomous Mode [Differentiator] (ChatGPT "end goal" / Devin)
- F23.1 State the end goal ("build me X that does Y"); the agent generates a game plan
  (requirements -> design -> tasks, via F6) before executing.
- F23.2 Autonomous execution loop: work through the plan, verifying each step (run
  tests, check the app), and DO NOT stop until the goal's success criteria are met.
- F23.3 Success criteria + self-verification: define "done" up front; the agent checks
  against it (tests, preview, property checks) before declaring completion.
- F23.4 Optional human checkpoints: pause for approval at phase boundaries or risky
  steps; otherwise grind autonomously ("grind until done").
- F23.5 Progress visibility: live plan/task status, artifacts, and a resumable state so
  a long goal survives restarts.
- F23.6 Budget/safety guardrails: time/step/spend limits + the permission engine (F7)
  still apply during autonomous runs.

### F24. Interactive Preview & Annotation Loop [Differentiator]
- F24.1 The agent pops up a live browser preview of what you're working on (the app/site
  it's building), in the GUI preview panel or a detached window.
- F24.2 The user draws/writes/annotates directly on the preview (markup canvas: arrows,
  text, highlights, freehand) - like redlining a design.
- F24.3 Send it back: the annotated screenshot (image + markup text) is sent to the LLM
  as multimodal feedback; the agent applies the requested changes.
- F24.4 Iterate: preview -> annotate -> send -> agent edits -> re-preview, until right.
- F24.5 Built on browser-in-loop / computer-use (F17.5) + multimodal input; pairs with
  the GUI preview panel (F9.1) and verification (F17).

### F25. Mobile Remote Access [Differentiator]
- F25.1 iOS + Android apps (React Native/Flutter, or responsive PWA) that connect to the
  home Go daemon so you can talk to it / it works for you while you're away.
- F25.2 Secure tunnel from device to home server: Tailscale / Cloudflare Tunnel / a
  hosted relay; encrypted, authenticated, paired (rotating code / QR).
- F25.3 Full chat + steering from mobile: send prompts, approve actions, view progress/
  artifacts, review diffs - same sessions as desktop (F9.6 portable sessions).
- F25.4 Push notifications for approvals/completion; voice input optional.
- F25.5 (Monetization) The hosted relay/tunnel + off-machine execution = a paid service
  (the local P2P/Tailscale path stays free).

### F26. Self-Learning Memory Engine (recovered from `agence`) [Differentiator - the AGI core]
The crown jewel from the user's prior `agence` work, rebuilt cleanly on the Cortex
(F22) in SurrealDB. This is what makes the framework self-improving.
- F26.1 Three memory tiers: **Memory** (short ranked learnings injected into prompts),
  **Knowledge** (long-form wiki notes with `[[wikilinks]]` + backlinks), **Heartbeat**
  (scheduled maintenance) - all in SurrealDB.
- F26.2 Five memory layers (LobeHub model): activity, context, experience, identity,
  preference - each tagged + importance-ranked (low/medium/high/critical).
- F26.3 Forgetting curve: importance-scaled half-life decay + access reinforcement
  (spaced-repetition boost per recall) + explicit expiry (`computeDecayScore`).
- F26.4 Outcome-driven auto-capture (regex, no extra LLM call): user preferences, user
  corrections, and **tool-failure -> "avoid repeating" lesson**. Stored with provenance
  + confidence (lift pattern).
- F26.5 Global vs project scope: identity/preferences/critical items live in a
  `__global__` cortex (transfer across projects); the rest stays project-local.
- F26.6 Consolidation "sleep" pass (scheduled): merge near-dupes (embedding cosine
  > 0.9), prune redundant (0.86-0.9 band), prune stale (decay < threshold).
- F26.7 Associative linking + concept graph: cross-layer links (cosine > 0.72) + a
  concept map - SurrealDB synapse edges make this natural (beats agence's flat SQLite).
- F26.8 Procedural memory: `reflect` (distill completed work into a reusable skill) +
  `quality_gate` (auto-create a prevention skill from a recurring failure).
- F26.9 Recall into prompt: semantic search re-ranked by `score * decay`, injected as a
  `<past_learnings>` block with layer tags + link counts.
- F26.10 Heartbeat scheduler: a human-readable checklist (`HEARTBEAT.md`) driving
  periodic maintenance/ingest/briefs while the agent is idle.

### F27. Governance & Deterministic Safety (from agent-governance-toolkit) [Differentiator]
Prompt-level safety is "a polite request to a stochastic system." This makes misbehavior
structurally impossible, not merely discouraged.
- F27.1 Deterministic policy gate in front of EVERY tool/plugin call - fail-closed.
  Policy as YAML/Cedar-like rules (allow/deny/require_approval + default_action).
- F27.2 Privilege rings for tools/plugins (4 tiers) - a clean capability model for the
  plugin system (F19).
- F27.3 Tamper-evident audit log + Decision BOM: hash-chained (Merkle) records of active
  policy + request + allow/deny rationale - provable replay. SurrealDB append-only graph.
- F27.4 Agent SRE: kill switch, SLOs/error budgets, circuit breakers, chaos testing -
  the reliability half of autonomy.
- F27.5 Reversibility verification / execution-plan validation before acting (Hypervisor).
- F27.6 MCP security gateway: tool-poisoning detection, drift monitoring, hidden-
  instruction scanning (Mímir is plugin/MCP-oriented).
- F27.7 Human-in-the-loop confirm before irreversible actions (agentic-inbox: draft
  freely, never send without approval).
- F27.8 (Stretch) Plugin trust scoring + RL violation penalties during learning (ties to
  F26.8).

### F28. Native Sandbox Runtime (from vercel/sandbox) [Differentiator]
vercel/sandbox is only a client SDK over a hosted Firecracker control plane - Mímir
builds the control plane + runtime natively in the Go daemon (concrete impl of F7.12).
- F28.1 Tiered runtime: containers (fast, trusted code) vs microVMs (Firecracker/gVisor/
  Kata - strong, untrusted code). Agent picks per task.
- F28.2 In-guest agent: a tiny daemon in each VM/container exposing exec, file read/
  write, PTY - one protocol regardless of backend.
- F28.3 Command exec + log streaming: spawn in-guest, stream stdout/stderr as tagged
  frames; detached handles with kill + exit codes.
- F28.4 Filesystem API: write/read/mkdir/download against the in-guest agent.
- F28.5 Port routing: local reverse proxy maps localhost:N -> guest port; the GUI
  previews apps at http://localhost:N (no public subdomains).
- F28.6 Snapshot/fork: VM memory+disk snapshots (Firecracker) or container commits;
  pause/resume + branch-from-snapshot; metadata in SurrealDB. Branch an environment,
  try a change, roll back or fork cheaply.
- F28.7 Egress network policy: per-sandbox allow/deny + header transforms (iptables/
  nftables or userspace proxy).
- F28.8 Per-user Linux isolation: one Linux user per agent (useradd/chmod), setgid
  shared group dirs for collaboration - the most transferable idea for multi-agent work.
- F28.9 Warm pool: pre-booted microVMs (Firecracker boots ~125ms) for sub-second
  on-the-fly provisioning; auto-reap on timeout.

### F29. Universal Ingestion & Research (Gjallarhorn) [Core/Differentiator]
The Gjallarhorn: ingest ANY content, convert it, and store it as neurons in the Cortex.
This is the Open Notebook / NotebookLM capability set, built in natively (not a separate
app).
- F29.1 Universal file ingestion - drop any file; Mímir detects the type, converts to
  text, chunks, embeds, and stores as neurons with provenance:
  - Web pages / URLs -> clean markdown
  - PDF -> text extraction (pdftotext / pdfplumber-equivalent)
  - YouTube videos -> transcripts/subtitles (yt-dlp style)
  - Audio / video -> transcription (local Whisper or provider STT)
  - Office docs (DOCX/XLSX/PPTX) -> text/structured extraction
  - Images -> OCR + vision description
  - Text / markdown / code / CSV / JSON -> direct
- F29.2 Conversion pipeline: detect -> extract -> clean -> chunk (header/size-based) ->
  embed (F41) -> store as neurons (idempotent via content hash for re-ingest).
- F29.3 Metasearch (SearXNG-style, built in): fire one query to MULTIPLE engines in
  parallel (Google, Bing, DuckDuckGo, Brave, etc.), aggregate/dedupe/rank - no
  single-engine lock-in, privacy-respecting. Native `metasearch` tool.
- F29.4 Web scraping tool: fetch + extract clean markdown from sites (extends F3.7).
- F29.5 Notebooks: a notebook is a scoped collection of sources (neurons) around a
  topic/project - the user-facing unit; the Cortex is the underlying brain. Multi-notebook.
- F29.6 All ingestion auto-stores into the Cortex (F22) with provenance, available to RAG.
- F29.7 Implemented as native tools + skills (F11); respect robots.txt + rate limits.

### F30. Task & Build Engine (from turborepo) [Differentiator]
- F30.1 Declarative task graph: tasks + `dependsOn` + inputs/outputs per project, in
  SurrealDB.
- F30.2 Scheduler: topological sort + parallel execution of independent tasks (bounded
  by CPU), each in a sandbox (F28).
- F30.3 Content-addressed caching: hash task inputs -> cache key -> store/restore
  outputs (local NVMe, optionally SurrealDB/S3). Skip unchanged work - fast/cheap
  agent-driven rebuilds.
- F30.4 Filtered/scoped runs: rebuild only a target + its dependents.

### F31. Small-Model Mode & Structured Task Workflow [Differentiator - core to the local-first mission]
Designed so models **<=30B** can code smoothly without getting lost or sidetracked.
- F31.1 Mandatory game-plan-first: before writing code, the agent produces a game plan
  (spec: requirements -> design -> tasks). The framework enforces plan-before-code.
- F31.2 Game plan -> to-do list: the plan is converted into an ordered, concrete to-do
  list (each item small and independently verifiable).
- F31.3 Robust to-do list tool: persistent task tracker (id, content, status
  pending/in_progress/completed/blocked), subtasks, dependencies, tags; stored in
  SurrealDB; survives sessions; `task_search` (cross-session) + `todo_carry` (forward).
  MANDATORY: every persona always uses it for any non-trivial task (plan the steps
  first, work through them one at a time, update status as it goes); the base prompt
  instructs the model to always use it.
- F31.4 One-task-at-a-time execution: the agent works a single to-do item -> implement
  -> debug -> test -> mark done -> next. Prevents drift/sidetracking.
- F31.5 Verification at each step: after each item, run tests/checks; if failing, debug
  before advancing.
- F31.6 Lean tool surface for small models: expose a focused tool set (read/write/edit/
  bash/todowrite) instead of all tools - reduces tool-selection confusion.
- F31.7 Focused context: load only what the current task needs; re-inject the plan +
  to-do list each step so the model stays oriented.
- F31.8 Anti-derailment: doom-loop detection (F4.3), step budgets, "re-read the plan"
  re-orientation prompts.
- F31.9 Model-tier awareness: detect small models (by size/family) and auto-enable
  small-model mode + lean tools; larger models get the full toolset.
- F31.10 Target: smooth coding on <=30B models (Qwen3.6-35B-A3B, Qwen3.5-9B, small
  Codex/Gemma/Llama), local or API.

### F32. Telemetry & Privacy [Core]
- F32.1 Telemetry **ON by default**: collect anonymized usage metrics + tool success/
  failure rates + errors + latency, to improve the system.
- F32.2 Privacy settings: users turn telemetry OFF (opt-out) in settings; respected
  immediately and persisted.
- F32.3 What's collected: usage counts, model/tool used, success/error rates, latency.
  NOT code/content/prompts by default (that needs a separate explicit opt-in).
- F32.4 Transparent: documented what is collected; no hidden collection.
- F32.5 Local-first: telemetry is the only outbound call besides the chosen LLM provider;
  fully disable-able for air-gapped use.

### F33. Norse Agent Personas [Differentiator - identity]
The agent modes and subagents are named Norse figures, giving Mímir a coherent mythic
identity and clear, memorable role names.
- F33.1 Built-in specialist personas (the build team):
  - **Odin** - orchestrator/planner: runs Discovery, asks the questions, brainstorms,
    makes the plan, and hands off to the specialists.
  - **Thor** - builder: writes the code, one section at a time.
  - **Loki** - tester & debugger: end-to-end tests + debugging (the trickster who breaks
    things to find the bugs).
  - **Forseti** - code auditor: reviews code quality, correctness, security (the judge).
  - **Heimdall** - visual auditor: screenshots the running app + vision-verifies the UI
    matches the approved mock (the all-seeing watchman).
  - **Bragi** - skald: docs, README, marketing copy.
- F33.2 Scout personas: Huginn & Muninn (research - "thought" & "memory"), Ratatoskr
  (messenger / inter-agent comms).
- F33.3 Planning personas: the Norns (Urd/Verdandi/Skuld - spec past/present/future).
- F33.4 Each persona bundles a prompt + tool allowlist + default model (custom modes,
  F16); users can add their own Norse (or any) personas.
- F33.5 Persona names surface in the GUI/TUI (agent switcher, chat, subagent cards).

### F34. Code Intelligence (LSP) [Differentiator - table stakes for a coding agent]
Language-server integration so the agent understands code, not just text.
- F34.1 Diagnostics (compile/lint errors) fed to the model so it fixes REAL errors.
- F34.2 Go-to-definition, find-references, hover, document/workspace symbols.
- F34.3 Auto-start the right language server per file type; expose as tools.
- F34.4 (opencode/Cursor/Copilot all have this; a coding agent without it is blind.)

### F35. Git & PR Automation [Differentiator - the "autonomous developer" story]
- F35.1 Git workflow tools: status, diff, commit, branch, stash, log.
- F35.2 Create pull requests with generated descriptions; push branches.
- F35.3 Issue -> PR: assign an issue, the agent plans + implements + opens a draft PR
  (Copilot agent / Devin style).
- F35.4 Automated PR review (BugBot-style): review a diff, leave comments, flag risks.
- F35.5 Auto-fix CI failures on the agent's own PRs.

### F36. Deep Codebase Indexing [Differentiator]
- F36.1 AST-aware code indexing (tree-sitter) + symbol graph - not just naive text RAG.
- F36.2 Hybrid retrieval: semantic (embeddings) + keyword (FTS) + structural (symbols).
- F36.3 Repo-map / codebase tour so the agent orients in large / multi-repo projects.
- F36.4 Incremental re-index on file change.

### F37. Usage & Cost Tracking [Core - trust + monetization]
- F37.1 Track tokens + cost per session/model/provider (from streaming usage events).
- F37.2 Usage dashboard in the GUI; budget alerts/limits per session/workspace.
- F37.3 Per-model cost config; show estimated cost before expensive actions.
- F37.4 Feeds the business model: demonstrates the value of the hosted gateway pricing.

### F38. Context Mastery [Differentiator - the "context engineering" edge]
- F38.1 Output styles: customize the agent's tone/format (concise, explanatory, learning).
- F38.2 Steering files (Kiro-style): product/tech/structure context, global vs workspace.
- F38.3 Flow awareness (Windsurf-style): passively fold the user's recent edits/terminal/
  navigation into context ("fix this" already knows the failing test you just ran).
- F38.4 Prompt caching: cache stable prompt prefixes to cut cost/latency.
- F38.5 Context visualization (/context): show what's consuming the context window.

### F39. Inline Tab Completion [Differentiator - adoption driver]
- F39.1 Predictive inline autocomplete (next-edit prediction, not just insertion).
- F39.2 Requires an editor surface: IDE extension (VS Code/JetBrains via ACP) or the
  built-in editor in the GUI.
- F39.3 Cursor's moat; pairs with our small-model focus (a fast local model for tabs).

### F40. Sharing & Collaboration [Differentiator - growth + team]
- F40.1 Session share links: publish a session as a public/read-only URL (opencode-style).
- F40.2 Published artifacts: share the agent's deliverables (plans, diffs, previews).
- F40.3 Team sessions / multiplayer (later): shared workspace, shared Cortex.

### F41. Embedding Strategy [Core - powers RAG]
How content becomes searchable vectors. Hybrid: local-first (free/private) + optional cloud.
- F41.1 Bundled local embedding model (default, zero-config): a small ONNX model
  (e.g. all-MiniLM-L6-v2, ~22M params, ~80MB) runs on CPU via ONNX Runtime - private,
  free, no key. Ships with the installer.
- F41.2 Ollama embeddings: use a local Ollama embedding model (nomic-embed-text,
  mxbai-embed-large) if Ollama is present (preferred when available).
- F41.3 Provider embeddings: OpenAI text-embedding-3-small, etc. (requires a key).
- F41.4 Cloud embeddings (paid tier): hosted, higher-quality embeddings without local
  compute - part of the Cloud plan (monetization).
- F41.5 Pluggable embedder interface; configurable per notebook; dimension stored in schema.

### F42. Marketplace: MCP + Skills + Personas (one-click install) [Differentiator + monetization]
A built-in marketplace browser (GUI + CLI) to discover and one-click install extensions,
aggregating existing registries plus a Mímir-curated registry.
- F42.1 One-click install of **MCP servers** (Smithery/MCPfinder-style): generate config +
  install + auth handling; aggregate Official MCP Registry, Smithery, Glama, mcpmarket.
- F42.2 One-click install of **Skills** (skills.sh / LobeHub-style): download SKILL.md +
  resources into the skills dir.
- F42.3 One-click install of **Personas** (LobeHub open-persona-style): download a persona
  (prompt + tool allowlist + model) into the personas config.
- F42.4 Discovery: search/browse by category/keyword across registries; ratings, install
  counts, verified publishers.
- F42.5 Security: trust scoring + security scanning before install (ties to F27.6 MCP
  security gateway); "will install" permission preview.
- F42.6 Mímir-curated registry + featured listings (monetization: featured/verified slots).
- F42.7 Publish flow: authors publish MCP/Skills/Personas to the Mímir registry.
- F42.8 Paid listings: creators set a price (one-time or subscription) for skills/hooks/
  MCP/personas/scrapers/workflows (the Apify-Actor-Store model for agent components).
- F42.9 Revenue share + payouts: e.g. 80/20 (creator/Mímir); creator dashboard with
  sales analytics + payouts (Stripe Connect).
- F42.10 Creator tooling: versioning, listing pages, reviews/ratings, "sell your
  scraper/hook" storefronts; verified-publisher badges.

### F43. Guided Development Protocol [Core differentiator - "hold the LLM's hand"]
A mandatory, gated, phase-based workflow that holds the model's hand so it builds
software ONE section at a time without getting lost (critical for small models,
disciplined for all). This is the BMAD method baked into Mímir as the default build
protocol. Lots of prep FIRST, then build.

Phases (each ends with a user checkpoint gate):
- F43.1 **Discovery (Odin) - deep requirements interview**: before anything is built,
  Odin interviews the user with a structured, adaptive questionnaire AND brainstorms
  until the concept is airtight: What kind of app (web/mobile/desktop/SaaS/CLI)? What
  UI style/look? If SaaS: logins/auth? subscriptions/payments? multi-tenant? roles?
  Database? API? integrations? Then it produces complete docs (PRD + design + tech
  spec) and gets them PERFECT before any build starts.
- F43.2 **Design - mock first**: generate a UI mock/wireframe for the user to review and
  APPROVE before any code. Visual sign-off gate.
- F43.3 **Research (Huginn & Muninn)**: research the domain/tech via the Cortex +
  metasearch; gather and ground context.
- F43.4 **Plan**: convert the approved design into an ordered to-do list of sections
  (game plan -> to-do list, F31).
- F43.5 **Build + specialist handoff - one section at a time**: Odin hands each section
  to the specialists in sequence and does NOT advance until all pass:
  Thor (build) -> Loki (end-to-end test + debug) -> Forseti (code audit) -> Heimdall
  (visual audit: screenshot + vision-verify the UI) -> user checkpoint -> next section.
- F43.6 **Polish (Bragi)**: docs, comments, README, marketing copy.
- F43.7 **Checkpoint gates**: each phase ends with a user sign-off; the framework does
  NOT proceed until the user checks it off. Configurable auto-advance for trusted users.
- F43.8 **Three verification roles** per section:
  - **Loki** = tester & debugger: runs end-to-end tests, reproduces + fixes bugs.
  - **Forseti** = code auditor: reviews code quality, correctness, security.
  - **Heimdall** = visual auditor: screenshots the running app and uses a vision model
    to verify the UI matches the approved mock and looks right (no visual bugs).
- F43.9 **Prep-first enforcement**: the framework refuses to write app code until
  Discovery + Design (mock approved) + Research + Plan are complete.
- F43.10 Mock generator: produce UI mockups (wireframe -> high-fidelity) the user can
  annotate (F24) and approve; the approved mock drives the build.
- F43.11 **Adaptive questionnaire**: the Discovery interview adapts to the project type
  (SaaS triggers auth/payment/tenant questions; a static site triggers design/content
  questions); brainstorm mode until the user is satisfied.
- F43.12 **Visual verification**: capture screenshots of the running app (from the
  sandbox/browser preview) and compare against the approved mock via a vision model;
  flag visual regressions.
- F43.13 **Debugging built in**: debug logging, error capture, and Loki's debugging
  tools are first-class throughout the whole build (not bolted on).
- F43.14 **Document-first gate**: the build cannot start until the Discovery docs (PRD +
  design + tech spec) are complete and user-approved.

### F44. Default Voice & Communication Style [Core - identity]
How Mímir talks. Baked into the base system prompt every persona inherits.
- F44.1 **Default voice = 10th-grade casual**: Mímir talks like a 10th-grade high school
  student. Simple, relatable, everyday lingo. Clear and chill, like explaining to a
  friend. Not stiff, not corporate, not stuffed with jargon.
- F44.2 **Hard rule - never use em dashes.** Use hyphens (-) or colons (:) instead.
  Enforced in the base prompt AND a post-output filter.
- F44.3 Lives in the base system prompt all personas inherit (Odin through Bragi).
- F44.4 Configurable output styles (F38.1): users can change the voice, but the
  no-em-dash rule stays on by default.
- F44.5 Output filter: a post-processing pass strips any em dashes that slip through.

### F45. Question Tool [Core - like opencode's question]
A structured question tool so the agent asks the user clear, clickable questions (like
opencode's `question` tool).
- F45.1 The agent poses questions with multiple-choice options (label + description);
  the GUI/chat renders them as clickable choices.
- F45.2 Single or multiple selection; a "type your own answer" option is always added.
- F45.3 Used heavily in Discovery (F43.11) for the requirements interview, and whenever
  the agent needs a decision, preference, or clarification.
- F45.4 The agent gets the selected answer(s) back and proceeds.
- F45.5 Group multiple questions in one call.

### F46. Thinking Mode [Core - visible reasoning]
- F46.1 The agent's reasoning/thinking shows in the chat as a separate, light-gray,
  collapsible block - readable but visually de-emphasized (not the focus).
- F46.2 Thinking levels: low / medium / high (default medium), controlling how much
  reasoning the model does and shows.
- F46.3 Off switch: thinking can be turned off entirely in the options.
- F46.4 The thinking block is clearly distinct from the agent's spoken reply.

### F47. Chat Composer (input bar) [Core]
The chat input bar has everything the user needs to direct the agent:
- F47.1 **+ attach button**: add documents, images, files, and other context to the
  message (fed into the Cortex + the model).
- F47.2 **Build / Plan toggle**: switch between Build mode (agent writes code) and Plan
  mode (agent only plans, no building) - so the user can have it just plan.
- F47.3 **Model picker**: choose which model handles the message.
- F47.4 **Thinking level selector**: low / medium / high (ties to F46.2).
- F47.5 **Send button** to send the message.

### F48. Computer Use & OS Automation (cross-platform) [Differentiator]
Tools that let the agent read screens, click UI, inject input, manage the OS, and
trigger app actions across Windows, macOS, Linux, Android, and iOS. This powers
Heimdall's visual auditing (F43.12) and gives the agent real control of the computer
and phone.
- F48.1 **UI automation (screen reading + clicking)** - read the UI tree, click/tap:
  - Windows: UI Automation (UIA).
  - macOS: Accessibility API (AXUIElement: AXUIElementCreateApplication,
    AXUIElementCopyActionNames).
  - Linux: AT-SPI2 (over the D-Bus API, accessible widget trees).
  - Android: AccessibilityService (AccessibilityNodeInfo + dispatchGesture for
    taps/swipes).
  - iOS: UIAccessibility protocol / XCUITest.
- F48.2 **Low-level event hooks (input injection / interception)**:
  - Windows: SetWindowsHookEx (Win32 hooks).
  - macOS: Quartz Event Taps (CGEventTapCreate).
  - Linux: uinput + libevdev (write /dev/uinput, read raw hardware input).
  - Android: restricted without root (AccessibilityService for high-level events;
    /dev/input/eventX needs root).
  - iOS: blocked by the sandbox (jailbreak + MobileSubstrate/SpringBoard only) - flagged
    as not feasible on stock devices.
- F48.3 **System instrumentation (background OS management)**:
  - Windows: WMI & COM.
  - macOS: sysctl (kernel/hardware state) + OSA (AppleScript/JXA).
  - Linux: D-Bus (systemd services) + /sys + /proc virtual filesystems.
  - Android: system services (ConnectivityManager/BatteryManager), DevicePolicyManager
    (Device Owner), adb shell.
  - iOS: restricted, permission-gated frameworks (CoreMotion), MDM payloads for system
    changes.
- F48.4 **Modern AI app actions (agent runtimes)**:
  - Windows: Windows App Actions / Copilot+ actions.
  - macOS: App Intents (AppIntent protocol + perform()).
  - Linux: standard D-Bus interfaces.
  - Android: App Actions + App Functions API (XML-defined; Gemini Nano / Assistant).
  - iOS: App Intents (AppIntent + AppEntity for Foundation Models / Siri).
- F48.5 **Feasibility tiers + permission handling**: desktop (Windows/macOS/Linux) is
  fully feasible; Android needs the accessibility permission (root for low-level); iOS is
  sandbox-restricted (XCUITest for testing, MDM for system, jailbreak for low-level). The
  agent reports what's possible per platform and asks the user for the needed permission.
- F48.6 **Screenshot + vision loop** (ties to Heimdall F43.12): capture the screen, feed
  it to a vision model to understand the UI, then act via the UI automation tools.

---

## 7. Epics Roadmap

The roadmap is tiered. Tier 1 (Core v1) delivers a usable, safe, modular agent.
Tier 2 (Differentiators) is what makes it great. Tier 3 (Stretch) is ambitious.

### Tier 1 - Core v1
| Epic | Title | Maps to | Depends | Size |
|---|---|---|---|---|
| E1 | Modular Foundation (monorepo skeleton, `@mimirmind/agent-core`, config, PluginRegistry + plugin-SDK contract shape, provider registry, credential store) | F1.1-F1.6, F19.1-F19.5, F19.8 | - | L |
| E2 | Conversation & Agent Loop (5-Piece-Kit execution state machine, streaming, tool calling, sessions, lanes/queues) | F2, F19.8, F20.8 | E1 | L |
| E3 | Built-in System Tools (shell, filesystem, process, code-exec, edit_block, web) | F3 | E1 | L |
| E4 | Permission Engine & Core Guardrails (allow/ask/deny, shell-AST, blocklist, doom-loop, audit, approvals) | F7.1-F7.6 | E3 | M |
| E5 | Checkpoints / Undo / Rewind (git-backed) | F13 | E2 | M |
| E6 | Memory & Persona (AGENTS.md, layered config, persona files) | F15.1, F15.6, F20.7 | E1 | M |
| E7 | Skills System (SKILL.md, progressive disclosure, snapshots) | F11 | E2 | M |
| E8 | Plan Mode & Permission Modes | F14.1-F14.2 | E2, E4 | S |
| E9 | Subagents (core isolation + built-in types) | F5.1-F5.3 | E2 | M |
| E10 | Two-Tier Hooks (gateway + plugin lifecycle) | F12 | E2, E11 | M |
| E11 | Plugin SDK & Registry (full register* contract, manifest, load pipeline, contract tests) | F19.3-F19.7 | E1 | L |
| E12 | Primary GUI (web app: static top bar, collapsible left nav, center workspace, right chat panel, far-right quick-launch rail) + TUI (Bubble Tea) | F9.1, F9.2 | E2 | L |
| E13 | Packaging & Distribution (single binary, installer, self-update) | F10 | E12 | S |
| E14 | MCP Integration (optional 3rd-party tool servers) | F4 | E2, E3 | S |

### Tier 2 - Differentiators (v1.x)
| Epic | Title | Maps to | Depends | Size |
|---|---|---|---|---|
| E15 | Pluggable Runtimes / Harness Registry | F20.1-F20.3 | E11 | M |
| E16 | Multi-Agent Routing (per-agent workspace/session/auth, bindings) | F20.4-F20.6 | E6, E11 | M |
| E17 | Custom Modes / Role Personas | F16 | E2 | M |
| E18 | Auto Mode + Auto Memory + Path-scoped Rules | F14.3, F15.2-F15.5 | E4, E6 | M |
| E19 | Advanced Orchestration (dynamic subagents, worktrees, waves, coordinator/worker/validator) | F5.4-F5.9 | E9 | L |
| E20 | Background Tasks & Automation (Await, cron, events, artifacts) | F8 | E9 | M |
| E21 | Spec-Driven Workflow + BMAD integration | F6 | E2, E9 | M |
| E22 | Verification & Artifacts + Model Routing | F17.1-F17.4, F1.7 | E2 | M |
| E23 | Sandbox Tiers + On-the-Fly Sandboxing (OS-native + Docker created/destroyed per task, two-phase, network allowlists, autonomy levels) | F7.7-F7.12 | E4 | L |
| E24 | Multi-surface Runtime (Headless + SDK + ACP) | F9.2-F9.5 | E2, E12 | M |
| E25 | Plugins Marketplace | F18 | E11 | M |

### Tier 3 - Stretch
| Epic | Title | Maps to | Depends | Size |
|---|---|---|---|---|
| E26 | Property-based Testing + Computer Use / browser-in-loop | F6.5, F17.5 | E21, E22 | L |
| E27 | Self-improving Memory + Skill/Tool Synthesis | F15.5 | E18 | L |
| E28 | Best-of-n / Arena + Flow Awareness | F5.10 | E19 | M |
| E29 | Messaging Channels (Telegram/Discord/Slack) | F9.7 | E16, E24 | M |
| E30 | Desktop App (Wails packaging of the GUI) + Voice input | F9.1 | E12, E24 | M |

**Suggested Core v1 build order:** E1 -> E2 -> E3 -> E4 -> E5 -> E6 -> E7 -> E8 ->
E9 -> E11 -> E10 -> E12 -> E13 -> E14. (Establish the modular skeleton and plugin
contract first, then a usable, safe TUI agent with tools, permissions, checkpoints,
memory, skills, plan mode, subagents, and hooks.)

### Tier 4 - Knowledge, Platform & Monetization
These turn Mímir into a sustainable, differentiated product (see Section 11).
The knowledge/learning/sandbox/governance/research/small-model epics (E31, E38, E40, E41, E43-E59)
are product differentiators (schedule alongside Tier 2); E32-E37 and E42 are the hosted
platform/backend that earns.
| Epic | Title | Maps to | Depends | Size |
|---|---|---|---|---|
| E31 | Knowledge Layer (gitmcp on-the-fly spawning, `knowledge.*` namespace, local cache/index, curated packs) | F21 | E2, E14 | M |
| E32 | Hosted AI Gateway (OpenAI-compatible proxy, API keys, token metering, model catalog, limits) | Sec 11 | backend | L |
| E33 | Console & Billing (signup, Stripe, usage dashboard, API key + workspace management) | Sec 11 | E32 | L |
| E34 | Subscriptions & Plans (pay-as-you-go credits + monthly subscription + $-limits + upsell hooks) | Sec 11 | E33 | M |
| E35 | Cloud Agent Execution (hosted sandboxes for background/parallel agents) | Sec 11 | E19, E20, E23, E32 | L |
| E36 | Team Workspaces & Enterprise (SSO/RBAC, spend limits, analytics, self-hosted gateway) | Sec 11 | E33 | L |
| E37 | Hosted Knowledge Bases & Notebooks (indexed/private repo + notebook knowledge, curated packs, team-shared) | F21.9, F22.11 | E31, E38, E33 | M |
| E38 | The Cortex - Knowledge & Memory Brain (neurons/synapses/engrams in SurrealDB, multi-modal ingestion, RAG, notes, transformations, citations) | F22 | E2, E31 | L |
| E39 | (Stretch) Podcast / Audio Overviews (multi-speaker TTS summaries of a cortex) | F22.11 | E38 | M |
| E40 | Goal-Driven Autonomous Mode (end goal -> game plan -> execute + verify until done) | F23 | E8, E21 | L |
| E41 | Interactive Preview & Annotation (live browser preview, markup canvas, send back to LLM) | F24 | E12 | M |
| E42 | Mobile Remote Access (iOS/Android app or PWA + secure tunnel to home daemon) | F25 | E12, E24 | L |
| E43 | Self-Learning Memory Engine (3-tier memory, 5 layers, forgetting curve, auto-capture, consolidation, concept graph, reflect/quality_gate, heartbeat) | F26 | E38 | L |
| E44 | Governance & Deterministic Safety (fail-closed policy gate, privilege rings, tamper-evident audit, agent SRE, MCP security gateway) | F27 | E4, E11 | L |
| E45 | Native Sandbox Runtime (Firecracker/gVisor/containers, in-guest agent, exec+streaming, fs API, port routing, snapshot/fork, per-user isolation, warm pool) | F28 | E3, E23 | L |
| E46 | Universal Ingestion & Research / Gjallarhorn (any file -> convert -> Cortex; metasearch, YouTube, scraping, audio/video transcription) | F29 | E38, E41 | L |
| E47 | Task & Build Engine (declarative task graph, parallel topo execution, content-addressed caching) | F30 | E45 | M |
| E48 | Small-Model Mode & Structured Task Workflow (game plan -> to-do list -> one-task-at-a-time, lean tools, verification, anti-derailment) | F31 | E2, E3, E6 | L |
| E49 | Telemetry & Privacy (on by default, anonymized, opt-out in privacy settings) | F32 | E1 | M |
| E50 | Norse Agent Personas (Odin/Thor/Loki/Heimdall/Bragi modes, Huginn & Muninn scouts, Ratatoskr, Norns/Forseti) | F33 | E9, E16 | M |
| E51 | Code Intelligence / LSP (diagnostics, go-to-def, find-refs, symbols) | F34 | E3 | M |
| E52 | Git & PR Automation (commit/branch/diff, Issue->PR, PR review, CI auto-fix) | F35 | E3, E4 | L |
| E53 | Deep Codebase Indexing (tree-sitter AST + hybrid retrieval + repo-map) | F36 | E6 | L |
| E54 | Usage & Cost Tracking (per-session/model cost, dashboard, budget alerts) | F37 | E2 | M |
| E55 | Context Mastery (output styles, steering files, flow awareness, prompt caching, /context) | F38 | E2, E6 | L |
| E56 | Inline Tab Completion (next-edit prediction; IDE extension or built-in editor) | F39 | E12 | L |
| E57 | Sharing & Collaboration (session share links, published artifacts, team sessions) | F40 | E12 | M |
| E58 | Embedding Strategy (bundled local ONNX model + Ollama + provider + cloud embeddings) | F41 | E6 | M |
| E59 | Marketplace (one-click install of MCP + Skills + Personas; registry aggregation; security scanning) | F42 | E11, E12 | L |
| E60 | Guided Development Protocol (Discovery -> Design mock -> Research -> Plan -> build one section at a time with auditor gates + checkpoints) | F43 | E2, E8, E9 | L |
| E61 | Default Voice & Style (10th-grade casual voice + never-use-em-dashes rule + output filter) | F44 | E2 | S |
| E62 | Question Tool (structured multiple-choice questions with options, like opencode; used in Discovery + clarifications) | F45 | E2, E12 | S |
| E63 | Thinking Mode (light-gray collapsible thinking blocks, low/medium/high levels, off switch) | F46 | E2, E12 | S |
| E64 | Chat Composer (+ attachments, Build/Plan toggle, model picker, thinking level, send button) | F47 | E12 | M |
| E65 | Computer Use & OS Automation (UI automation, input hooks, system instrumentation, app actions across Windows/macOS/Linux/Android/iOS; desktop first, mobile gated by permissions) | F48 | E3, E28 | L |

---

## 8. BMAD Execution Plan

BMAD v6.10.0 is installed at `E:\agent-hub\_bmad`. Run each workflow in a FRESH chat.
Outputs go to `E:\agent-hub\_bmad-output\planning-artifacts\` and
`...\implementation-artifacts\`.

| Phase | Agent | Workflow | Produces |
|---|---|---|---|
| 1 | Mary (Analyst) | `bmad-brainstorming` | Refined concept, name, scope decisions |
| 2 | Mary (Analyst) | `bmad-product-brief` | Product brief (this doc seeds it) |
| 3 | John (PM) | `bmad-prd` | PRD: personas, user stories, F1-F20 requirements |
| 4 | Winston (Architect) | `bmad-architecture` | Architecture doc (stack, structure, ADRs) |
| 5 | Winston (Architect) | `bmad-check-implementation-readiness` | Gate: is the PRD+arch ready to build? |
| 6 | Sally (UX) | `bmad-ux` | UX/UI spec for the TUI (and later GUI) |
| 7 | John (PM) | `bmad-create-epics-and-stories` | Epics E1-E65 broken into stories |
| 8 | Amelia (Dev) | `bmad-sprint-planning` | Sprint 1 plan (start with E1) |
| 9 | Amelia (Dev) | `bmad-create-story` then `bmad-dev-story` | Implement story by story |
| 10 | Amelia (Dev) | `bmad-code-review` | Review each completed story |

**Rule:** one fresh chat per workflow. Carry artifacts forward by file path.

---

## 9. Risks and Decisions to Resolve (in brainstorming)

- **D1 Fork vs fresh build.** Recommended: fresh on AI SDK + MCP SDK + Bun, with an
  opencode-config import layer. Confirm in Phase 1.
- **D2 UI first-class or TUI-first.** Recommended: TUI-first (E8), GUI deferred (E11).
- **D3 Native tools vs MCP for system access.** DECIDED: system access is BUILT IN
  natively (Desktop-Commander-class tool set reimplemented in our code). We do NOT
  depend on the Desktop Commander MCP server. MCP is optional, for other 3rd-party
  tools only.
- **D4 Sandbox stance.** Recommended: guardrails by default, optional Docker isolation.
- **D5 Name.** Pick during brainstorming.
- **R1 Scope creep.** The combined feature set of 6 tools is large. The epic order
  above front-loads a usable core; hold E6/E9/E11 until the core is solid.
- **R2 Provider parity.** "All providers opencode has" is achieved via AI SDK, but
  some opencode providers use bespoke SDKs (Bedrock, Vertex, Copilot OAuth). Scope
  those as stretch unless needed.
- **R3 Security.** Shipping terminal + filesystem + code-exec is powerful and
  dangerous. E7 (guardrails) must land before any broad distribution.

---

## 10. Immediate Next Steps

1. Open a fresh chat and run `bmad-brainstorming` (Mary). Paste this document as
   context. Resolve D1-D5 and pick a name.
2. Run `bmad-product-brief` (Mary) to formalize the brief.
3. Run `bmad-prd` (John) to turn Section 6 into user stories.
4. Run `bmad-architecture` (Winston) to finalize Section 4-5 into ADRs.
5. Scaffold the repo (E1) once architecture passes the readiness check.

---

## 11. Commercialization & SaaS Strategy

Mímir is local-first and open-source, but it can sustain a business the same
way opencode does: **open-core + a hosted AI gateway**, plus cloud/enterprise
upsells. The CLI is free; the hosted services make money.

### How opencode does it (the model to copy)
- **OpenCode Zen** (`https://opencode.ai/zen`): a pay-as-you-go AI gateway.
  OpenAI-compatible API. Users add credit ($20), get an API key from
  `opencode.ai/auth`, paste it via `/connect`. Charged per token at cost + card
  fees (~4.4% + $0.30). Auto-reload at $5. Monthly spend limits. ~50 models
  curated and benchmarked for coding agents.
- **OpenCode Go** (`https://opencode.ai/zen/go`): a $5-then-$10/month subscription
  for curated open coding models, with usage limits denominated in dollars
  (5hr/$12, weekly/$30, monthly/$60) and a ~6x value multiplier from bulk
  discounts / reserved GPU capacity. Falls back to Zen balance over limits.
- **Teams/Workspaces** (beta): roles, model curation, per-member spend limits,
  SSO, centralized billing ($25-$150/seat/month).
- **BYOK + no lock-in**: use your own OpenAI/Anthropic keys alongside Zen; Zen
  works with any agent, and opencode works with any provider.
- **Mechanism (technical)**: Zen/Go are just PROVIDERS in the config - an
  OpenAI-compatible `baseURL` (`https://opencode.ai/zen` or `/zen/go`) + an API
  key. The open-source CLI talks to a proprietary hosted backend
  (console/api/app.opencode.ai) that does auth, routing, metering, and billing.
  The CLI even upsells Go when you hit limits (`GO_UPSELL_URL`).

### Mímir revenue streams (in order of fit)
1. **Hosted AI Gateway** (primary; the Zen/Go model). `gateway.mimirmind.com`,
   OpenAI-compatible. Curated/benchmarked models, unified billing, pay-as-you-go
   credits + a cheap subscription tier. Margin from bulk discounts/reserved
   capacity + small markup. Works with any agent (no lock-in).
2. **Cloud Agent Execution** (Cursor/Devin/Copilot/Kiro model). Local agents are
   free; optional hosted sandboxes for background/parallel agents are paid compute.
   Maps to E19/E20/E23 + the sandbox tiers.
3. **Pro / Team subscription**. Shared workspaces, per-member spend limits,
   SSO/RBAC, usage analytics, centralized billing, model curation.
4. **Hosted Knowledge Bases** (ties to F21). Free = public gitmcp passthrough;
   Paid = indexed/cached/private repo knowledge, curated knowledge packs,
   team-shared knowledge.
5. **Marketplace cut**. Curated MCP/Skills/Personas marketplace with one-click install
   + revenue share + featured/verified listings (F42/E59).
6. **Enterprise / self-hosted**. Air-gapped gateway, compliance, on-prem, support
   (Factory/IBM Bob/Kiro model).

### Pricing - two product lines + a creator marketplace

Mímir monetizes two things: **coding** (model access, like opencode's Zen/Go) and
**knowledge** (notebooks/Cortex, vs NotebookLM). Plus a **creator marketplace** with
revenue share. Self-hosted is always free and unlimited.

#### A. Coding plans (model access - like opencode Zen/Go)
| Plan | Price | What you get |
|---|---|---|
| **Gateway (pay-as-you-go)** | per-token at cost + small fee | Curated/benchmarked models, one key, unified billing, works with any agent (like Zen) |
| **Mímir Coder** | $10/mo | Cheap subscription for curated open coding models with $-denominated usage limits (~6x value, like Go) |

#### B. Knowledge plans (notebooks - vs NotebookLM: cheaper AND double)
NotebookLM 2026: Free = 50 sources/notebook; Pro ($19.99) = 300; Ultra ($199.99) = 600.
| Plan | Price | Notebooks | Sources/notebook | What you get |
|---|---|---|---|---|
| **Self-hosted (free)** | $0 | Unlimited | Unlimited | Full framework, BYOK or local, private, local embeddings |
| **Cloud Sync** | $9/mo | 200 | 200 | Sync Cortex across devices + mobile + hosted embeddings |
| **Cloud Pro** | **$15/mo** | 1,000 | **600 (2x NotebookLM Pro, $5 cheaper)** | Hosted embeddings, cloud agents, advanced orchestration |
| **Cloud Ultra** | $39/mo | Unlimited | **1,200 (2x NotebookLM Ultra, $160 cheaper)** | Hosted, priority, watermark-free outputs |
| **Team / Enterprise** | custom / ~$25+ seat | custom | custom | Shared Cortex, SSO/RBAC, spend limits, self-host, support |

#### C. Creator Marketplace (revenue share)
Creators publish and **sell** skills, hooks, MCP servers, personas, scrapers, and
workflows. Free + paid listings; revenue share (e.g. 80/20 creator/Mímir); creator
payouts + dashboard; featured/verified listings. Security scanning gates every listing
(the Apify-Actor-Store model, but for agent components).

**The headlines:**
- Coding: "Curated coding models for $10/mo - or pay only for what you use."
- Knowledge: "NotebookLM charges $20/mo for 300 sources. Mímir is unlimited free on your
  machine, or $15/mo synced everywhere with double the sources."
- Creators: "Build skills, hooks, and scrapers. Sell them. Keep 80%."

### What to build (the platform backend - separate from the open-source CLI)
- **Gateway service**: OpenAI-compatible proxy that authenticates API keys, routes
  to upstream providers, meters tokens, enforces limits. (Not open-sourced.)
- **Console + billing web app** (console.mimirmind.com): signup, Stripe billing,
  API key management, usage dashboards, team/workspace management.
- **CLI integration**: a built-in `mimirmind` provider (baseURL = gateway, apiKey
  from console) + `auth login` flow + upsell hooks on limit-hit.
- **Knowledge service** (paid tier): index/cache repo docs, serve private bases.

> The open-source CLI stays free and fully functional with BYOK. The hosted
> gateway, cloud execution, and knowledge services are the paid layer. This is
> the exact open-core playbook opencode, Cursor, and Kiro run.

---

## Appendix A: Reference URLs
- OpenClaw (PRIMARY architectural reference): https://github.com/openclaw/openclaw , https://docs.openclaw.ai , https://openclawlab.com/en/docs/deep-dive/framework-focus/ai-five-kit/
- opencode: https://opencode.ai/ (source: https://github.com/anomalyco/opencode)
- opencode Zen (pay-as-you-go gateway): https://opencode.ai/zen , https://opencode.ai/docs/zen/
- opencode Go ($10/mo subscription): https://opencode.ai/go , https://opencode.ai/docs/go/
- GitMCP (repo -> MCP knowledge server): https://gitmcp.io/ (source: https://github.com/idosal/git-mcp)
- Desktop Commander (reference spec for our native tools - NOT a dependency): https://github.com/wonderwhy-er/DesktopCommanderMCP
- Claude Code: https://docs.claude.com/en/docs/claude-code
- Agent Zero: https://github.com/agent0ai/agent-zero
- Gemini CLI: https://github.com/google-gemini/gemini-cli
- Kiro: https://kiro.dev/
- Google Antigravity: https://antigravity.google/product/antigravity-2
- IBM Bob: https://bob.ibm.com/
- Cursor: https://cursor.com/
- Windsurf: https://docs.windsurf.com/ ; Cline: https://docs.cline.bot/ ; Roo Code: https://docs.roocode.com/
- Devin: https://devin.ai/ ; GitHub Copilot coding agent: https://docs.github.com/ ; OpenAI Codex: https://openai.com/codex ; Factory: https://factory.ai/
- BMAD Method: https://github.com/bmad-code-org/BMAD-METHOD
- Vercel AI SDK: https://ai-sdk.dev/
- Model Context Protocol: https://modelcontextprotocol.io/
- Agent Skills standard: https://agentskills.io/ ; Agent Client Protocol: https://agentclientprotocol.io/
