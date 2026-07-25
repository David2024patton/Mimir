# Competitive Landscape - AI Coding Tools (July 2026)

**Analyst:** Mímir competitive-analysis workflow (F72)
**Date:** 2026-07-23
**Competitors analyzed:** OpenCode, Kiro, Devin, Cursor, Windsurf, CodeRabbit, Zencoder,
Mintlify, Paralect, Blitzy, Air (JetBrains), Merget, Qoder, Telerik
**Companion:** `opencode.md` (deep dive), `PROJECT-MASTER-PLAN.md`

---

## Executive Summary - The White Space

Every competitor optimizes for **speed of code generation** and/or **autonomy**. **None**
owns the combination Mímir is built on:

> **Durable, user-owned memory + a guided, gated full-SDLC partnership + local-first
> sovereignty (your models, no lock-in).**

| Mímir differentiator | The gap across the field |
|---|---|
| **Neural knowledge graph (remembers everything)** | Everyone is stateless or session-scoped. Blitzy has a codebase graph (enterprise, $500K+/yr). Merget records context (team/cloud). Nobody has a *personal, persistent, growing* memory. |
| **Guided, user-gated SDLC** | Kiro/Zencoder/Qoder have specs but rigid/delegation-oriented. Devin/Cursor are autonomy-first. Nobody *walks with you* step-by-step, gated. |
| **Local-first / your models / no lock-in** | Kiro (AWS), Devin (cloud), Cursor (cloud + SpaceX), Qoder (Alibaba) are all locked. Mímir is sovereign. |
| **Computer use + futuristic UI** | Devin has desktop QA (cloud). Nobody combines coding agent + computer use in a distinctive, alive UI. |

**The two biggest competitive threats:** Cursor (DX leader, $4B ARR, but pricing-trust
deficit + SpaceX lock-in) and Qoder (price + distribution + spec-driven, but
Alibaba/privacy concerns). **Merget** validates the memory thesis from the team/cloud
angle - a potential ally or cautionary tale.

---

## Per-Competitor Summaries

### OpenCode (opencode.ai) - see `opencode.md`
Open-source AI coding agent (160K+ stars, 7.5M monthly devs). Free CLI + Zen
(pay-as-you-go, zero markup) + Go ($10/mo). **Gap: stateless, no memory, no guided
SDLC, cloud for the good stuff.**

### Kiro (kiro.dev) - AWS
Spec-driven IDE (successor to Amazon Q Developer, retires Apr 2027). Credit-based
(Free 50 / Pro $20 / Pro+ $40 / Pro Max $100 / Power $200). Generates
requirements/design/tasks before code; property-based testing; agent hooks.
**Gap: AWS/Bedrock lock-in, no BYOK, credit opacity, rigid specs, no memory.**
News: GA May 2026; replaces Amazon Q; Pro Max + iOS app Jun 2026.

### Devin (devin.ai) - Cognition
First-mover autonomous SWE agent. $1B+ Series D at **$26B** (May 2026); ~$492M
run-rate. Free / Pro $20 / Max $200 / Teams $80+ / Enterprise. Full autonomy,
dynamic re-planning, desktop QA via computer use, fleets of parallel Devins. Acquired
Windsurf (Jul 2025). **Gap: cloud-only, no BYOK, opaque metering, weak on ambiguous
work, no persistent memory.** 89% of Cognition's own code now committed by Devin.

### Cursor (cursor.com) - Anysphere
Dominant AI-native editor. **SpaceX/xAI $60B all-stock acquisition** announced Jun 2026;
$4B ARR. Hobby free / Pro $20 / Pro+ $60 / Ultra $200 / Teams. Best-in-class Tab
autocomplete, Composer 2.5, background/cloud agents, Bugbot. **Gap: no SDLC methodology,
no memory, metered-pricing trust deficit (Jul 2025 backlash), cloud + SpaceX lock-in.**

### Windsurf (windsurf.dev) - Cognition (acquired Dec 2025, ~$250M)
AI IDE, now Devin's front-end. Free / Pro $20 / Max $200 / Teams $40. Cascade agent,
SWE-1.6 (950 tok/s), Devin Cloud, Codemaps, Agent Command Center. **Gap: cloud-dependent,
session memory only, no guided SDLC, 3 owners in 18 months, proprietary model.**

### CodeRabbit (coderabbit.ai)
AI code review leader. $88M raised, **$550M** valuation, ~$40M ARR (700% YoY). Free /
Pro $24 / Pro Plus $48 / Enterprise. CodeGraph, 40+ linters, learnings, Slack/Discord
agent. **Gap: reactive (post-PR), not a builder, no memory graph, cloud, per-seat cost.**

### Zencoder (zencoder.ai)
Multi-model orchestration (Andrew Filev, ex-Wrike $2.25B exit). Free / Starter $19 /
Core $49 / Advanced $119 / Max $250. Multi-model cross-validation, spec-driven with
checkpoints, BYOK, Zenflow desktop app. **Gap: cloud SaaS, no memory, small/unfunded,
no computer use.** Closest to Mímir's gated-SDLC concept.

### Mintlify (mintlify.com)
Docs/knowledge platform pivoting to "AI knowledge infrastructure." $45M Series B at
**$500M** (Apr 2026, a16z). Free / Pro $250/mo / Enterprise. Docs-as-code, AI assistant,
auto MCP server, Agent Score. **Gap: cloud-only, $0→$250 cliff, documents not guides,
no memory/SDLC.**

