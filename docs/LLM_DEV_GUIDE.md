# Yuyan 项目开发指导文档

> **版本**: v1.0（2026-09）
> **用途**: 本文档面向 AI 编程助手与人类开发者，定义 Yuyan 项目的技术边界、架构约束与开发规范。
> **优先级**: 当任务描述与本文档冲突时，**以本文档为准**；如需变更规范，先修改本文档再改代码。

---

## 1. 项目概览

### 1.1 产品定义

Yuyan（语燕）是一款即时通讯 App，一期 MVP 目标：**实现单聊纯文本消息的完整闭环**——用户注册登录、添加好友、发送/接收文本消息、消息持久化、多端同步。

### 1.2 MVP 范围（一期）

**包含（In Scope）：**
- 手机号 + 验证码注册登录
- 用户资料（昵称、头像）
- 好友关系（搜索、申请、同意）
- 单聊文本消息收发（WebSocket 实时推送）
- 消息本地持久化与历史拉取（分页）
- 离线消息（上线后补发）
- 未读数与会话列表
- Android + iOS 双端（Flutter 一套代码）

**不包含（Out of Scope，禁止 LLM 主动实现）：**
- 群聊、语音、视频、图片/文件消息
- 消息撤回、已读回执、端到端加密
- 推送（APNs/FCM）、动态、朋友圈
- 后台管理系统、运营工具
- 微服务拆分、K8s、多机房部署

### 1.3 非功能目标

| 指标 | 目标值 |
|---|---|
| 在线连接数 | 1 万（单机 Go） |
| 消息端到端延迟 | P99 < 500ms |
| App 冷启动 | < 2.5s（中端机） |
| 服务可用性 | 99.5%（一期） |
| 团队规模 | 2 名客户端（Flutter）+ 2 名服务端（Go，Java 转型中） |

---

## 2. 技术栈（已定案，不可更改）

| 层 | 技术 | 版本基线 |
|---|---|---|
| 客户端 | Flutter (Dart) | Flutter 3.2x+ / Dart 3.x |
| 客户端状态管理 | Riverpod | 2.x |
| 客户端本地存储 | drift (SQLite) | — |
| 客户端网络 | WebSocket + Dio | — |
| 服务端 | Go 单体 | Go 1.22+ |
| 服务端 Web 框架 | Gin | — |
| 服务端 ORM | GORM（仅用于 CRUD，复杂查询写原生 SQL） | — |
| 数据库 | MySQL 8.x（InnoDB，utf8mb4） | — |
| 缓存 | Redis 7+（在线状态、未读数、验证码） | — |
| 消息协议 | 自定义 JSON 帧（见 §5） | v1 |
| 部署 | Docker Compose（单机） | — |

**⚠️ LLM 注意：不要引入以下技术，除非任务明确要求：**
Kafka / RabbitMQ / gRPC / 微服务 / Kubernetes / MongoDB / PostgreSQL（已选 MySQL）。
如需引入新依赖，先在 PR 描述中说明理由，等待人工确认。

---

## 3. 仓库结构与命名规范

### 3.1 Monorepo 结构

```
yuyan/
├── apps/
│   ├── app/                # Flutter App（package: yuyan_app）
│   └── server/             # Go 单体（module: github.com//yuyan/server）
├── packages/
│   └── protocol/           # 共享协议定义（JSON Schema + 双端代码生成）
├── docs/                   # 技术方案、API 文档、排期、决策记录
├── deploy/                 # docker-compose、CI、部署脚本
├── scripts/                # gen-protocol.sh 等辅助脚本
├── Makefile
└── README.md
```

### 3.2 命名速查

| 项 | 值 |
|---|---|
| Go module | `github.com/<org>/yuyan/server` |
| Flutter package | `yuyan_app` |
| Android applicationId / iOS Bundle ID | `com.yuyan.app` |
| API Base | `https://api.yuyan.im/api/v1` |
| WS 端点 | `wss://api.yuyan.im/ws` |

### 3.3 Go 服务端内部分层（单体分层，禁止跳层调用）

