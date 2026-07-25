@echo off
setlocal
rem ============================================================
rem  Mímir launcher - type anything, no env-var setup needed.
rem  Defaults to LOCAL Ollama (free, no API key, your machine).
rem
rem  Examples:
rem    test-mimir                       show status (tools + provider)
rem    test-mimir "explain this repo"   one chat turn
rem    test-mimir trace "list the files here with bash, then reply DONE"
rem
rem  To use a CLOUD provider instead (e.g. your Qwen token plan), set these
rem  BEFORE running this file:
rem    set MIMIR_BASE_URL=https://token-plan.ap-southeast-1.maas.aliyuncs.com/compatible-mode/v1
rem    set MIMIR_API_KEY=your-sk-sp-...-key
rem    set MIMIR_MODEL=qwen3.6-flash
rem ============================================================
if not defined MIMIR_BASE_URL set MIMIR_BASE_URL=http://localhost:11434/v1
if not defined MIMIR_MODEL    set MIMIR_MODEL=qwen3:8b
"%~dp0bin\mimir.exe" %*
endlocal
