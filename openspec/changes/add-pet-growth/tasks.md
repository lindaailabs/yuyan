## 1. 文档与变更

- [x] 1.1 新增 `add-pet-growth` change：proposal / design / tasks / specs（pet-growth）
- [x] 1.2 更新 docs/MILESTONES.md：W4 交付状态与 TEST-GAP 留档

## 2. 服务端（apps/server）

- [x] 2.1 migration `20260912090000_add_pet_growth_events`、`20260912100000_add_pet_daily_stats`（up/down）
- [x] 2.2 `internal/model/growth.go`：事件与每日统计实体、GrowthEventItem DTO、事件类型与心情常量
- [x] 2.3 `internal/repo/growth_repo`：事件幂等写入（唯一键）、倒序查询、每日统计 upsert
- [x] 2.4 `internal/service/growth_service`：亲密度 +2/条、`level = 1 + intimacy/20`、自然日连续互动（3/7/30 里程碑）、距上次互动推导心情；单测覆盖升级/跨天连续/断天重置/重复不加经验/心情变化/越权/纯函数
- [x] 2.5 对话链路接入：宠物回复落库后结算一次成长（幂等键 = 消息 id，AI 失败也计互动但原因不同）
- [x] 2.6 `internal/api/handler_growth` + 路由注册 + `main.go` 装配
- [x] 2.7 测试：service 单测 + api 端点 happy/error httptest；`go test` 全绿（根包/表结构、service、api、ai）+ `go vet` 0 issue

## 3. Flutter App（apps/app）

- [x] 3.1 `data/model/pet_growth.dart` 与 `data/repository/pet_growth_repository.dart`
- [x] 3.2 `core/growth/growth_controller.dart`：加载事件、失败重试
- [x] 3.3 `features/growth/growth_page.dart`：时间线 + 空/错/加载态 + 下拉刷新
- [x] 3.4 路由 `/growth/:petId` + 主页「成长」入口接真实跳转 + l10n 文案 + providers 装配
- [x] 3.5 测试：controller 3 例、成长页 widget 2 例；`flutter test` 60 例全绿 + `flutter analyze` 0 issue

## 4. 联调与收尾

- [x] 4.1 `scripts/smoke-pet-growth.ps1`：10 轮互动 → 亲密度 20/等级 2 → 事件可查且含升级事件 → 2303/1001/1002；W2 对话与 W3 记忆冒烟回归通过
- [x] 4.2 里程碑提交 + push
