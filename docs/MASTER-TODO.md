# Mímir Polyglot Software Engineering & Creation Harness -- Master Todo Blueprint

> **Purpose of This Document**: This is the authoritative, step-by-step implementation manual for Mímir's **4-Stack Polyglot Software Engineering & Creation Harness** (`Go + Rust + Mojo + Flutter`) and **Level 1-3 Software Factory Architecture**. Every phase, task, and sub-task is specified down to exact file paths, structs, CLI flags, FFI signatures, verification commands, and SEO/GEO/Web Creation rules so that any engineer or autonomous coding agent can execute the build from start to finish without ambiguity.

---

## Part I: How Mímir's Engineering Process Works Right Now (Zero-Bloat 7-Phase SDLC Pipeline)

Mímir is **not a passive chatbot**. It is a **Software Engineering & Creation Harness** that follows the **7-Phase Software Development Life Cycle (SDLC)** across **Agile**, **Waterfall**, and **DevOps** execution models.

To prevent prompt bloat for both frontier and tiny local LLMs (`SmolMode`), **Mímir never dumps all engineering or SEO rules into the base prompt at once**. Instead, [`internal/sdlc/sdlc.go`](../internal/sdlc/sdlc.go) and [`internal/sdlc/phases.go`](../internal/sdlc/phases.go) load **only the ~40-75 token micro-prompt, Norse Persona, and tool whitelist for the currently active SDLC step**:

| Step | Phase ID | Core Goal | Norse Persona | Primary Deliverable (`.mimir/sdlc/`) | Phase-Scoped Micro-Prompt & Web Overlay |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1** | `planning` | Define project scope & feasibility | `Odin` | `01-PLAN.md` (*Project Plan & Charter*) | Scope, feasibility, cost/schedule, technical risk, and initial `todo` breakdown. |
| **2** | `analysis` | Gather exact system requirements | `Huginn` | `02-SRS.md` (*Software Requirement Specification*) | Yggdrasil Mind Map inspection (`ingest_repo`), functional/security constraints (+ Web Entity/NAP/Lead fields if web task). |
| **3** | `design` | Blueprint the system architecture | `Odin` | `03-SDD.md` (*System Design Document*) | Modular folder layout, DB schemas, C-ABI/API contracts, UI wireframes (+ Semantic HTML5, JSON-LD `Schema.org`, `/llms.txt` blueprint if web task). |
| **4** | `implementation` | Write the actual source code | `Thor` | `04-BUILD.md` (*Working Source Code & Builds*) | Surgical `edit_block` diffs (<15 lines), modular packages, zero em dashes (+ Semantic HTML5, JSON-LD, Anti-Slop copy, `[vendor claim]` provenance, and first-party lead capture if web task). |
| **5** | `testing` | Verify quality & squash bugs | `Loki` | `05-QA-REPORT.md` (*QA Test Reports & Bug-Free Build*) | Linter, static analysis, and `SelectAffectedTests` against Step 2 SRS (+ Schema.org & `/llms.txt` validation if web task). |
| **6** | `deployment` | Release software to production | `Heimdall` | `06-RELEASE.md` (*Live Production Release*) | CI/CD pipelines, container/server verification, rollback readiness (+ canonical tags, GBP indexability, live lead beacon if web task). |
| **7** | `maintenance` | Monitor, patch, and iterate | `Muninn` | `07-MAINTENANCE.md` (*Updates, Patches & Logs*) | Telemetry/logs, vulnerability patching, regression fixes, and Yggdrasil Mind Map refresh. |

### The 3 SDLC Execution Models (`mimir sdlc agile|waterfall|devops`)
- **`agile` (Default -- Iterative Sprints)**: Loops smoothly through steps `1 -> 6`, auto-selecting the matching SDLC phase micro-prompt per turn (`DetectPhaseFromPrompt`) while keeping the base prompt razor-thin.
- **`waterfall` (Linear Strict Gate)**: Disables auto-jumping and requires each phase (`1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7`) to write its `.mimir/sdlc/0X-*.md` deliverable and be explicitly signed off (`sdlc` tool `save_deliverable` / `sign_off`) before the next phase can begin.
- **`devops` (Continuous CI/CD)**: Automates the transition chain `Implementation (4) -> Testing (5) -> Deployment (6)` so code is linted, tested, and release-gated immediately after implementation edits.

