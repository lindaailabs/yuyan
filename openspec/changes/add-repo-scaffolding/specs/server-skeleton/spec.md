## ADDED Requirements

### Requirement: 服务进程启动与配置加载
Go 单体服务 SHALL 通过 `cmd/server/main.go` 启动；配置从环境变量加载（HTTP_PORT、MYSQL_DSN、REDIS_ADDR、LOG_LEVEL、APP_ENV），未设置时 MUST 使用安全默认值（端口 8080、日志 info、env=dev）。配置加载与依赖装配 MUST 全部位于 main.go，业务代码不得读取环境变量。

#### Scenario: 使用默认配置启动
- **WHEN** 不设置任何环境变量执行服务二进制
- **THEN** 进程以默认配置启动，日志输出当前生效配置（不含敏感明文），HTTP 监听 8080

#### Scenario: 环境变量覆盖默认值
- **WHEN** 设置 HTTP_PORT=9090 后启动
- **THEN** 服务监听 9090 端口

### Requirement: 健康检查端点
服务 SHALL 暴露 `GET /healthz`，无条件返回 HTTP 200 与统一响应包裹 `{"code":0,"msg":"ok","data":null}`。

#### Scenario: 健康检查
- **WHEN** 任意时刻请求 GET /healthz
- **THEN** 返回 200，响应体 code 为 0

### Requirement: 结构化日志与 trace_id 注入
服务 MUST 使用 slog（JSON handler）输出结构化日志，禁止 fmt.Println。HTTP 中间件 SHALL 为每个请求生成或透传 trace_id（请求头 `X-Trace-Id` 存在则复用，否则生成 uuid），MUST 写入响应头并在该请求的所有日志行携带。

#### Scenario: 请求携带 trace_id
- **WHEN** 请求头携带 X-Trace-Id: abc123 访问 /healthz
- **THEN** 响应头 X-Trace-Id 为 abc123，且该请求日志行含 trace_id=abc123

#### Scenario: 请求未携带 trace_id
- **WHEN** 请求头不含 X-Trace-Id
- **THEN** 响应头返回新生成的 trace_id，日志行携带同一值

### Requirement: MySQL 连接与 migration 自动执行
服务启动时 SHALL 通过 GORM 建立 MySQL 连接（utf8mb4），并使用 golang-migrate 自动执行 `migrations/` 目录下全部待执行 migration（时间戳前缀 up/down SQL 文件）。已执行的 migration MUST NOT 被重复执行；migration 文件历史 MUST NOT 被修改或删除。

#### Scenario: 首次启动
- **WHEN** 空数据库上启动服务
- **THEN** 全部 migration 依序执行，schema_migrations 表记录版本，日志输出执行数量

#### Scenario: 重复启动幂等
- **WHEN** 已完成 migration 的数据库再次启动服务
- **THEN** 不执行任何 migration，服务正常就绪

### Requirement: Redis 连接
服务启动时 SHALL 初始化 go-redis 客户端并对 Redis 执行 PING；连接失败 MUST 以 error 级别日志记录（进程不因 Redis 不可用退出，W2 起再按业务决定失败策略）。

#### Scenario: Redis 可用
- **WHEN** Redis 正常运行时启动服务
- **THEN** 启动日志记录 Redis PING 成功

#### Scenario: Redis 不可用
- **WHEN** Redis 无法连接时启动服务
- **THEN** 日志输出 error 级别 Redis 连接失败信息，进程继续运行

### Requirement: 统一错误码体系
`internal/pkg/errcode` SHALL 定义错误码分段：0 成功、1xxx 参数/鉴权、2xxx 业务、5xxx 服务端，并提供实现 error 接口的带码错误类型；api 层 SHALL 统一转换为 REST 响应包裹。

#### Scenario: 错误码分段定义
- **WHEN** 检查 errcode 包预定义错误
- **THEN** 各错误所属段位与规范一致且不可被业务随意越段定义（提供 New 构造函数校验段位）

### Requirement: 分层目录与调用方向
服务端 MUST 保持 `internal/{api,ws,service,repo,model,pkg}` 分层，调用方向仅允许 api/ws → service → repo；W1 各业务层为空占位（包声明文件 + 注释说明职责），ws 仅含空 Hub 占位类型。

#### Scenario: 分层占位存在
- **WHEN** 检查 apps/server/internal 目录
- **THEN** 六个子包均存在且可被 main.go 无错误引用
