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

$targets = @(
    @{ OS = "windows"; Arch = "amd64"; Suffix = "_windows_amd64.exe" },
    @{ OS = "linux";   Arch = "amd64"; Suffix = "_linux_amd64" },
    @{ OS = "darwin";  Arch = "arm64"; Suffix = "_darwin_arm64" }
)

foreach ($e in $engines) {
    if ($Only -and $e.Key -ne $Only) { continue }
    $wt = Join-Path $wtBase $e.Key
    if (Test-Path $wt) { git worktree remove --force $wt | Out-Null }
    Write-Host "checking out $($e.Branch) into worktree"
    git worktree add --detach $wt $e.Branch | Out-Null
    try {
        Push-Location $wt
        # 1. Compile host binary
        Write-Host "  -> building host binary $($e.Out)"
        go build -trimpath -ldflags="-s -w" -o (Join-Path $binDir $e.Out) $e.Pkg
        if ($LASTEXITCODE -ne 0) { throw "host build failed for $($e.Branch)" }

        # 2. Compile cross-platform matrix
        $baseName = [System.IO.Path]::GetFileNameWithoutExtension($e.Out)
        foreach ($t in $targets) {
            $crossOut = "$baseName$($t.Suffix)"
            Write-Host "  -> building cross-platform binary $crossOut ($($t.OS)/$($t.Arch))"
            $env:GOOS = $t.OS
            $env:GOARCH = $t.Arch
            $env:CGO_ENABLED = "0"
            go build -trimpath -ldflags="-s -w" -o (Join-Path $binDir $crossOut) $e.Pkg
            if ($LASTEXITCODE -ne 0) { throw "cross-build failed for $($e.Branch) ($($t.OS)/$($t.Arch))" }
        }
        $env:GOOS = ""
        $env:GOARCH = ""
    } finally {
        $env:GOOS = ""
        $env:GOARCH = ""
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