### Modular Folder Structure (`G:\Mimir`)
```text
G:\Mimir\
├── cmd/mimir/main.go                # CLI & Daemon entrypoint (setup, ingest, sdlc, providers, serve)
├── internal/
│   ├── sdlc/                        # 7-Phase SDLC State Machine & Zero-Bloat Phase Prompt Loader
│   │   ├── phases.go                # Step 1..7 specs, micro-prompts, allowed tools, Web overlays
│   │   ├── sdlc.go                  # Agile/Waterfall/DevOps state machine & deliverable writer
│   │   └── sdlc_test.go             # Unit tests for zero-bloat prompts & Waterfall gates
│   ├── decision/                    # Huginn-Decider (Liquid AI d1-omni-600M System-One Router)
│   ├── ingest/                      # Gjallarhorn-Ingest (.mimir/digest.txt + Yggdrasil Mind Map)
│   ├── agent/                       # Gather -> Infer -> Act -> Verify loop, SmolMode & OnStop gate
│   ├── config/                      # Zero-config discovery + OpenCode BYOK & Subscription catalog
│   ├── cortex/                      # Persistent Neuron & Session memory store
│   ├── tools/                       # Modular tools (shell, edit_block, todo, ingest_repo, sdlc, skills)
│   └── tui/                         # Interactive Charmbracelet TUI (/sdlc, /ingest, /mindmap, /decide)
├── crates/mimir-rs/                 # Stack 2 (Rust): Tree-sitter AST, Tantivy BM25, Ratatoskr-FS Sandbox
├── mojo/mimir_mojo/                 # Stack 3 (Mojo): .itf mmap, SIMD Cosine, 10k-bit VSA, d1-omni-600M
├── ui/mimir_flutter/                # Stack 4 (Flutter): Desktop Studio + Happy E2EE Mobile Companion
└── .mimir/
    ├── cortex.json                  # Persistent Cortex neurons + Yggdrasil Mind Map graph
    ├── digest.txt                   # Latest GitIngest repository digest
    └── sdlc/                        # Phase deliverables (01-PLAN.md through 07-MAINTENANCE.md)
```

---

## Part II: The 4-Stack Polyglot Architecture Overview

| Stack Layer | Language | Codename | Directory | Primary Responsibility |
| :--- | :--- | :--- | :--- | :--- |
| **Stack 1: Control Plane & Orchestrator** | **Go 1.25+** | `Bifrost-Core` | `cmd/mimir/`, `internal/` | 7-Phase SDLC Engine, Agent loop, 12+ BYOK & Subscription LLM routers (Meta Muse, MiMo, OpenAI, Anthropic, Ollama), Bubble Tea TUI, Web/SSE Daemon, E2EE Happy Relay, GitIngest + Yggdrasil Mind Map. |
| **Stack 2: Deterministic Systems & Sandbox Core** | **Rust (2024)** | `Tyr-Core` | `crates/mimir-rs/` | Tree-sitter multi-language AST structural editor, Tantivy BM25 lexical index, OS process/filesystem jail (`Crabbox` / Job Objects), git worktree snapshotter (`Ratatoskr-FS`), exposed via C-ABI (`include/mimir_rs.h`). |
| **Stack 3: Bare-Metal Tensor & VSA Compute** | **Mojo (`24.6+`)** | `Mjolnir-Compute` | `mojo/mimir_mojo/` | Zero-copy `.itf` (iTaK Tensor Format) mmap loader, SIMD AVX-512 / NEON cosine similarity & 10,000-bit Hyperdimensional VSA associative binding, local `Liquid AI d1-omni-600M` System-One Decision Engine accelerator, exposed via C-ABI (`include/mimir_mojo.h`). |
| **Stack 4: Cross-Platform 60fps Omnichannel UI** | **Flutter / Dart 3** | `Gjallarhorn-UI` | `ui/mimir_flutter/` | Native Desktop Studio (Windows/macOS/Linux) + Mobile Companion App (iOS/Android) with QR-code E2EE pairing (`Happy` pattern), 1-tap `Approve / Reject` diff cards, and 2D/3D Software Factory Swarm visualizer. |

