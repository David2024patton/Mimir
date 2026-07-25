# Competitive Landscape - Batch 2 (July 2026)

**Analyst:** Mímir competitive-analysis workflow (F72)
**Date:** 2026-07-23
**Companion:** `COMPETITIVE-LANDSCAPE.md` (batch 1), `opencode.md` (deep dive + dashboard)
**Competitors:** GitHub Copilot, OpenAI Codex, Cline, Roo Code, Amazon Q, Tabnine,
Sourcegraph Cody/Amp, Replit Agent, Augment, Trae, Factory, Poolside, Aider, Bolt.new,
Lovable, Continue

---

## Per-Competitor Summaries

### GitHub Copilot (github.com/features/copilot) - Microsoft/GitHub
The market incumbent; deepest distribution (bundled into the world's largest code host).
Moving to usage-based **AI Credits** (1 credit = $0.01) June 1, 2026. Free / Student $0 /
Pro $10 / Pro+ $39 / Max $100 / Business $19 / Enterprise $39. Completions + next-edit
stay unlimited; chat/agent/review burn credits. Agent mode + cloud coding agent + code
review + Spaces + Spark. **Gap: cloud-locked (GitHub/MS infra), no memory, no guided
SDLC, credits expire monthly, billing backlash.** News: Max tier $100; paused self-serve
Business signups Apr 2026; Opus 5 / GPT-5.6 / Gemini 3.x added.

### OpenAI Codex (openai.com/codex)
OpenAI's coding agent; unified cloud agent + open CLI + IDE ext + desktop + GitHub/Slack;
3M+ weekly devs (Apr 2026). Included with ChatGPT Plus $20 / Pro $200 / Business /
Enterprise; codex-only pay-as-you-go seats. Cloud sandboxed parallel tasks, verifiable
citations, **computer use** (Apr 2026), in-app browser, image gen, **memory preview**
(Apr 2026), automations, proactive suggestions, 90+ plugins. **Gap: OpenAI-locked,
memory is a shallow preference-recall preview (not a neural graph), pushes full autonomy
(opposite of gated SDLC), code in OpenAI containers.** News: desktop app Feb 2026;
"Codex for almost everything" Apr 2026; ChatGPT Business cut $25→$20 Jun 2026.

### Cline (cline.bot)
Leading open-source agent (Apache 2.0); 5M+ installs, 65k stars; $32M raised (Emergence).
Free BYOK / ClinePass $9.99 / Teams $20 / Enterprise. Plan/Act loop, every step gated
behind human approval, model/IDE-agnostic, MCP, checkpoints, SDK. **Gap: no memory graph
(lossy condensation), no guided full-SDLC, no computer use, utilitarian UX.** Closest
direct competitor on local-first/no-lock-in - Mímir beats it on memory + SDLC + UI.
News: $32M Jul 2025; Cline Enterprise Oct 2025; JetBrains/CLI/SDK 2026; absorbing Roo
refugees.

### Roo Code (roocode.com) - SUNSET
Prominent Cline fork (custom modes, Boomerang orchestration); 3M installs. **SHUT DOWN
May 15, 2026** (extension/Cloud/Router terminated; repo archived → community team);
pivoting to Roomote (cloud Slack agent, ~$20/mo + $5/agent-hr). **Gap: dead product;
no memory; no guided SDLC.** Validates local-first IDE agents AND creates a 3M-install
refugee pool. News: sunset Apr 21 2026; users steered to Cline/Roomote.

### Amazon Q Developer (aws.amazon.com) - being replaced by Kiro
AWS AI dev assistant; **sunset for Kiro** (end-of-support Apr 30, 2027; new signups
blocked May 15, 2026). Free / Pro $19. Agentic coding, completions, vuln scan,
Java/.NET transform. **Gap: hard AWS lock-in, no memory, forced Q→Kiro migration =
churn window to poach.** News: Kiro CLI Nov 2025; Opus 4.6 dropped from Q Pro May 2026.

### Tabnine (tabnine.com)
Original AI code assistant (2018); **pivoted enterprise-only 2025, killed all free/
individual tiers** (Apr 2, 2025). $57M raised. Code Assistant $39 / Agentic $59 /
Context Engine $59 per user/mo; BYO-LLM + 5% handling fee. Context Engine (org knowledge
graph via MCP), air-gapped/on-prem, code provenance. **Gap: enterprise-only floor prices
out individuals; Context Engine is retrieval not a learned personal memory; no guided
SDLC; no computer use.** News: free tier killed Apr 2025; CLI Jan 2026; Context Engine
Feb/Mar 2026. **Tabnine abandoning the prosumer/indie segment = Mímir's opening.**

### Sourcegraph Cody / Amp
Two products, now two companies. Cody = context-aware in-IDE chat on Sourcegraph search
(**Free/Pro killed Jul 23, 2025**; only Enterprise ~$59/user or $16K/yr platform). Amp =
agentic frontier tool, **spun out Dec 2025** (already profitable); Free $10/day +
pay-as-you-go + subscriptions ~$20/$200 (Jul 2026); Orbs = remote always-on machines.
**Gap: brand fragmentation; no persistent personal memory (per-query RAG); Amp
cloud/orb-dependent; no guided SDLC.** News: Cody free/pro killed Jul 2025; Amp
subscriptions Jul 2026; Orbs event-reactive.

### Replit Agent (replit.com)
"Vibe-coding" leader; natural-language app builder for non-programmers. $400M Series D
at **$9B** (Mar 2026); $150M ARR; 40M+ users. Free / Core $20 / Pro $100. Plain-language
→ full-stack (web/mobile/DB/auth/deploy), self-testing reflection loop, parallel agents.
**Gap: total cloud lock-in, not for real repos, no memory, autonomy-first (opposite of
gated), effort-pricing runs hot.** News: effort-based pricing Jun 2025; Pro $100 Feb 2026;
Agent 4 Mar 2026.

### Augment Code (augmentcode.com)
Enterprise AI coding for large codebases; $252M raised ($977M val); $20M ARR. Restructured
Jun 2026 to **Business $100/mo flat** (≤50 seats) / Enterprise. Context Engine (400K+
files), **Cosmos** (agentic SDLC OS: shared memory, Expert Registry, self-improving
agents), Intent (Coordinator-Implementor-Verifier), Prism routing. **Gap: cloud-dependent,
no plan under $100, Expert Registry is org not personal memory, no computer use, credit
unpredictability.** News: Cosmos GA Jun 2026; Intent Feb 2026.

### Trae (trae.ai) - ByteDance
AI-native IDE (VS Code fork); 12M users. Free / Lite $3 / Pro $10 / Pro+ $30 / Ultra $100
(token-based since Feb 2026). SOLO multi-agent (Orchestrator/Arch/Dev/QA/DevOps), Builder
2.0 (NL→MVP ~10 min), voice debugging, Figma-to-code. **Gap: ByteDance telemetry/privacy
concerns (telemetry persists in Privacy Mode; 5-yr retention), cloud-dependent, no memory,
SOLO fully autonomous (opposite of gated), weak on huge repos.** News: SOLO→TRAE Work Jun
2026; 12M users; Doubao 2.1 Pro.

### Factory (factory.ai)
Agent-native platform with autonomous "Droids"; $220M raised ($1.5B val, Apr 2026). Pro $20
/ Plus $100 / Max $200 / Business / Enterprise. Droids (Code/Reliability/Tutorial),
Missions (multi-week multi-agent), **Factory Desktop** (Apr 2026, full system access),
Droid Computers (cloud sandboxes), model-agnostic, #1 Terminal-Bench. **Gap: no free tier,
no personal memory, autonomy-first not human-gated, cloud-dependent.** News: $150M Series C
Apr 2026; Factory Desktop Apr 2026; #1 Terminal-Bench.

### Poolside (poolside.ai)
Foundation-model lab for agentic coding (ex-GitHub CTO); $626M raised ($12B val). Open-
weight models (Apache 2.0): Laguna M.1 (225B/23B), S 2.1, **XS.2 (33B/3B, runs on a Mac)**.
pool (terminal agent) + Shimmer (cloud dev) research previews; on-prem/air-gapped;
fine-tune on your code. **Gap: model company not product company; no UI/memory/SDLC/
computer use; enterprise/gov focus.** **Opportunity: XS.2 is a perfect local model to
bundle/support for Mímir's sovereignty story.** News: M.1 + XS.2 Apr 2026; Nvidia up to $1B.

### Aider (aider.chat)
Open-source terminal pair programmer (Paul Gauthier, solo); 44k stars, 6.8M pip installs.
Free + BYOK (real cost = your tokens; Ollama = $0). Git-first (auto-commits every edit),
tree-sitter repo map, architect mode, 75+ providers, local/offline. **Gap: no memory, no
MCP (deliberate), terminal-only no GUI, single-agent, no guided SDLC.** **Maintainer went
quiet Oct-Nov 2025; no tagged release ~1 yr → local-first terminal niche under-served.**
News: v0.86.0 Aug 2025; reduced cadence since.

### Bolt.new (bolt.new) - StackBlitz
In-browser full-stack app builder (WebContainers); $135M raised ($700M val); $40M ARR.
Free / Pro ~$25 / Teams ~$30 / Enterprise. NL→full-stack + instant deploy + DB (Supabase),
Figma-to-app. **Gap: cloud-locked, token anxiety, vendor lock-in, no memory/SDLC, forced
V1→Claude Agent migration (Aug 2026) clears history = churn risk.** News: $105M Series B
Jan 2025; MS Azure partnership May 2026.

### Lovable (lovable.dev)
"Vibe coding" platform; fastest-growing software startup on record. $330M Series B at
**$6.6B** (Dec 2025); in talks ~$12-13B mid-2026; $500M+ ARR, ~146 staff. Free / Pro $25 /
Business $50 / Enterprise. Prompt→full app + hosting/DB/auth/payments. **Gap: cloud SaaS
lock-in, dual-layer billing opacity (#1 complaint), no local models/sovereignty, no
memory/SDLC, security growing pains.** News: $400M ARR Feb 2026; AIUC-1 cert Jul 2026.

### Continue (continue.dev) - ACQUIRED/SHUTTERED
Open-source AI code assistant (VS Code + JetBrains + CLI); 34k stars, 3.6M installs.
**Acquired by Cursor (acqui-hire) ~June 16, 2026; product shuttered** (final v2.0.0 Jun 19;
cloud data deleted Jul 15; repo read-only). Pre-acq: Solo free BYOK / Teams ~$10. Per-
feature model routing, 100+ providers, local-first, CLI "Continuous AI" (PR status checks).
**Gap: now defunct → 3.6M-install orphaned community needs a home (Apache 2.0 forkable but
unmaintained).** News: acquired by Cursor Jun 2026 (itself being bought by SpaceX/xAI).

---

## Updated White-Space Synthesis (after 29 competitors total)

The intersection **durable personal memory + guided gated SDLC + local-first sovereignty**
remains **completely unoccupied**. New strategic reads from batch 2:

1. **Three refugee pools just opened:** Continue (3.6M, shuttered by Cursor), Roo Code
   (3M, sunset), Aider (stalled). A maintained, memory-first, local-first successor can
   absorb them. Cline is the current magnet - Mímir competes with memory + SDLC + UI.
2. **Memory is heating up - move fast.** Codex shipped a "memory preview" (Apr 2026) -
   the first incumbent to touch it, but it's shallow preference recall, not a neural
   graph. Augment's Expert Registry is org-level, not personal. The personal-memory
   wedge is still open but closing.
3. **The prosumer/indie segment is being abandoned:** Tabnine killed free/individual
   tiers; Augment dropped its sub-$100 plan; Copilot's credit migration angered users.
   Mímir's free-self-hosted + cheap sync owns this.
4. **Cloud lock-in + billing opacity is the universal flank:** Replit ($9B), Lovable
   ($6.6-13B), Bolt ($700M) prove huge demand for "build with AI" but all have token
   anxiety + dual-layer billing + no data sovereignty. "Your models, your machine,
   cancel anytime" attacks them directly.
5. **Local sovereignty has a model ally:** Poolside's open-weight XS.2 (Apache 2.0, runs
   on a Mac) is a perfect local model to bundle/support - strengthens the no-cloud story.
6. **Computer use is table-stakes-soon:** Codex (Mac) + Factory Desktop (Apr 2026) +
   Replit self-test all have it. Mímir's F48 must match, not just claim.
7. **Autonomy vs guidance is the philosophical split:** the market races to "let go of
   the IDE" (Codex, Roomote, SOLO, Missions). Mímir's gated step-by-step is a deliberate
   contrarian bet for devs/teams who want control + auditability - and it's uncontested.

## Marketing / Positioning (refined)

- **Lead:** "the agent that remembers + walks the SDLC with you + you own it."
- **Recruit the refugees:** "your open, local, memory-first home" aimed at Continue/Roo/
  Aider orphans.
- **Attack billing opacity:** transparent, predictable, free self-hosted vs the credit/
  dual-layer chaos of Copilot/Cursor/Replit/Lovable.
- **Own the contrarian lane:** guided + human-in-the-loop, against the autonomy stampede.
- **Bundle a sovereign local model** (Poolside XS.2 / Qwen / nomic) to make "no cloud"
  real out of the box.

## TODO (next passes)
- Deep-dive authenticated dashboards of Kiro/Cursor/Cline (log in).
- Track news monthly (F72.7) - funding/pricing move fast (weekly changes observed).
- Build the Mímir dashboard (F52) using opencode's Usage page as the reference, improved
  with most-to-least model sorting + cloud/local tags.
