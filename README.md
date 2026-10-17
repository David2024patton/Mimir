# Mímir

> **The agent that remembers.** A local-first, self-learning polyglot agentic coding framework with a neural knowledge brain, built-in **Liquid AI `d1-omni-600M`** System-One decision routing, **GitIngest + Yggdrasil Mind Map** memory, and a **4-Stack Architecture (Go + Rust + Mojo + Flutter)**.

Named for **Mímir** of Norse mythology - "the rememberer" / "the wise one" - who guards **Mímisbrunnr**, the Well of Wisdom beneath Yggdrasil. The well grants knowledge to those who drink from it; after Mímir was beheaded, Odin preserved his head so it would keep giving counsel - wisdom that outlives its keeper. Mímir drinks from the well and never forgets. So does this agent.

- **Brand name:** Mímir (ASCII: Mimir)
- **Domain:** mimirmind.com
- **Current Version:** `0.2.0-dev` (Go Core + Charmbracelet TUI + Liquid AI `d1-omni-600M` Decision Engine + GitIngest Mind Map + Book-Spine Web UI live)

---

## 1. The 4-Stack Polyglot System & Why Each Language Is Used

Mímir does not force a single language to do jobs it wasn't designed for. Instead, it uses a **4-Stack Polyglot Architecture** where each language owns its strongest hardware/UI lane:

| Stack Layer | Language | Module / Crate | Why This Language Owns This Lane |
| :--- | :--- | :--- | :--- |
| **1. The Brain & Harness Daemon** | **Go (`golang`)** | `cmd/mimir` & `internal/*` | **Concurrency + Single Binary + Charmbracelet TUI.** Juggles SSE streams, MCP servers, mid-turn steering channels, local named pipes, and the interactive **Charmbracelet** terminal UI (`Bubble Tea`, `Lipgloss`, `Huh`, `Glamour`, `Wish` SSH) in a ~15 MB static binary. |
| **2. The Iron Substrate** | **Rust (`rust`)** | `native/rust` (`mimir-substrate`) | **Zero-GC Systems Speed + Tree-Sitter + E2EE Crypto.** Powers sub-millisecond `fff` git-aware fuzzy file/ripgrep search, `tree-sitter` AST symbol editing (`ast_replace_symbol`), `<5ms` copy-on-write workspace snapshots/rollbacks (`space`), and `X25519 + XChaCha20-Poly1305` E2EE encryption shared with Flutter via `flutter_rust_bridge`. |
| **3. The Metal Cortex & System-One Engine** | **Mojo (`mojo`)** | `native/mojo` (`mimir-cortex-metal`) | **Direct Silicon Execution (AVX-512 / NEON / GPU Warps).** Talks directly to CPU SIMD lanes and GPU registers for **10,000-bit Hyperdimensional VSA memory**, `<0.2ms` SIMD cosine similarity over 100k+ Cortex neurons, zero-copy `.itf` tensor loading, and local hardware execution for the **Liquid AI `d1-omni-600M`** decision model. |
| **4. The Glass (Desktop & Phone App)** | **Flutter (`dart`)** | `app/mimir_flutter` | **One 120fps Native Codebase for Desktop + iOS/Android.** Replaces bloated Electron + React Native with a single Skia/Impeller app: a `happy-desktop` multi-column workbench on Windows/macOS/Linux (`Ctrl+K` palette, `Ctrl+J` terminal drawer, Book-Spine session rail) and an E2EE paired **Phone App** (`mimir pair` QR code, 1-tap diff/command approvals, and **"Call Odin" live voice mode**). |

---

## 2. Core Innovations Built Into Mímir

### A. Built-in "System One" Decision Model: Liquid AI `d1-omni-600M` (`Huginn-Decider`)
Before Mímir spends a single token on a large generative LLM ("System Two"), its built-in **Liquid AI `d1-omni-600M`** (~400 MB multimodal text/audio/image LFM, auto-bootstrapped via `mimir setup`) runs in `<10ms` for `$0.00` using three primitives (`Choice`, `Score`, `Noul`):
1. **Tier & Norse Persona Routing (`RoutePrompt`)**: Automatically picks `Odin` (Architect, `T1`), `Thor` (Executor, `T3`), `Loki` (Debugger/Red-Team, `T1`), or `Heimdall` (Security Reviewer, `T3`) and decides whether extended Chain-of-Thought is needed.
2. **Tool Schema Pruning (`PruneToolNames`)**: Prunes 25+ registered tools down to the top 5 needed for the current turn, saving ~70% of prompt schema tokens.
3. **Yggdrasil Mind Map File & Test Selection (`SelectMindMapFiles` & `SelectAffectedTests`)**: Queries the ingested repository Mind Map to pick the exact files to read and the exact subset of tests affected by a diff (inspired by the `Superlinear` / `Jev` verification optimization).
4. **Deterministic + Probabilistic Safety Gate (`AssessCommandSafety`)**: Scores shell commands (`0.0` to `1.0`) so safe commands (`go test ./...`) auto-execute while risky mutations push a 1-tap approval card to the TUI or Flutter Phone App.