---

## Part III: Master Step-by-Step Todo Checklist (With Granular Sub-Lists)

### Phase 1: Go Control Plane, 7-Phase SDLC Engine, Decision Engine & GitIngest Mind Map (`Bifrost-Core`)
- [x] **1.1 Build System-One Decision Engine (`Huginn-Decider` -- `Liquid AI d1-omni-600M`)**
  - [x] `1.1.a` Create `internal/decision/decider.go` setting `PrimaryDecisionModel = "liquid/d1-omni-600m"` (~400 MB multimodal LFM2.5) with fallback catalog (`jev-1.7b`, `lfm2.5-1.2b-thinking`, `qwen3-0.6b`, `meta/muse-glimmer`).
  - [x] `1.1.b` Implement core primitives: `Choice(ctx, query, candidates)`, `Score(ctx, query, criteria)`, and `Noul(ctx, taskDescription)` with deterministic zero-latency heuristic fallback when the local server is offline.
  - [x] `1.1.c` Implement harness decision gates: `RoutePrompt` (persona + tier + `Think` toggle), `PruneToolNames`, `AssessCommandSafety`, `SelectMindMapFiles`, and `SelectAffectedTests` (`Superlinear` test-selection pattern).
  - [x] `1.1.d` Implement `EnsureLocalDecisionModel` and CLI command `mimir setup` in `cmd/mimir/main.go` to auto-pull `d1-omni-600m` via Ollama or HuggingFace GGUF.
  - [x] `1.1.e` Verify with unit tests in `internal/decision/decider_test.go` (`go test ./internal/decision/...`).
- [x] **1.2 Build Native GitIngest + Yggdrasil Mind Map (`Gjallarhorn-Ingest`)**
  - [x] `1.2.a` Create `internal/ingest/ingest.go` to walk local directories or clone remote GitHub repos (`https://github.com/owner/repo`), using a built-in directory skip list (`node_modules`, `vendor`, `dist`, `build`, and similar) and binary filters (`.gitignore` files are not read).
  - [x] `1.2.b` Extract multi-language symbols (Go, Rust, Python, TS/JS; no Mojo or Dart symbol patterns yet) and import edges into `MindMap` (`Node`, `Edge`, `IsTest`, `TokenEstimate`).
  - [x] `1.2.c` Generate `.mimir/digest.txt` (GitIngest format) + `.mimir/mindmap.json` and persist linked `KindCode` neurons (`PutNeuron` + `Relate`) into Cortex memory.
  - [x] `1.2.d` Expose `ingest_repo` tool (`internal/tools/ingest_tool.go`), CLI subcommand `mimir ingest <path|url>`, and TUI slash commands `/ingest` and `/mindmap`.
  - [x] `1.2.e` Verify with unit tests in `internal/ingest/ingest_test.go` and live self-ingestion (`.\bin\mimir.exe ingest .`).
- [x] **1.3 Build OpenCode-Style Provider Catalog, Meta Muse Support & Subscription Tiers**
  - [x] `1.3.a` Create `internal/config/catalog.go` defining 3 subscription tiers: **Mímir Sovereign (`$0/mo` Local + BYOK)**, **Mímir Well Pro (`$20/mo` `MIMIR_SUBSCRIPTION_KEY`)**, and **Mímir Odin Max (`$100/mo` parallel swarms + remote `Crabbox` test offloading)**.
  - [x] `1.3.b` Register 12+ BYOK providers in `internal/config/catalog.go` and `internal/config/discovery.go`, including **Meta Muse (`META_API_KEY`: `muse-wonder-1.0`, `muse-spark-1.0`, `muse-glimmer`, `muse-voice-transcribe`, `sam-3.1`)** and **Xiaomi MiMo (`MIMO_API_KEY`)**.
  - [x] `1.3.c` Wire `mimir providers` CLI subcommand and `/providers` TUI command.
