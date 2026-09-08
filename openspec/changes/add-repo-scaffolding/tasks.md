## 1. Go 服务端骨架（apps/server）

- [ ] 1.1 初始化 Go module（github.com/lindaailabs/yuyan/server，go 1.22），建立 cmd/server 与 internal/{api,ws,service,repo,model,pkg} 目录及包声明占位（含职责注释），.golangci.yml 基础配置
- [ ] 1.2 实现 internal/pkg/config：Config struct + 环境变量加载（HTTP_PORT/MYSQL_DSN/REDIS_ADDR/LOG_LEVEL/APP_ENV，含默认值），main.go 中打印生效配置（脱敏）
- [ ] 1.3 实现 pkg 日志与 trace_id：slog JSON handler 初始化（pkg/logger）、Gin 中间件（X-Trace-Id 生成/透传，google/uuid）、响应头回写与 context 注入
- [ ] 1.4 实现 internal/pkg/errcode：错误码分段（0/1xxx/2xxx/5xxx）、带码 Error 类型、New 构造校验段位、api 层统一响应包裹转换
- [ ] 1.5 接入 GORM（mysql driver，utf8mb4 DSN）与 go-redis v9（启动 PING，失败记 error 日志不退出）
- [ ] 1.6 接入 golang-migrate：migrations/ 目录 + go:embed + 启动自动 Up；提交首个 migration（空基线 000001_init.up/down.sql 或等价时间戳前缀）
- [ ] 1.7 实现 GET /healthz（统一包裹返回）与 api 路由装配；httptest：healthz 200、trace_id 透传、未带 trace_id 生成
- [ ] 1.8 server 单测：config 默认值/覆盖、errcode 段位校验、migration 幂等（重复 Up 不报错）；make test-server 全绿

## 2. 协议包（packages/protocol）

- [ ] 2.1 初始化独立 Go module（github.com/lindaailabs/yuyan/protocol），go:embed 暴露 schema/ 目录
- [ ] 2.2 编写 7 个 WS 命令字 JSON Schema（schema/ws/）：cmd 必填、seq 必填整数、data 对象、client_msg_id 仅 msg.send 必填（帧字段语义严格对齐 guide §5.1/§5.2）
- [ ] 2.3 编写 REST 统一响应包裹 Schema（schema/rest/response.json）
- [ ] 2.4 校验用例（santhosh-tekuri/jsonschema/v6）：每帧类型 ≥1 合法 + ≥1 非法样例；REST 包裹同样；go test 全绿
- [ ] 2.5 go.work 联编：apps/server 通过 workspace 引用 protocol 并编写一个编译级用例（读 embed FS 成功）

## 3. Flutter 骨架（apps/app）

- [ ] 3.1 flutter create（--org com.yuyan.app --project-name yuyan_app），按 §3.4 建立 core/data/features/shared 目录
- [ ] 3.2 core：主题（theme.dart）、l10n 中文常量（l10n/zh.dart）、go_router 路由（登录/主页占位路由）
- [ ] 3.3 data/local：drift AppDatabase（版本 1 空迁移），启动打开/退出释放验证
- [ ] 3.4 data/remote + data/repository：Dio、WebSocket 客户端接口与 repository 抽象占位（空实现）
- [ ] 3.5 main.dart：ProviderScope 包裹、路由壳接入、ProviderScope 示例 provider；模拟器启动走查（3 秒内显示主页、文案均出自 core/l10n）

## 4. 环境与工具（deploy/ + 根目录）

- [ ] 4.1 apps/server/Dockerfile（多阶段：golang:1.22-alpine → alpine）
- [ ] 4.2 deploy/docker-compose.yml：mysql8（utf8mb4 显式参数、healthcheck、volume）+ redis7（appendonly）+ server（depends_on healthy、env 传 DSN/Redis）+ db-only profile
- [ ] 4.3 Makefile：run/test-server/lint/gen-protocol 四目标 + scripts/make.ps1 Windows 等价物
- [ ] 4.4 scripts/gen-protocol 占位（执行 protocol 校验测试）并验证可跑
- [ ] 4.5 README 补充本地启动指南（compose 一键拉起、make run、Windows 说明）

## 5. 验收与收尾

- [ ] 5.1 W1 验收清单走查：docker compose up 后 /healthz 200、migration 自动执行、make test-server 全绿、Flutter 壳页启动、golangci-lint 无告警
- [ ] 5.2 依 guide §6.3 整理 PR 材料：变更文件清单、新依赖声明（gin/gorm/go-redis/golang-migrate/google/uuid/jsonschema + Flutter 五件套）、测试清单
