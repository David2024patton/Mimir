# Relational map: nodes

Every node in Mimir: features, modules, engines, services, screens,
external systems, and the planned pieces, each with its kind, its state,
and the source file or place it was verified from on 2026-11-18.

States: LIVE means it exists in code and works as described. PLANNED
means it is named work with a TODO id and no code yet. ABSENT means
paper in this repo claims or assumes it and the code does not have it;
each ABSENT node names the disagreement.

This file is canonical. [wires.md](wires.md) holds the connections,
[diagram.md](diagram.md) renders the same nodes and wires, and
[../MAP.md](../MAP.md) is the entry point.

| Node | Kind | State | Source | Owns |
| :--- | :--- | :--- | :--- | :--- |
| Operator | external | LIVE | external | The human: plans with Mimir, approves plans, and holds the hard gates at verify and ship |
| CLI: cmd/mimir | module | LIVE | cmd/mimir/main.go | The entry point and the full subcommand set: version, help, setup (init), ingest (gitingest), providers (subscribe), sdlc (phase), audit (critic), audit-web (web-audit), engines (status, publish, run), update, swarm (factory: status, add, run, approve), review, ship, handoff, pair, serve, chat, trace, and the TUI launch; flags include --web, --addr, --mode, --effort, --model, --prompt |
| TUI: internal/tui | screen | LIVE | internal/tui | The terminal interface for interactive sessions; also runs ingest, review and ship gates, and a swarm factory panel |
| Server and Web UI: internal/server | service | LIVE | internal/server | The HTTP server on :8420 (default 127.0.0.1:8420): OTP auth, sessions, chat and chat stream, memory, usage, engines and swarm pages and APIs, mobile stream, steer and mode endpoints, and the pairing endpoints |
| Relay: internal/relay | service | LIVE | internal/relay | Phone pairing and the E2EE relay sessions (Happy pattern): pairing token required with constant time compare, unpaired sessions expire and are torn down, replay rejected by counter nonces |
| Gjallarhorn UI: engine/gjallarhorn-ui branch | screen | LIVE | branch engine/gjallarhorn-ui at 9d1cae0 (on origin) | The Flutter mobile companion, held on its own branch: pairing, remote diff approval cards, mid-turn steering for the same sessions the server hosts |
| Agent loop: internal/agent | engine | LIVE | internal/agent | The tool loop: Norse personas (odin, thor, loki, heimdall, in internal/agent/personas.go), steering, compaction, modes, AutoVerify, handoff generation, cross-run memory use |
| Policy gate: internal/policy | module | LIVE | internal/policy | The fail-closed gate every agent action passes: allow, deny, or approval required; approval with no approver denies, and bypass modes never override a deny |
| Huginn-Decider: internal/decision | module | LIVE | internal/decision | Routing decisions for the agent: which model, which path; primary decision model liquid/d1-omni-600m served locally (Ollama), with a heuristic fallback |
| SDLC engine: internal/sdlc | engine | LIVE | internal/sdlc | The 7-phase SDLC (planning, analysis, design, implementation, testing, deployment, maintenance) with Norse personas per phase, agile, waterfall, and devops methodologies, and phase tool enforcement at execution time |
| Epistemic Critic: internal/critic | module | LIVE | internal/critic | Deterministic epistemic audit of drafts against retrieved context; backs the audit CLI, the epistemic audit tool, the run engine gates, and a swarm persona gate |
| Tools: internal/tools | module | LIVE | internal/tools | The tool registry the agent calls: shell, read file, write file, edit block, list directory, todo, question, skill, ingest repo, SDLC, epistemic audit, web audit |
| Provider catalog: internal/config | module | LIVE | internal/config | The 12 built-in providers, 3 subscription tiers (Sovereign, Well Pro, Odin Max), provider auto-discovery, and MIMIR_ environment configuration |
| LLM clients: internal/llm | module | LIVE | internal/llm | The provider clients behind one Provider interface, plus the usage tracker every surface reports through |
| Cortex: internal/cortex | service | LIVE | internal/cortex | Memory and identity: the file store (default) and the SurrealDB store (MIMIR_CORTEX_BACKEND=surreal, sidecar or external server), sessions, OTP auth store, credential resolution with no built-in default password |
| GitIngest and Yggdrasil Mind Map: internal/ingest | module | LIVE | internal/ingest | Ingests a local repo or a GitHub URL (git clone) into the Cortex as a typed, queryable mind map; digest export and LatestMindMap reads |
| Run engine gates: internal/runengine (main) | module | LIVE | internal/runengine/runengine.go | Main's gate and test library: the file jail, the single-run lock (.mimir/run.lock), the argv-only test executor with two-tier testing (targeted in-loop, full suite at the gate), ReviewGate and ShipGate, and RunUnattended, which today has no production caller |
| Engines catalog and updater: internal/engines | module | LIVE | internal/engines | Engine status, manifest build, engine invocation, and updates: SHA-256 manifest comparison plus ed25519 signature verification, ACTIVE since 2026-11-04 with MIMIR_ENGINE_PUBKEY pinned; an unsigned manifest is refused when the key is pinned |
| Engine binaries and manifest: engines/bin | service | LIVE | engines/bin/manifest.json | The precompiled proprietary engines and their signed manifest, under LICENSE-ENGINES.md: tyr-core 0.2.0, mjolnir-compute 0.2.0, alphaflux-core 0.2.0, each for darwin arm64, linux amd64, and windows amd64 |
| Swarm factory: internal/swarm | engine | LIVE | internal/swarm | The multi-agent task factory (Odin, Thor, Loki, Heimdall pipeline in isolated worktrees); tasks stop at AwaitingMerge and merge only through Factory.ApproveMerge; Thor blocks honestly when no provider is configured |
| Plugin API draft: internal/plugins | module | LIVE | internal/plugins/api.go | A Go-native plugin API contract, present in the tree and imported by nothing: an unwired draft, inert today, documented in its own package comment as superseded by the planned Wasm plugin host |
| Marketing site: site | page | LIVE | site/index.html | The static public site for Mimir (index and app mockup) with its own Dockerfile; no code connection to the Go binary |
| Engine publish script: scripts | module | LIVE | scripts/publish-engines.ps1 | The release helper that rebuilds the engine binaries from their source branches in throwaway worktrees and refreshes engines/bin/manifest.json |
| Run engine: scaffold branch | engine | LIVE | branch run-engine-scaffold at 7c4ff32 (2026-10-21), internal/runengine on that branch | The unattended run engine, held on its own branch and NOT on origin (local checkout only): RunContext and run-spec frontmatter limits, the shell-less argv jail, the Executor seam, the pluggable verify registry, atomic patches, the run lock, the checklist done-gate, and the review and ship CLI split |
| SurrealDB | external | LIVE | external | The database backend for the persistent Cortex in docker installs and surreal-backend local installs |
| Git forges | external | LIVE | external | GitHub, GitLab, Bitbucket, Gitea and Forgejo: where user repositories live; GitHub URLs are ingestable today |
| Model provider APIs | external | LIVE | external | The 12 catalog providers' endpoints, plus local Ollama serving the decision model and local providers |
| Plan approval gate | module | PLANNED | TODO.md T-6.4 | The run graph node where the human approves the plan and execution is unleashed |
| 10-stage run graph | engine | PLANNED | TODO.md T-6.4 | Discover, contract, repo, research, architecture, build, verify, review, ship, handover as an inspectable graph with typed done gates |
| Verify gate | module | PLANNED | TODO.md T-6.4 | The run graph hard human gate at verify; the scaffold branch already carries a code-level exit-code verify gate and checklist done-gate (TODO T-0.14, done) |
| Ship gate | module | PLANNED | TODO.md T-6.4 | The run graph hard human gate at ship; today's partial forms are the run engine ShipGate checklist in internal/runengine and the swarm AwaitingMerge stop, neither of which is the run graph gate |
| RunUnattended production caller | module | PLANNED | TODO.md T-2.2 | A production caller that starts unattended runs through internal/runengine RunUnattended; today the only caller is its own test |
| Mind map ingestion pipeline to spec | module | PLANNED | TODO.md T-6.3 | Parse, entity and relation extraction, deterministic dedupe, merge with contradiction flagging, and an integration receipt, to the GOAL spec |
| Forge OAuth registration | service | PLANNED | TODO.md T-6.1 | OAuth apps for the git forges so users register their own accounts; stalled on credentials arriving through the secure card |
| Wasm plugin host | engine | PLANNED | TODO.md T-5.5 | The planned plugin host (any language to Wasm, capability granted, sandboxed, hot swappable) that supersedes the internal/plugins draft |
| Engine source branches (engine/tyr-rust, engine/mjolnir-mojo) | module | ABSENT | engines/bin/manifest.json (the claim) | Named as the source branches for tyr-core and mjolnir-compute in the manifest, and assumed by scripts/publish-engines.ps1, but neither branch exists on origin or in the local checkouts as of 2026-11-18; both engines are Go builds today (rust port planned, mojo port planned) |
