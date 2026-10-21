<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: every link David posts about this project, and the
answer to it. When David drops a link in any chat (a repo, an article, a
prompt collection, a product), it lands here with what it is and what we
took from it, so the research is never lost and never repeated.

EXACT STRUCTURE: one entry per link, newest first:
## <YYYY-MM-DD HH:MM TZ>: <link>
- **What it is:** <one or two lines, factual>
- **License:** <license or "n/a">
- **Lulu's answer:** <verdict: TAKE, RETROFIT, REFERENCE, or SKIP, with
  the reason in plain words>
- **Went to:** <where the takeaway lives now: a TODO item id, an IDEA
  entry, the shelf file, or "nowhere, verdict only">

RULES:
- Every posted link gets an entry the same day, even when the verdict
  is SKIP; the skip and its reason are the value.
- Facts only: license and contents verified from the source, not from
  the link's reputation.
- No secrets, passwords, or tokens in this file, ever.
-->
# REFERENCES: Mimir

Entries from David's 2026-10-20 link batch (posted evening EDT; entries
carry the date and batch only, no precise times were recorded). Facts
taken from each repo's GitHub page on 2026-10-20.

## 2026-10-20 (evening EDT): https://github.com/ishandutta2007/Open-Laudable
- **What it is:** A local-first open-source Lovable/v0/Bolt alternative packaged as an Electron desktop app: prompt-to-app generation of React/TypeScript + Tailwind, Supabase/SQLite backends, multi-model support (OpenAI, Anthropic, Gemini, Ollama local models). The root layout (`packages/@dyad-sh`, makers, scaffold, workers) shows it is built on the Dyad codebase. 1,111 commits, 54 stars.
- **License:** Conflicted on the page: GitHub reports "Other (NOASSERTION)" while the README claims Apache 2.0. Treat as unconfirmed until the LICENSE file itself is read.
- **Lulu's answer:** REFERENCE for B, SKIP for A. For A it is the wrong form factor entirely (a desktop Electron app, not a hosted multi-tenant builder). For B it is worth a look as the Dyad lineage in the wild: local-first execution, BYO keys and local-model support, visual component inspector and multi-file diff view are the features Mimir's own UI roadmap can be compared against. Do not take code until the license conflict is resolved by reading the actual LICENSE file.
- **Went to:** TODO T-5.3 (Flutter stack, Gjallarhorn-UI): the visual component inspector and multi-file diff view are the comparison set for the UI roadmap

## 2026-10-20 (evening EDT): https://github.com/emanueleielo/deepagents-open-lovable
- **What it is:** An open Lovable-style platform built on DeepAgents + LangGraph: a Python agent backend generates complete React/Next.js apps from chat (live Sandpack preview, file browser, Vercel one-click deploy), with specialized sub-agents for design and image research and SKILL.md-based skills the agent loads. 71 commits, 112 stars.
- **License:** Not pinned down on the page: the README says "open source" and defers to "the individual components for their respective licenses"; no LICENSE file in the root listing.
- **Lulu's answer:** REFERENCE for B, REFERENCE (weaker) for A. For Mimir the interesting part is the shape, not the stack (Python/LangGraph does not drop into a Go harness): a main agent plus named specialist sub-agents, and skills as SKILL.md files in a skills folder the agent reads, which independently confirms the skill pattern already in use here. For A it is one more proof that sub-agent specialization (a designer agent separate from the builder agent) produces better generated sites. License ambiguity alone rules out taking code.
- **Went to:** TODO T-6.4 (run-graph execution model): the main agent plus named specialist sub-agents shape is the comparison point for stage and agent design

## 2026-10-20 (evening EDT): https://github.com/beam-cloud/lovable-clone
- **What it is:** A Lovable clone demonstrating Beam's platform: a Python agent (BAML-defined prompts, FastMCP) streams edit requests over websockets to a React app preview running in a Beam cloud sandbox (Node 20, React + Vite + shadcn/ui template). 20 commits, 293 stars.
- **License:** No license shown on the repo page and no LICENSE file in the root listing.
- **Lulu's answer:** REFERENCE for A and B. Do not take the code (no license, and it is welded to Beam's paid cloud). The architecture is the lesson: prompts defined as typed BAML functions with a `CodeChanges` schema and test cases, a websocket edit stream, and a sandboxed preview. For A that validates the typed-edit-contract shape between the builder UI and the generation engine; for B the BAML idea (prompts as schema-checked functions with tests) is a pattern worth comparing against Mimir's stage contracts.
- **Went to:** TODO T-6.4 (run-graph execution model): prompts as schema-checked functions with tests is the pattern to compare against the run graph's typed done gates

