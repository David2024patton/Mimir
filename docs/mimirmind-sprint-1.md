# Mímir - Sprint 1 Plan

**BMAD Artifact - Phase 8 (Amelia, Senior Engineer)**
Date: 2026-07-23 | Goal: **Walking skeleton** - prove the end-to-end path.
Inputs: Core stories, Readiness review (conditions C1-C3).

---

## Sprint Goal

A user can: install one binary -> configure a provider (or Ollama) -> open a minimal
GUI -> send a prompt -> see the agent stream a response and call one tool -> see a
to-do item. This proves daemon <-> GUI <-> provider <-> tool <-> SurrealDB end-to-end,
then we widen.

---

## Sprint Backlog (~44 pts)

| Story | Title | Pts | Notes |
|---|---|---|---|
| E1.1 | Go module + single-binary scaffold | 5 | go.mod, cobra, build matrix, CI |
| E1.2 | Config loader + interpolation | 5 | mimir.json, {env}/{file}, opencode import |
| E1.3 | Credential store (encrypted) | 5 | auth.json, AES-GCM, `mimir auth login` |
| E1.4 | SurrealDB managed sidecar **(Spike C1)** | 8 | auto-start/health/reconnect; schema bootstrap |
| E2.1 | Provider abstraction + OpenAI dialect | 8 | Provider/Dialect, SSE streaming, registry |
| E3.5 | To-do tool | 5 | the `todo` table + todowrite/todoread |
| E11.1 | Plugin SDK + registry **(freeze C2)** | 8 | MimirPluginApi interface + contract |
| E12.1 | Daemon HTTP + WS server (minimal) | 5 | REST /chat + WS token stream (thin) |

**Out this sprint:** Anthropic dialect (E2.2), full agent loop tool-dispatch (E2.3
beyond one tool), policy gate (E4), GUI shell (E12.2 - use a minimal page), TUI (E12.4).

---

## Spikes / de-risking (from the readiness conditions)

- **C1 (SurrealDB sidecar):** timebox 2 days. Deliverable: daemon starts SurrealDB
  (Docker), connects, runs the schema, reconnects after a kill. If Docker is absent,
  fall back to a bundled `surreal` binary.
- **C2 (plugin-SDK freeze):** define `MimirPluginApi` + write the contract test stubs
  this sprint so nothing builds against a moving target.
- **C3 (small-model validation):** stretch - if Ollama + a <=30B model is available,
  run one plan -> to-do round-trip to sanity-check the loop. Full tuning later (E48).

---

## Definition of Done (sprint)

- One binary builds for the dev's OS; `mimir` starts the daemon + serves a minimal page.
- A prompt to an OpenAI-compatible provider (or Ollama) streams back into the page.
- One tool call executes (e.g. read_file) and its result returns to the model.
- A to-do item can be written and read back (persisted in SurrealDB).
- SurrealDB auto-starts and survives a restart.
- CI builds the binary + runs the plugin contract test stub.
- All committed + pushed to the Mimir repo.

---

## Sequence

1. E1.1 (scaffold) -> E1.2 (config) -> E1.3 (creds)  [foundation]
2. E1.4 (SurrealDB spike, parallel with #1 where possible)
3. E2.1 (provider) -> wire a one-shot stream to stdout
4. E3.5 (to-do) + E12.1 (minimal server) -> stream into a minimal web page
5. E11.1 (plugin SDK freeze) + CI

---

## Next sprints (preview)

- **Sprint 2:** full agent loop + tool dispatch (E2.3), shell+fs tools (E3.1/E3.2),
  policy gate (E4.1), Anthropic dialect (E2.2).
- **Sprint 3:** GUI shell (E12.2) + chat panel (E12.3), Cortex store (E6.2),
  auto-capture (E6.3).
- **Sprint 4:** skills (E7), plan mode (E8), subagents (E9), checkpoints (E5).
- **Sprint 5:** hooks (E10), plugin loader (E11.2), MCP (E14), audit (E4.3).
- **Sprint 6:** packaging (E13), TUI (E12.4), compaction (E2.5), doom-loop (E4.4),
  persona (E6.4), edit_block (E3.3), code-exec (E3.4).

After the Core: Tier 2 differentiators (small-model mode E48, governance E44, sandbox
E45, memory engine E43, telemetry E49, research tools E46, etc.).
