# Mímir - Architecture Document

**BMAD Artifact - Phase 4 (Winston, System Architect)**
Date: 2026-07-23 | Status: Draft for implementation-readiness review
Inputs: `mimirmind-prd.md`, `PROJECT-MASTER-PLAN.md`
Feeds: `bmad-check-implementation-readiness`, then epics/stories + build.

---

## 1. Architecture Overview

Mímir is a **modular, local-first agentic coding framework**. A single Go binary runs a
long-lived **daemon** that owns the agent loop, tools, the Cortex (SurrealDB brain),
sandboxing, plugins, and a local server. A **web GUI** (Solid) is served by the daemon
(also packaged as a Wails desktop app); a **Bubble Tea TUI** and (later) mobile/chat
clients all talk to the same daemon over a local REST + WebSocket API.

**Style:** modular monolith (clean package boundaries + a typed plugin SDK) running as a
local service; plugin-extensible; provider-agnostic. Not microservices - one process,
many well-separated packages, with SurrealDB as the only external companion.

```
                 +--------------------------------------------------+
   GUI (Solid) --|                                                  |
   TUI (BubbleTea)|              MIMIR DAEMON (Go binary)           |
   Mobile/Chat ---|  +----------+  +--------+  +-----------------+  |
        |        |  | Agent    |  | Tools  |  | Cortex (SurrealDB)| |
   REST + WS ---->|  | Loop     |->| (native|->| neurons/synapses | |
                 |  | (state   |  | + MCP) |  | engrams + RAG    | |
                 |  |  machine)|  +--------+  +-----------------+  |
                 |  +----------+                                    |
                 |  +----------+  +--------+  +-----------------+  |
                 |  | Plugin   |  | Policy |  | Sandbox Manager |  |
                 |  | Registry |  | Gate   |  | (Docker/Firecr.)|  |
                 |  +----------+  +--------+  +-----------------+  |
                 |  +--------------------------------------------+ |
                 |  | LLM Provider Abstraction (OpenAI-compat HTTP)| |
                 |  +--------------------------------------------+ |
                 +--------------------------------------------------+
                                   |
                    +--------------+--------------+
                    |                             |
            Any cloud provider             Local model
            (OpenAI/Anthropic/...)        (Ollama/llama.cpp/vLLM)
```

---

## 2. Resolved Decisions (PRD open questions)

| # | Question | Decision | Rationale |
|---|---|---|---|
| Q1 | SurrealDB embedded vs sidecar | **Managed sidecar** (daemon auto-starts a local SurrealDB; Docker or bundled binary) | SurrealDB's Go support is client/server; a managed sidecar keeps the brain durable and upgradable. Embedded revisited later. |
| Q2 | GUI framework | **Solid** (SolidJS) | Fast, fine-grained reactivity, small bundles; matches the user's prior SolidJS experience (agence). React is an acceptable fallback. |
| Q3 | Default zero-config model | **Ollama** if detected locally; else prompt for a provider key | Ollama is the most common local runtime; gives a true "runs with no key" path. |
| Q4 | Telemetry | **Off by default**; opt-in anonymous only | Local-first, privacy-first brand promise. |

---

## 3. Architecture Decision Records (ADRs)

### ADR-001: Go for the core runtime
- **Context:** Need a cross-platform single binary, a home-server daemon, parallel
  agents, and mobile clients.
- **Decision:** Implement the daemon, agent loop, tools, sandbox, plugins, and server
  in Go. The GUI is a separate web frontend (TypeScript/Solid).
- **Consequences:** (+) single static binary per OS/arch, great concurrency, robust
  daemon. (-) thinner AI ecosystem than TS - mitigated because LLM access is plain
  OpenAI-compatible HTTP and a Go MCP SDK exists.

### ADR-002: SurrealDB as the Cortex store
- **Context:** The brain needs graph (relationships), vector (RAG), and document
  (content) queries in one store.
- **Decision:** Use SurrealDB (multi-model: graph + vector + document + KV).
- **Consequences:** (+) one store for neurons/synapses/engrams + sessions + audit;
  native `RELATE` graph + vector search. (-) a companion service to manage (see ADR-003).

### ADR-003: SurrealDB as a managed sidecar
- **Decision:** The daemon auto-starts/manages a local SurrealDB (Docker container or
  bundled binary) and connects over `ws://localhost:8000/rpc`.
