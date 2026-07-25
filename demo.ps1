# Mímir live demo - watch the agent actually DO things on your machine.
# Runs against LOCAL Ollama (free, no key). Takes ~1 minute (the model thinks between steps).
# Honest by design: it shows exactly what the model did (the trace), and verifies any
# file changes on disk. If the model varies and skips a step, you see that too.
$ErrorActionPreference = "Continue"
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$env:MIMIR_BASE_URL = if ($env:MIMIR_BASE_URL) { $env:MIMIR_BASE_URL } else { "http://localhost:11434/v1" }
$env:MIMIR_MODEL    = if ($env:MIMIR_MODEL)    { $env:MIMIR_MODEL }    else { "qwen3:8b" }
$exe = Join-Path $here "bin\mimir.exe"
$note = Join-Path $here "demo-note.txt"

function Step($n, $title) { Write-Host ""; Write-Host "=== [$n] $title ===" -ForegroundColor Cyan }
function Note($t) { Write-Host "    $t" -ForegroundColor DarkGray }

if (-not (Test-Path $exe)) { Write-Host "ERROR: $exe not found. Run: go build -o bin\mimir.exe ./cmd/mimir" -ForegroundColor Red; exit 1 }

Write-Host "Mímir live demo  (provider base: $env:MIMIR_BASE_URL  model: $env:MIMIR_MODEL)" -ForegroundColor Green
Note "If Ollama isn't running, start it first. This uses your local machine only."

Step 1 "STATUS - what Mímir has wired up"
& $exe

Step 2 "TOOL CALL - the model runs a real shell command (bash)"
Note "Asking the model to use bash. Watch [step 1] tool=bash below."
& $exe trace "You have a bash tool. Use bash to run exactly this command: echo HELLO_FROM_MIMIR  Then reply with the single word DONE."

Step 3 "WRITE + READ - the model creates a real file on your disk"
Remove-Item $note -Force -ErrorAction SilentlyContinue
& $exe trace "Use write_file to create a file named exactly demo-note.txt (in the current directory) whose content is the single line: MIMIR_WAS_HERE  Then use read_file to read demo-note.txt back. Then reply DONE."
if (Test-Path $note) {
    Write-Host "    [VERIFIED on disk] demo-note.txt exists. Contents:" -ForegroundColor Green
    Get-Content $note | ForEach-Object { Write-Host "        $_" -ForegroundColor Green }
} else {
    Note "[observed] the model did not create demo-note.txt this run (model variance). The trace above shows what it actually did."
}

Step 4 "EDIT - the model surgically edits that file (edit_block)"
if (Test-Path $note) {
    & $exe trace "Use edit_block on demo-note.txt to replace the text MIMIR_WAS_HERE with MIMIR_EDITED_THIS. Then reply DONE."
    $c = Get-Content $note -Raw
    if ($c -match "MIMIR_EDITED_THIS") { Write-Host "    [VERIFIED on disk] the edit landed: file now contains MIMIR_EDITED_THIS." -ForegroundColor Green }
    else { Note "[observed] the edit did not land this run (model variance). File currently: $c" }
} else {
    Note "skipped (no demo-note.txt from step 3)."
}
Note "Note: edit_block is exact-match. If a model mangles non-ASCII chars while copying, the error now echoes the real file content so the model can self-correct and retry."

Remove-Item $note -Force -ErrorAction SilentlyContinue
Write-Host ""
Write-Host "=== demo complete ===" -ForegroundColor Green
Note "Honest note: memory (the Cortex) is currently in-memory, so it resets between separate"
Note "runs of mimir.exe. Within ONE trace/chat call the recall->act->remember loop is live."
Note "Cross-run memory (remembering between sessions) is the next build step: swap in SurrealDB."
Note "Ad-hoc use anytime:  test-mimir `"your prompt`"   or   test-mimir trace `"...`""
