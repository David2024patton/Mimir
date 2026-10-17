# End-to-End Scenario: Building & Shipping a SaaS Platform With Mímir (Raw Idea to Production)

This walkthrough demonstrates how a developer (or solo founder) takes a SaaS product from a **raw brainstorm idea** all the way to **live production and post-launch maintenance** using Mímir's **7-Phase Zero-Bloat SDLC Engine** (`internal/sdlc`), **Liquid AI `d1-omni-600M` Decision Router** (`internal/decision`), **Yggdrasil Mind Map** (`internal/ingest`), and **4-Layer Web/SEO/GEO Rules** (`internal/agent/rules.go`).

---

## The Scenario Idea: "MapPulse" (Local Service Lead-Response & GBP Review SaaS)
The user wants to build **MapPulse** -- a micro-SaaS for local home-service businesses (pest control, HVAC, plumbing) that:
1. Catches missed inbound phone calls and web leads within 10 seconds via SMS + 1-click human callback escalation.
2. Sends FTC-compliant post-job Google Business Profile (GBP) review requests (never gating on positive sentiment).
3. Publishes a high-converting public SaaS marketing site with transparent pricing tiers, JSON-LD `Schema.org`, and `/llms.txt` for AI search engines.

```mermaid
flowchart LR
    P1["1. Planning (Odin)\n01-PLAN.md"] --> P2["2. Analysis (Huginn)\n02-SRS.md"]
    P2 --> P3["3. Design (Odin)\n03-SDD.md"]
    P3 --> P4["4. Implementation (Thor)\n04-BUILD.md"]
    P4 --> P5["5. Testing (Loki)\n05-QA-REPORT.md"]
    P5 --> P6["6. Deployment (Heimdall)\n06-RELEASE.md"]
    P6 --> P7["7. Maintenance (Muninn)\n07-MAINTENANCE.md"]
```

---

## Step 1: Planning (`[SDLC 1/7: PLANNING]` -- Brainstorming, Scope & Charter)

### 1. User Action
The user creates an empty folder `G:\MapPulse`, launches Mímir (`mimir --web`), and types:
> *"I want to build MapPulse, a SaaS for local trades (HVAC, pest control, plumbers) that texts back missed calls in 10 seconds with human escalation and sends FTC-compliant Google Maps review requests after a job closes. Let's brainstorm the project plan, scope, and feasibility charter first."*

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase Detection (`internal/sdlc/sdlc.go`)**: `DetectPhaseFromPrompt` sees `"project plan"`, `"scope"`, and `"charter"` and locks into **Step 1: `planning`**.
- **Zero-Bloat Prompt Loaded (`internal/sdlc/phases.go`)**: Instead of loading 2,500 chars of coding/testing/deployment rules, Mímir injects **only** the 29-word Step 1 micro-prompt:
  `[SDLC 1/7: PLANNING (Odin)] Define scope, feasibility, risks, and modular milestones. Record a step-by-step checklist in todo and write the Project Plan & Charter via sdlc before coding.`
- **Tool Whitelist Active**: Only `read_file`, `list_dir`, `ingest_repo`, `todo`, `question`, and `sdlc` are exposed (write/edit tools are held back until Design/Implementation).
- **Persona Activated**: `Odin` (Architect & Planner).

### 3. Mímir's Execution & Deliverable
1. `Odin` calls the `question` tool (`internal/tools/question.go`) to ask 3 quick multiple-choice product questions ( rendered as clickable `[choice]` buttons in the TUI/Web/Phone UI):
   - **Billing stack**: `A) Stripe Checkout + Webhooks` | `B) LemonSqueezy`
   - **Telephony provider**: `A) Twilio Programmable SMS/Voice` | `B) Telnyx`
   - **Database**: `A) SQLite (WAL mode, zero external container)` | `B) Postgres`
