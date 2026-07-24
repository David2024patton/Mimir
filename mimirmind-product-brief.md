# Mímir - Product Brief

**BMAD Artifact - Phase 2 (Mary, Business Analyst)**
Date: 2026-07-23 | Status: Draft for review | Feeds: PRD (John, PM)
Companion docs: `PROJECT-MASTER-PLAN.md` (full plan), `RESEARCH-FEATURE-LANDSCAPE.md` (research)

---

## 1. Product Identity

- **Name:** Mímir
- **Domain:** mimirmind.com (reserved; not purchased until the build is proven working)
- **Tagline candidates:** "The agent that remembers." - "Drink from the Well of Wisdom."
  - "Wisdom that outlives the session."
- **One-liner:** A local-first, self-learning agentic coding framework with a neural
  knowledge brain (the **Cortex**), built in Go, that remembers everything and keeps
  improving - named for **Mímir**, the Norse "rememberer" who guards the Well of Wisdom.

**The myth (our identity):** Mímir guards Mímisbrunnr, the Well of Wisdom beneath
Yggdrasil, granting knowledge to those who drink. Odin sacrificed an eye for a single
draught. After Mímir was beheaded, Odin preserved his head so it would keep giving
counsel - wisdom that outlives its keeper. Mímir drinks from the Well through the
Gjallarhorn, grows a Yggdrasil of connected knowledge, lays down engrams (memories),
and - like Mímir's head - keeps advising long after each session ends.

---

## 2. Vision

A sovereign, self-improving AI development partner that runs on your own hardware,
remembers everything it learns across every project, and works for you from anywhere
(your desk, your phone, your chat apps) - without locking you into any model vendor or
cloud. The agent that drinks from the well of knowledge and never forgets.

---

## 3. Problem Statement

Current AI coding tools fail the people who need them most (power users, tinkerers,
privacy-conscious developers, and teams building AGI-oriented systems):

1. **They forget.** Every session starts from zero. Corrections, preferences, and
   hard-won lessons are lost. There is no durable, self-organizing memory.
2. **Vendor lock-in.** Most tools route you through their models/gateway and their
   cloud. Your keys, your data, your money - on their terms.
3. **Closed & non-extensible.** You can't reshape the agent, add deep system access,
   or wire in your own knowledge sources without fighting the tool.
4. **Shallow knowledge.** They read the current file, not your whole research corpus
   (docs, repos, videos, web). No grounded, citable, multi-modal knowledge base.
5. **Tethered to the desk.** You can't steer the agent from your phone or a chat app
   while away.
6. **Unsafe autonomy.** As agents get more autonomous, prompt-level safety is a "polite
   request to a stochastic system." There's no deterministic governance.

The user's own prior attempt (`agence`) proved the self-learning idea works but broke
under fork debt - validating both the demand and the need for a clean foundation.

---

## 4. Target Audience / Personas

- **The Sovereign Builder (primary):** Technical, privacy-conscious, runs their own
  hardware (home server / homelab). Wants a capable agent they fully own and control,
  with BYOK and local-first data. Willing to self-host.
- **The AGI Tinkerer:** Building toward self-learning / AGI systems. Wants the memory,
  governance, and learning primitives exposed and hackable. (The user themselves.)
- **The Mobile Delegator:** Wants to kick off work and steer/approve it from their phone
  or a chat app while away from the desk.
- **The Team (later):** Small teams wanting shared knowledge, spend controls, SSO, and
  governance across members.

---

## 5. Value Proposition

Mímir gives you an agent that:
- **Remembers and learns** - a self-organizing neural memory (the Cortex) with a
  forgetting curve, outcome-driven learning from failures, and skill synthesis.
- **You own completely** - open-source Go core, BYOK, local-first, no vendor lock-in.
- **Knows your world** - multi-modal knowledge (repos via gitmcp, web, PDFs, video,
  audio) grounded with citations via RAG.
- **Acts safely** - deterministic governance (fail-closed policy gate), tiered
  sandboxing, tamper-evident audit.
- **Works anywhere** - desktop GUI, TUI, mobile app, and chat channels (Discord/Slack/
  Telegram), all against one home daemon.
- **Goes until done** - goal-driven autonomous mode: state the goal, it plans and
  executes until success criteria are met.

