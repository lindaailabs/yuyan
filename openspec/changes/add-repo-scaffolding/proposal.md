## Why

一期开发没有可运行的工程基座：`apps/server`、`apps/app`、`packages/protocol`、`deploy/` 均不存在，W2~W6 的所有功能任务都无处落脚。需要先按 LLM_DEV_GUIDE.md §3 的 monorepo 结构搭出三端骨架与一键环境（MILESTONES.md W1），让后续每个功能任务在固定分层内增量开发。

## What Changes

- 新增 `apps/server`：Go 单体分层骨架（`cmd/server/main.go` + `internal/{api,ws,service,repo,model,pkg}`），含配置加载、`pkg/errcode`、slog 日志封装（HTTP 中间件注入 trace_id）、GORM+MySQL 连接、migration 机制（时间戳前缀 SQL）、`GET /healthz` 健康检查
- 新增 `apps/app`：Flutter 工程骨架，按 §3.4 建立 `core/data/features/shared` 目录，Riverpod 路由壳 + 主题、drift 数据库初始化、Dio/WS client 与 repository 接口占位
- 新增 `packages/protocol`：7 个 WS 命令字 + REST 统一响应包裹的 JSON Schema（§5.2/§5.3 全集）及校验用例
- 新增 `deploy/docker-compose.yml`：MySQL 8（utf8mb4）+ Redis 7 + server，一键拉起
- 新增根 `Makefile`：`run` / `test-server` / `lint` / `gen-protocol` 目标

不修改任何现有代码（仓库目前仅有文档）；无对外契约变更。

## Capabilities

### New Capabilities

- `server-skeleton`: Go 单体服务骨架——进程启动、配置加载、结构化日志（trace_id）、MySQL/Redis 连接、migration 自动执行、健康检查端点
- `app-skeleton`: Flutter 客户端骨架——路由壳、主题、drift 初始化、网络层与 repository 接口占位（无业务功能）
- `protocol-schema`: WS 帧与 REST 包裹的 JSON Schema 定义及校验用例（v1 全集，仅定义不实现传输）

### Modified Capabilities

（无——`openspec/specs/` 目前为空，本变更为首批能力）

## Impact

- **代码**：全部为新增目录（apps/、packages/、deploy/、Makefile），不触碰 docs/ 既有内容
- **依赖（需 PR 声明的新第三方引入）**：
  - Go：gin、gorm、go-redis、golang-jwt（W2 用，W1 仅 go.mod 占位引入 gin/gorm/go-redis）、golang-migrate（或等价 migration 方案）
  - Flutter：flutter_riverpod、drift、dio、web_socket_channel、go_router
  - 校验用例用 Go 标准库 + `santhosh-tekuri/jsonschema`（或等价，design 定案）
- **环境**：要求本机具备 Go 1.22+、Flutter 3.2x、Docker；Windows 为主要开发环境（Makefile 需兼容或提供 .ps1 等价脚本，design 决策）
- **风险**：低——纯骨架无业务逻辑；主要不确定性是 Windows 下 Make/工具链兼容性
