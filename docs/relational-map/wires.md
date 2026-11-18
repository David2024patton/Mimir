# Relational map: wires

Every connection between the nodes in [nodes.md](nodes.md), traced both
ways: the wire, and the backlink (the reverse relationship, named).
Every wire on main was verified on 2026-11-18 against an actual import,
call site, route registration, or command dispatch in the code at
f9fc0a8. Wires into PLANNED nodes describe the planned shape and are
marked (planned). Wires the old docs/MAP.md claimed that the code does
not evidence were dropped; the drops are listed in
[README.md](README.md), verification notes.

| From | To | Kind | Backlink |
| :--- | :--- | :--- | :--- |
| Operator | CLI: cmd/mimir | calls | CLI prints results, gate reports, and run artifacts back to the Operator |
| Operator | TUI: internal/tui | navigates | TUI renders session state back to the Operator |
| Operator | Server and Web UI: internal/server | navigates | Server serves the Web UI and session state back |
| Operator | Marketing site: site | navigates | Site presents Mimir back to the Operator (static, no return path) |
| Operator | Plan approval gate | approves (planned) | Plan approval gate unleashes the run for the Operator |
| Operator | Verify gate | approves (planned) | Verify gate escalates failures back to the Operator |
| Operator | Ship gate | approves (planned) | Ship gate holds the release until the Operator approves |
| CLI: cmd/mimir | TUI: internal/tui | launches | TUI returns control and errors to the CLI process |
| CLI: cmd/mimir | Server and Web UI: internal/server | launches (serve, --web) | Server reports serve errors back to the CLI |
| CLI: cmd/mimir | Agent loop: internal/agent | calls | Agent returns run results and traces to the CLI |
| CLI: cmd/mimir | GitIngest and Yggdrasil Mind Map: internal/ingest | calls (ingest command) | Ingest reports neurons saved and the digest path to the CLI |
| CLI: cmd/mimir | Epistemic Critic: internal/critic | calls (audit command) | Critic returns the audit report; a failed audit exits nonzero |
| CLI: cmd/mimir | Engines catalog and updater: internal/engines | calls (engines and update commands) | Updater reports status, plan, and applied updates to the CLI |
| CLI: cmd/mimir | Swarm factory: internal/swarm | calls (swarm commands) | Factory lists task state back; approve merges through the CLI |
| CLI: cmd/mimir | Run engine gates: internal/runengine (main) | calls (review and ship commands) | Gates return a GateReport; a failed gate exits nonzero |
| CLI: cmd/mimir | Huginn-Decider: internal/decision | calls (setup command) | Decider reports the local decision model state to the CLI |
| CLI: cmd/mimir | Relay: internal/relay | calls (pair command, through the server pairing endpoint) | Server returns the pairing payload and QR for the CLI to print |
| CLI: cmd/mimir | Cortex: internal/cortex | hosts (builds the stores in buildCortex) | Cortex returns the store, session store, and auth store every surface shares |
| CLI: cmd/mimir | Tools: internal/tools | hosts (builds the registry with tools.Default) | Tools return the shared registry the CLI hands to the agent, swarm, and gates |
| CLI: cmd/mimir | Provider catalog: internal/config | calls (providers command, AutoDiscoverProvider) | Catalog returns tiers, providers, and the discovered endpoint to the CLI |
| CLI: cmd/mimir | LLM clients: internal/llm | calls (provider construction, usage tracker) | LLM clients return the provider handle and usage totals to the CLI |
| TUI: internal/tui | Agent loop: internal/agent | calls | Agent streams session events back to the TUI |
| TUI: internal/tui | Run engine gates: internal/runengine (main) | calls (review and ship gates) | Gates return GateReports the TUI renders |
| TUI: internal/tui | Swarm factory: internal/swarm | calls (factory panel) | Factory returns task state to the TUI |
| TUI: internal/tui | GitIngest and Yggdrasil Mind Map: internal/ingest | calls (ingest.Run, LatestMindMap) | Ingest returns run results and the latest map summary to the TUI |
| TUI: internal/tui | Tools: internal/tools | calls | Tools return results into the TUI session |
| TUI: internal/tui | Cortex: internal/cortex | reads and writes | Cortex returns memory and session state to the TUI |
| TUI: internal/tui | Provider catalog: internal/config | calls | Catalog returns provider state to the TUI |
| TUI: internal/tui | LLM clients: internal/llm | reads (usage tracker) | LLM clients return usage totals to the TUI |
| Server and Web UI: internal/server | Agent loop: internal/agent | calls | Agent streams session and steer results back to the Server |
| Server and Web UI: internal/server | Relay: internal/relay | hosts (relay.NewManager, /api/pair routes) | Relay returns pairing sessions and E2EE stream state to the Server |
| Server and Web UI: internal/server | Cortex: internal/cortex | reads and writes | Cortex returns memory, sessions, and auth state to the Server |
| Server and Web UI: internal/server | Engines catalog and updater: internal/engines | reads (LoadManifest, attribution) | Engines return the manifest and attribution notice to the Server pages |
| Server and Web UI: internal/server | Swarm factory: internal/swarm | calls (swarm pages and APIs) | Factory returns task state to the Server |
| Server and Web UI: internal/server | LLM clients: internal/llm | reads (usage tracker) | LLM clients return usage totals to the Server |
| Gjallarhorn UI: engine/gjallarhorn-ui branch | Relay: internal/relay | calls (pairing, mobile stream, steer) | Relay delivers pairing state and session events back to the phone |
| Agent loop: internal/agent | Policy gate: internal/policy | calls (every action) | Policy gate returns allow, deny, or approval-required verdicts |
| Agent loop: internal/agent | Huginn-Decider: internal/decision | calls | Decider returns the routing choice |
| Agent loop: internal/agent | SDLC engine: internal/sdlc | calls | SDLC returns phase state and gate rules |
| Agent loop: internal/agent | Tools: internal/tools | calls | Tools return results into the loop |
| Agent loop: internal/agent | Cortex: internal/cortex | reads and writes | Cortex returns recalled memory and session state |
| Agent loop: internal/agent | LLM clients: internal/llm | calls | LLM clients return model responses and usage |
| Tools: internal/tools | Cortex: internal/cortex | reads and writes | Cortex returns stored state to the tools (todo store, ingest target) |
| Tools: internal/tools | Epistemic Critic: internal/critic | calls (epistemic audit tool) | Critic returns the audit report into the loop |
| Tools: internal/tools | GitIngest and Yggdrasil Mind Map: internal/ingest | calls (ingest repo tool) | Ingest returns what was integrated |
| Tools: internal/tools | SDLC engine: internal/sdlc | calls (SDLC tool) | SDLC returns board and phase state to the tool |
| GitIngest and Yggdrasil Mind Map: internal/ingest | Cortex: internal/cortex | reads and writes | Cortex serves map queries and LatestMindMap reads back to Ingest callers |
| Cortex: internal/cortex | SurrealDB | reads and writes | SurrealDB returns stored records to the Cortex |
| Run engine gates: internal/runengine (main) | Epistemic Critic: internal/critic | calls (gate epistemic check) | Critic returns the AuditReport embedded in the GateReport |
| Run engine gates: internal/runengine (main) | Huginn-Decider: internal/decision | calls | Decider returns routing for gate runs |
| Run engine gates: internal/runengine (main) | SDLC engine: internal/sdlc | reads (phase state) | SDLC returns the phase verdicts the gates enforce |
| Run engine gates: internal/runengine (main) | Tools: internal/tools | reads (todo checklist store) | Tools return the checklist state the gates check |
| Swarm factory: internal/swarm | Agent loop: internal/agent | calls (worker passes) | Agent returns task results to the Factory |
| Swarm factory: internal/swarm | Epistemic Critic: internal/critic | calls (persona gate check) | Critic returns the deterministic audit verdict to the Factory |
| Swarm factory: internal/swarm | Tools: internal/tools | calls (per-worktree registry, web audit) | Tools return results and audit reports to the Factory |
| Swarm factory: internal/swarm | Run engine gates: internal/runengine (main) | reads (run Limits type) | Run engine gates return the Limits shape the Factory config carries |
| Swarm factory: internal/swarm | Ship gate | emits (AwaitingMerge stop) (planned) | Ship gate approval returns through Factory.ApproveMerge |
| Engines catalog and updater: internal/engines | Engine binaries and manifest: engines/bin | reads and writes (verified swaps only) | Manifest reports SHA-256 and version per engine back |
| Engine publish script: scripts | Engine binaries and manifest: engines/bin | writes (rebuild and manifest refresh) | Manifest carries the new hashes back to the repo |
| Engine publish script: scripts | Engine source branches (engine/tyr-rust, engine/mjolnir-mojo) | reads (build inputs) | none: the branches are ABSENT, so this wire cannot be served today |
| Run engine: scaffold branch | Policy gate: internal/policy | calls (argv jail verdicts) | Policy gate returns verdicts to the scaffold run engine |
| Run engine: scaffold branch | LLM clients: internal/llm | calls (loop provider) | LLM clients return model responses to the scaffold loop |
| RunUnattended production caller | Run engine gates: internal/runengine (main) | calls (planned) | Run engine gates return the UnattendedRunSummary |
| Plan approval gate | 10-stage run graph | emits (unleash) (planned) | Run graph returns stage state to the gate |
| 10-stage run graph | SDLC engine: internal/sdlc | calls (build stage runs the inside-run phases) (planned) | SDLC returns phase verdicts to the graph |
| 10-stage run graph | Verify gate | emits (planned) | Verify gate returns pass or escalation to the graph |
| 10-stage run graph | Ship gate | emits (planned) | Ship gate returns the release decision to the graph |
| 10-stage run graph | GitIngest and Yggdrasil Mind Map: internal/ingest | reads (planned) | Mind Map returns knowledge for the research and architecture stages |
| Mind map ingestion pipeline to spec | GitIngest and Yggdrasil Mind Map: internal/ingest | writes (extends) (planned) | Mind Map returns integration receipts to the pipeline |
| Forge OAuth registration | Git forges | calls (OAuth) (planned) | Forges return account and repository access |
| Forge OAuth registration | Server and Web UI: internal/server | writes (session repo link) (planned) | Server returns session state for the linked repo |
| Git forges | GitIngest and Yggdrasil Mind Map: internal/ingest | reads (repo ingest by URL, git clone) | Ingest returns the integration result |
| Wasm plugin host | Agent loop: internal/agent | hosts (plugins) (planned) | Agent returns plugin results to the host |
| Wasm plugin host | Policy gate: internal/policy | calls (capability checks) (planned) | Policy gate returns verdicts to the host |
| Plugin API draft: internal/plugins | Wasm plugin host | superseded by (documented in internal/plugins/api.go) | Wasm plugin host replaces this contract when it lands (planned) |
| LLM clients: internal/llm | Model provider APIs | calls | Providers return model responses, errors, and usage |
| Huginn-Decider: internal/decision | Model provider APIs | calls (local decision model over HTTP) | The local model endpoint returns routing decisions |