### Paralect (paralect.com)
Product dev company / venture studio (not a product). Bootstrapped, services from
$2,500/bi-weekly. Full-SDLC services, "no lock-in" messaging, Ship boilerplate. **Gap:
it's a services business (time-for-money), not a scalable product; no memory product.**
Validates demand for a no-lock-in full-SDLC partner.

### Blitzy (blitzy.com)
Autonomous enterprise dev platform. $200M at **$1.4B** (May 2026, Northzone). Enterprise
only ($50K entry, $500K-$50M/yr). Reverse-engineers codebases into a knowledge graph,
orchestrates 3,000+ agents. 66.5% SWE-Bench Pro. **Gap: enterprise-only pricing,
batch/black-box non-cancellable runs, their infra, no personal memory.** Validates the
knowledge-graph thesis at enterprise scale.

### Air (air.dev) - JetBrains
Agentic Development Environment (ADE). Free preview (requires JetBrains AI sub or BYOK).
Parallel multi-agent (Claude/Codex/Gemini/Junie + ACP agents), isolated execution
(Docker/worktrees/cloud), co-developing ACP protocol. **Gap: no memory, no guided SDLC,
JetBrains-tied, utilitarian UI.** Local-friendly (partial overlap with Mímir).

### Merget (merget.ai)
"System of record for AI-native dev." $1M seed (IDrive), v0 Jul 2026. Auto-captures
prompts/diffs/goals across tools, D.O.M (Deep Organizational Memory), Merget Map/Stats.
**Gap: team-observability not an individual agent, cloud/GitHub-centric, very early.**
Closest conceptual cousin to Mímir's memory thesis - validates it from the team angle.

### Qoder (qoder.com) - Alibaba-backed
Agentic coding platform, 5M+ users. Free / Pro $20 / Pro+ $60 / Ultra $200 / Teams $40.
Quest Mode (spec-driven), RepoWiki, Knowledge Engine, QoderWork (local-first), Qwen3-Coder.
**Gap: Alibaba/China privacy concerns, cloud/credit-metered, per-repo knowledge, delegation
not guided.** Big threat on price + distribution.

### Telerik (telerik.com) - Progress Software
.NET/JS UI component library (1,250+ components), not an agent. $649-$1,649/dev/yr.
Now has AI Coding Assistant + Agentic UI Generator via MCP. **Not a direct rival - an
integration candidate (MCP).** Highlights Mímir's local-first/model-agnostic identity.

---

## Pricing Landscape (quick reference)

| Tool | Free | Entry paid | Top tier | Model |
|---|---|---|---|---|
| OpenCode | yes (CLI) | Go $10/mo | Zen pay-as-you-go | free CLI + paid gateway |
| Kiro | 50 credits | Pro $20 | Power $200 | credits |
| Devin | light quota | Pro $20 | Max $200 / Teams $80 | credits |
| Cursor | Hobby | Pro $20 | Ultra $200 | credits |
| Windsurf | quota | Pro $20 | Max $200 | quota |
| CodeRabbit | OSS free | Pro $24 | Pro Plus $48 | per-seat |
| Zencoder | 30 calls/day | Starter $19 | Max $250 | daily calls |
| Mintlify | 1 editor | Pro $250/mo | Enterprise | seats + AI credits |
| Blitzy | reverse-eng free | PoC $50K | $50M/yr | enterprise |
| Air | preview free | JetBrains AI sub | TBD | sub/BYOK |
| Qoder | basic + BYOK | Pro $20 | Ultra $200 | credits |
| Telerik | n/a | $649/dev/yr | $1,649/dev/yr | per-seat |

**Mímir's pricing wedge:** free self-hosted (unlimited) + Cloud Sync $9 + Cloud Pro $15
(2x NotebookLM, $5 less) + Ultra $39. Undercuts the $20 entry standard and the cloud
lock-in.

---

## Marketing / Positioning Takeaways

1. **Own the wedge:** "the agent that remembers + walks the SDLC with you + you own it."
   Nobody occupies this intersection.
2. **Attack the lock-in:** Kiro (AWS), Devin/Cursor (cloud + SpaceX), Qoder (Alibaba)
   are all lock-in stories. "Your models, your machine, cancel anytime" is timely.
3. **Attack the pricing opacity:** Kiro credits, Cursor's Jul 2025 backlash, Devin's
   ACU confusion. "Transparent, predictable, free self-hosted" wins trust.
4. **Memory is the moat:** every competitor is stateless or session-scoped. Lead with
   "it remembers."
5. **Guided > autonomous for the messy middle:** Devin/Cursor own autonomy; Mímir owns
   the guided, gated partnership for ambiguous, real-world work.
6. **Watch Merget:** it validates the memory thesis from the team angle - potential ally
   or acquisition target.
7. **Telerik is an integration, not a rival** (MCP).

---

## TODO (next passes)
- Deep-dive each competitor's authenticated dashboard/layout (log in).
- Screenshot each site for the design reference.
- Track news monthly (funding, launches, pricing) - F72.7 periodic refresh.
- Deep-dive Factory, GitHub Copilot coding agent, OpenAI Codex, Cline, Roo (not yet covered).