- [x] **1.4 Build Modular 7-Phase SDLC Framework & Zero-Bloat Phase-Scoped Prompts (`internal/sdlc`)**
  - [x] `1.4.a` Create `internal/sdlc/phases.go` defining the 7 SDLC phases (`1: planning`, `2: analysis`, `3: design`, `4: implementation`, `5: testing`, `6: deployment`, `7: maintenance`), each with its own `<300 char` `MicroPrompt`, `AllowedTools`, Norse `Persona`, `.mimir/sdlc/0X-*.md` deliverable, and optional 1-line `WebOverlay`.
  - [x] `1.4.b` Create `internal/sdlc/sdlc.go` implementing `Manager` with support for `MethodAgile`, `MethodWaterfall` (strict linear sign-off gates), and `MethodDevOps` (continuous `4 -> 5 -> 6` pipeline), plus `DetectPhaseFromPrompt` and `BuildPhasePrompt`.
  - [x] `1.4.c` Create `internal/tools/sdlc_tool.go` (`sdlc` tool) and wire `/sdlc` into `internal/tui/tui.go` and `mimir sdlc` into `cmd/mimir/main.go`.
  - [x] `1.4.d` Refactor `internal/agent/rules.go` (`BuildPhaseScopedRulesBlock`) and `internal/agent/agent.go` (`buildPromptForInput`) so only the active SDLC phase micro-prompt loads on each turn, keeping the main prompt small for both big and tiny LLMs.
  - [x] `1.4.e` Verify with unit tests in `internal/sdlc/sdlc_test.go` and `internal/agent/rules_test.go`.

---

### Phase 2: Level 1-3 Harness Engineering & Autonomous Software Factory Upgrades (`Bifrost-Core`)
> *Inspired by Leon van Zyl's 3 Levels of Harness Engineering (`YouTube_Video__z1EaBe9oIME_.txt`), `AutoForge`, and `Superlinear`.*

- [x] **2.1 Implement Level 1 Harness Lifecycle Hooks (`internal/agent/hooks.go`)**
  - [x] `2.1.a` Define `HookType` constants: `HookPrePrompt`, `HookPreTool`, `HookPostTool`, and `HookOnStop`.
  - [x] `2.1.b` Create `HookRunner` struct in `internal/agent/hooks.go` that reads optional `.mimir/hooks.json` shell or internal rules and runs them deterministically at each lifecycle boundary.
  - [x] `2.1.c` Wire `HookPostTool` after `write_file` / `edit_block` to automatically trigger fast syntax/linter checks (`gofmt`, `cargo check`, `tsc --noEmit`) and append any compiler errors directly to the tool output.
  - [x] `2.1.d` Wire `HookOnStop` in `runLoop` and `RunStream` (`internal/agent/agent.go`): if `EvaluateStopGate` detects unverified file mutations or incomplete `[pending]` Todo items when `AutoVerify` is enabled, inject a system reminder forcing the agent to run verification tests before emitting `EventDone`.
  - [x] `2.1.e` Write unit tests in `internal/agent/hooks_test.go` verifying that `HookPostTool` catches syntax errors and `HookOnStop` blocks premature completion.
- [x] **2.2 Implement Session Handoff & Context Continuity (`HANDOFF.md` Generator)**
  - [x] `2.2.a` Create `internal/agent/handoff.go` with `GenerateHandoffMarkdown(sessionID, history, todoLines, mindMapSummary)`.
  - [x] `2.2.b` Structured sections required in `.mimir/HANDOFF.md`:
    1. `## Current Objective & Architectural State`
    2. `## Completed Tasks & Verified Files`
    3. `## Remaining Todo Checklist (In-Progress & Pending)`
    4. `## Failed Approaches & Guardrails (Do Not Repeat)`
    5. `## Exact Next Command to Run`
  - [x] `2.2.c` Expose `/handoff` slash command in `internal/tui/tui.go` and `mimir handoff` CLI command in `cmd/mimir/main.go`.
  - [x] `2.2.d` Auto-trigger `GenerateHandoffMarkdown` whenever token usage crosses 80% of the model's context window (`CompactionThreshold`).