- **Consequences:** (+) durable, upgradeable brain; consistent with Open Notebook. (-)
  one more process; the daemon must health-check/restart it.

### ADR-004: OpenAI-compatible HTTP for providers (no SDK lock-in)
- **Decision:** A single provider abstraction calls OpenAI-compatible `/chat/completions`
  (streaming) over HTTP; per-provider adapters only adjust base URL, auth, and quirks.
- **Consequences:** (+) any provider or local model via one code path; no vendor SDK
  dependency. (-) we own retry/streaming/usage parsing (small, well-understood).

### ADR-005: Modular packages + typed plugin SDK + central registry
- **Decision:** Clean Go package boundaries; plugins register capabilities via a typed
  `MimirPluginApi` into a central `PluginRegistry`; plugins import only
  `@mimirmind/plugin-sdk` barrels, never core internals.
- **Consequences:** (+) extensible without forking; testable modules (the lesson from
  agence's fork debt). (-) requires a stable SDK contract + contract tests.

### ADR-006: Solid for the GUI
- **Decision:** Web GUI in SolidJS, served by the daemon; packaged as a desktop app via
  Wails.
- **Consequences:** (+) fast, small, fine-grained; user familiarity. (-) smaller
  ecosystem than React (acceptable).

### ADR-007: Tiered sandbox (containers + microVMs)
- **Decision:** Sandbox Manager offers containers (fast, trusted code) and
  Firecracker/gVisor microVMs (strong, untrusted code), with an in-guest agent exposing
  exec/files/PTY; the agent picks per task.
- **Consequences:** (+) strong isolation on demand; snapshot/fork; per-user isolation.
  (-) complexity; containers ship first, microVMs later.

### ADR-008: Neural data model (neurons / synapses / engrams)
- **Decision:** Model knowledge as a neural graph: neurons (nodes), synapses (edges),
  engrams (durable memories), stored in SurrealDB.
- **Consequences:** (+) associative, self-organizing memory; distinctive identity. (-)
  more schema design than flat rows (worth it).

### ADR-009: Fail-closed policy gate (deterministic governance)
- **Decision:** Every tool/plugin call passes a deterministic policy gate BEFORE
  execution; default-deny with allow/require_approval rules; unbypassable.
- **Consequences:** (+) misbehavior is structurally impossible, not prompt-discouraged.
  (-) policy authoring overhead (ship sensible defaults).

### ADR-010: Ollama as the default local model; telemetry off
- **Decision:** Zero-config path uses Ollama if present; telemetry defaults off.
- **Consequences:** (+) true "no key, no cloud" first run; privacy promise kept.

---

## 4. Go Module Layout

```
mimir/                          # module: github.com/David2024patton/Mimir
  cmd/
    mimir/                      # main: daemon + CLI dispatch (cobra)
  internal/
    agent/                      # agent loop state machine
      loop.go                   # run/attempt/subscribe; intake->infer->act->verify
      state.go                  # states: idle/running/awaiting_approval/compacting/done/error
      subagent.go               # isolated-context subagents (goroutines)
      compaction.go             # context compaction
      prompts.go                # system prompt assembly (AGENTS.md + skills + memory)
    llm/                        # provider abstraction
      provider.go               # Provider interface (Generate/Stream)
      registry.go               # provider registry + routing + model catalog
      openai_compat.go          # OpenAI-compatible HTTP transport (SSE streaming)
      credentials.go            # auth.json + encrypted key store
    tools/                      # built-in native tools
      registry.go               # tool registry (native + MCP)
      shell.go filesystem.go process.go codeexec.go web.go
      sandbox.go                # sandbox_create/exec/destroy tools
    cortex/                     # the brain (SurrealDB)
      store.go                  # SurrealDB client + queries
      neuron.go synapse.go engram.go
      ingest.go                 # extract/transcribe/chunk/embed pipeline
      rag.go                    # hybrid full-text + vector + graph retrieval
      memory.go                 # decay, auto-capture, consolidation, recall
    plugins/                    # plugin system
      registry.go               # PluginRegistry (central)
      api.go                    # MimirPluginApi (register* contract)
      loader.go                 # load pipeline (Go plugins / wasm / subprocess)
    policy/                     # governance
      gate.go                   # fail-closed policy gate
      rules.go                  # YAML/Cedar-like rules + privilege rings
      audit.go                  # tamper-evident (hash-chained) audit + Decision BOM
    sandbox/                    # sandbox manager
      manager.go                # tiered: container vs microVM; warm pool
      guest.go                  # in-guest agent protocol (exec/files/PTY)
      snapshot.go               # snapshot/fork; metadata in SurrealDB
    sessions/                   # session manager + lanes/queues
    hooks/                      # two-tier hooks (gateway + plugin lifecycle)
    skills/                     # skills discovery + snapshots
    spec/                       # spec-driven workflow + goal/autonomous mode
    server/                     # HTTP + WebSocket API + GUI serving
      rest.go ws.go gui.go
    channels/                   # chat channel adapters (Discord/Slack/Telegram) - later
    tunnel/                     # secure remote access - later
    config/                     # mimir.json loader + {env}/{file} interpolation
  gui/                          # Solid web frontend (separate TS project)
  desktop/                      # Wails packaging
  plugins/                      # bundled plugins
  mimir.json                    # default config
  .env.example
  go.mod
```

---

## 5. Cortex Schema (SurrealDB / SurrealQL)

```sql
-- Neurons: the knowledge nodes (sources, notes, concepts, memories, skills)
DEFINE TABLE neuron SCHEMAFULL;
DEFINE FIELD kind        ON neuron TYPE string
  ASSERT $value IN ["source","note","concept","memory","skill"];
DEFINE FIELD layer       ON neuron TYPE string
  ASSERT $value IN ["activity","context","experience","identity","preference"];
DEFINE FIELD scope       ON neuron TYPE string DEFAULT "project";   -- or "__global__"
DEFINE FIELD title       ON neuron TYPE string;
DEFINE FIELD content     ON neuron TYPE string;
DEFINE FIELD summary     ON neuron TYPE option<string>;
DEFINE FIELD embedding   ON neuron FLEXIBLE TYPE option<array<float>>;  -- vector
DEFINE FIELD importance  ON neuron TYPE string DEFAULT "medium"
  ASSERT $value IN ["low","medium","high","critical"];
DEFINE FIELD decay_score ON neuron TYPE float DEFAULT 1.0;            -- forgetting curve
DEFINE FIELD access_count ON neuron TYPE int DEFAULT 0;               -- reinforcement
DEFINE FIELD last_accessed ON neuron TYPE option<datetime>;
DEFINE FIELD expires_at  ON neuron TYPE option<datetime>;
DEFINE FIELD provenance  ON neuron FLEXIBLE TYPE option<object>;      -- {docPath,docHash,url}
DEFINE FIELD confidence  ON neuron TYPE option<float>;
DEFINE FIELD created_at  ON neuron TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at  ON neuron TYPE datetime DEFAULT time::now();
DEFINE INDEX neuron_embedding ON neuron FIELDS embedding MTREE DIMENSION 1536;
DEFINE INDEX neuron_fts ON neuron FIELDS title, content SEARCH ANALYZER ascii BM25;

-- Synapses: typed relationships between neurons (graph edges)
DEFINE TABLE synapse SCHEMAFULL;
DEFINE FIELD in    ON synapse TYPE record<neuron>;
DEFINE FIELD out   ON synapse TYPE record<neuron>;
DEFINE FIELD kind  ON synapse TYPE string
  ASSERT $value IN ["references","derives_from","relates_to","contradicts","cross_layer_link"];
DEFINE FIELD weight ON synapse TYPE float DEFAULT 1.0;
DEFINE FIELD created_at ON synapse TYPE datetime DEFAULT time::now();
-- usage: RELATE neuron:a->synapse->neuron:b SET kind="relates_to", weight=0.8;

-- Engrams: durable memories (a neuron kind=memory promoted to long-term)
DEFINE TABLE engram SCHEMAFULL;
DEFINE FIELD neuron   ON engram TYPE record<neuron>;
DEFINE FIELD strength ON engram TYPE float DEFAULT 1.0;
DEFINE FIELD consolidated ON engram TYPE bool DEFAULT false;

-- Sessions / messages
DEFINE TABLE session SCHEMAFULL;
DEFINE FIELD project ON session TYPE string;
DEFINE FIELD agent   ON session TYPE string DEFAULT "default";
DEFINE FIELD created_at ON session TYPE datetime DEFAULT time::now();
DEFINE TABLE message SCHEMAFULL;
DEFINE FIELD session ON message TYPE record<session>;
DEFINE FIELD role    ON message TYPE string ASSERT $value IN ["user","assistant","tool","system"];
DEFINE FIELD content ON message TYPE string;
DEFINE FIELD tool_calls ON message FLEXIBLE TYPE option<array>;
DEFINE FIELD created_at ON message TYPE datetime DEFAULT time::now();

-- Audit: tamper-evident, hash-chained
DEFINE TABLE audit SCHEMAFULL;
DEFINE FIELD prev_hash ON audit TYPE string;
DEFINE FIELD hash      ON audit TYPE string;        -- hash(prev_hash + tool + args + decision)
DEFINE FIELD tool      ON audit TYPE string;
DEFINE FIELD args      ON audit FLEXIBLE TYPE object;
DEFINE FIELD policy_decision ON audit TYPE string ASSERT $value IN ["allow","deny","require_approval"];
DEFINE FIELD result_code ON audit TYPE option<int>;
DEFINE FIELD created_at ON audit TYPE datetime DEFAULT time::now();

-- Skills, tasks, sandboxes
DEFINE TABLE skill SCHEMAFULL;
DEFINE FIELD name ON skill TYPE string;
DEFINE FIELD content ON skill TYPE string;
DEFINE FIELD source ON skill TYPE string ASSERT $value IN ["reflected","quality_gate","manual"];
DEFINE TABLE task SCHEMAFULL;
DEFINE FIELD goal ON task TYPE string;
DEFINE FIELD status ON task TYPE string DEFAULT "pending";
DEFINE FIELD plan ON task FLEXIBLE TYPE option<object>;
DEFINE TABLE sandbox SCHEMAFULL;
DEFINE FIELD backend ON sandbox TYPE string ASSERT $value IN ["container","microvm"];
DEFINE FIELD status ON sandbox TYPE string DEFAULT "creating";
DEFINE FIELD snapshot_id ON sandbox TYPE option<string>;
```

---

## 6. Agent-Loop State Machine

States: `idle -> running -> (awaiting_approval) -> running -> (compacting) -> done | error`.

```
run(sessionKey, prompt):
  1. INTAKE        - resolve session, persist user message
  2. ASSEMBLE      - build system prompt: AGENTS.md + active skills + recalled
                     engrams (<past_learnings>) + tool schemas
  3. INFER (stream)- call provider; stream tokens to UI
  4. ACT           - on tool_call: policy gate -> (approve?) -> execute in
                     sandbox/native -> persist result -> capture memory if outcome
  5. LOOP          - feed result back; goto 3 until model stops
  6. COMPACT       - if near limit, compact (re-inject memory + skills)
  7. DONE/ERROR    - persist final; emit lifecycle event
```
- Runs are serialized per session lane (no tool/session races); subagents run on their
  own goroutines with isolated context, returning only a summary.
- The policy gate (ADR-009) wraps every ACT step; a `require_approval` transitions to
  `awaiting_approval` and pauses until the user responds (GUI/TUI/mobile).

---

## 7. Plugin SDK Contract (Go)

```go
// MimirPluginApi is the only surface plugins use to extend Mimir.
type MimirPluginApi interface {
    RegisterProvider(p ProviderDef) error
    RegisterTool(t ToolDef) error
    RegisterHook(h HookDef) error
    RegisterChannel(c ChannelDef) error          // chat channels
    RegisterHarness(h HarnessDef) error          // pluggable agent runtimes
    RegisterCapability(c CapabilityDef) error    // embedding/speech/image/etc.
    RegisterCommand(cmd CommandDef) error
    Runtime() RuntimeHelpers                     // config, logger, cortex access
}

// A plugin exports a single registration entrypoint:
type Plugin interface {
    Manifest() Manifest                          // id, version, contracts
    Register(api MimirPluginApi) error
}
```
- Plugins import only `@mimirmind/plugin-sdk`; never `internal/**`.
- The `PluginRegistry` collects registrations; core reads the registry (one-way).
- **Contract tests** assert each plugin registers exactly what its manifest declares.
- Plugin runtimes (v1): in-process Go plugins; later: wasm / subprocess / MCP.

---

## 8. Provider Abstraction

```go
type Provider interface {
    ID() string
    Generate(ctx, req GenerateRequest) (GenerateResponse, error)
    Stream(ctx, req GenerateRequest) (<-chan StreamEvent, error)  // SSE deltas + usage
}
```
- One `openaiCompatProvider` covers all OpenAI-compatible endpoints (cloud + local).
- Provider registry routes by model string (`anthropic/claude-...`, `ollama/...`).
- Base URL + API key from `mimir.json` / env / `auth.json` (encrypted).
- Default local: Ollama at `http://localhost:11434` (ADR-010).

---

## 9. GUI <-> Daemon Protocol

- **REST** (`/api/...`): CRUD for projects, sessions, cortex (neurons/search), config,
  auth, tasks, sandboxes. The GUI uses an **SWR-style** cache (keyed dedup +
  stale-while-revalidate + optimistic updates).
- **WebSocket** (`/ws`): push stream of typed events - `token` (model deltas),
  `tool` (start/update/end), `lifecycle` (start/end/error/awaiting_approval),
  `status` (sandbox/agent state). The GUI renders these live.
- **GUI serving:** the daemon serves the built Solid app at `http://localhost:<port>`;
  Wails wraps the same app for desktop.

---

## 10. Sandbox Architecture

- **Sandbox Manager** maintains a warm pool; `sandbox_create` boots a container (fast)
  or microVM (strong) in ~ms-s; `sandbox_exec` runs commands via the in-guest agent;
  `sandbox_destroy` reaps. Snapshots (container commit / Firecracker snapshot) stored
  with metadata in SurrealDB; `fork` clones from a snapshot.
- **In-guest agent:** a tiny daemon inside each sandbox exposing exec / file read-write
  / PTY over a socket - one protocol regardless of backend.
- **Port routing:** local reverse proxy maps `localhost:N` -> guest port; the GUI
  previews apps at `http://localhost:N`.
- **Per-user isolation:** one Linux user per agent; setgid shared group dirs.
- **Egress policy:** per-sandbox allow/deny via nftables or a userspace proxy.

---

## 11. Security Architecture

- **Policy gate (fail-closed):** every tool/plugin call -> evaluate rules (default-deny)
  -> allow / deny / require_approval. Unbypassable, even in bypass mode.
- **Privilege rings:** tools/plugins assigned to 4 tiers; higher tiers need more trust.
- **Audit:** hash-chained (Merkle) records of policy + request + decision (Decision BOM).
- **Secrets:** API keys encrypted at rest (MIMIR_ENCRYPTION_KEY); `.env` gitignored.
- **Sandboxing:** untrusted code runs in containers/microVMs (ADR-007).
- **MCP security gateway** (Tier 2): tool-poisoning/drift/hidden-instruction scanning.

---

## 12. Request Lifecycle (end to end)

1. User types in GUI -> POST `/api/sessions/:id/messages` + WS subscribe.
2. Daemon `run()`: intake -> assemble (AGENTS.md + skills + recalled engrams).
3. Provider.Stream() -> tokens pushed over WS -> GUI renders.
4. Tool call -> policy gate -> (approval?) -> execute (native or sandbox) -> result.
5. Outcome capture: correction/preference/failure -> engram in Cortex.
6. Loop until done -> persist -> lifecycle `end` over WS.

---

## 13. Deployment Architecture

- **Default (local-first):** one `mimir` binary + a managed SurrealDB sidecar (Docker or
  bundled). GUI at `localhost:<port>`. No cloud required.
- **Optional hosted layer (business):** gateway.mimirmind.com (model gateway),
  console.mimirmind.com (billing/keys), relay.mimirmind.com (mobile tunnel), hosted
  Cortex sync. All optional; the local install is fully functional without them.

---

## 14. Implementation Readiness Notes

- **Ready to build:** Go module skeleton (E1), provider abstraction + agent loop (E2),
  native tools (E3), policy gate (E4), plugin SDK + registry (E11), GUI shell (E12).
- **Spike first:** SurrealDB sidecar management (auto-start/health/restart) and the
  in-guest sandbox agent protocol - these are the two least-proven mechanisms.
- **De-risk:** the plugin SDK contract must freeze early (everything depends on it);
  write contract tests alongside E11.
- **Sequencing:** E1 -> E2 -> E3 -> E4 -> E11 -> E12 -> E5/E6/E7/E8/E9 -> E10/E14.

Next: `bmad-check-implementation-readiness` to gate the build, then John breaks the
Core epics into stories.
