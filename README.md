# Yuyan

An AI pet app: a virtual companion with memory, growth, and conversation.

This project has shifted from an instant messaging app to an AI companion pet product. The existing Go + Flutter + MySQL + Redis foundation remains useful, while account, messaging, and WebSocket capabilities are treated as infrastructure. The product loop is now human-to-AI-pet companionship rather than human-to-human chat.

See [docs/LLM_DEV_GUIDE.md](docs/LLM_DEV_GUIDE.md) for product boundaries, architecture rules, AI pipeline guidance, payments, analytics, and testing requirements. Milestones: [docs/MILESTONES.md](docs/MILESTONES.md).

[中文文档](README-zh.md)

## Phase 1 MVP

Phase 1 goal: a user can sign up, create an AI pet, talk to it through text, see it remember meaningful facts, and watch its state and growth evolve. The foundation should also leave room for voice and subscription payments.

Priorities:

- Account and user profile
- Pet creation, state, and growth system
- Text conversation through an AI Gateway
- Short-term and long-term memory
- Basic analytics, usage, and cost dashboards
- Server-side entitlement model for subscriptions

Friend relationships, multi-user chat, group chat, and rich media messages are no longer Phase 1 product priorities.

## Local Development

### Prerequisites

- Go 1.25+ (or use the portable SDK in `.tools/go`)
- Flutter 3.47+ (or use the portable SDK in `.tools/flutter`)
- Docker (Docker Desktop) for MySQL / Redis / integration tests

### Start Everything

```bash
cd deploy
docker compose up -d --build
curl http://127.0.0.1:8080/healthz
```

Database-only profile:

```bash
cd deploy
docker compose --profile db-only up -d
```

### Common Tasks

| Task | Linux/macOS | Windows (PowerShell) |
|---|---|---|
| Run server locally | `make run` | `powershell -File scripts/make.ps1 -Target run` |
| Server tests | `make test-server` | `powershell -File scripts/make.ps1 -Target test-server` |
| Lint | `make lint` | `powershell -File scripts/make.ps1 -Target lint` |
| Protocol schema checks | `make gen-protocol` | `powershell -File scripts/make.ps1 -Target gen-protocol` |

### Flutter App

```bash
cd apps/app
flutter pub get
dart run build_runner build --delete-conflicting-outputs
flutter test
flutter run
```
