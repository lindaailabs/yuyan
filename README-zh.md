# Yuyan（语燕）

AI 宠物应用：一只有记忆、会成长、能对话的虚拟宠物。

当前项目从“即时通讯 App”调整为“AI 陪伴型宠物 App”。既有 Go + Flutter + MySQL + Redis 技术底座继续保留，账号、消息、WebSocket 等能力作为基础设施复用；产品主线从“人与人单聊闭环”改为“人与 AI 宠物陪伴闭环”。

开发规范见 [docs/LLM_DEV_GUIDE.md](docs/LLM_DEV_GUIDE.md)（产品边界、技术栈、架构分层、AI 链路、支付与数据要求）。里程碑见 [docs/MILESTONES.md](docs/MILESTONES.md)。

[English](README.md)

## 一期 MVP

一期目标：用户可以注册登录，创建自己的 AI 宠物，通过文本对话建立关系；宠物能记住关键信息，拥有可解释的状态和成长变化，并具备后续接入语音与订阅支付的技术地基。

一期优先级：

- 账号与用户资料
- 宠物创建、状态与成长系统
- 文本对话与 AI Gateway
- 长短期记忆管理
- 基础埋点、用量、成本看板
- 订阅权益模型的服务端地基

暂不把好友关系、多人聊天、群聊、多媒体消息作为主线。

## 本地开发

### 前置条件

- Go 1.25+（或使用 `.tools/go` 便携 SDK）
- Flutter 3.47+（或使用 `.tools/flutter` 便携 SDK）
- Docker（Docker Desktop），用于 MySQL / Redis / 集成测试

### 一键启动（Docker 全栈）

```bash
cd deploy
docker compose up -d --build
curl http://127.0.0.1:8080/healthz
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
dart run build_runner build --delete-conflicting-outputs
flutter test
flutter run
```
