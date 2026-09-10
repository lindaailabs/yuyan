# Yuyan（语燕）

AI 宠物应用：一只有记忆、会成长、能对话的虚拟宠物。

当前项目从"即时通讯 App"调整为"AI 陪伴型宠物 App"。既有 Go + Flutter + MySQL + Redis 技术底座继续保留，账号、消息、WebSocket 等能力作为基础设施复用；产品主线从"人与人单聊闭环"改为"人与 AI 宠物陪伴闭环"。

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

## 后端环境配置

后端配置按环境拆成独立文件，位于 `apps/server/`：

- `config.yaml` — dev 兜底（`APP_ENV=dev` 或 `config.{env}.yaml` 缺失时回退）
- `config.test.yaml` — 本地开发（用 `config.test.yaml.example` 复制而来，**已被 .gitignore 忽略，不提交**）
- `config.prod.yaml` — 生产（密钥用 `${ENV_VAR}` 引用，无明文）

加载优先级：**同名环境变量 > 配置文件字段 > 内置默认**；路径选择：`CONFIG_FILE` 显式指定 > `config.{APP_ENV}.yaml` > `config.yaml`。

## 本地调试（开发联调）

目标：本机起存储 + 后端 + Flutter（Chrome）跑通全流程，快速验证 UI / 接口。

### 前置条件

- Go 1.25+（或使用 `.tools/go` 便携 SDK）
- Flutter 3.47+（或使用 `.tools/flutter` 便携 SDK）
- Docker（Docker Desktop），用于本地 MySQL / Redis

> 若 `go` 报 `cannot find GOROOT directory: D:\db\go`（系统 `GOROOT` 指向了已不存在的路径），把 GOROOT 指到便携 SDK 即可：
> ````powershell
> $env:GOROOT = "d:/workspace/yuyan/.tools/go"
> $env:PATH = "d:/workspace/yuyan/.tools/go/bin;" + $env:PATH
> ````

### 0. 一键启动（脚本，推荐）

不想分别敲命令，可用脚本一次起后端 + 前端（自动用 `.tools` 便携 SDK，进程独立常驻）：

```bash
powershell -ExecutionPolicy Bypass -File scripts/dev.ps1
```

- 后端：`go run ./cmd/server -env test` → `:8089`（含 dev-only CORS）
- 前端：`flutter run -d web-server --web-port 8081` → 浏览器开 `http://localhost:8081`
- 日志：`_backend_run.log` / `_flutter_run.log`；停止就结束 `go` / `flutter(dart)` 进程
- 本机 `flutter run -d chrome` 拉不起浏览器时，web-server 模式最稳（手动开 URL 即可）

> 前端固定 8081，避免每次随机端口；后端测试端口为 8089（避开 docker server 容器占用的 8080）。

### 1. 起存储（MySQL / Redis）

compose 未配置 profile，请用服务名只起数据库，避免 server 容器占用 8080 与本地 `go run` 冲突：

```bash
cd deploy
docker compose up -d mysql redis
```

> 想一键起整套（含后端容器）也可：`docker compose up -d --build`，但会占用 8080。

### 2. 起后端

```bash
cd apps/server
cp config.test.yaml.example config.test.yaml   # 编辑 ai_base_url / ai_api_key / ai_model
go run ./cmd/server -env test                  # 命令行参数指定环境，自动加载 config.test.yaml（无需设环境变量）
```

- 指定环境的两种方式（效果等价，都不需要设环境变量）：
  - 按环境名：`go run ./cmd/server -env test` → 加载 `config.test.yaml`
  - 直接指定文件：`go run ./cmd/server -config config.test.yaml`（`-config` 优先级最高）
- 不接真实模型：`config.test.yaml` 里 `ai_provider` 设为 `mock`（默认），后端不触网、确定性回复。
- 接真实模型：`ai_provider: openai`，填 `ai_base_url`(url) / `ai_api_key`(key) / `ai_model`(name)。
- 健康检查：`curl http://127.0.0.1:8089/healthz`
- 模型 key 默认写在 `config.test.yaml` 里（文件已被忽略不提交）；也可用环境变量 `AI_API_KEY` 覆盖。