---

## 6. Key Feature Pillars

(Full detail in `PROJECT-MASTER-PLAN.md`, F1-F30.)

1. **The Cortex - Knowledge & Memory Brain** (F22, F26): SurrealDB-backed neural graph
   (neurons/synapses/engrams). Three memory tiers, five layers, forgetting curve,
   outcome-driven auto-capture, consolidation, concept graph, `reflect`/`quality_gate`
   skill synthesis, heartbeat scheduler.
2. **Built-in Deep System Access** (F3): Desktop-Commander-class native tools -
   terminal, filesystem, process control, in-memory code exec, edit_block.
3. **Governance & Deterministic Safety** (F7, F27): fail-closed policy gate, privilege
   rings, tamper-evident audit + Decision BOM, agent SRE, MCP security gateway.
4. **Native Sandbox Runtime** (F28): on-the-fly containers/microVMs (Firecracker),
   snapshot/fork, per-user isolation, port routing, egress policy.
5. **Goal-Driven Autonomous Mode** (F23): end goal -> game plan -> execute + verify
   until done.
6. **Interactive Preview & Annotation** (F24): live browser preview the user draws on
   and sends back to the LLM.
7. **Research & Ingestion (Gjallarhorn)** (F29): built-in SearXNG-style metasearch,
   YouTube transcripts, web scraping - all ingested into the Cortex.
8. **Multi-Surface + Mobile Remote** (F9, F25): web GUI (top bar / left nav / right
   chat / quick-launch rail), TUI, mobile app over a secure tunnel, chat channels.
9. **Modular Plugin Architecture** (F19, F20): typed plugin SDK, central registry,
   pluggable runtimes (harnesses), multi-agent routing, two-tier hooks.
10. **Spec-Driven Workflow + BMAD** (F6): requirements -> design -> tasks; integrates
    the BMAD method.

---

## 7. Differentiators (vs the field)

| vs | Mímir's edge |
|---|---|
| **opencode / Claude Code / Gemini CLI** | A genuine self-learning memory (Cortex) + deterministic governance + native sandboxing + mobile - not just a stateless loop. |
| **Cursor / Windsurf** | Local-first & open (no cloud tether), BYOK, fully extensible, multi-modal knowledge brain. |
| **Kiro** | Same spec-driven rigor, plus self-learning memory + sovereignty + mobile. |
| **Devin / Copilot agent** | Self-hosted (your hardware, your data), not a hosted black box; you keep the keys. |
| **NotebookLM / Open Notebook** | The knowledge brain is wired into an *acting* coding agent (not just Q&A), with a neural graph + forgetting curve. |
| **Agent Zero** | Cleaner modular foundation (Go + SurrealDB + plugin contract) instead of a heavy Docker/Python monolith. |

**The moat:** the combination of (a) self-learning memory, (b) deterministic governance,
(c) native sandboxing, and (d) sovereignty - in one open, modular framework. No current
tool has all four.

---

## 8. Business Model (open-core)

(Detail in master plan Section 11.) The CLI/core is free & open-source. Revenue:
1. **Hosted AI Gateway** (primary) - curated models, unified billing, pay-as-you-go +
   cheap subscription (the opencode Zen/Go model).
2. **Cloud Agent Execution** - optional hosted sandboxes for background/parallel agents.
3. **Hosted Knowledge / Cortex** - synced, indexed, private, team-shared knowledge brains.
4. **Pro / Team subscription** - workspaces, SSO/RBAC, spend limits, analytics.
5. **Mobile relay / tunnel** - hosted relay for off-machine access (local P2P stays free).
6. **Marketplace cut** - curated plugins/skills revenue share.
7. **Enterprise / self-hosted** - air-gapped, compliance, support.

No lock-in is the trust hook: it works with any provider/agent; the hosted layer is
convenience, not a cage.

---

## 9. Scope

**v1 (Core):** Modular Go foundation, agent loop + multi-provider, built-in system
tools, permission engine, the Cortex (knowledge + self-learning memory), skills,
plan/auto modes, subagents, hooks, the web GUI + TUI, MCP client, packaging.

