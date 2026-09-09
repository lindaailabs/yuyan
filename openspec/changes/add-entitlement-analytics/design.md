# W5 权益与数据看板设计

## Context

前序已交付 W1~W4。现在要为商业化与运营观测打地基：服务端必须能表达「这个用户有什么权益、还剩多少额度」，AI 对话必须受额度约束，关键行为必须可查询。

权威约束：`docs/LLM_DEV_GUIDE.md` §4（entitlements 表基线、历史 migration 不可改）、§7（客户端只展示权益、服务端判定；消耗型权益必须服务端扣减或校验；支付回调必须幂等且订单号唯一键；不得硬编码商户密钥）、§8（埋点范围；A/B 先做服务端分桶与事件上报，不引入平台）、§9/§10/§12（错误码、日志脱敏、测试门槛、红线）；`docs/MILESTONES.md` W5；`pivot-to-ai-pet-mvp` D6。

## Goals / Non-Goals

**Goals**
- 服务端可查询用户权益与当日剩余额度
- AI 对话前校验额度；超额**明确拒绝**（25xx），不写消息、不扣额度
- 支付回调按订单号幂等；沙盒可模拟开通/续费（仅非生产环境）
- 关键事件落库，客户端可批量上报，内部可查询

**Non-Goals**
- 不接 Apple IAP / Google Play / 微信 / 支付宝生产支付，不存商户密钥
- 不引入消息队列、实时数仓、复杂实验平台
- 不扩展好友/IM；不让模型决定权益与扣费

## Decisions

### D1 权益与额度分离存储
- `entitlements`：plan（free/pro）、status、`quota_json`（每日 AI 对话条数、记忆容量、高级模型开关）、`renew_at`
- `usage_counters`：按 `(user_id, action, period)` 唯一键记录 `used`，`period` 为 Asia/Shanghai 自然日（`2026-09-12`）
- 理由：额度窗口随时间滚动，用量与权益解耦，便于重置与对账

### D2 扣减用条件 UPDATE（防并发超卖）
```
INSERT IGNORE usage_counters(user_id, action, period, used) VALUES (?,?,?,0);
UPDATE usage_counters SET used = used + ? 
 WHERE user_id=? AND action=? AND period=? AND used + ? <= ?;
```
- 影响 1 行 → 扣减成功；0 行 → 已达上限，返回 `ErrQuotaExceeded`（2501）
- 避免「先查后写」的并发超卖，单测覆盖并发累加不超过上限

### D3 对话接入点在 Gateway 之前
`SendMessage`：内容校验 → 宠物归属 → **额度校验与扣减** → 写用户消息 → 记忆抽取/召回 → Gateway → 落 assistant 消息。
超额时直接返回 2501，**不写任何消息、不产生 AI 调用**（符合「客户端不能作为权益事实源」，也避免无效成本）。

### D4 超额语义（用户已拍板）
明确拒绝 + 业务错误码 25xx；客户端展示「今日额度用完」并引导订阅，不做静默降级。

### D5 支付幂等靠订单号唯一键
`payment_orders(order_no)` 唯一键：重复回调命中唯一键即视为已处理，返回当前权益，**不重复发放**；沙盒开通同样写入订单记录（order_no 由服务端生成，前缀 `sandbox-`）。

### D6 沙盒与内部接口均有环境门禁
- `POST /entitlements/sandbox-purchase`：仅 `APP_ENV != prod` 可用，否则返回 2504
- `GET /admin/events`：仅非生产环境注册/放行，避免内部数据外泄

### D7 埋点脱敏与上限
- 事件名白名单长度 ≤64；`props` JSON 长度上限（如 2KB）；**禁止**写入手机号/token/prompt 全文（红线 §12）
- 批量上报单次 ≤50 条；超限参数错误
- 服务端在登录/建宠/首轮对话/记忆生成与删除/升级/额度拒绝等处落事件；端上曝光与点击由客户端上报

### D8 错误码
- 2501 额度不足
- 2502 权益不存在或无权限
- 2503 订单参数非法
- 2504 沙盒支付未启用（生产环境）

## Risks / Trade-offs

- [额度窗口跨时区] → 固定 Asia/Shanghai 自然日，注入时钟单测
- [用量与权益短暂不一致] → 事件与订单可回放对账；一期无分布式事务
- [沙盒被误用于生产] → 环境门禁 + 明确错误码 + 启动日志提示
- [埋点写放大] → 关键路径少量事件 + 批量上报，失败仅告警不阻断业务

## Migration Plan

1. 新增四张表 migration（时间戳前缀，up/down 成对）
2. `docker compose up -d --build` 自动执行；回滚 = down 迁移 + 回退镜像
3. 里程碑提交：`feat(entitlement): W5 权益、支付地基与数据看板`

## Open Questions

（无——额度语义、埋点范围、支付地基深度均已由用户拍板）