- [x] **2.3 Build Built-In Web/SEO/GEO + Impeccable Anti-Slop Auditor (`internal/tools/web_audit_tool.go`)**
  - [x] `2.3.a` Create `web_seo_audit` tool in `internal/tools/web_audit_tool.go` (registered in `internal/tools/default.go`) + durable `PRODUCT.md` / `DESIGN.md` context loader (`LoadDesignContext`).
  - [x] `2.3.b` Check and score the **4 Layers of Modern Search & Conversion + Impeccable Anti-Slop UI Rules**:
    - Presence and validity of `<title>`, `<meta name="description">`, single `<h1>` hierarchy, and semantic landmarks (`<main>`, `<header>`, `<nav>`, `<footer>`).
    - Presence and JSON validity of `<script type="application/ld+json">` (`LocalBusiness`, `SoftwareApplication`, `FAQPage`, `Product`).
    - Presence of `/robots.txt`, `/sitemap.xml`, and `/llms.txt`.
    - Impeccable deterministic detectors for banned AI slop phrases ("in today's fast-paced world", "unleash the power"), em dashes (`\u2014`), default purple/indigo gradient clichés (`from-purple-`, `from-indigo-`), nested `.card` soup, and unlabeled statistical claims (`[vendor claim]`).
  - [x] `2.3.c` Add CLI command `mimir audit-web [dir]` and TUI command `/audit-web`.
  - [x] `2.3.d` Verify with unit tests in `internal/tools/web_audit_test.go`.
- [x] **2.4 Build Level 3 Multi-Agent Software Factory Queue (`internal/swarm/factory.go`)**
  - [x] `2.4.a` Create `internal/swarm/factory.go` implementing a kanban state machine (`Backlog -> Spec -> Implementing -> Testing -> Security/SEO Audit -> Done`) stored in a JSON state file (`.mimir/swarm/tasks.json`).
  - [x] `2.4.b` Current state of the pipeline: `Odin` (Architect) writes a templated spec document from the task title and description (no LLM task decomposition yet), and each task runs sequentially in its own isolated git worktree. The `Thor` step currently records a templated build record only and does not write code. `Loki` runs the worktree test suite and `Heimdall` runs the web audit plus an epistemic critic pass over the task description; after those checks the branch merges automatically, with no human approval gate.
  - [x] `2.4.c` Expose swarm task submit/list endpoints on the Web/SSE daemon (`internal/server`). Real-time worker state event streaming (`WorkerSpawned`, `WorktreeCreated`, `ToolRunning`, `TestsPassed`, `Merged`) for the Web & Flutter UIs is not implemented yet.
- [x] **2.5 Build Unattended Run Engine & Two-Tier Testing Law (`internal/runengine/runengine.go`)**
  - [x] `2.5.a` Implement shell-less workspace file `Jail` (`Resolve`, `CheckWrite`, `ReadFile`, `WriteFile`, `ListDir`) blocking `..` traversal, symlink escapes, and writes to `.git/` or `.mimir/run.lock`.
  - [x] `2.5.b` Implement atomic single-run lockfile (`.mimir/run.lock` with PID, task, timestamp, and automatic stale lock recovery).
  - [x] `2.5.c` Implement `Limits` config (`MaxSteps`, `MaxTokens`, `Timeout`) and `Executor` seam interface (`GoTestExecutor` using discrete argv without shell interpolation).
  - [x] `2.5.d` Enforce the **Two-Tier Testing Law**: Tier-1 inner edit loops run fast affected tests via `Huginn-Decider` (`RunInnerLoopTests`), while Tier-2 outer gates (`ReviewGate` / `ShipGate` via `mimir review` and `mimir ship`) always run the full test suite (`go test ./...`) + Checklist Gate + SDLC Deliverable Gate; the Epistemic Critic Gate runs only when a draft report is supplied to the gate (the unattended path supplies none).
  - [x] `2.5.e` Verify with unit tests in `internal/runengine/runengine_test.go`.

