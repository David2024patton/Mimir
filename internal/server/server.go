package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/David2024patton/Mimir/internal/config"
)

// Serve starts the daemon's HTTP + WebSocket API and serves the GUI (E12.1).
func Serve(cfg *config.Config) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// TODO(E12.1): REST handlers (projects/sessions/cortex/config/auth/tasks/sandboxes)
	// and a WebSocket hub for token/tool/lifecycle/status events.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Mimir GUI: http://localhost:%d\n", cfg.Port)
	return http.ListenAndServe(addr, mux)
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Mimir</title>
  <style>
    :root { color-scheme: dark; }
    body { margin: 0; font-family: system-ui, sans-serif; background: #0e1116; color: #e6e6e6; }
    .topbar { display:flex; align-items:center; gap:12px; padding:10px 16px; background:#161b22; border-bottom:1px solid #30363d; }
    .brand { font-weight: 700; }
    .layout { display: grid; grid-template-columns: 220px 1fr 360px 56px; height: calc(100vh - 45px); }
    .nav { background:#161b22; border-right:1px solid #30363d; padding:12px; }
    .nav div { padding:6px 4px; color:#8b949e; }
    .workspace { padding:16px; overflow:auto; }
    .chat { background:#161b22; border-left:1px solid #30363d; padding:12px; }
    .rail { background:#0d1117; border-left:1px solid #30363d; }
    h1 { font-size: 18px; }
    .muted { color:#8b949e; font-size: 13px; }
  </style>
</head>
<body>
  <div class="topbar">
    <span class="brand">Mimir</span>
    <span class="muted">project &#9662;</span>
    <span class="muted">agent &#9662;</span>
    <span style="flex:1"></span>
    <span class="muted" id="status">connecting&hellip;</span>
  </div>
  <div class="layout">
    <nav class="nav">
      <div>Projects</div><div>Cortex</div><div>Sessions</div>
      <div>Agents</div><div>Skills</div><div>Settings</div>
    </nav>
    <main class="workspace">
      <h1>The agent that remembers.</h1>
      <p class="muted">Walking skeleton: daemon + GUI shell (5-region layout).
      The Cortex, agent loop, and tools land next.</p>
    </main>
    <aside class="chat">
      <strong>Chat</strong>
      <p class="muted">Streaming + tool calls + approvals + to-do list (E12.3).</p>
    </aside>
    <div class="rail"></div>
  </div>
  <script>
    fetch('/api/health').then(r=>r.json()).then(d=>{
      document.getElementById('status').textContent='daemon: '+d.status;
    });
  </script>
</body>
</html>`
