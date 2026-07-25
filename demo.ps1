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
$cortex = Join-Path $here ".mimir\cortex.json"
$env:MIMIR_CORTEX = $cortex

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

Step 5 "MEMORY ACROSS RUNS - run 1 plants a code on disk, run 2 recalls it"
$cortex = Join-Path $here ".mimir\cortex.json"
Remove-Item $cortex -Force -ErrorAction SilentlyContinue
Note "run 1: plant the launch code KIWI55 (stored to .mimir\cortex.json)."
& $exe trace "Remember this exactly: the launch code is KIWI55. Reply OK." | Out-Null
Note "run 2: ask for the code; the [memory] line proves it was recalled from disk."
$out2 = & $exe trace "What is the launch code KIWI55? Answer from memory." 2>&1 | Out-String
if ($out2 -match '\[memory\][^\r\n]*KIWI55') {
    Write-Host "    [VERIFIED] run 2 recalled run 1's memory from disk (cross-run memory works)." -ForegroundColor Green
} else {
    Note "[observed] recall line not matched this run. run 2 output:"
    ($out2 -split "`n") | Where-Object { $_ -match 'memory|KIWI|step|---|DONE' } | ForEach-Object { Note $_ }
}
Step 6 "LIVE BOOK-SPINE UI - a browser talks to the brain over HTTP (E70)"
$addr = ":8420"
Remove-Item $cortex -Force -ErrorAction SilentlyContinue
$proc = Start-Process -FilePath $exe -ArgumentList @("serve","--addr",$addr) -WorkingDirectory $here -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 2
try {
    Invoke-RestMethod -Uri "http://localhost:8420/health" -TimeoutSec 8 | Out-Null
    Note "GET  /health  -> ok"
    $page = [string](Invoke-RestMethod -Uri "http://localhost:8420/" -TimeoutSec 8)
    $isBookSpine = ($page -match 'CORTEX') -and ($page -match 'spine')
    Note ("GET  /        -> book-spine UI served: " + $isBookSpine)
    $chat = Invoke-RestMethod -Uri "http://localhost:8420/chat" -Method Post -ContentType "application/json" -Body '{"prompt":"Reply with the single word WIRED and nothing else.","mode":"chat"}' -TimeoutSec 180
    Note ("POST /chat  -> reply: " + (($chat.reply -replace "`r`n"," ") -replace "`n"," "))
    $mem = Invoke-RestMethod -Uri "http://localhost:8420/memory" -TimeoutSec 8
    Note ("GET  /memory -> " + $mem.count + " neuron(s) persisted")
    if ([string]::IsNullOrWhiteSpace($chat.reply) -eq $false -and [int]$mem.count -ge 1 -and $isBookSpine) {
        Write-Host "    [VERIFIED] the book-spine UI is live end-to-end: browser -> server -> model -> Cortex." -ForegroundColor Green
        Note "Open it in a browser: http://localhost:8420/  (spines open/close + drag; chat + Cortex are live)"
    } else {
        Note "[observed] the wire responded but the soft assertion was not met (model variance)."
    }
} catch {
    Note ("GUI wire error: " + $_.Exception.Message)
} finally {
    Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
}
Remove-Item $note -Force -ErrorAction SilentlyContinue
Remove-Item $cortex -Force -ErrorAction SilentlyContinue
Write-Host ""
Write-Host "=== demo complete ===" -ForegroundColor Green
Note "Memory persists across runs in .mimir/cortex.json (file-backed Cortex; SurrealDB upgrades to vector+graph)."
Note "The GUI wire (browser <-> server <-> brain) is live at http://localhost:8420/ via: mimir serve"
Note "The polished book-spine UI (E70) is the live face: vertical spines, open/close + drag panels, live chat + Cortex."
Note "Ad-hoc use anytime:  test-mimir `"your prompt`"   or   test-mimir trace `"...`""
