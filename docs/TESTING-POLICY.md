# Mímir Testing Policy

**Standing rule (non-negotiable):** after every build and after every feature, run a
**live test** against a real model - not only unit tests. Unit tests prove the code is
correct in isolation; a live test proves the whole stack (provider wire format, streaming,
the recall -> infer -> remember loop) actually works end-to-end against a real backend.

## The three tiers

1. **Unit tests** - `go test ./...`. Always run, always green, offline. Includes a fake
   HTTP server that speaks the OpenAI-compatible wire format, so the provider + loop are
   tested without network or keys.
2. **Live test** - `go test ./internal/agent -run TestLiveRun -v` with the env vars below.
   Hits a real provider. **Skipped automatically** when `MIMIR_LIVE_BASE_URL` is unset, so
   it never breaks offline/CI runs.
3. **CLI smoke** - `go run ./cmd/mimir "prompt"` against a real provider; eyeball the reply.

## Running the live test

Local, free, no key (Ollama):

```powershell
$env:MIMIR_LIVE_BASE_URL = "http://localhost:11434/v1"
$env:MIMIR_LIVE_MODEL    = "qwen2.5:0.5b"   # or qwen3:8b, gemma4:e2b, etc.
go test ./internal/agent -run TestLiveRun -v
```

Cloud (any OpenAI-compatible endpoint):

```powershell
$env:MIMIR_LIVE_BASE_URL = "https://api.openai.com/v1"   # or OpenRouter, Groq, the Qwen compatible-mode, Mímir's own gateway
$env:MIMIR_LIVE_API_KEY  = $env:SOME_KEY                 # never echo the value
$env:MIMIR_LIVE_MODEL    = "gpt-5-nano"
go test ./internal/agent -run TestLiveRun -v
```

## Discipline

- A feature is **not done** until both `go test ./...` (green) and a live test (pass) have
  run. The live test does not need to be in CI (keys/cost), but it must be run by the
  developer/agent after the change.
- The live test must never print secrets: keys come from the environment and are never
  logged. Our provider's error path prints the server response body, never the request
  Authorization header.
- If the live test fails on a credential/401, that is a *credential* problem, not a code
  defect - but the network path still counts as "exercised." Fall back to a local Ollama
  model to get a clean pass that isolates the code from the credential.
- `go vet ./...` must also be clean after every change.

## Tool loop (when you add/change tools or the loop)

- **Deterministic (offline):** `go test ./internal/agent -run TestRunToolLoop -v` - a fake
  model asks to run `bash`, the loop runs the real tool, feeds the result back, and the
  model's final reply is returned. `TestRunUnknownTool` proves a bad tool name degrades
  gracefully. These never need network or keys.
- **Live:** `MIMIR_LIVE_TOOL=1 MIMIR_LIVE_BASE_URL=http://localhost:11434/v1
  MIMIR_LIVE_MODEL=qwen3:8b go test ./internal/agent -run TestLiveToolCall -v` - a real
  model drives a real tool end-to-end. Needs a tool-capable model (qwen3:8b works; tiny
  models usually won't tool-call, so this is gated off by default).
- **CLI trace (watch it act):** `mimir trace "your prompt"` prints each tool call
  (name + args + result) then the final reply - the human-observable proof and a useful
  transparency/observability feature (F60).

A tool feature is not done until `TestRunToolLoop` passes (offline) AND a live tool test
or `mimir trace` shows a real model calling a real tool.