2. Once the user clicks `A, A, A`, `Odin` populates the persistent `todo` checklist (`internal/tools/todo.go`) with 7 milestones matching the SDLC phases.
3. `Odin` calls the `sdlc` tool (`action: "save_deliverable"`, `phase: "planning"`) to write **`.mimir/sdlc/01-PLAN.md`** (*Project Plan & Charter*), defining:
   - MVP scope (Missed-Call SMS Responder, 6-Field Lead Record, FTC-Compliant Review Sender, Public SaaS Landing Page).
   - Unit economics (`$49/mo` Starter tier, `$129/mo` Pro tier, estimated SMS/voice pass-through cost).
   - Risk register (TCPA/FCC consent compliance, FTC Oct 2024 review-gating ban, webhook idempotency).

---

## Step 2: Analysis (`[SDLC 2/7: ANALYSIS]` -- Exact System & SEO Requirements)

### 1. User Action
The user runs `/sdlc next` (or says: *"Great, now let's gather the exact SRS requirements and security constraints."*).

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase**: Advances to **Step 2: `analysis`** (`Persona: Huginn`).
- **Zero-Bloat Prompt Loaded**: Step 1's prompt is unloaded. Only Step 2's micro-prompt + the 1-line `[Web/SEO Analysis]` overlay (~95 tokens total) are loaded:
  - `[SDLC 2/7: ANALYSIS (Huginn)] Inspect the Yggdrasil Mind Map (ingest_repo/read_file), clarify ambiguities with question, and define exact functional/security constraints in the SRS.`
  - `[Web/SEO Analysis]: Define target entity facts (NAP, pricing units, service areas) and first-party lead conversion fields (source, response_time, contact, outcome).`

### 3. Mímir's Execution & Deliverable
1. `Huginn` maps the exact functional and compliance requirements:
   - **Lead Record Schema (6 Mandatory Fields)**: `lead_id`, `source`, `response_time_ms`, `contact_phone`, `booked_outcome`, `next_action`.
   - **Human Escalation Requirement**: Every automated SMS must disclose it is MapPulse's assistant for the business and offer `"Reply 0 or press 1 to speak directly with the on-call technician"` (preventing undisclosed-AI drop-off).
   - **FTC Review Compliance**: Review requests trigger on `job_status == "completed"` for **all** customers equally -- zero sentiment screening or review gating.
   - **Public SaaS Entity Facts**: Exact pricing tiers (`$49/mo` includes 250 SMS + 1 tracking number; `$129/mo` includes 1,000 SMS + 3 numbers), refund policy, and company contact facts.
2. `Huginn` calls `sdlc` (`action: "save_deliverable"`, `phase: "analysis"`) to write **`.mimir/sdlc/02-SRS.md`** (*Software Requirement Specification*).

---

## Step 3: Design (`[SDLC 3/7: DESIGN]` -- Architecture, DB Schema & Anti-Slop Wireframe)

### 1. User Action
The user types: *"Blueprint the system design, SQLite schema, API routes, and folder structure."*

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase**: Auto-switches to **Step 3: `design`** (`Persona: Odin`).
- **Zero-Bloat Prompt Loaded**: Loads only Step 3's micro-prompt + `[Web/SEO Design]` overlay (~105 tokens):
  - `[SDLC 3/7: DESIGN (Odin)] Blueprint modular package/folder boundaries, database schemas, C-ABI/API contracts, and UI wireframes in the SDD before implementation.`
  - `[Web/SEO Design]: Blueprint semantic HTML5 landmarks, JSON-LD Schema.org types (LocalBusiness/SoftwareApplication/FAQPage), /llms.txt structure, and non-formulaic UI layout.`

### 3. Mímir's Execution & Deliverable
1. `Odin` designs a clean, modular folder structure for `G:\MapPulse`:
   ```text
   G:\MapPulse\
   ├── cmd/mappulse/main.go          # Single-binary Go server + static asset embed
   ├── internal/
   │   ├── db/schema.sql             # SQLite WAL tables: tenants, leads, sms_logs, review_requests, beacon_events
   │   ├── telephony/twilio.go       # Missed-call webhook + 10s SMS responder + human escalation router
   │   ├── reviews/dispatcher.go     # FTC-compliant post-job GBP review link sender
   │   ├── billing/stripe.go         # Subscription checkout + webhook handler
   │   └── web/                      # Server-rendered HTML5 + First-Party Retention Beacon
   └── public/
       ├── index.html                # Semantic HTML5 landing page + JSON-LD SoftwareApplication & FAQPage
       ├── llms.txt                  # Machine-readable entity & pricing summary for ChatGPT/Perplexity/Google AI
       ├── robots.txt
       └── sitemap.xml
   ```