## 2026-10-20 (evening EDT): https://github.com/x1xhlol/system-prompts-and-models-of-ai-tools
- **What it is:** The largest public collection of leaked and open system prompts (about 144,000 stars, 34,700 forks): one folder per tool holding the full system prompt, internal tool definitions, and model information for roughly 30 AI tools, including Augment Code, Claude Code, Cursor, Devin AI, Junie, Kiro, Leap.new, Lovable, Manus (agent tools and prompt), NotionAI, Orchids.app, Perplexity, Qoder, Replit, Same.dev, Trae, Traycer AI, VSCode Agent, Warp.dev, Windsurf, Xcode, Z.ai Code, Dia, and v0 (prompts and tools), plus an "Open Source prompts" folder. The README leads with a security notice warning AI startups that exposed prompts are an extraction target. Where CL4R1T4S gives one prompt per product, this repo's value is breadth and the tool-definition files: how each agent's tools are named, described, and schema'd for the model.
- **License:** GPL-3.0 on the repository, but the contents are the respective vendors' proprietary prompts and tool specs published without their permission; the wrapper license does not clear the underlying IP. Study only, copy nothing.
- **Lulu's answer:** REFERENCE for both. For (A) the v0, Same.dev, Leap.new, and Lovable folders are the comparator set for the builder's generation prompts, worth a side-by-side study when the builder's prompt is next revised. For (B) the Devin, Manus, and Claude Code folders are the comparator set for Mimir's agent tool design: tool naming, description wording, and schema shapes from three production coding agents in one place. No TAKE or RETROFIT: the asset is the library itself, consulted during design work in original words.
- **Went to:** this file (verdict on record)

## 2026-10-20 (evening EDT): https://github.com/eyaltoledano/claude-task-master
- **What it is:** A real, actively maintained open-source tool (about 28,300 stars, 92 releases), not a prompt dump: an AI task-management system for AI-driven development that drops into Cursor, Lovable, Windsurf, Roo, VS Code, and Claude Code via an MCP server (36 tools) or a CLI (npm package task-master-ai). Core flow: start from a detailed PRD (recommended at .taskmaster/docs/prd.txt), parse it into a structured task graph (tasks.json) with dependencies, priorities, and test strategies; expand any task into subtasks on demand; run a complexity analysis pass that scores tasks and recommends which to expand; ask for the next task and get the dependency-aware next unblocked item; organize parallel workstreams with tags and move tasks across tags with or without their dependencies. It defines three model roles (main, research, fallback) swappable across many providers (Anthropic, OpenAI, Gemini, Perplexity, xAI, OpenRouter, plus Claude Code and Codex CLI with no API key), and a built-in research command that answers with project context. One notable engineering detail: selective MCP tool loading to control context cost (all 36 tools about 21,000 tokens, standard 15 about 10,000, core 7 about 5,000).
- **License:** MIT with Commons Clause. Use, modify, distribute, and sell products built USING Task Master are allowed; selling Task Master itself, offering it as a hosted service, or creating competing products based on it are NOT allowed.
- **Lulu's answer:** RETROFIT the patterns for (B), REFERENCE for (A), TAKE nothing. For Mimir the decomposition model is directly on point: PRD-first intake (a detailed spec produces a better task graph), expand-task as an explicit operation, complexity scoring that decides where to decompose further, and next-task resolution over a dependency graph are patterns Mimir's planning layer should match in its own Go code. The code itself stays out: the Commons Clause bars competing products and hosted services, which is exactly Mimir's shape, so patterns yes, source no. For the AlphaFlux builder it is reference grade: the PRD-first lesson (a detailed customer brief yields a better build plan) and the tiered tool-loading trick (expose agents only the tool subset the step needs) both transfer as ideas.
- **Went to:** TODO T-6.4 (run-graph execution model): PRD-first intake, expand-task, complexity scoring, and dependency-aware next-task resolution are the planning patterns to match in Mimir's own Go code

