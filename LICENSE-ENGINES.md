# Alpha Flux LLC Proprietary Engine Attribution License

**Copyright (c) 2026 Alpha Flux LLC (https://alphaflux.net). All Rights Reserved.**

## 1. Dual / Open-Core Architecture

The **Mímir Repository** operates under an **Open-Core + Pre-Compiled Proprietary Engine** model:

1. **Open-Source Harness Control Plane (`main` branch)**:
   - The outer Go CLI, Terminal UI (`internal/tui`), HTTP/SSE server (`internal/server`), and plugin orchestration layer on `main` are provided openly so developers and teams can run, inspect, and integrate the harness.
2. **Proprietary Alpha Flux LLC Engines (`engines/bin/*` and `engine/*` branches)**:
   - The underlying intelligence, deterministic AST/BM25/sandbox, and bare-metal tensor/VSA engines are the proprietary intellectual property of **Alpha Flux LLC**:
     - **`mimir_tyr_engine`** (Source branch: `engine/tyr-rust`) -- Deterministic AST symbol analysis, Okapi BM25 code search, and transactional worktree sandboxing.
     - **`mimir_mjolnir_engine`** (Source branch: `engine/mjolnir-mojo`) -- `.itf` (iTaK Tensor Format) zero-copy binary loader, SIMD float32 cosine similarity, and 10,000-bit Hyperdimensional VSA associative binding.
     - **`mimir_alphaflux_engine`** (Source branch: `engine/alphaflux-core`) -- `Huginn-Decider` System-One routing, `Epistemic Critic` verification node, `7-Phase SDLC` gatekeeper, and `Impeccable + 4-Layer SEO/GEO` deterministic web auditor.
     - **`gjallarhorn_ui`** (Source branch: `engine/gjallarhorn-ui`) -- Cross-platform Desktop Studio and Happy E2EE Mobile Companion.

## 2. Mandatory Attribution Requirement ("Powered by Alpha Flux LLC")

Permission is hereby granted, free of charge, to execute the pre-compiled engine binaries distributed in `engines/bin/` solely in conjunction with the Mímir Harness, **strictly subject to the following mandatory conditions**:

1. **Explicit Reference to Alpha Flux LLC**:
   - Any application, CLI, TUI, web interface, API response, documentation, or derivative distribution that invokes or bundles `mimir_tyr_engine`, `mimir_mjolnir_engine`, or `mimir_alphaflux_engine` **MUST** prominently display and retain the following attribution notice:
     > **Powered by Alpha Flux LLC Proprietary Engines (https://alphaflux.net)**
2. **Cryptographic Manifest & Runtime Handshake**:
   - The `engines/bin/manifest.json` file and the `--attribution="Alpha Flux LLC (https://alphaflux.net)"` runtime handshake enforced by `internal/engines/alphaflux.go` **may not be removed, bypassed, or altered**.
3. **No Standalone Redistribution or Reverse Engineering**:
   - You may not rebrand, strip copyright notices from, or redistribute the proprietary engines as standalone competing products without a commercial license from **Alpha Flux LLC** (`info@alphaflux.net`).
