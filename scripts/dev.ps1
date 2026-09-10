# Local one-click startup: backend (Go, :8089) + frontend (Flutter web-server, :8081)
# Usage (repo root):
#   powershell -ExecutionPolicy Bypass -File scripts/dev.ps1
#
# Notes:
# - Uses portable Go / Flutter SDK under .tools; no global install needed.
# - Backend loads apps/server/config.test.yaml via -env test (port 8089, dev-only CORS on).
# - Frontend fixed on 8081; open http://localhost:8081 in browser; API base from config.test.json -> :8089.
# - Logs: _backend_run.log / _flutter_run.log (processes are independent; closing this window won't stop them).
# - Stop: end the go / flutter(dart) processes in Task Manager, or close their windows.

$ErrorActionPreference = 'Stop'
$root = Resolve-Path "$PSScriptRoot/.."
$tools = Join-Path $root '.tools'

$goBin = Join-Path $tools 'go/bin/go.exe'
$flutterBin = Join-Path $tools 'flutter/bin/flutter.bat'

if (-not (Test-Path $goBin)) { Write-Error "portable Go not found: $goBin"; exit 1 }
if (-not (Test-Path $flutterBin)) { Write-Error "portable Flutter not found: $flutterBin"; exit 1 }

# Portable SDK into PATH; point GOROOT at .tools/go (avoid system GOROOT pointing to a deleted path).
$env:PATH = "$tools\go\bin;$tools\flutter\bin;$env:PATH"
$env:GOROOT = Join-Path $tools 'go'

# Backend
Write-Host '[dev] starting backend  go run ./cmd/server -env test  -> :8089' -ForegroundColor Cyan
Start-Process -FilePath $goBin -ArgumentList 'run ./cmd/server -env test' -WorkingDirectory (Join-Path $root 'apps/server') -RedirectStandardOutput (Join-Path $root '_backend_run.log') -RedirectStandardError (Join-Path $root '_backend_err.log') -NoNewWindow

# Frontend: web-server on fixed port. Most reliable when chrome cannot launch on this machine.
Write-Host '[dev] starting frontend flutter run -d web-server --web-port 8081  -> :8081' -ForegroundColor Cyan
Start-Process -FilePath $flutterBin -ArgumentList 'run -d web-server --web-port 8081' -WorkingDirectory (Join-Path $root 'apps/app') -RedirectStandardOutput (Join-Path $root '_flutter_run.log') -RedirectStandardError (Join-Path $root '_flutter_err.log') -NoNewWindow

Start-Sleep -Seconds 2
Write-Host '[dev] done. open http://localhost:8081  (backend :8089)' -ForegroundColor Green
Write-Host '[dev] logs: _backend_run.log / _flutter_run.log' -ForegroundColor DarkGray
