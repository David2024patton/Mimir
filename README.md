# Mímir

> **The agent that remembers.** A local-first, self-learning agentic coding framework
> with a neural knowledge brain.

Named for **Mímir** of Norse mythology - "the rememberer" / "the wise one" - who guards
**Mímisbrunnr**, the Well of Wisdom beneath Yggdrasil. The well grants knowledge to
those who drink from it; after Mímir was beheaded, Odin preserved his head so it would
keep giving counsel - wisdom that outlives its keeper. Mímir drinks from the well and
never forgets. So does this agent.

- **Brand name:** Mímir (ASCII: Mimir)
- **Domain:** mimirmind.com
- **Status:** Planning / BMAD method phase (concept complete; PRD next)

---

## What it is

Mímir is a sovereign, self-improving AI development partner that runs on your own
hardware, remembers everything it learns across every project, and works for you from
anywhere (desktop, terminal, phone, chat apps) - with no lock-in to any model vendor
or cloud.

## Why

Current AI coding tools **forget** (every session starts from zero), **lock you in**
(their models, their cloud), and are **closed** to deep customization. Mímir is the
answer: an open, modular, local-first agent with a durable self-organizing memory and
deterministic safety.

## Core pillars

1. **The Cortex** - a neural knowledge brain on SurrealDB (neurons / synapses /
   engrams): multi-modal knowledge + a self-learning memory engine (forgetting curve,
   outcome-driven learning, skill synthesis).
2. **Built-in deep system access** - terminal, filesystem, process control, in-memory
   code execution (Desktop-Commander-class, reimplemented natively).
3. **Governance & deterministic safety** - fail-closed policy gate, privilege rings,
   tamper-evident audit, agent SRE.
4. **Native sandboxing** - on-the-fly containers / Firecracker microVMs, snapshot/fork,
   per-user isolation.
5. **Goal-driven autonomy** - state the goal; it plans and executes until done.
6. **Research tools (Gjallarhorn)** - built-in metasearch (SearXNG-style), YouTube
   transcripts, web scraping, all ingested into the Cortex.
7. **Multi-surface + mobile** - web GUI, TUI, mobile app over a secure tunnel, and chat
   channels (Discord/Slack/Telegram).
8. **Modular plugin architecture** - typed plugin SDK, central registry, pluggable
   runtimes, multi-agent routing.

## Architecture

- **Core:** Go (single static cross-platform binary; home-server daemon)
- **Brain / storage:** SurrealDB (graph + vector + document)
- **GUI:** TypeScript + React/Solid web app served by the Go daemon; Wails desktop;
  Bubble Tea TUI; mobile (React Native/Flutter/PWA)
- **AI:** OpenAI-compatible HTTP to any provider (no SDK lock-in); Go MCP SDK
- **Sandbox:** containers + Firecracker/gVisor microVMs, on the fly

## Mythology -> architecture

| Norse myth | Mímir architecture |
|---|---|
| Mímir ("the rememberer") | The self-learning agent |
| Mímisbrunnr (Well of Wisdom) | The Cortex (SurrealDB knowledge brain) |
| Yggdrasil (World Tree) | The knowledge graph (neurons/synapses) |
| Gjallarhorn (the horn) | The research/ingestion tools |
| Mímir's preserved head | Persistent memory / engrams across sessions |
| Runes | Skills (reusable procedural knowledge) |

## Documentation (planning phase, in `docs/`)

- [`PROJECT-MASTER-PLAN.md`](docs/PROJECT-MASTER-PLAN.md) - the full plan: vision,
  competitive research, architecture, feature requirements (F1-F32), 49 epics,
  commercialization strategy, and the mythology/identity.
- [`RESEARCH-FEATURE-LANDSCAPE.md`](docs/RESEARCH-FEATURE-LANDSCAPE.md) - research across
  20+ tools (opencode, Claude Code, Cursor, Kiro, Antigravity, Devin, Agent Zero,
  Open Notebook, Vercel, Microsoft AGT, and more).
- [`mimirmind-product-brief.md`](docs/mimirmind-product-brief.md) - BMAD Product Brief (Mary).
- [`mimirmind-prd.md`](docs/mimirmind-prd.md) - BMAD PRD (John): user stories + EARS
  acceptance criteria for the Core epics E1-E14.
- [`mimirmind-architecture.md`](docs/mimirmind-architecture.md) - BMAD Architecture
  (Winston): ADRs, Go module layout, SurrealDB Cortex schema, agent-loop state machine,
  plugin-SDK contract, GUI<->daemon protocol, sandbox + security architecture.
- [`mimirmind-readiness-review.md`](docs/mimirmind-readiness-review.md) - Implementation
  Readiness Review (Winston): PASS with conditions.
- [`mimirmind-ux-spec.md`](docs/mimirmind-ux-spec.md) - UX spec (Sally): 5-region GUI
  layout, screens, TUI, mobile, design system.
- [`mimirmind-core-stories.md`](docs/mimirmind-core-stories.md) - Core epics E1-E14 broken
  into stories (John).
- [`mimirmind-sprint-1.md`](docs/mimirmind-sprint-1.md) - Sprint 1 plan (Amelia): the
  walking skeleton.

## Building

Requires [Go](https://go.dev) 1.26+.

```bash
# build the single binary
go build -o mimir ./cmd/mimir

# run the daemon + GUI (default http://localhost:8420)
./mimir

# other commands
./mimir version
./mimir auth        # provider key management (E1.3)
```

Configuration: copy [`.env.example`](.env.example) to `.env` and/or create a
`mimir.json` (opencode-compatible). Provider keys can use `{env:VAR}` / `{file:path}`.

## Project structure

```
cmd/mimir/          # entrypoint (daemon + CLI)
internal/
  config/           # mimir.json loader + {env}/{file} interpolation
  llm/              # provider abstraction + OpenAI & Anthropic dialects + registry
  tools/            # tool registry + to-do tool (+ shell/fs/code-exec next)
  agent/            # agent-loop state machine
  cortex/           # the brain: neuron/synapse/engram store (SurrealDB next)
  plugins/          # plugin SDK contract + registry
  policy/           # fail-closed policy gate
  server/           # HTTP + WebSocket daemon + GUI
gui/                # Solid web frontend (next)
site/               # marketing landing page (mimirmind.com)
docs/               # BMAD planning artifacts
```

## Development method

This project follows the [BMAD Method](https://github.com/bmad-code-org/BMAD-METHOD):
brainstorming -> product brief -> PRD -> architecture -> epics/stories -> sprint -> dev.

## License

TBD (planned open-source; open-core business model).