2. `Odin` calls `sdlc` (`action: "save_deliverable"`, `phase: "design"`) to save **`.mimir/sdlc/03-SDD.md`** (*System Design Document*).

---

## Step 4: Implementation (`[SDLC 4/7: IMPLEMENTATION]` -- Writing the Code)

### 1. User Action
The user types: *"Build the MapPulse backend, SQLite store, Twilio/Stripe handlers, and the public SaaS site."*
*(If the user walks away from their desk to grab coffee, they run `mimir pair`, scan the terminal QR code with the Mímir Phone App, and approve file writes or steer from their phone.)*

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase**: Switches to **Step 4: `implementation`** (`Persona: Thor`).
- **Zero-Bloat Prompt Loaded**: Loads only Step 4's micro-prompt + `[Web/SEO Build]` overlay (~115 tokens):
  - `[SDLC 4/7: IMPLEMENTATION (Thor)] Read target files first, make surgical edit_block changes (<15 lines context), keep packages modular, and never use em dashes or placeholder stubs.`
  - `[Web/SEO Build]: Emit semantic HTML5, JSON-LD Schema.org, /llms.txt, explicit pricing/entity facts, and first-party lead capture with human handoff. Zero AI slop or fake stats ([vendor claim]).`
- **All Build Tools Unlocked**: `write_file`, `edit_block`, `run_shell`, `read_file`, `list_dir`, `todo`, `skills`, `sdlc`.
- **If using a local model (`qwen3:14b` or `muse-glimmer`)**: `⚡SmolMode` automatically activates, disabling chain-of-thought (`Effort: "off"`) on mechanical tool steps `1..N` and compacting older tool outputs to 1-line placeholders.

### 3. Mímir's Execution & Deliverable
1. `Thor` writes the modular Go packages (`internal/db`, `internal/telephony`, `internal/reviews`, `internal/billing`, `internal/web`).
2. `Thor` builds `public/index.html`, `public/llms.txt`, `public/robots.txt`, and `public/sitemap.xml` following every rule in `[Web/SEO Build]`:
   - Semantic `<header>`, `<nav>`, `<main>`, `<section id="pricing">`, `<section id="faq">`, `<footer>`.
   - Embedded `<script type="application/ld+json">` with `SoftwareApplication` (`Offer` tiers `$49` and `$129`) and `FAQPage`.
   - Zero generic "in today's fast-paced world" AI slop, zero em dashes, and zero fabricated ROI numbers (any external study is explicitly cited with `[vendor claim]`).
   - Built-in first-party conversion beacon (`POST /api/beacon`) logging every demo request and signup into SQLite.
3. `Thor` runs `mimir ingest .` (`ingest_repo` tool) so `G:\MapPulse`'s new files and symbols are stored in the **Yggdrasil Mind Map** in Cortex memory, then saves **`.mimir/sdlc/04-BUILD.md`**.

---

## Step 5: Testing (`[SDLC 5/7: TESTING]` -- Linter, Unit Tests & SEO/Schema QA)

### 1. User Action
The user types: *"Run all unit tests, linter checks, and verify the SEO/GEO requirements against our SRS."*

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase**: Switches to **Step 5: `testing`** (`Persona: Loki`, the adversarial QA & Test Critic).
- **Zero-Bloat Prompt Loaded**: Loads only Step 5's micro-prompt + `[Web/SEO QA]` overlay (~95 tokens):
  - `[SDLC 5/7: TESTING (Loki)] Run linter, static analysis, and affected unit/integration tests via run_shell. Fix every failure and record the QA Test Report via sdlc.`
  - `[Web/SEO QA]: Verify valid JSON-LD Schema.org, /robots.txt, /sitemap.xml, /llms.txt, zero em dashes, and FTC-compliant factual claims.`
- **Decision Model Test Selection (`internal/decision/decider.go`)**: `SelectAffectedTests` queries the Yggdrasil Mind Map to identify every `*_test.go` file covering `telephony`, `reviews`, `billing`, and `web`.

