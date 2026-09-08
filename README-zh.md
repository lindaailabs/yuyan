# Yuyan（语燕）

即时通讯 App。一期 MVP：单聊纯文本消息完整闭环（Flutter + Go + MySQL 8 + Redis）。

开发规范见 [docs/LLM_DEV_GUIDE.md](docs/LLM_DEV_GUIDE.md)（技术栈、架构分层、协议、红线清单）。里程碑：[docs/MILESTONES.md](docs/MILESTONES.md)。

[English](README.md)

## 本地开发

### 前置条件

- Go 1.25+（或使用 `.tools/go` 便携 SDK）
- Flutter 3.47+（或使用 `.tools/flutter` 便携 SDK）
- Docker（Docker Desktop），用于 MySQL / Redis / 集成测试

### 一键启动（Docker 全栈）

```bash
cd deploy
docker compose up -d --build
curl http://127.0.0.1:8080/healthz   # {"code":0,"msg":"ok","data":null}
```

仅数据库（本地直跑 Go 服务时）：

```bash
cd deploy
docker compose --profile db-only up -d
```

### 常用任务

| 任务 | Linux/macOS | Windows (PowerShell) |
|---|---|---|
| 本地跑服务端 | `make run` | `powershell -File scripts/make.ps1 -Target run` |
| 服务端测试 | `make test-server` | `powershell -File scripts/make.ps1 -Target test-server` |
| Lint | `make lint` | `powershell -File scripts/make.ps1 -Target lint` |
| 协议 Schema 校验 | `make gen-protocol` | `powershell -File scripts/make.ps1 -Target gen-protocol` |

### Flutter 客户端

```bash
cd apps/app
flutter pub get
dart run build_runner build --delete-conflicting-outputs  # drift 代码生成
flutter test
flutter run  # 需要 Android/iOS 设备或模拟器
```