### B. Built-in GitIngest + Yggdrasil Mind Map in Cortex (`Gjallarhorn-Ingest`)
Inspired by `coderamp-labs/gitingest`:
- Run `mimir ingest .` or `mimir ingest https://github.com/owner/repo` (or `/ingest` in the TUI, or let the agent call the `ingest_repo` tool).
- Mímir scans the repository, writes a clean `gitingest.com`-compatible `.mimir/digest.txt`, extracts every file's exported symbols (`func`, `type`, `class`, `def`, `fn`, `struct`), imports, and test relationships into a **Yggdrasil Mind Map**, and stores the Mind Map + module neurons + graph synapses directly in **The Cortex** (`cortex.json` or SurrealDB).
- Both the main coding LLM and **Liquid AI `d1-omni-600M`** (`/mindmap <query>`) can query the Yggdrasil Mind Map across sessions without re-reading the whole repository.

### C. Auto-Activated `⚡SmolMode` for Local & Compact Models
Inspired by `leonvanzyl/smolcoder`:
- Whenever Mímir detects a local model `<=32B` (`qwen3:8b`, `muse-glimmer`, `liquid/d1-omni-600m`) or a fast cloud tier (`flash`, `mini`, `nano`, `haiku`), it automatically activates **`⚡SmolMode`**:
  - Switches to a lean 2-paragraph system prompt with concrete JSON tool examples.
  - Tightens `PruneHistoryPlaceholders` to keep only the last 2 turns of raw file writes while preserving the harness-owned Todo checklist at the tail of the prompt.
  - Uses full thinking on Step 0 (planning) and automatically injects `/no_think` (`Effort: "off"`) on mechanical follow-up tool steps (`step > 0`) for 3-5x faster local execution.

### D. OpenCode-Style Provider Registry + BYOK & Mímir Well Subscriptions
Inspired by `anomalyco/opencode` and `truefoundry/trueforge`, Mímir supports both **100% Free Sovereign BYOK/Local** and **Managed Subscription Tiers** (`mimir providers` or `/providers`):
- **Mímir Sovereign (`$0 / mo`)**: Unlimited local Ollama (`:11434`), LM Studio (`:1234`), bundled **Liquid AI `d1-omni-600M`**, **Meta `Muse Glimmer`**, and unlimited BYOK across 12+ cloud providers.
- **Mímir Well Pro (`$20 / mo`)**: Single-key managed gateway (`MIMIR_SUBSCRIPTION_KEY`) with instant access to **Meta `Muse Wonder 1.0`**, **Meta `Muse Spark 1.0`**, **Xiaomi `MiMo-v2-Pro`**, **Claude Sonnet 4.6**, **GPT-5**, and **DeepSeek V3.2**, plus hosted E2EE Mobile Relay for the Flutter Phone App.
- **Mímir Odin Max (`$100 / mo`)**: 5x frontier token budget (`muse-wonder-full-1.0` / `claude-opus-4-6`), parallel Norse sub-agent swarms, and `Crabbox`-style remote build/test offloading.

---

## 3. Mythology -> Architecture Mapping

| Norse Myth | Mímir Component | Implementation |
| :--- | :--- | :--- |
| **Mímir** ("the rememberer") | The Sovereign Coding Agent | `internal/agent` + 4-Stack Polyglot Loop |
| **Mímisbrunnr** (Well of Wisdom) | The Cortex Knowledge Brain | `internal/cortex` (File + Mojo SIMD HDC + SurrealDB) |
| **Yggdrasil** (World Tree) | Repository Mind Map Graph | `internal/ingest` (`MindMap` nodes, symbols, test edges) |
| **Huginn** ("thought") | System-One Decision Model | `internal/decision` (**Liquid AI `d1-omni-600M`**) |
| **Muninn** ("memory") | Compaction & Recall Engine | `internal/agent/compaction.go` + Mojo VSA associative recall |
| **Gjallarhorn** (the horn) | GitIngest & Research Tools | `internal/ingest` + `internal/tools/ingest_tool.go` |
| **Runes** | Progressive Skills Library | `internal/tools/skills.go` (`H:\AI\skills` + `.mimir/skills`) |

