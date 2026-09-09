# 一期里程碑拆解（W1~W6）

> **依据**: [LLM_DEV_GUIDE.md](LLM_DEV_GUIDE.md) §1/§4/§5/§7
> **主线验收**: 多账号对话闭环（见 §0）
> **粒度约定**: 每个 W 对应一个 openspec change（如 `add-auth-module`），任务下发用 guide §8 模板；W 验收通过即为里程碑节点，自动 commit。

---

## 0. 闭环验收场景（一期完成的唯一标准）

两个真实账号 A、B，在两台设备/模拟器上：

1. A、B 各自注册登录（验证码）
2. A 搜索 B 手机号 → 发好友申请 → B 同意
3. A、B 在线互发文本消息（P99 < 500ms，WS 实时推送）
4. B 离线时 A 发 N 条消息 → B 上线后**恰好收到 N 条**（离线补发，不丢不重）
5. 任一端杀进程重开 → 历史消息完整（本地 drift + 服务端游标分页双源）
6. 会话列表正确：排序按最新消息、未读数准确、preview 为最后一条文本
7. B 换一台设备登录 → 好友、会话、消息全部一致（多端同步）
8. 压测：1 万连接、1000 msg/s 通过 `scripts/bench-ws.go`

---

## W1 骨架与环境基座

**目标**: 三端脚手架跑通，一条命令拉起完整环境。

### 服务端
- monorepo 目录：`apps/server`（Go module `github.com/lindaailabs/yuyan/server`）、`packages/protocol`
- 分层骨架：`cmd/server/main.go` + `internal/{api,ws,service,repo,model,pkg}`
- pkg：配置加载、`errcode`、slog 封装（HTTP 中间件注入 trace_id）
- GORM + MySQL 连接；migration 机制（时间戳前缀 SQL 文件，golang-migrate 或等价方案，**新依赖需 PR 确认**）
- 健康检查端点 `GET /healthz`

### 协议
- `packages/protocol`：7 个命令字 + REST 响应包裹的 JSON Schema（§5.2/§5.3 全集）+ 校验用例（§7）

### 客户端
- `apps/app`：Flutter 工程骨架，按 §3.4 建目录（core/data/features/shared）
- Riverpod 路由壳 + 主题；drift 初始化；Dio/WS client 空实现（repository 接口占位）

### 环境
- `deploy/docker-compose.yml`：MySQL 8（utf8mb4）+ Redis 7 + server
- `Makefile`：`run` / `test-server` / `lint` / `gen-protocol`

**验收**: `docker compose up` 一键拉起；`/healthz` 200；migration 自动执行；`make test-server` 全绿；Flutter app 可在模拟器启动显示壳页面。

---

## W2 账号体系（auth）

**目标**: A、B 能各自登录，能互相搜到。

- migration：`users` 表（§4 基线）
- REST：`POST /auth/sms-code`（生成 6 位数字验证码 → Redis `sms:code:{phone}` 5min TTL、限频 1/min；一期将验证码渲染为**图形验证码** base64 图片随响应下发，不对接短信服务商）、`POST /auth/login`（验证码校验 + 自动注册，JWT access 2h / refresh 30d）、`POST /auth/refresh`、`GET/PUT /users/me`、`GET /users/search?q=`
- App：登录页（手机号 + 图形验证码图片展示 + 验证码输入；响应无图片字段时按等待短信处理，两态兼容）、首次登录引导页（选预置头像 avatar_id 1~8 + 设昵称）、token 持久化与自动刷新、我的资料页（昵称/预置头像切换）、用户搜索页
- 测试：service 层单测 + 每端点 happy/error httptest（§7）
- 依赖：图片渲染拟用 `base64Captcha` 或 x/image 手绘（二选一，PR 中确认）

**验收**: 闭环场景 §0.1、§0.2 通过（搜索部分）；错误码符合 1xxx/2xxx 段位；验证码生成/校验/限频/TTL 各有用例。

**已决（2026-09）**: 一期验证码不做真实短信对接，以图形验证码代替——验证码仅存 Redis，响应携带渲染图片；生产接入短信时只改 `sms-code` 端点内部实现，契约不变。

