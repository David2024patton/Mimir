# publish-engines.ps1 -- rebuild Alpha Flux LLC engine binaries from their source
# branches and refresh engines/bin/manifest.json on main.
#
# Usage (from the repo root, on main):
#   .\scripts\publish-engines.ps1                 # build all engines
#   .\scripts\publish-engines.ps1 -Only tyr       # build one engine
#   .\scripts\publish-engines.ps1 -Commit         # also commit + push main
#
# Each engine is built in a throwaway git worktree so main's working tree never
# contains engine source. Users then pick up the new binaries with `mimir update`.

param(
    [string]$Only = "",
    [switch]$Commit
)

$ErrorActionPreference = "Stop"
$root = (git rev-parse --show-toplevel).Trim()
Set-Location $root

$branch = (git branch --show-current).Trim()
if ($branch -ne "main") { throw "Run this from main (currently on '$branch')." }

$engines = @(
    @{ Key = "tyr";       Branch = "engine/tyr-rust";       Pkg = "./cmd/tyr-engine";       Out = "mimir_tyr_engine.exe" },
    @{ Key = "mjolnir";   Branch = "engine/mjolnir-mojo";   Pkg = "./cmd/mjolnir-engine";   Out = "mimir_mjolnir_engine.exe" },
    @{ Key = "alphaflux"; Branch = "engine/alphaflux-core"; Pkg = "./cmd/alphaflux-engine"; Out = "mimir_alphaflux_engine.exe" }
)

$binDir = Join-Path $root "engines\bin"
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
$wtBase = Join-Path ([System.IO.Path]::GetTempPath()) "mimir-engine-build"

foreach ($e in $engines) {
    if ($Only -and $e.Key -ne $Only) { continue }
    $wt = Join-Path $wtBase $e.Key
    if (Test-Path $wt) { git worktree remove --force $wt | Out-Null }
    Write-Host "building $($e.Out) from $($e.Branch)"
    git worktree add --detach $wt $e.Branch | Out-Null
    try {
        Push-Location $wt
        go build -trimpath -ldflags="-s -w" -o (Join-Path $binDir $e.Out) $e.Pkg
        if ($LASTEXITCODE -ne 0) { throw "build failed for $($e.Branch)" }
    } finally {
        Pop-Location
        git worktree remove --force $wt | Out-Null
    }
}

go build -o bin\mimir.exe ./cmd/mimir
if ($LASTEXITCODE -ne 0) { throw "harness build failed" }
.\bin\mimir.exe engines publish
if ($LASTEXITCODE -ne 0) { throw "manifest publish failed" }

if ($Commit) {
    git add engines/bin
    git commit -m "chore(engines): publish rebuilt Alpha Flux LLC engine binaries"
    git push origin main
}