---

## 4. Quick Start & CLI Reference

Requires [Go](https://go.dev) 1.26+.

```powershell
# Build the single static binary
go build -o bin/mimir.exe ./cmd/mimir

# Bootstrap the local Liquid AI d1-omni-600M System-One Decision Model
.\bin\mimir.exe setup

# Ingest any local repo or GitHub URL into digest.txt + Yggdrasil Mind Map in Cortex
.\bin\mimir.exe ingest .
.\bin\mimir.exe ingest https://github.com/David2024patton/Mimir

# List built-in BYOK providers (Meta Muse, MiMo, Claude, OpenAI, Local) & Subscription tiers
.\bin\mimir.exe providers

# View or switch the 7-Phase SDLC board & execution methodology (agile | waterfall | devops)
.\bin\mimir.exe sdlc

# Launch the interactive Charmbracelet TUI (auto-discovers Ollama, LM Studio, or Cloud keys)
.\bin\mimir.exe

# Launch interactive TUI + live Book-Spine Web UI simultaneously
.\bin\mimir.exe --web
```

### Interactive TUI Slash Commands
- `/sdlc [1..7|next|signoff|agile|waterfall|devops]` - View or control the **7-Phase SDLC State Machine** (`Planning -> Analysis -> Design -> Implementation -> Testing -> Deployment -> Maintenance`) so only the active phase's compact micro-prompt & tools load
- `/model [name]` - Auto-discover local Ollama/LM Studio models or hot-swap active model
- `/providers` - View built-in BYOK providers and Mímir Well Subscription tiers
- `/ingest [path|git-url]` - GitIngest a repository into `.mimir/digest.txt` and store its **Yggdrasil Mind Map** in Cortex
- `/mindmap [query]` - Inspect the repository's Yggdrasil Mind Map or ask **Liquid AI `d1-omni-600M`** which files and tests match a task
- `/decide <prompt>` - Run a live **Liquid AI `d1-omni-600M`** System-One pass (Persona, Tier, Pruned Tools, Mind Map files, Safety score)
- `/smol [auto|on|off]` - Toggle `⚡SmolMode` optimizations for compact/local models
- `/mode [ro|edit|bypass]` - Cycle permission modes (`ro` hides write tools from the model schema)
- `/effort [off|low|med|high]` - Toggle reasoning budget (`off` sends `/no_think` for fast tool loops)
- `/persona [name]` - Switch Norse persona (`odin`, `thor`, `loki`, `heimdall`, `bragi`, `huginn`, `muninn`, `forseti`)
- `/todo` - Display the harness-owned Todo checklist
- `/memory` - Inspect neurons stored in The Cortex
- `/compact` - Prune old file-write payloads from conversation history

---

## 5. Master Implementation Task Checklist (4-Stack Roadmap)

> **Full Granular Step-by-Step Blueprint**: See [`docs/MASTER-TODO.md`](docs/MASTER-TODO.md) for the complete multi-level checklist (`1.1.a` through `6.2.c`) with exact file paths, C-ABI headers, and verification gates.

### Phase 1: Go Core, 7-Phase SDLC Framework, Charmbracelet TUI, SmolMode & Decision Engine (`COMPLETE`)
- [x] Zero-config local discovery (`Ollama :11434`, `LM Studio :1234`) + OpenCode-style BYOK & Subscription catalog (`internal/config/discovery.go`, `internal/config/catalog.go`)
- [x] Full **Meta Model API (`dev.meta.ai`)** support (`muse-wonder-1.0`, `muse-spark-1.0`, `muse-spark-open`, `muse-glimmer`, `muse-voice-transcribe`, `muse-image`, `sam-3.1`)
- [x] Built-in **Liquid AI `d1-omni-600M`** System-One Decision Engine (`Choice`, `Score`, `Noul`, `RoutePrompt`, `PruneToolNames`, `SelectMindMapFiles`, `SelectAffectedTests`, `AssessCommandSafety`) in `internal/decision/decider.go`
- [x] Built-in **GitIngest + Yggdrasil Mind Map** engine (`internal/ingest/ingest.go`) + `ingest_repo` tool (`internal/tools/ingest_tool.go`) persisting directly into Cortex memory
- [x] **Modular 7-Phase SDLC Engine (`internal/sdlc/phases.go`, `internal/sdlc/sdlc.go`, `internal/tools/sdlc_tool.go`)** across **Agile**, **Waterfall**, and **DevOps** methodologies -- loads ONLY the active phase's compact ~45-75 token micro-prompt and phase-scoped Web/SEO/GEO overlay (`internal/agent/rules.go`) to keep the main prompt bloat-free for both big and tiny LLMs
- [x] Auto-activated **`⚡SmolMode`** (`internal/agent/smolmode.go`) + 3-stage context compaction (`internal/agent/compaction.go`) + `charmbracelet/fantasy` malformed tool JSON repair (`internal/llm/jsonrepair.go`)
- [x] Hardened tools (`edit_block` CRLF/fuzzy fallback + `closestBlockHint`, 32 KB `shell` spill files, interactive `QuestionTool`, progressive `SKILL.md` loader)
- [x] Interactive terminal UI (`internal/tui/tui.go`) with `/sdlc`, `/model`, `/providers`, `/ingest`, `/mindmap`, `/decide`, `/smol`, `/mode`, `/effort`, `/persona`, `/todo`, `/memory`, `/compact`, `/web`

### Phase 2: Level 1-3 Harness Lifecycle Hooks, Web/SEO Auditor & Software Factory (`NEXT`)
- [ ] `2.1` Implement `HookPrePrompt`, `HookPreTool`, `HookPostTool` (auto-linter), and `HookOnStop` (`internal/agent/hooks.go`)
- [ ] `2.2` Implement `.mimir/HANDOFF.md` generator (`internal/agent/handoff.go`) + `/handoff` command for zero-loss context handoffs at 80% window fill
- [ ] `2.3` Implement `web_seo_audit` tool (`internal/tools/web_audit_tool.go`) + `mimir audit-web` CLI to score JSON-LD Schema.org, `/llms.txt`, semantic landmarks, and zero AI slop
- [ ] `2.4` Implement Level 3 Multi-Agent Software Factory Queue (`internal/swarm/factory.go`) where `Odin` dispatches isolated git-worktree tasks to parallel `Thor` workers with `Loki` + `Heimdall` sign-off

### Phase 3: Rust Iron Substrate (`crates/mimir-rs` -- Stack 2) (`PLANNED`)
- [ ] `3.1` Scaffold `crates/mimir-rs` (`cdylib` + `staticlib`) with C-ABI exports (`include/mimir_rs.h`) for Go (`internal/native/tyr_bridge.go`) and `flutter_rust_bridge` for Flutter
- [ ] `3.2` Implement `ast.rs` (`tree-sitter` `mimir_ast_outline` and `mimir_ast_patch_symbol` so small models edit whole functions without line-number drift)
- [ ] `3.3` Implement `search.rs` (Tantivy sub-millisecond BM25 lexical code & Cortex search + `fff` git-aware fuzzy file finder)
- [ ] `3.4` Implement `sandbox.rs` (`Ratatoskr-FS` transactional git worktree snapshots & `<5ms` rollback + OS process sandbox)

### Phase 4: Mojo Metal Cortex (`mojo/mimir_mojo` -- Stack 3) (`PLANNED`)
- [ ] `4.1` Scaffold `mojo/mimir_mojo` and C-ABI header `include/mimir_mojo.h`
- [ ] `4.2` Implement `itf_loader.mojo` (zero-copy `mmap` loader for 64-byte aligned `.itf` iTaK Tensor Format files)
- [ ] `4.3` Implement `vsa_cortex.mojo` (10,000-bit Hyperdimensional VSA associative binding + AVX-512 `vpopcntdq` Hamming & SIMD cosine similarity)
- [ ] `4.4` Implement `decider_kernel.mojo` (native bare-metal runner for **Liquid AI `d1-omni-600M`** and **Meta `Muse Glimmer`** draft decoding)

### Phase 5: Flutter Desktop Workbench & E2EE Phone App (`ui/mimir_flutter` -- Stack 4) (`PLANNED`)
- [ ] `5.1` Implement `internal/relay/happy_e2ee.go` + `mimir pair` terminal QR code pairing (`X25519` + `XChaCha20-Poly1305` E2EE envelopes)
- [ ] `5.2` Scaffold `ui/mimir_flutter` targeting Windows, macOS, Linux, iOS, and Android with Nordic Glassmorphism theme
- [ ] `5.3` Build **Desktop Workbench** (4-pane layout: Yggdrasil Mind Map Explorer, Live Agent Stream & Diff Editor, Level 3 Software Factory Swarm Board, and 4-Layer Web/SEO/GEO & Cost Dashboard)
- [ ] `5.4` Build **Mobile Companion App** (`mimir pair` QR scanner, 1-tap `Approve / Reject` diff cards, and mid-turn voice/text steering powered by `muse-voice-transcribe` + `d1-omni-600M`)

