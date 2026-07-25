# Competitive Analysis: OpenCode

**Analyst:** Mímir competitive-analysis workflow (F72)
**Date:** 2026-07-23
**Pages walked:** homepage, /zen, /go
**Not yet walked (TODO):** /docs, /data, /enterprise, /download, /changelog, /brand
**URL:** https://opencode.ai

> Template: copy this file to `docs/competitive/<competitor>.md` for each competitor.

---

## 1. Overview

OpenCode is an open-source AI coding agent (terminal, IDE, desktop) by Anomaly.
160K-190K GitHub stars, 900 contributors, 7.5M monthly devs. Free CLI + two paid
products (Zen, Go). Positioning: "the open source AI coding agent" - privacy-first,
use any model, no lock-in.

## 2. Products & Pricing

| Product | What | Price | Models | Hook |
|---|---|---|---|---|
| **CLI** | open-source agent (terminal/IDE/desktop) | free | any (75+ via Models.dev) | open source, privacy-first |
| **Zen** | curated model gateway | pay-as-you-go, add $20 (+$1.23 fee), per-request, **zero markup**, auto-top-up at $5→+$20 | curated/benchmarked models | "use with any agent, cancel anytime" |
| **Go** | low-cost subscription | **$10/mo ($5 first month)**, top-up credit | open-source models (Grok 4.5, GLM-5.2, Kimi K3, Qwen3.7, MiniMax, DeepSeek...) | "generous limits, $5 first month" |
| **Enterprise** | enterprise tier | custom | - | privacy, governance |

**Pricing model:** free CLI (top of funnel) → Zen (pay-as-you-go, power users) + Go
($10/mo, cost-conscious). Both: "use with any agent," cancel anytime, transparent
pricing, zero markups.

## 3. Features (from site)

- LSP enabled (auto-loads the right LSPs)
- Multi-session (parallel agents on the same project)
- Share links (share a session for reference/debug)
- GitHub Copilot login (use your Copilot account)
- ChatGPT Plus/Pro login (use your OpenAI account)
- Any model (75+ providers via Models.dev, incl. local)
- Any editor (terminal, desktop app, IDE extension)
- Privacy-first (no code/context stored)

## 4. Site Design Patterns

**Consistent page template (every product page):**
```
Hero (headline + subline + ONE CTA + price + reassurance line)
  → "What problem is X solving?" (3 bullets)
  → "How X works" (3 numbered steps)
  → Privacy reassurance
  → Testimonials (named, with titles)
  → FAQ (expandable, kills objections)
  → Waitlist email capture (lead gen)
  → Footer (GitHub [star count], Docs, Changelog, Discord, X, © Anomaly, Brand, Privacy, Terms)
```

**Nav (consistent):** GitHub · Docs · Data · Zen · Go · Enterprise · Login/Download

## 5. Positioning & Messaging

- **Core promise:** "the open source AI coding agent" - free, open, privacy-first.
- **No lock-in:** "use with any agent," "use with any model," "cancel anytime" - repeated
  on every page. This is the whole pitch.
- **Transparent pricing:** "zero markups," clear numbers, no surprises.
- **Social proof:** star counts, 7.5M monthly devs, named testimonials with titles.
- **Privacy-first** as a headline feature, not a footnote.

## 6. Strengths

- Massive adoption (160K+ stars, 7.5M monthly devs) - strong network effects.
- Open source + privacy-first - trust.
- "Use with any agent/model" - no lock-in is a strong, repeated promise.
- Transparent pricing (zero markup) - builds trust.
- Two-product funnel (Zen pay-as-you-go + Go subscription) covers both power users and
  cost-conscious devs.
- Models.dev as a free data layer drives adoption.

## 7. Weaknesses / Gaps (our opportunities)

- **No memory.** OpenCode is stateless across sessions - it doesn't remember you. Mímir's
  Cortex (neurons/synapses/engrams) is the direct counter.
