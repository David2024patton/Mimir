# Engines and Updates

Mimir ships as an open harness plus pre-compiled engines owned by **Alpha Flux LLC** (https://alphaflux.net). See [LICENSE-ENGINES.md](../LICENSE-ENGINES.md).

## Layout

| Where | What |
| :--- | :--- |
| `main` | Open harness source + compiled engine binaries in `engines/bin/` + `manifest.json` (SHA-256 + version per engine) |
| `engine/tyr-rust` | Tyr-Core source (symbol outline, BM25) |
| `engine/mjolnir-mojo` | Mjolnir-Compute source (.itf header, cosine, 10,048-bit VSA) |
| `engine/alphaflux-core` | Alpha Flux Intelligence engine entrypoint (decider, critic, web audit) |
| `engine/gjallarhorn-ui` | Flutter Studio scaffold (not compiled yet, no binary shipped) |

## For users: staying current

```powershell
mimir update --check   # show what changed, install nothing
mimir update           # git fast-forward main, rebuild bin/mimir, swap in changed engines
```

- Engines are compared by **SHA-256**, so any rebuilt engine is detected even if its version string did not change.
- Every download is verified against the published manifest before anything is replaced. A failed check leaves the old binary untouched.
- A remote manifest that does not name Alpha Flux LLC as owner is refused.
- `MIMIR_UPDATE_BASE` points the updater at a fork or mirror (default: this repo's `main` on raw.githubusercontent.com).
- The harness update uses `git pull --ff-only`, so it never merges or overwrites local edits. If it refuses, commit or stash first.

## For maintainers: shipping an engine change

1. Commit the change on its engine branch (for example `engine/mjolnir-mojo`) and bump `engineVersion` in its `main.go`.
2. From `main`, run:
   ```powershell
   .\scripts\publish-engines.ps1 -Commit
   ```
   This builds each engine in a temporary git worktree (engine source never lands on `main`), copies the binaries into `engines/bin/`, rewrites `manifest.json`, and pushes.
3. Users get it on their next `mimir update`.

Use `-Only tyr|mjolnir|alphaflux` to rebuild a single engine.

## Honest status

- All three shipped engines are currently compiled from **Go**. The Rust (`crates/mimir-rs`) and Mojo (`mojo/mimir_mojo`) sources on the engine branches are ports in progress and are not what the binaries run yet.
- Binaries are Windows x64 only for now. Linux and macOS builds need a cross-compile matrix in the publish script.
- The attribution handshake is a license notice enforced in code, not DRM. It makes the terms explicit and visible; it cannot stop a determined person from patching a binary.