## 2026-10-20 (evening EDT): https://github.com/elder-plinius/CL4R1T4S/tree/main/REPLIT
- **What it is:** The REPLIT folder holds exactly one file, Replit_Initial_Code_Generation_Prompt.md: Replit's one-shot initial code generation prompt. Input contract: a goal, a task description, and success criteria; output: the complete source of a working app in a single response. Structure: an Input Description section; Output Rules covering directory structure (minimal, flat, frontend and backend coexist rather than splitting into service folders), complete code per file with comments on complex logic, and security hygiene (avoid SQL injection and XSS); a strict markdown output format with a `# Thoughts` section, then a `# directory_structure` section as a JSON array of {path, status: "new" or "overwritten"}, then each file's full contents under a `## file_path:` heading; code-generation rules stating the runtime contract up front (unprivileged Linux container, bind 0.0.0.0, frontend on port 5000, backend on port 8000, API keys from environment variables with fallback, no Docker); development constraints (favor web apps, SVG first for graphics, named asset libraries: Feather Icons and Font Awesome, Bootstrap, Fabric.js and Two.js, Chart.js and D3.js, tone-js); a restricted-files deny list (never generate package.json or requirements.txt because the platform owns them, never emit binary assets like images, audio, or fonts, use libraries and CDNs instead); and a worked example output for a TODO app.
- **License:** The repo carries AGPL-3.0 on the collection; the content is Replit's proprietary prompt, published without Replit's permission; the repo license does not clear the underlying IP. Study only, copy nothing.
- **Lulu's answer:** REFERENCE for both, RETROFIT the output contract. For (A) the manifest-first format is the extractable gem: the generator emits a machine-readable file manifest (path plus new/overwritten status) before any file bodies, which is exactly the contract an AlphaFlux builder needs to show the customer a plan, diff against an existing site, and apply changes safely. The platform-owns-the-manifest deny list (package.json and friends) is the same seam. For (B) it is a reference point only: Mimir's staged SDLC with hard gates is deliberately not a one-shot generator, but this prompt is a clean model for a greenfield scaffolding subtask inside a stage.
- **Techniques worth extracting:**
  - Success criteria are part of the input contract, stated before any generation.
  - Manifest-first output: a JSON directory structure with a per-file status flag precedes all file contents, so a machine can plan, diff, and verify.
  - A separate Thoughts section keeps reasoning out of the artifacts.
  - State the runtime contract (ports, bind host, privileges, available tooling) inside the prompt instead of letting the model guess the environment.
  - Deny list of platform-owned files so the generator and the platform never fight over the same artifacts.
  - Asset strategy: vector and library/CDN assets only, no generated binaries.
  - One complete worked example of the output format at the end of the prompt.
- **Went to:** this file (verdict on record)

## 2026-10-20 (evening EDT): https://github.com/elder-plinius/CL4R1T4S/blob/main/LOVABLE%2FLovable_2.0.txt
- **What it is:** The leaked Lovable 2.0 system prompt, 353 lines, internally dated 2025-04-25. One file in elder-plinius/CL4R1T4S, a collection of extracted system prompts from most major AI products (OpenAI, Anthropic, Google, xAI, Cursor, Devin, Manus, Replit, Vercel v0, and more). The Lovable prompt is a complete app-builder agent spec in XML-tagged sections: a `<role>` block (AI editor with live iframe preview and console-log access); a `<response_format>` block with a triage flow (check whether the request is already implemented before editing; answer questions without touching code; edit only on explicit action words); one single `<lov-code>` block per response wrapping ALL changes, with `<lov-write>`, `<lov-rename>`, `<lov-delete>`, and `<lov-add-dependency>` tags inside it; two full worked examples showing the exact expected output shape; a `<guidelines>` block pinning the stack (React, Tailwind CSS always, shadcn/ui, lucide-react icons, recharts, @tanstack/react-query in object format, errors left to bubble up instead of try/catch, "do not overengineer, don't do more than the user asks"); a `<tools>` block (never name tools to users); per-integration blocks for OpenAI models, Perplexity, and Runware (API keys go in Supabase Edge Function secrets, or a temporary user-entered key field if not connected, never hardcoded); a `<supabase-integration>` gate that refuses to write ANY backend code until the user connects the native Supabase integration; a `<session_replay>` block injecting the user's mouse movements, clicks, and text selections as structured context; and a closing micro-lesson section on JSX quote escaping.
- **License:** The repo carries AGPL-3.0, but the file's content is Lovable's proprietary system prompt, published without Lovable's permission; the repo license does not clear the underlying IP. Study only, copy nothing.
- **Lulu's answer:** REFERENCE for both, with techniques worth RETROFIT into our own original prompts. For (A) this is the closest public blueprint of how the market-leading builder agent behaves: the triage-before-editing flow, the one atomic change block with a file plan up front and a completeness checklist before closing (all imports present, dependencies installed), and the integration gate that stops and routes the user through a native connection instead of generating backend code against missing infrastructure all map directly onto the AlphaFlux builder and its iframe-guided provider setup. For (B) the response discipline (already-implemented check, one-sentence plain summary after the technical block) is a useful pattern for Mimir's user-facing reporting.
- **Techniques worth extracting:**
  - Triage first: verify the request is not already implemented; answer pure questions with no code changes; treat action words as the edit trigger.
  - One atomic change block per turn containing every file operation (write, rename, delete, add dependency), opened with a step-by-step file plan and closed with a completeness review (imports resolve, packages installed).
  - Few-shot by construction: embed two complete worked examples of the desired output format inside the prompt itself.
  - Pin one opinionated stack (framework, CSS system, component library, icon set, chart lib, data-fetching convention) to cut model variance.
  - Secrets pattern: keys live in the platform's secret store; fallback is a temporary user-entered field; keys are never written into generated code.
  - Integration gate: when a capability needs a connected platform service, refuse to generate around it and walk the user through connecting it first.
  - Dynamic context sections: structured session replay (clicks, selections) injected into the prompt alongside the code state.
  - Named failure-mode micro-lessons appended to the prompt as their own sections (the JSX quote-escaping section is the model).
  - Let errors bubble back to the agent instead of swallowing them in generated try/catch, so failures stay visible and fixable.