---

### Phase 3: Rust Deterministic Systems & Sandbox Core (`Tyr-Core` -- Stack 2)
> *Located in `G:\Mimir\crates\mimir-rs\`, compiled to a C-ABI library (`mimir_rs.dll` / `libmimir_rs.so` / `libmimir_rs.a`) and bridged into Go via `internal/native/tyr_bridge.go`.*

- [ ] **3.1 Scaffold the Rust Workspace & C-ABI Header**
  - [ ] `3.1.a` Create `G:\Mimir\crates\mimir-rs\Cargo.toml` with `crate-type = ["cdylib", "staticlib"]` and dependencies: `tree-sitter`, `tree-sitter-go`, `tree-sitter-rust`, `tree-sitter-python`, `tree-sitter-typescript`, `tantivy`, `git2`, `ignore`, `serde`, `serde_json`.
  - [ ] `3.1.b` Create `G:\Mimir\include\mimir_rs.h` declaring C-ABI functions:
    - `char* mimir_ast_outline(const char* repo_path);`
    - `char* mimir_ast_patch_symbol(const char* file_path, const char* symbol_name, const char* new_body);`
    - `char* mimir_bm25_search(const char* index_path, const char* query, int top_k);`
    - `char* mimir_worktree_snapshot(const char* repo_path, const char* task_id);`
    - `int mimir_worktree_rollback(const char* repo_path, const char* snapshot_id);`
    - `void mimir_free_string(char* ptr);`
- [ ] **3.2 Implement Tree-Sitter Structural AST Editor (`crates/mimir-rs/src/ast.rs`)**
  - [ ] `3.2.a` Parse source files into concrete syntax trees and extract exact byte/line ranges for functions, methods, structs, traits, and classes.
  - [ ] `3.2.b` Implement `mimir_ast_patch_symbol`: replace a function or struct body by symbol name and validate that the resulting file has zero Tree-sitter `ERROR` or `MISSING` nodes before committing the write to disk (eliminating broken line-number edits on small models).
- [ ] **3.3 Implement Tantivy Sub-Millisecond BM25 Code & Cortex Index (`crates/mimir-rs/src/search.rs`)**
  - [ ] `3.3.a` Index repo files, AST symbol names, and Cortex memory neurons into a memory-mapped Tantivy index under `.mimir/tantivy/`.
  - [ ] `3.3.b` Expose `mimir_bm25_search` returning ranked JSON hits in `<2ms` for hybrid RRF fusion with Mojo's dense/VSA vector scores.
- [ ] **3.4 Implement `Ratatoskr-FS` Transactional Git Worktree Snapshotter & OS Sandbox (`crates/mimir-rs/src/sandbox.rs`)**
  - [ ] `3.4.a` Create shadow git worktrees per agent task so parallel `Thor` workers never collide on the working tree and any failed tool step can roll back in `<5ms`.
  - [ ] `3.4.b` Enforce OS-level process sandboxing (Windows Job Objects with memory/process caps; Linux `landlock` + `seccomp` + `cgroups v2`; macOS `sandbox-exec`) for `run_shell`.
- [ ] **3.5 Wire Go FFI Bridge (`internal/native/tyr_bridge.go`)**
  - [ ] `3.5.a` Create `internal/native/tyr_bridge.go` with build tags (`//go:build tyr_ffi`) and pure-Go fallback (`tyr_fallback.go`) so `go build ./...` always succeeds even on machines without Cargo installed.
  - [ ] `3.5.b` Verify with `cargo test` in `crates/mimir-rs` and `go test ./internal/native/...`.

---

