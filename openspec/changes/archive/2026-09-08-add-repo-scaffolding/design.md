## Context

仓库现状：仅文档（docs/、openspec/、README），无任何工程代码。本变更建立 W2~W6 所有功能任务的工程地基。主开发环境为 **Windows**（PowerShell），CI/部署目标为 Linux Docker；权威规范 docs/LLM_DEV_GUIDE.md §3（结构/分层）、§5（协议）、§6（编码规范）、§7（测试）。

## Goals / Non-Goals

**Goals:**
- 三端骨架可运行：Go server 能起进程并连上 MySQL/Redis；Flutter app 能在模拟器启动显示壳页面；protocol 包含 v1 全集 Schema 且校验用例全绿
- `docker compose up`（deploy/）一键拉起 MySQL+Redis+server，migration 自动执行
- Makefile 四目标（run/test-server/lint/gen-protocol）+ Windows 等价脚本
- 分层目录就位，后续功能只在既有分层内加代码

**Non-Goals:**
- 任何业务功能（登录、好友、消息——W2+）
- WS 网关实现（仅目录+空 hub 占位，W4 实现）
- CI 流水线（W6 收尾时评估）
- 代码生成器真实实现（gen-protocol 仅做 Schema 校验的占位实现）

## Decisions

### D1. Migration 工具：golang-migrate（iofs 嵌入式）
- **选择**：SQL 文件放 `apps/server/migrations/`（`YYYYMMDDHHMMSS_name.up.sql`/`.down.sql`），go:embed 嵌入二进制，启动时 `migrate.Up()` 自动执行
- **理由**：guide §4 要求时间戳前缀 SQL 文件且禁改历史；嵌入式使容器无需挂载额外文件；自动执行满足 W1 验收"migration 自动执行"
- **备选**：GORM AutoMigrate（否——不产生可评审的 SQL 历史，违反 §4）；shell 脚本手动执行（否——不满足自动执行验收）
- **新依赖声明**：`golang-migrate/migrate/v4`

### D2. 配置加载：环境变量 + 纯标准库
- **选择**：`internal/pkg/config` 定义 Config struct（HTTP 端口、MySQL DSN、Redis Addr、日志级别、Env），`os.Getenv` + 默认值，无第三方库
- **理由**：docker-compose 传 env 天然契合；符合 §6.1"Java 工程师能读懂"的朴素标准
- **备选**：viper（否——文件分层配置一期过度）

### D3. 日志与 trace_id：slog（JSON handler）+ Gin 中间件
- **选择**：`pkg/logger` 初始化 slog；Gin 中间件生成/透传 `X-Trace-Id`（无则 uuid 生成）并注入 `context` 与响应头；handler 内用 `slog.InfoContext` 带 trace_id 输出
- **理由**：§6.1 指定 slog；WS 帧透传 trace_id 留到 W4 在帧解码处接同一 context
- **新依赖声明**：无（uuid 用标准库或 google/uuid，倾向 google/uuid——成熟、单文件，PR 一并声明）

### D4. Windows 兼容：Makefile 为准 + scripts/*.ps1 等价物
- **选择**：Makefile（Linux/macOS/CI 用）+ `scripts/make.ps1 -Target <name>`（Windows 用），目标：run/test-server/lint/gen-protocol
- **理由**：guide §8 完成标准引用 make；主开发机是 Windows，双入口成本仅 4 个薄封装
- **备选**：仅 Makefile（否——Windows 无 make 时验收无法本地执行）；Taskfile/just（否——引入新工具链，违反朴素原则）

### D5. protocol 包形态：独立 Go module + go:embed Schema
- **选择**：`packages/protocol` 为独立 module `github.com/lindaailabs/yuyan/protocol`，JSON Schema 文件放 `schema/ws/*.json`、`schema/rest/*.json`，embed 后暴露 FS；校验用例（合法/非法帧各若干）作为包内 `_test.go`
- **理由**：满足 §7"每个帧类型必须有 JSON Schema 校验用例"；独立 module 双端可引用，为 gen-protocol 生成 Dart/Go 类型留位
- **备选**：纯静态目录无代码（否——无法跑校验用例）；并入 server module（否——App 侧将来也要引用）
- **新依赖声明**：`santhosh-tekuri/jsonschema/v6`（校验用例用；PR 中说明）

### D6. Go 依赖最小集（W1 实际引入）
gin、gorm+mysql driver、go-redis/v9、golang-migrate、google/uuid、santhosh-tekuri/jsonschema；**golang-jwt 留到 W2 再引**（避免未使用依赖）。golangci-lint 配置 `.golangci.yml`（govet/staticcheck/errcheck/revive 基础集）。

### D7. Flutter 骨架形态
- `flutter create --org com.yuyan.app --project-name yuyan_app`；目录按 §3.4 补齐 core/data/features/shared
- 路由用 go_router（登录/主页两个占位路由）；`ProviderScope` 包根；主题收敛 core/theme；drift 建 `AppDatabase`（W1 仅打开+版本 1 空迁移，验证代码生成链路）；Dio 与 WS client 以接口形式放 data/remote，repository 层空实现占位
- **新依赖声明**：flutter_riverpod、go_router、drift(+dev: drift_dev/build_runner)、dio、web_socket_channel
- l10n：`core/l10n/zh.dart` 常量类占位（一期硬编码禁令的落点）

### D8. docker-compose 拓扑
mysql8（utf8mb4 显式 `command` 参数、healthcheck、volume）+ redis7（appendonly）+ server（多阶段 Dockerfile：golang:1.22-alpine 构建 → alpine 运行，depends_on healthy）。DSN/Redis 走 env。本地 `make run` 亦可直连本机 Go 进程 + compose 只起 mysql/redis（profile: `db-only`）。

## Risks / Trade-offs

- [Windows make 缺失] → D4 的 ps1 等价物；验收时两入口各跑一次
- [golang-migrate 嵌入后无法单独 down（线上）] → 一期单机，回滚 = git revert + 新增补偿 migration（符合 §9.2 不改历史）
- [drift build_runner 在 Windows 首次拉依赖慢] → 可接受，一次性成本
- [protocol 独立 module 增加 go.work/replace 维护] → 用 go.work 统一本地解析（apps/server 通过 workspace 引用 protocol，提交的 go.mod 用 replace 指令固定相对路径，CI 无需额外配置）

## Migration Plan

纯新增，无存量迁移：合入即生效；回滚 = 删除新增目录（数据库层面仅一张 `schema_migrations` 表，由 golang-migrate 管理）。

## Open Questions

（无——本变更全部决策已在上述 D1~D8 定案；实现期若 PR 评审推翻某项，更新本文件后实施。）
