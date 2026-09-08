# Yuyan

An instant messaging app. Phase 1 MVP: a complete one-to-one text messaging loop (Flutter + Go + MySQL 8 + Redis).

See [docs/LLM_DEV_GUIDE.md](docs/LLM_DEV_GUIDE.md) for development guidelines (tech stack, architecture layering, protocol, hard rules). Milestones: [docs/MILESTONES.md](docs/MILESTONES.md).

[中文文档](README-zh.md)

## Local Development

### Prerequisites

- Go 1.25+ (or use the portable SDK in `.tools/go`)
- Flutter 3.47+ (or use the portable SDK in `.tools/flutter`)
- Docker (Docker Desktop) for MySQL / Redis / integration tests

### Start everything (Docker)

```bash
cd deploy
docker compose up -d --build
curl http://127.0.0.1:8080/healthz   # {"code":0,"msg":"ok","data":null}
```

Database-only profile (when running the Go server natively):

```bash
cd deploy
docker compose --profile db-only up -d
```

### Common tasks

| Task | Linux/macOS | Windows (PowerShell) |
|---|---|---|
| Run server locally | `make run` | `powershell -File scripts/make.ps1 -Target run` |
| Server tests | `make test-server` | `powershell -File scripts/make.ps1 -Target test-server` |
| Lint | `make lint` | `powershell -File scripts/make.ps1 -Target lint` |
| Protocol schema checks | `make gen-protocol` | `powershell -File scripts/make.ps1 -Target gen-protocol` |

### Flutter app

```bash
cd apps/app
flutter pub get
dart run build_runner build --delete-conflicting-outputs  # drift codegen
flutter test
flutter run  # requires an Android/iOS device or emulator
```