```
apps/server/
├── cmd/server/main.go      # 入口：只做配置加载与依赖装配
├── internal/
│   ├── api/                # HTTP handler（Gin），只做参数校验与调用 service
│   ├── ws/                 # WebSocket 网关（连接管理、帧编解码、心跳）
│   ├── service/            # 业务逻辑层（唯一允许写事务的地方）
│   ├── repo/               # 数据访问层（GORM / 原生 SQL）
│   ├── model/              # 实体与 DTO（entity 与 dto 分开定义）
│   └── pkg/                # 内部工具（错误码、日志、配置）
└── migrations/             # SQL 迁移文件（时间戳前缀）
```

**调用方向只允许：`api/ws → service → repo`。** 反向调用、跨层调用（api 直接调 repo）视为架构违规。

### 3.4 Flutter 客户端分层

```
apps/app/lib/
├── main.dart
├── core/                   # 与业务无关：主题、路由、常量、工具
├── data/
│   ├── remote/             # API client、WebSocket client
│   ├── local/              # drift 数据库、DAO
│   └── repository/         # 仓库实现（remote+local 合并策略）
├── features/               # 按功能垂直切分
│   ├── auth/               # 每个 feature 内含 ui/ + providers/ + models/
│   ├── contacts/
│   ├── chat/
│   └── conversation/
└── shared/                 # 跨 feature 复用的 widget
```

**架构要求（为二期可能的迁移预留）：**
- Riverpod provider 层不得直接引用 widget；UI 不得直接调用 `data/remote`。
- 所有业务逻辑收敛在 repository 与 provider，UI 层只负责渲染与事件转发。

---

## 4. 数据库设计基线

> LLM 修改表结构前必须检查本节；新增字段必须同步写 migration 文件。

```sql
-- 用户
users(id BIGINT UNSIGNED AUTO_INCREMENT PK, phone VARCHAR(20) UNIQUE,
      nickname, avatar_url, created_at, updated_at)

-- 好友关系（双向各存一行；status 用 SMALLINT 常量：1=pending 2=accepted 3=blocked）
friendships(id, user_id, friend_id, status SMALLINT, created_at,
            UNIQUE KEY uk_user_friend(user_id, friend_id))

-- 会话（单聊会话，双方各一行）
conversations(id, user_id, peer_id, last_msg_id, last_msg_preview,
              unread_count INT DEFAULT 0, updated_at,
              UNIQUE KEY uk_user_peer(user_id, peer_id))

-- 消息（服务端持久化；游标分页走 idx_conv_id_id 二级索引，量级上来后再评估分区表）
messages(id BIGINT UNSIGNED AUTO_INCREMENT PK,  -- 同时作为时序游标
         conv_id, sender_id, receiver_id,
         msg_type SMALLINT DEFAULT 1,           -- 1=文本（一期仅此一种）
         content TEXT, status SMALLINT,         -- 1=已发送 2=已送达
         client_msg_id CHAR(36),                -- 客户端去重键（UUID v4 字符串）
         created_at BIGINT,                     -- 毫秒时间戳
         UNIQUE KEY uk_client_msg_id(client_msg_id),
         KEY idx_conv_id_id(conv_id, id))

-- 离线消息索引（用户上线后拉取，确认送达后删除）
offline_msgs(id, user_id, msg_id, created_at,
             UNIQUE KEY uk_user_msg(user_id, msg_id))
```

**关键约定：**
1. 消息 ID 用数据库自增（或雪花 ID），**兼作时序排序与增量拉取游标**，禁止用时间戳排序去重。
2. 客户端发送消息必须携带 `client_msg_id`（UUID），服务端以此做幂等去重。
3. 所有表禁止物理删除，需要删除语义时用 status 标记。

---

## 5. 通信协议（帧格式 v1）

### 5.1 WebSocket 帧（JSON，UTF-8）

```json
{
  "cmd": "msg.send",           // 命令字，见下表
  "seq": 12345,                // 客户端递增序号，响应原样带回
  "data": { ... },             // 业务负载
  "client_msg_id": "uuid"      // 仅 msg.send 必填
}
```

响应帧：

```json
{ "cmd": "msg.send", "seq": 12345, "code": 0, "msg": "ok", "data": { ... } }
```

