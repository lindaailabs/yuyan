# W5 权益、支付地基与数据看板（add-entitlement-analytics）变更提案

## Why

W0~W4 已交付陪伴闭环的主体（档案、对话、记忆、成长），但产品还不具备商业化与运营观测的最小地基：服务端没有权益概念，AI 对话无成本约束，关键行为无法追溯。`docs/MILESTONES.md` W5 的验收是「服务端能按用户返回权益和剩余额度；AI 对话能记录和检查额度；关键事件可查询」。`pivot-to-ai-pet-mvp` D6 也要求「支付地基先于真实支付，客户端不能成为权益事实源」。

## What Changes

- 数据库：新增 `entitlements`、`usage_counters`、`event_logs`、`payment_orders` 四张表（migration up/down + 真容器结构断言）
- 服务端：
  - `internal/model`：权益、用量计数器、事件日志、支付订单实体与 DTO
  - `internal/repo`：权益 upsert/查询、用量条件扣减（原子）、事件批量写入与查询、支付订单幂等写入
  - `internal/service/entitlement_service`：权益查询（含剩余额度）、额度校验与扣减、沙盒开通（非生产环境）、按订单号幂等的支付回调；错误码 25xx
  - `internal/service/analytics_service`：事件落库、批量上报（限条数与长度、脱敏）、内部查询出口
  - `internal/service/conversation_service`：调用 AI Gateway 前校验额度；关键路径埋点
  - `internal/api`：`GET /entitlements/me`、`POST /entitlements/sandbox-purchase`、`POST /entitlements/payments/callback`、`POST /events`、内部 `GET /admin/events`
- Flutter App：
  - 权益模型与 repository、`EntitlementController`
  - 订阅页（权益状态、剩余额度、沙盒开通、额度耗尽引导）
  - 聊天页捕获额度错误并引导订阅
  - `AnalyticsRepository` 批量上报（订阅页曝光/点击等端上事件）
- 脚本：`scripts/smoke-entitlement.ps1` 全栈冒烟（额度耗尽拒绝 → 沙盒恢复 → 事件可查）

## Capabilities

### New Capabilities

- `entitlement`：服务端权益模型、额度校验与扣减、沙盒开通与幂等支付回调（客户端仅展示）
- `analytics`：关键行为事件落库、客户端批量上报与内部查询出口

### Modified Capabilities

- `pet-conversation`：发送前校验额度，超额返回 25xx 且不写消息；关键路径埋点
- `pet-profile` / `pet-memory` / `pet-growth`：创建宠物、记忆生成与删除、升级等处埋点

## Impact

- **数据库**：四张新表（utf8mb4/InnoDB，纯新增）
- **服务端代码**：新增 entitlement/analytics 域 model/repo/service/api；`conversation_service` 改动；`router.go`/`main.go` 装配；`config.go` 新增沙盒与内部接口开关
- **Flutter 代码**：`data/model`、`data/repository`、`core/entitlement`、`features/subscription`、`core/router`、`features/chat`、`features/home`、`core/l10n`
- **新依赖**：无
- **测试**：额度边界与并发安全、沙盒环境门禁、回调幂等、事件上报上限与脱敏；api 每端点 happy + error；客户端 controller/页面测试；全栈冒烟
- **不做**：真实支付渠道接入与商户密钥、复杂实验平台、好友/IM 扩展、模型决定权益或扣费
