# scripts/make.ps1 — Makefile 的 Windows PowerShell 等价物。
# 用法：powershell -File scripts/make.ps1 -Target <run|test-server|lint|gen-protocol>
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("run", "test-server", "lint", "gen-protocol")]
    [string]$Target
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

# 便携工具链（存在则启用）
$goDir = Join-Path $root ".tools\go\bin"
if (Test-Path (Join-Path $goDir "go.exe")) {
    $env:GOROOT = Join-Path $root ".tools\go"
    $env:PATH = "$goDir;$env:PATH"
}
if (-not $env:GOPROXY) { $env:GOPROXY = "https://goproxy.cn,direct" }

function Run-In($dir, $cmd, $cmdArgs) {
    Push-Location (Join-Path $root $dir)
    try {
        & $cmd @cmdArgs
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally { Pop-Location }
}

switch ($Target) {
    "run" {
        # 前置：docker compose --profile db-only up -d
        Run-In "apps\server" "go" @("run", "./cmd/server")
    }
    "test-server" {
        Run-In "apps\server" "go" @("test", "./...")
    }
    "lint" {
        $lint = Get-Command golangci-lint -ErrorAction SilentlyContinue
        if ($lint) {
            Run-In "." "golangci-lint" @("run", "./apps/server/...", "./packages/protocol/...")
        }
        else {
            Write-Host "golangci-lint 未安装，回退 go vet"
            Run-In "apps\server" "go" @("vet", "./...")
        }
    }
    "gen-protocol" {
        Run-In "packages\protocol" "go" @("test", "./...")
    }
}