### 3. Flutter（Chrome 走流程）

接口前缀按环境放在配置文件里（对应后端 `config.test.yaml` / `config.prod.yaml`），集中在 `apps/app/lib/core/app_config.dart`：
- `apps/app/assets/config.test.json`（本地调试，默认 `http://localhost:8089/api/v1`）
- `apps/app/assets/config.prod.json`（正式，填你的服务器地址）

取值优先级：`--dart-define=API_BASE_URL` > `assets/config.<APP_ENV>.json` > `assets/config.json` > 内置默认。**本地调试零参数**，直接读 `config.test.json`：

```bash
cd apps/app
flutter pub get
dart run build_runner build --delete-conflicting-outputs
flutter run -d chrome
```

- `APP_ENV` 默认 `test`，所以 `flutter run` 自动加载 `config.test.json`，不用带任何参数。
- 想临时覆盖某个地址：`flutter run -d chrome --dart-define=API_BASE_URL=http://localhost:8089/api/v1`。

### 4. 登录（手机号 + 密码）

已改为手机号 + 密码，注册与登录分离（不再有图形验证码）：

- 首次使用先在 App 内走「注册」建号（`POST /api/v1/auth/register`，手机号 + 密码）；旧验证码账号无密码，需重新注册。
- 登录（`POST /api/v1/auth/login`，手机号 + 密码）成功后返回双 token，自动跳主页。
- 本地联调后端开启了 dev-only CORS，浏览器跨域调用正常；生产镜像不含 CORS（原生 app 也不受浏览器 CORS 限制）。

## 正式打包与部署

目标：产出可安装的应用（Android / iOS）与对应的生产后端镜像。

### 后端（生产镜像）

```bash
# 构建上下文为仓库根（需要 packages/protocol 本地 replace）
docker build -f apps/server/Dockerfile -t yuyan-server .
```

镜像内置 `config.prod.yaml`，运行时指定 prod 并注入密钥（密钥经 `${ENV_VAR}` 引用，不落明文）：

```bash
APP_ENV=prod \
MYSQL_DSN=... REDIS_ADDR=... JWT_SECRET=... \
AI_BASE_URL=... AI_API_KEY=... AI_MODEL=... \
docker run -p 8080:8080 yuyan-server
```

### Flutter（打包成 app）

把 `apps/app/assets/config.prod.json` 的 `apiBaseUrl` 改成正式服务器地址，再用短参数指定环境打包（与本地调试只差一个 `APP_ENV`）：

```bash
# config.prod.json 已填好服务器地址后：
flutter build apk --dart-define=APP_ENV=prod        # Android
flutter build ipa --dart-define=APP_ENV=prod        # iOS
# 或临时覆盖（不改文件）：
flutter build apk --dart-define=API_BASE_URL=http://<服务器地址>:8080/api/v1
```

- `APP_ENV` 默认 `test`；打包正式环境务必传 `--dart-define=APP_ENV=prod`，否则会打进 test 配置。
- 模拟器：`http://10.0.2.2:8080/api/v1`（Android）/ `http://localhost:8080/api/v1`（iOS Simulator）。
- 真机同局域网调试：填开发机局域网 IP，如 `http://192.168.1.10:8080/api/v1`。
- 原生 app 不受 CORS 限制；若发布 Web 版则需另配生产 CORS。

### 常用任务

| 任务 | Linux/macOS | Windows (PowerShell) |
|---|---|---|
| 本地跑服务端 | `make run` | `powershell -File scripts/make.ps1 -Target run` |
| 服务端测试 | `make test-server` | `powershell -File scripts/make.ps1 -Target test-server` |
| Lint | `make lint` | `powershell -File scripts/make.ps1 -Target lint` |
| 协议 Schema 校验 | `make gen-protocol` | `powershell -File scripts/make.ps1 -Target gen-protocol` |