### 5.2 命令字清单（一期全集，LLM 不得自行新增）

| cmd | 方向 | 说明 |
|---|---|---|
| `conn.auth` | C→S | 连接后 5s 内鉴权，携带 token |
| `conn.heartbeat` | C→S | 每 30s，服务端 90s 未收到则断开 |
| `msg.send` | C→S | 发送消息，data: `{conv_id, content}` |
| `msg.push` | S→C | 推送新消息（对方在线时实时下发） |
| `msg.ack` | S→C | 服务端确认，data: `{client_msg_id, msg_id, created_at}` |
| `msg.pull` | C→S | 拉取离线消息，data: `{cursor, limit}` |
| `conv.unread` | S→C | 未读数变更推送 |

### 5.3 REST API（一期全集）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/auth/sms-code` | 发送验证码（Redis 存储，5min 过期，限频 1/min） |
| POST | `/api/v1/auth/login` | 验证码登录/注册，返回 JWT（access 2h + refresh 30d） |
| POST | `/api/v1/auth/refresh` | 刷新 token |
| GET/PUT | `/api/v1/users/me` | 个人资料 |
| GET | `/api/v1/users/search?q=` | 按手机号精确搜索 |
| POST | `/api/v1/friends/requests` | 发起好友申请 |
| GET | `/api/v1/friends/requests` | 申请列表 |
| POST | `/api/v1/friends/requests/{id}/accept` | 同意申请 |
| GET | `/api/v1/friends` | 好友列表 |
| GET | `/api/v1/conversations` | 会话列表 |
| GET | `/api/v1/messages?conv_id=&cursor=&limit=20` | 历史消息（游标分页，禁止 offset） |

**REST 统一响应包裹：** `{ "code": 0, "msg": "ok", "data": {...} }`
错误码段：`0` 成功；`1xxx` 参数/鉴权；`2xxx` 业务；`5xxx` 服务端。

### 5.4 消息可靠性模型（LLM 实现消息逻辑时必须遵守）

```
发送方: 本地落库(status=pending) → WS msg.send
        ← msg.ack → 更新 msg_id/status=sent；超时 5s 重试(同 client_msg_id)，3 次失败标记 failed
接收方: 在线 → msg.push → 回 ack(隐式) → 服务端删 offline_msgs
        离线 → 写 offline_msgs → 上线后 conn.auth 成功 → 客户端主动 msg.pull 分页拉取
去重: 客户端以 msg_id 去重；服务端以 client_msg_id 幂等
```

---

## 6. 编码规范

### 6.1 Go

- 遵循 `gofmt` + `golangci-lint`（配置见仓库根目录）。
- 错误处理：service 层返回带错误码的自定义 error（`pkg/errcode`），api 层统一转换为响应包裹；**禁止吞错误、禁止 `panic` 处理业务异常**。
- 每个 handler 必须做参数校验（binding tag + 显式业务校验）。
- 日志用 `slog`（结构化），禁止 `fmt.Println`。日志必须含 `trace_id`（HTTP 中间件注入，WS 帧透传）。
- 团队为 Java 转型：**避免炫技式 Go 写法**（过度泛型、反射魔法、channel 杂技），以"Java 工程师能读懂"为第一标准；并发原语使用处必须写注释说明。

### 6.2 Dart/Flutter

- 遵循 `dart format` + `flutter_lints`。
- 状态管理只用 Riverpod：`StateNotifier/Notifier + provider`；**禁止 setState 管理跨页状态、禁止 GetX、禁止 Provider(旧库)**。
- 网络层与业务层之间通过 repository 接口隔离，widget 树中禁止出现 Dio/WebSocket 直接调用。
- 列表一律 `ListView.builder` + 分页加载；聊天页反转列表用 `reverse: true`。
- 所有用户可见文案收敛到 `lib/core/l10n/`（一期可只有中文，但禁止硬编码在 widget 中）。

### 6.3 Git 与提交