### Phase 4: Mojo Bare-Metal Tensor, VSA & Liquid AI Engine (`Mjolnir-Compute` -- Stack 3)
> *Located in `G:\Mimir\mojo\mimir_mojo\`, compiled via `mojo build --emit shared-lib` and bridged into Go via `internal/native/mjolnir_bridge.go`.*

- [ ] **4.1 Scaffold Mojo Package & C-ABI Header**
  - [ ] `4.1.a` Create `G:\Mimir\mojo\mimir_mojo\mojoproject.toml` and `G:\Mimir\include\mimir_mojo.h`.
  - [ ] `4.1.b` Declare C-ABI exports:
    - `int mimir_itf_open(const char* path);`
    - `int mimir_simd_topk_cosine(const float* query, const float* matrix, int rows, int dim, int k, int* out_indices, float* out_scores);`
    - `int mimir_vsa_bind_hamming(const uint64_t* a, const uint64_t* b, uint64_t* out, int words);`
    - `char* mimir_decider_infer(const char* prompt, int max_tokens);`
- [ ] **4.2 Implement Native `.itf` (iTaK Tensor Format) Zero-Copy Mmap Loader (`mojo/mimir_mojo/itf_loader.mojo`)**
  - [ ] `4.2.a` Implement 64-byte cache-line aligned header parsing (`ITF\x01` magic bytes) and `mmap` pointer binding using Mojo `UnsafePointer` with zero heap copies.
  - [ ] `4.2.b` Support persistent sections for Cortex dense embeddings, 10,000-bit Hyperdimensional VSA concept vectors, and `Liquid AI d1-omni-600M` quantized weights.
- [ ] **4.3 Implement SIMD Vector Search & 10,000-Bit VSA Associative Memory (`mojo/mimir_mojo/vsa_cortex.mojo`)**
  - [ ] `4.3.a` Write vectorized SIMD `DType.float32` / `DType.bfloat16` dot-product and cosine similarity kernels across all CPU SIMD lanes (`simdwidthof[DType.float32]()`).
  - [ ] `4.3.b` Implement 10,000-bit bipolar/boolean Hyperdimensional Computing (`XOR` binding + `POPCNT` Hamming distance) so Mímir can bind `(Repo Symbol * Bug Pattern * Fix Commit)` into a single associative vector with zero backpropagation.
- [ ] **4.4 Implement Bare-Metal `Liquid AI d1-omni-600M` Kernel (`mojo/mimir_mojo/decider_kernel.mojo`)**
  - [ ] `4.4.a` Accelerate `Liquid AI d1-omni-600M` state-space / conv1d + gated attention forward pass directly in Mojo so `Huginn-Decider` (`choice`, `score`, `noul`, `RoutePrompt`, `SelectAffectedTests`) runs in `<5ms` on CPU/iGPU with zero external daemon required.
- [ ] **4.5 Wire Go FFI Bridge (`internal/native/mjolnir_bridge.go`)**
  - [ ] `4.5.a` Create `internal/native/mjolnir_bridge.go` with automatic fallback to `internal/decision/decider.go` and `internal/cortex/sqlite.go` when the Mojo shared library is not present.

---

### Phase 5: Flutter Desktop Studio & Happy Mobile E2EE Companion (`Gjallarhorn-UI` -- Stack 4)
> *Located in `G:\Mimir\ui\mimir_flutter\`, targeting Windows, macOS, Linux, iOS, and Android from a single 60fps Impeller codebase.*

- [x] **5.1 Build E2EE Happy Relay & QR Pairing in Go Daemon (`internal/relay/happy_e2ee.go`)**
  - [x] `5.1.a` Implement X25519 ephemeral key exchange + ChaCha20-Poly1305 payload encryption in `internal/relay/happy_e2ee.go`.
  - [x] `5.1.b` Add CLI command `mimir pair` and TUI command `/pair` that renders an ASCII QR code in the terminal containing `mimir://pair?host=<ip:port>&pubkey=<base64>&token=<nonce>`.
  - [x] `5.1.c` Relay frame encryption helpers exist in `internal/relay`, but streaming encrypted `EventToken`, `EventTool`, `EventSteer`, `StopGateStatus`, and `ApprovalRequest` frames over WebSocket/SSE (`/api/v1/mobile/stream`) is not wired into the agent loop yet; nothing in the binary sends or receives those frames today.
