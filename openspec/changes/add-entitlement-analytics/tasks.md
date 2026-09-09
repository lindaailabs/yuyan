## 1. 文档与变更

- [x] 1.1 新增 `add-entitlement-analytics` change：proposal / design / tasks / specs（entitlement、analytics）
- [x] 1.2 更新 docs/MILESTONES.md：W5 交付状态与 TEST-GAP 留档

## 2. 服务端（apps/server）

- [x] 2.1 migration `20260913090000_add_entitlements`、`20260913093000_add_usage_counters`、`20260913100000_add_event_logs`、`20260913103000_add_payment_orders`（up/down）
- [x] 2.2 `internal/model/entitlement.go`：权益/用量/事件/订单实体与 DTO、plan 与状态常量
- [x] 2.3 `internal/repo`：entitlement_repo（upsert/查询）、usage_repo（条件扣减原子）、event_repo（批量写入/查询）、payment_order_repo（订单号幂等）
- [x] 2.4 `internal/service/entitlement_service`：权益查询（含剩余额度）、额度扣减（并发安全）、沙盒开通（非生产）、幂等回调；25xx 错误码 + 单测
- [x] 2.5 `internal/service/analytics_service`：事件落库、批量上报（限条数/长度/脱敏）、内部查询 + 单测
- [x] 2.6 对话链路接入：Gateway 前校验额度（超额不写消息不调用 AI）；关键路径埋点（建宠、首轮对话、记忆生成/删除、升级、额度拒绝）
- [x] 2.7 `internal/api`：GET /entitlements/me、POST /entitlements/sandbox-purchase、POST /entitlements/payments/callback、POST /events、GET /admin/events（非生产）+ 路由与装配
- [x] 2.8 测试：service 单测 + api 每端点 happy/error httptest；`go test` 全绿 + `go vet` 0 issue

## 3. Flutter App（apps/app）

- [x] 3.1 `data/model/entitlement.dart` 与 `data/repository/entitlement_repository.dart`、`analytics_repository.dart`
- [x] 3.2 `core/entitlement/entitlement_controller.dart`：加载权益、沙盒开通、额度耗尽状态
- [x] 3.3 `features/subscription/subscription_page.dart`：权益与剩余额度、沙盒开通、耗尽引导 + 空/错/加载态
- [x] 3.4 路由 `/subscription` + 主页权益入口 + 聊天页额度错误引导 + 订阅页曝光埋点 + l10n 文案
- [x] 3.5 测试：controller 与订阅页 widget；`flutter test` 全绿 + `flutter analyze` 0 issue

## 4. 联调与收尾

- [x] 4.1 `scripts/smoke-entitlement.ps1`：额度耗尽拒绝 2501（不写消息）→ 沙盒恢复 → 回调幂等 → 事件可查
- [ ] 4.2 里程碑提交 + push