- **Went to:** this file (verdict on record)

## 2026-10-20 (evening EDT): https://github.com/dyad-sh/dyad
- **What it is:** A local, open-source AI app builder for power users: a cross-platform desktop app (Electron, built with Electron Forge; TypeScript/React/Vite) positioned as a local v0/Lovable/Bolt alternative. Fast, private, no lock-in, bring your own API keys, with a wide provider set including local models (Ollama appears in its topics alongside OpenAI, Anthropic, Gemini, DeepSeek, Qwen). It is by far the most mature project in this batch: 22,011 stars, 2,785 forks, 2,169 commits, 151 releases, created 2025-04-11.
- **License:** Apache 2.0 for all code outside `src/pro`; code inside `src/pro` is fair-source under the Functional Source License 1.1 (FSL-1.1, converts to Apache 2.0). GitHub's license detector reports it as "Other (NOASSERTION)" because of the split.
- **Lulu's answer:** REFERENCE for B. The relevant study material for Mimir is how a local-first harness handles many model providers including on-device models, bring-your-own-keys with no vendor account, and app scaffolding, all proven at real scale with a heavy release cadence. The codebase is an Electron desktop app, not a Go harness, and the `src/pro` fair-source portion is off limits for direct reuse, so this is a read-and-learn source, not a code source. Minor value for A beyond confirming the local/BYO-key positioning customers recognize.
- **Went to:** this file (verdict on record)

## 2026-10-20 (evening EDT): https://github.com/freestyle-sh/Adorable
- **What it is:** An open-source Lovable clone where every project is a pair of persistent Freestyle VMs: a dev VM the agent edits (hot-reloading dev server behind a preview URL, terminal sessions) and a production VM created on first publish. Publishing rsyncs source to the production VM in about 4 seconds; every publish snapshots the dev VM so any earlier release can be restored as a rollback. A pre-built base snapshot (template cloned, dependencies installed, dev server already compiled) makes a new project's preview live in about 1.7 seconds. State is JSON files on the dev VM; there is no database. Stack: Next.js (App Router, TypeScript, Turbopack), Vercel AI SDK, Tailwind + shadcn/ui, Freestyle VMs. 829 stars, 318 forks, 277 commits, created 2025-04-10.
- **License:** MIT
- **Lulu's answer:** RETROFIT for B. The persistent per-project VM with snapshot-based publish and rollback is a concrete, working model for Mimir's executor seam: a workspace that survives between sessions, with releases as snapshots and rollback as a boot from an earlier snapshot. The base-snapshot trick (pre-compiled template snapshot so startup skips the cold build) is a second pattern worth stealing for harness cold starts. It is bound to Freestyle's VM API and Next.js, so the architecture transfers, the code does not. Little direct value for A.
- **Went to:** TODO T-2.1 (merge the run-engine scaffold into main) and T-2.2 (give RunUnattended a production caller): the persistent workspace with snapshot publish and rollback is the working model for the executor seam the run engine lands on