- 分支：`main`（可发布）/ `dev`（集成）/ `feat/<模块>-<描述>` / `fix/<描述>`。
- Commit 遵循 Conventional Commits：`feat(chat): 支持消息分页拉取`。
- **LLM 生成代码的每个 PR 必须附带**：变更文件清单、对应的测试、以及"是否引入新依赖"声明。

---

## 7. 测试要求（LLM 交付代码的验收门槛）

| 层 | 要求 |
|---|---|
| Go service 层 | 核心逻辑（消息收发、幂等、离线补发）单测覆盖率 ≥ 70%，用 `testify` + `sqlmock`/testcontainers |
| Go api 层 | 每个端点至少 1 个 happy path + 1 个错误路径的 httptest 用例 |
| Flutter | repository 与 provider 层必须有单测；widget 测试仅覆盖聊天页气泡渲染 |
| 协议 | `packages/protocol` 中每个帧类型必须有 JSON Schema 校验用例 |
| 压测 | 消息链路需通过 `scripts/bench-ws.go`（1 万连接、每秒 1000 条消息）验证后才可合并 |

**LLM 不得以"测试较复杂"为由跳过测试**；确实无法编写时，必须在 PR 中显式标注 `TEST-GAP: 原因`，由人工评审。

---

## 8. 给 LLM 的任务下发模板

> 每次向 LLM 下发任务时，使用以下模板，可显著减少来回澄清：

```
【任务】feat(chat): 实现消息历史分页拉取
【涉及端】server / app / 两者
【依据文档】LLM_DEV_GUIDE.md §4、§5.3
【输入】
- 现有代码：apps/server/internal/repo/message_repo.go
- 接口契约：GET /api/v1/messages?conv_id=&cursor=&limit=20
【交付物】
1. repo 层方法 GetMessagesByCursor(convID, cursor, limit)
2. service 层 + api handler
3. migration（如需）
4. 单测（见 §7）
【禁止】
- 修改帧协议或新增命令字
- 引入新第三方依赖
- 改动本任务范围外的文件
【完成标准】make test-server 全绿；golangci-lint 无新增告警
```

---

## 9. LLM 红线清单（任何时候不得违反）

1. ❌ 不得修改 `packages/protocol` 的已有字段语义（只能新增可选字段，且需人工确认）。
2. ❌ 不得删除或重写 migration 历史文件。
3. ❌ 不得在代码中硬编码密钥、token、生产环境地址。
4. ❌ 不得将 Out of Scope（§1.2）功能"顺手"实现。
5. ❌ 不得绕过分层约束（§3.3/§3.4）调用。
6. ❌ 不得使用 offset 分页、时间戳排序做消息游标。
7. ❌ 不得引入 §2 禁用清单中的中间件/框架。
8. ⚠️ 所有涉及"消息丢失/重复/乱序"的改动，必须在 PR 中单独说明推理过程。

---

## 10. 二期演进预案（LLM 知悉即可，一期不实现）

- **鸿蒙适配**：评估 Kuikly 或 OpenHarmony Flutter fork；因此一期 Flutter 业务逻辑必须与 UI 严格解耦（§3.4）。
- **消息类型扩展**：`msg_type` 字段已预留，新增类型走 protocol 包版本化流程。
- **推送**：APNs/FCM 接入时，离线消息模型（offline_msgs）不变，推送仅作为"唤醒"手段。
- **规模化**：1 万连接以上再评估 WS 网关独立部署（Redis Pub/Sub 做跨节点路由），一期单机内存路由即可。

---

*本文档由架构负责人维护，修改需走 PR 评审。最后更新：2026-09。*

---

### 使用建议

1. **存放位置**：`docs/LLM_DEV_GUIDE.md`，同时在仓库根目录 `README.md` 里加一行链接指向它。
2. **注入方式**：用 Claude Code / Cursor 时，将其配置为项目级规则文件（如 `.cursorrules` 或 `CLAUDE.md` 引用此文档）；用对话式 LLM 时，每次新会话先粘贴 §2、§3、§9 三节（约 1500 token，性价比最高的最小子集）。
3. **迭代节奏**：每完成一个周里程碑（W1~W6），回顾一次文档与实际代码的偏差，及时更新——**文档过期比没有文档更危险**。
