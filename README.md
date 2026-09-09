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

## Backend Environment Config

Backend config is split per environment, under `apps/server/`:

- `config.yaml` — dev fallback (used when `APP_ENV=dev` or `config.{env}.yaml` is missing)
- `config.test.yaml` — local development (copied from `config.test.yaml.example`; git-ignored, not committed)
- `config.prod.yaml` — production (secrets referenced via `${ENV_VAR}`, no plaintext)

Precedence: **env var > config file field > built-in default**; path: `CONFIG_FILE` > `config.{APP_ENV}.yaml` > `config.yaml`.

## Local Debugging

Goal: run storage + backend + Flutter (Chrome) locally to walk the whole flow.

### Prerequisites

- Go 1.25+ (or use the portable SDK in `.tools/go`)
- Flutter 3.47+ (or use the portable SDK in `.tools/flutter`)
- Docker (Docker Desktop) for local MySQL / Redis

> If `go` fails with `cannot find GOROOT directory: D:\db\go` (the system `GOROOT` points to a missing path), point GOROOT at the portable SDK:
> ````powershell
> $env:GOROOT = "d:/workspace/yuyan/.tools/go"
> $env:PATH = "d:/workspace/yuyan/.tools/go/bin;" + $env:PATH
> ````

### 1. Start storage (MySQL / Redis)

The compose file has no profiles, so name the services explicitly to avoid the `server` container occupying port 8080:

```bash
cd deploy
docker compose up -d mysql redis
```

> One-click full stack is also available (`docker compose up -d --build`), but it occupies 8080.

### 2. Start the backend

```bash
cd apps/server
cp config.test.yaml.example config.test.yaml   # edit ai_base_url / ai_api_key / ai_model
go run ./cmd/server -env test                  # pass env as a CLI flag, auto-loads config.test.yaml (no env var needed)
```

- Two ways to pick the environment (equivalent, neither needs an env var):
  - By env name: `go run ./cmd/server -env test` → loads `config.test.yaml`
  - Explicit file: `go run ./cmd/server -config config.test.yaml` (`-config` wins)
- No real model: set `ai_provider: mock` (default) in `config.test.yaml` — deterministic, no network.
- Real model: `ai_provider: openai`, fill `ai_base_url` (url) / `ai_api_key` (key) / `ai_model` (name).
- Health check: `curl http://127.0.0.1:8089/healthz`

### 3. Flutter (Chrome walkthrough)

The base URL prefix is managed per environment in config files (mirroring the backend `config.test.yaml` / `config.prod.yaml`), centered in `apps/app/lib/core/app_config.dart`:
- `apps/app/assets/config.test.json` (local debug, defaults to `http://localhost:8089/api/v1`)
- `apps/app/assets/config.prod.json` (production; fill in your server address)

Precedence: `--dart-define=API_BASE_URL` > `assets/config.<APP_ENV>.json` > `assets/config.json` > built-in default. **Local debugging needs zero flags** — it reads `config.test.json` by default:

```bash
cd apps/app
flutter pub get
dart run build_runner build --delete-conflicting-outputs
flutter run -d chrome
```

- `APP_ENV` defaults to `test`, so `flutter run` automatically loads `config.test.json` with no flags.
- Temporary override: `flutter run -d chrome --dart-define=API_BASE_URL=http://localhost:8089/api/v1`.

### 4. Login

In the app, tap "Send code" → the code is stored in Redis; fetch it locally:

```bash
docker exec yuyan-redis-1 redis-cli get "sms:code:138xxxx"
```

> Local debugging enables dev-only CORS so Chrome cross-origin calls work; the production image has no CORS (native apps are not subject to browser CORS either).

## Production Build & Deploy

Goal: produce an installable app (Android / iOS) and the corresponding production backend image.

### Backend (production image)

```bash
# Build context is the repo root (needs the packages/protocol local replace)
docker build -f apps/server/Dockerfile -t yuyan-server .
```

The image bundles `config.prod.yaml`; set prod at runtime and inject secrets (referenced via `${ENV_VAR}`, no plaintext):

```bash
APP_ENV=prod \
MYSQL_DSN=... REDIS_ADDR=... JWT_SECRET=... \
AI_BASE_URL=... AI_API_KEY=... AI_MODEL=... \
docker run -p 8080:8080 yuyan-server
```

### Flutter (package as app)

Set `apiBaseUrl` in `apps/app/assets/config.prod.json` to the production server address, then package with a short env flag (only one `APP_ENV` differs from local debugging):

```bash
# After setting the server address in config.prod.json:
flutter build apk --dart-define=APP_ENV=prod        # Android
flutter build ipa --dart-define=APP_ENV=prod        # iOS
# Or override without editing the file:
flutter build apk --dart-define=API_BASE_URL=http://<server-address>:8080/api/v1
```

- `APP_ENV` defaults to `test`; always pass `--dart-define=APP_ENV=prod` for production builds, otherwise the test config gets bundled.
- Emulator default: `http://10.0.2.2:8089/api/v1` (Android) / `http://localhost:8089/api/v1` (iOS Simulator).
- Real device on the same LAN: use the dev machine's LAN IP, e.g. `http://192.168.1.10:8080/api/v1`.
- Native apps are not bound by CORS; a web release would need production CORS configured separately.

### Common Tasks

| Task | Linux/macOS | Windows (PowerShell) |
|---|---|---|
| Run server locally | `make run` | `powershell -File scripts/make.ps1 -Target run` |
| Server tests | `make test-server` | `powershell -File scripts/make.ps1 -Target test-server` |
| Lint | `make lint` | `powershell -File scripts/make.ps1 -Target lint` |
| Protocol schema checks | `make gen-protocol` | `powershell -File scripts/make.ps1 -Target gen-protocol` |