**v1.x (Differentiators):** Pluggable runtimes, multi-agent routing, custom modes,
advanced orchestration, background tasks/automation, spec-driven workflow + BMAD,
verification/artifacts, model routing, sandbox tiers, multi-surface runtime, native
sandbox runtime, research tools, build engine, governance, goal mode, preview+annotation.

**Stretch / later:** Property-based testing + computer use, self-improving memory
synthesis, best-of-n/arena, messaging channels, mobile app, Tauri desktop + voice,
podcasts/audio overviews, hosted platform (gateway/console/cloud).

**Non-goals:** Not a hosted-only SaaS. Not a sandbox by default (guardrails + optional
isolation). Not an IDE/editor competitor (we orchestrate; editor integration via ACP).

---

## 10. Success Metrics

- **Activation:** user runs a real coding task end-to-end on day one (BYOK or local model).
- **Memory value:** agent correctly recalls/applies a learned preference or lesson in a
  later session (the "it remembers" moment).
- **Retention:** weekly active use; growth of the user's Cortex (neurons/engrams).
- **Autonomy success rate:** % of goal-driven runs that reach their success criteria.
- **Safety:** zero policy-violating tool calls reach execution (fail-closed holds).
- **Conversion (later):** free -> gateway/subscription conversion; hosted-Cortex adoption.

---

## 11. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| **Scope creep** (30 feature groups, 47 epics) | Tiered roadmap; ship the Core (Tier 1) first; hold Tier 2/3 until the core is solid. |
| **Go AI ecosystem is thinner than TS** | LLM access is just OpenAI-compatible HTTP; official Go MCP SDK exists; build the agent loop natively. |
| **Memory that rots / hallucinated recall** | Forgetting curve + consolidation + provenance + confidence; RAG grounded in sources with citations. |
| **Autonomous agent causes harm** | Deterministic fail-closed policy gate + sandboxing + audit + human checkpoints (governance is a first-class pillar, not an afterthought). |
| **Repeating `agence`'s fork-debt failure** | Clean Go core + SurrealDB + thin plugin contract; build ideas as small testable modules, never a monolithic fork. |
| **Sandboxing complexity (Firecracker)** | Tiered: containers (fast path) first, microVMs later; in-guest agent abstracts the backend. |
| **Sustaining open-source + business** | Open-core with a genuinely useful free tier; hosted layer is convenience, no lock-in. |

---

## 12. Technical Foundation (summary)

- **Core:** Go (single static cross-platform binary, concurrency, home-server daemon).
- **Brain/storage:** SurrealDB (graph + vector + document) - the Cortex.
- **GUI:** TypeScript + React/Solid web app served by the Go daemon; Wails desktop;
  Bubble Tea TUI; mobile (React Native/Flutter/PWA).
- **AI:** OpenAI-compatible HTTP to any provider (no SDK lock-in); Go MCP SDK.
- **Sandbox:** containers + Firecracker/gVisor microVMs, on the fly.
- **Architecture:** modular packages (agent-core, llm, tools, cortex, plugins/sdk,
  server, channels, tunnel) + typed plugin SDK + central registry.

---

## 13. Decisions Resolved in Brainstorming

- **D1 Fork vs fresh:** Fresh build in Go (not a fork). DECIDED.
- **D2 UI:** GUI-primary (web app), TUI as alternative. DECIDED.
- **D3 System access:** Built-in native tools (not MCP-dependent). DECIDED.
- **D4 Sandbox:** Tiered, on-the-fly (containers + microVMs). DECIDED.
- **D5 Name:** **Mímir** (mimirmind.com). DECIDED.
- **Knowledge subsystem:** the **Cortex** (neurons/synapses/engrams) on SurrealDB;
  Norse naming for subsystems (Well/Yggdrasil/Gjallarhorn/Mímir's Head). DECIDED.
- **Database:** SurrealDB (per the user's choice). DECIDED.
- **Self-learning:** recovered from `agence` (3-tier memory, 5 layers, forgetting curve,
  outcome capture, consolidation, skill synthesis). DECIDED.

---

## 14. Next BMAD Step

Hand off to **John (PM)** for `bmad-prd`: turn this brief into a full PRD with detailed
user stories, acceptance criteria (EARS), and per-persona journeys - starting with the
Core (Tier 1) epics E1-E14.