---

## W3 好友关系（contacts）

**目标**: 好友申请-同意闭环，双方通讯录就绪。

- migration：`friendships` 表（双向各一行，SMALLINT 状态常量）
- REST：`POST /friends/requests`（含防重复申请、不能加自己）、`GET /friends/requests`、`POST /friends/requests/{id}/accept`（事务内写双向两行）、`POST /friends/requests/{id}/reject`（状态标记 4=rejected，不物理删除）、`GET /friends`
- App：好友申请列表页（同意/拒绝）、通讯录页、加好友入口（复用搜索页）
- 测试：accept 事务原子性单测；防重复/非法参数 error path

**验收**: 闭环场景 §0.2 全通过；重复申请返回业务错误码而非脏数据。

---

## W4 消息在线链路（chat 核心）

**目标**: 在线双端实时收发 + 持久化 + 历史分页。**一期核心，风险最高，预留缓冲。**

- migration：`messages`、`conversations` 表（含 uk_client_msg_id、idx_conv_id_id）
- WS 网关：连接管理（内存路由表）、`conn.auth`（5s 窗口，JWT 校验）、`conn.heartbeat`（30s/90s 超时断开）、JSON 帧编解码
- 消息链路：`msg.send` → client_msg_id 幂等（唯一键冲突即返回原结果）→ 落 messages + 双方 conversations（事务）→ `msg.ack` 回发送方 → `msg.push` 推接收方（在线时）
- REST：`GET /messages?conv_id=&cursor=&limit=20`（游标分页，禁 offset）
- App：聊天页（`reverse: true` + `ListView.builder`）、气泡渲染（widget 测试覆盖）、drift 消息表、发送状态机 pending→sent/failed（超时 5s 重试同 client_msg_id，3 次失败）、打开会话先读本地再拉服务端增量、msg_id 去重
- 测试：幂等/事务/游标分页 service 单测（覆盖率 ≥70%）；ws 帧编解码用例

**验收**: 闭环场景 §0.3、§0.5 通过；杀进程重开历史完整；同 client_msg_id 重发不产生重复消息。

---

## W5 离线可靠性与会话视图（conversation）

**目标**: 离线不丢不重，未读数/会话列表正确，多端一致。

- migration：`offline_msgs` 表
- 离线链路：接收方不在线 → 写 offline_msgs；`conn.auth` 成功后客户端 `msg.pull`（cursor+limit 分页）→ 拉完确认 → 服务端删 offline_msgs
- 未读数：conversations.unread_count 维护 + `conv.unread` 推送（对方发消息/自己已读时）
- REST：`GET /conversations`（按 updated_at 排序，含 last_msg_preview/unread_count）
- App：会话列表页（未读角标、preview、排序）、登录后离线拉取流程、已读上报（进入聊天页清零）
- 测试：离线补发"恰好一次"单测；未读数并发变更单测

**验收**: 闭环场景 §0.4、§0.6、§0.7 通过。

---

## W6 联调、压测与收尾

**目标**: 达到 §7 全部门槛，闭环全绿。

- `scripts/bench-ws.go`：1 万连接、1000 msg/s 压测，P99 延迟 < 500ms，内存无泄漏
- 全链路 E2E：按 §0 场景 1~7 完整走查（自动化脚本或手工清单留档）
- 补测试至门槛：service ≥70%、api 每端点 happy+error、协议 Schema 全覆盖
- docker-compose 生产化（资源限制、重启策略）；README 完善（本地启动指南）
- 修复联调发现的全部 P0/P1

**验收**: §0 场景 1~8 全绿；`make test-server` 全绿；golangci-lint 无告警。

---

## 风险与待决清单

| # | 事项 | 影响 | 建议 |
|---|---|---|---|
| 1 | W4 消息链路复杂度 | 幂等+事务+推送交织，最易出丢/重/乱序 | 单独 PR 分步交付：先持久化+ack，再 push；§9.8 要求 PR 说明推理 |
| 2 | Flutter Windows 环境差异 | 2 名客户端为虚拟配置 | 实际以本机模拟器为准，问题随 W1 暴露 |
