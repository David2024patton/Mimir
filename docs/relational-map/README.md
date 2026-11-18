# Relational map: Mimir

This folder is Mimir's relational diagram: every feature and module as
a node, every connection as a wire, built from the code. It exists so
anyone opening this repo can see the whole system at once and stay on
track: what is live, what is planned, and what the paper claims but the
code does not have.

## How to read it

- [nodes.md](nodes.md): the node table. Node, kind, state, source,
  what it owns. States: LIVE (in code, works as described), PLANNED
  (named work with a TODO id, no code yet), ABSENT (claimed on paper,
  not in the code; the disagreement is named on the node).
- [wires.md](wires.md): the wire table. From, to, kind, and the
  backlink (the reverse relationship, named).
- [diagram.md](diagram.md): the same nodes and wires as a Mermaid
  flowchart.
- [../MAP.md](../MAP.md): the entry point, with the overview and the
  main flows in words.

nodes.md and wires.md are canonical. The diagram and MAP.md follow
them; if they ever disagree, the tables win and the drawing is the
drift to fix.

## The update rule, in short

The full rule lives in [../RULES.md](../RULES.md). Short form: this map
is updated in the same push cycle as any feature build or any change
to any part of the project, and a feature is not fully landed until
the diagram matches it. The map is built from the code: code beats
paper, and disagreements are named on the map, never silently smoothed
over. No secrets, passwords, or tokens in these files, ever.

## Verification notes 2026-11-18

This set was built on 2026-11-18 by verifying the previous docs/MAP.md
(a single self-contained map) against the code at f9fc0a8: the package
list and per-package imports from go list, the subcommand dispatch in
cmd/mimir/main.go, the route registrations in internal/server,
engines/bin/manifest.json, the TODO.md ids, the run-engine-scaffold
branch in its local checkout, and git ls-remote for the remote
branches. Drift found and fixed:

1. Engines signing was stale. The old map said ed25519 verification
   was implemented but not yet active, and flow 4 repeated it. Since
   commit f9fc0a8 (2026-11-04) the manifest carries a signature,
   MIMIR_ENGINE_PUBKEY is pinned in .env.example, both compose files,
   the Dockerfile, and docs/RUNBOOK.md, and
   internal/engines/manifest_signature_test.go gates it. Fixed in the
   nodes, the wires, and flow 4.
2. The scaffold run engine was cited at 9e0b5ad. Its true tip is
   7c4ff32 (2026-10-21), five branch-unique commits: 22d2d3e, 9e0b5ad,
   664fa97, 63c170f, 7c4ff32. Fixed.
3. Main's own internal/runengine package had no node at all; only the
   scaffold engine was mapped. Added as its own LIVE node (file jail,
   run lock, two-tier test executor, ReviewGate and ShipGate,
   RunUnattended), with the scaffold engine kept as a separate
   branch-held node.
4. The CLI node listed 7 subcommands (setup, ingest, providers, sdlc,
   audit, update, trace). The real dispatch in cmd/mimir/main.go also
   has version, help, audit-web, engines, swarm, review, ship,
   handoff, pair, serve, and chat, plus the TUI launch and flags.
   Fixed; the review, ship, swarm, pair, serve, and handoff wires are
   new.
5. internal/plugins had no node; only the planned Wasm host appeared.
   Added the draft as a node in its true state: present in the tree,
   imported by nothing, inert, superseded by the Wasm host per its own
   package comment.
6. The old map marked planned nodes with (planned) in the name and had
   no state or source columns. Every node now carries State and Source.
7. Unmapped surfaces found and added: the site/ marketing site, the
   scripts/publish-engines.ps1 release helper, and the three engines
   named in the manifest (tyr-core, mjolnir-compute, alphaflux-core,
   each 0.2.0 with darwin arm64, linux amd64, and windows amd64
   builds). The Gjallarhorn branch tip of 9d1cae0 was verified by
   ls-remote and was accurate.
8. One wire was reversed. The old map drew the Relay calling the
   Server; the code has the Server constructing the relay manager
   (relay.NewManager in internal/server/server.go) and serving its
   routes. The wire now reads: Server hosts Relay.
9. Dropped wire: Agent loop to Epistemic Critic as a direct call.
   internal/agent does not import internal/critic. The agent reaches
   the critic through the Tools registry (the epistemic audit tool),
   and that wire is mapped instead.
10. Dropped wire: Engine binaries to Agent loop (hosts engine
    capabilities). Nothing in internal/agent imports
    internal/engines; engines are invoked from the CLI through
    internal/engines (the engines run command). No wire.
11. Retargeted wire: the RunUnattended production caller pointed at
    the scaffold engine. RunUnattended lives in main's
    internal/runengine, so the planned wire now targets the main gates
    node (TODO T-2.2).
12. Dropped wire: scaffold run engine to Agent loop. The scaffold
    runengine imports internal/llm and internal/policy only; its loop
    talks to providers through the LLM clients directly. That wire is
    mapped instead.
13. Missing wires added from the import graph: CLI to TUI, CLI to
    Server, CLI to Relay (the pair command), CLI to Cortex, CLI to
    Tools, CLI to Provider catalog, CLI to LLM clients, CLI to
    Decider (setup), TUI to the run engine gates, TUI to Swarm, TUI
    to Ingest, TUI to Provider catalog, TUI to LLM clients, Server to
    Engines, Server to Swarm, Server to LLM clients, the run engine
    gates to Critic, Decider, SDLC, and Tools, Swarm to Critic, Swarm
    to Tools, Swarm to the run engine gates (the Limits type), Tools
    to Critic, Tools to SDLC, and Decider to Model provider APIs.
14. New ABSENT node: the engine source branches engine/tyr-rust and
    engine/mjolnir-mojo are named as source branches in
    engines/bin/manifest.json and assumed by
    scripts/publish-engines.ps1, but neither exists on origin or in
    the local checkouts (ls-remote, 2026-11-18). Both engines are Go
    builds today; the manifest itself says rust port planned and mojo
    port planned.
15. The old map said the scaffold engine is "not on main". The
    stronger truth: the run-engine-scaffold branch is not on origin at
    all; it exists only in its local checkout. Named on the node.
16. The provider catalog was described as 12 plus. The catalog holds
    exactly 12 providers and 3 subscription tiers; the node states
    the exact counts.
17. The server bind: the old map said the Web UI on :8420. The default
    bind in cmd/mimir/main.go is 127.0.0.1:8420; the node states it.
18. Ship gate nuance: the old map treated the swarm AwaitingMerge stop
    as the ship gate itself. The planned run graph Ship gate is its
    own node (TODO T-6.4); today's partial forms (the run engine
    ShipGate checklist and the swarm stop) are named as partial forms
    on the node, not as the gate.

Disagreements named on the map itself, in nodes.md: the ABSENT engine
source branches, the plugin draft that exists but is inert, the
scaffold branch held only in a local checkout, and the Ship gate's
partial forms.