- [x] **5.2 Scaffold Flutter Multi-Platform Project (`ui/mimir_flutter/`)**
  - [x] `5.2.a` Initialize `ui/mimir_flutter/pubspec.yaml` with `flutter_riverpod`, `go_router`, `cryptography` (X25519 + AES-GCM), `mobile_scanner` (QR pairing), and `flutter_highlight`.
  - [x] `5.2.b` Configure Nordic Glassmorphism dark theme (obsidian `#0B0F17`, Bifrost cyan `#00F0FF`, Odin gold `#F59E0B`, Thor emerald `#10B981`, Loki crimson `#EF4444` -- zero generic AI slop).
- [x] **5.3 Build Desktop Studio Views (`ui/mimir_flutter/lib/screens/desktop/`)**
  - [x] `5.3.a` **4-Pane Engineering Workbench**:
    - Left Pane: Interactive **Yggdrasil Mind Map Explorer** (visualizing `MindMap` nodes, imports, and affected tests from `internal/ingest`).
    - Center Pane: **Live Agent Stream & Diff Editor** with clickable `[choice]` pills, `⚡SmolMode` badge, and inline `Huginn-Decider` routing telemetry (`Persona | Model | Latency`).
    - Right Pane: **Level 3 Software Factory Swarm Board** (`Odin`, `Thor`, `Loki`, `Heimdall` workers + live Todo checklist + `OnStop` verification gate status).
    - Bottom Drawer: **4-Layer Web/SEO/GEO & Provider Cost Dashboard** (live token spend, subscription tier status, and `web_seo_audit` scores).
- [x] **5.4 Build Mobile Companion Views (`ui/mimir_flutter/lib/screens/mobile/`)**
  - [x] `5.4.a` **QR Scanner Pairing Screen**: Scan `mimir pair` terminal QR code to establish E2EE session in `<2 seconds`.
  - [x] `5.4.b` **Remote Permission & Diff Approval Cards**: Push notifications + 1-tap `Approve` / `Reject` / `Switch Mode (RO / Edit / Bypass)` cards when `Thor` requests a file write or `CAUTION` shell command while you are away from your desk.
  - [x] `5.4.c` **Voice & Text Mid-Turn Steering Bar**: Send live `Steer(msg)` instructions from your phone directly into the running desktop agent loop (with optional `meta/muse-voice-transcribe` voice input).

---

### Phase 6: End-to-End Build, Benchmark & Release Verification Gates
- [ ] **6.1 Master Build & Test Gate (`Makefile` & `scripts/build-all.ps1`)**
  - [ ] `6.1.a` Run Go static analysis and full test suite: `go vet ./... ; go test -count=1 ./...`.
  - [ ] `6.1.b` Run Rust crate tests and release build: `cargo test --manifest-path crates/mimir-rs/Cargo.toml ; cargo build --release --manifest-path crates/mimir-rs/Cargo.toml`.
  - [ ] `6.1.c` Run Mojo shared library build and SIMD/VSA tests: `mojo test mojo/mimir_mojo`.
  - [ ] `6.1.d` Run Flutter analyzer and widget tests: `flutter analyze ui/mimir_flutter ; flutter test ui/mimir_flutter`.
- [ ] **6.2 Self-Hosted Dogfooding Gate**
  - [ ] `6.2.a` Run `./bin/mimir.exe setup` to verify `Liquid AI d1-omni-600M` local readiness.
  - [ ] `6.2.b` Run `./bin/mimir.exe ingest .` to verify `.mimir/digest.txt` and Cortex Yggdrasil Mind Map generation (`<1s`).
  - [ ] `6.2.c` Run `./bin/mimir.exe providers` to verify Sovereign, Well Pro, Odin Max, Meta Muse, and MiMo provider discovery.