- **No guided workflow.** OpenCode is a raw agent; it doesn't walk you through the SDLC.
  Mímir's gated SDLC (Discovery→Design→Build→Debug→Deploy→Maintain) is the counter.
- **Cloud-dependent for the good stuff.** Zen/Go are cloud. Mímir is local-first,
  air-gapped capable.
- **No computer use / OS automation** mentioned. Mímir has it (F48).
- **No RGB/alive feel, no voice, no marketplace of forkable experiments.**

## 8. How Mímir Wins

| OpenCode | Mímir counter |
|---|---|
| stateless | **the Cortex - it remembers** (the namesake) |
| raw agent | **guided SDLC, gated, loops back** |
| cloud for the good stuff | **local-first, air-gapped** |
| no computer use | **computer use + OS automation** |
| no alive feel | **RGB sync, futuristic theme** |
| no forkable experiments | **WASM marketplace + forkable experiments** |

## 9. Marketing / Site Takeaways (apply to Mímir's site)

1. **Use their page template:** hero → "what problem is Mímir solving?" → "how it works"
   (3 steps) → features → pricing → FAQ → waitlist → footer.
2. **Lead with the differentiator:** "the agent that remembers" + "your models, your
   machine, no lock-in."
3. **One CTA per page** + waitlist email capture for lead gen.
4. **Transparent pricing:** free self-hosted + simple paid sync, "use with any model,
   cancel anytime."
5. **Social proof** when we have it (stars, users, testimonials).
6. **Privacy/local-first as a headline**, not a footnote.
7. **FAQ that kills objections** (do I need subscriptions? is it really free? privacy?).

---

## TODO (next pass)
- Walk /docs, /data, /enterprise, /download, /changelog, /brand.
- Screenshot each page for the design reference.

---

## 10. Authenticated Dashboard (walked Jul 2026, logged in)

Workspace URL pattern: `opencode.ai/workspace/wrk_<id>`. Left nav (vertical text links):
**Zen · Go · Usage · API Keys · Members · Billing · Settings**. Top bar: workspace
switcher ("Default") + account email (top-right).

### Zen tab (model management)
- Header: "Reliable optimized models for coding agents." + **Current balance $X.XX** (top-right).
- **Models table**: columns `MODEL` (name + provider logo + provider name) and `ENABLED`
  (green toggle per model). ~50 models listed (Anthropic Claude family, OpenAI GPT-5.x +
  Codex variants, Google Gemini 3.x, DeepSeek V4, GLM 5.x, Kimi K2.x, Qwen3.x, Grok 4.5,
  MiniMax M2/M3, plus free "Stealth" models like Big Pickle, MiMo, Nemotron, Laguna).
  Admins toggle which models workspace members can use.
- **Bring Your Own Key** section below: rows per provider (OpenAI / Anthropic / Google
  Gemini) with a `Configure` button to paste your own key.

### Usage tab (the cost/usage view - key reference for Mímir's F52 dashboard)
- **Cost** section: "Usage costs broken down by model." Month picker (‹ July 2026 ›) +
  `All Models` + `All Keys` filter dropdowns. (Chart area; "No usage data" if empty.)
- **Usage History** table: columns `DATE · MODEL · INPUT · OUTPUT · COST · SESSION`.
  Per-request rows (e.g. glm-5.2, 229164 input / 594 output, Go ($0.3233), session id).
  Paginated.

### Takeaways for Mímir's dashboard (F52)
- opencode's layout is clean and minimal (monospace, lots of whitespace, green toggles).
- Their Usage view is per-request chronological. **Mímir's improvement:** aggregate by
  model and **sort most-tokens-used to least** (the user's explicit ask), plus a total +
  cost estimate, plus cloud-vs-local tagging.
- The BYOK section is a great pattern: list providers with a `Configure` button so users
  can drop in their own keys next to the managed gateway. Mímir should mirror this.
- Balance top-right + per-model enable toggles = a tight, scannable admin surface.
