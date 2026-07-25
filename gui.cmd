@echo off
setlocal
title Mimirmind GUI
cd /d "%~dp0"

REM === Provider (local Ollama by default; override these to point elsewhere) ===
if not defined MIMIR_BASE_URL set "MIMIR_BASE_URL=http://localhost:11434/v1"
if not defined MIMIR_MODEL set "MIMIR_MODEL=qwen3:8b"

REM === Cortex: SurrealDB brain if the binary is here, else file-backed ===
if exist "bin\surreal.exe" (
  if not defined MIMIR_CORTEX_BACKEND set "MIMIR_CORTEX_BACKEND=surreal"
  if not defined MIMIR_SURREAL_ADDR set "MIMIR_SURREAL_ADDR=127.0.0.1:8088"
)

REM === Build if the binary is missing ===
if not exist "bin\mimir.exe" (
  echo Building mimir...
  go build -o bin\mimir.exe ./cmd/mimir
  if errorlevel 1 ( echo Build failed. & pause & exit /b 1 )
)

echo.
echo   Mimir - the agent that remembers
echo   --------------------------------------------------
echo   GUI      http://localhost:8420/
echo   Model    %MIMIR_MODEL%   (%MIMIR_BASE_URL%)
echo   Cortex   %MIMIR_CORTEX_BACKEND%
echo   --------------------------------------------------
echo   Your browser will open when the server is ready.
echo   Close this window (or Ctrl+C) to stop Mimir.
echo.

REM Poll /health in the background and open the browser once the server is up.
start "" /b powershell -NoProfile -Command "for ($i=0;$i -lt 40;$i++){ Start-Sleep -Milliseconds 500; try { $null=Invoke-RestMethod -Uri 'http://localhost:8420/health' -TimeoutSec 2; Start-Process 'http://localhost:8420/'; break } catch {} }"

bin\mimir.exe serve --addr :8420