### 3. Mímir's Execution & Deliverable
1. `Loki` executes `go vet ./... ; go test -v -race -count=1 ./...` via `run_shell`.
2. `Loki` tests edge cases against Step 2's SRS:
   - Simulates duplicate Twilio missed-call webhooks to verify idempotency (ensuring a customer is never double-texted).
   - Simulates a `"1-star angry customer"` job completion to verify the FTC compliance gate (confirming the review dispatcher never blocks or filters unhappy customers).
   - Parses `public/index.html` JSON-LD to verify valid JSON and checks `public/llms.txt` and `public/sitemap.xml`.
3. `Loki` calls `EvaluateStopGate` (`internal/agent/rules.go`) to confirm 0 pending todos and 100% passing verification, then writes **`.mimir/sdlc/05-QA-REPORT.md`**.

---

## Step 6: Deployment (`[SDLC 6/7: DEPLOYMENT]` -- Production Container & Live Release Gate)

### 1. User Action
The user types: *"Package MapPulse for production deployment with a self-healing SQLite volume and run the release gate."*

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase**: Switches to **Step 6: `deployment`** (`Persona: Heimdall`, Guardian of Release & Security).
- **Zero-Bloat Prompt Loaded**: Loads only Step 6's micro-prompt + `[Web/SEO Release Gate]` overlay (~95 tokens):
  - `[SDLC 6/7: DEPLOYMENT (Heimdall)] Verify release binaries/containers, environment configs, and rollback readiness. Never expose secrets. Record the Release manifest via sdlc.`
  - `[Web/SEO Release Gate]: Confirm live canonical tags, /llms.txt, Search Console / GBP indexability, and active first-party lead beacon.`

### 3. Mímir's Execution & Deliverable
1. `Heimdall` writes a multi-stage `Dockerfile` and `docker-compose.yml` (using a non-conflicting 5-digit port such as `:48219` and an idempotent entrypoint that self-heals `/data/mappulse.sqlite` volume permissions on boot).
2. `Heimdall` builds and boots the production container locally (`docker compose up -d --build`), curls `http://localhost:48219/healthz`, `http://localhost:48219/llms.txt`, and `http://localhost:48219/api/beacon`, and verifies container logs (`docker compose logs`).
3. `Heimdall` writes **`.mimir/sdlc/06-RELEASE.md`** (*Live Production Release Manifest*) with exact rollback instructions and deploys to the production VPS.

---

## Step 7: Maintenance (`[SDLC 7/7: MAINTENANCE]` -- Telemetry, Patches & Next Sprint Loop)

### 1. User Action
Two weeks after launch, the user opens the Mímir Phone App (or TUI) and says:
> *"Check our production logs and first-party beacon conversion stats, patch any slow queries, and queue the next sprint items."*

### 2. What Mímir's Harness Does Under the Hood
- **SDLC Phase**: Switches to **Step 7: `maintenance`** (`Persona: Muninn`, Keeper of Long-Term Memory & Telemetry).
- **Zero-Bloat Prompt Loaded**: Loads only Step 7's 29-word micro-prompt:
  - `[SDLC 7/7: MAINTENANCE (Muninn)] Inspect logs/telemetry, apply targeted regression patches, update the Yggdrasil Mind Map (ingest_repo), and queue new sprint items in todo.`

### 3. Mímir's Execution & Deliverable
1. `Muninn` inspects the SQLite `beacon_events` and `leads` tables, reporting actual measured conversion metrics (e.g., median SMS response time `420ms`, `18%` callback escalation rate, `0` failed webhooks).
2. `Muninn` adds an index on `leads(tenant_id, created_at)` to keep dashboard queries under `2ms`, verifies with `go test ./...`, refreshes the Yggdrasil Mind Map (`mimir ingest .`), and writes **`.mimir/sdlc/07-MAINTENANCE.md`**.
3. When the user is ready for the next feature sprint, `/sdlc next` in `agile` mode loops cleanly back to **Step 1 (`Planning`)** with all architectural lessons preserved in `.mimir/cortex.json`.
