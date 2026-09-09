# W4 成长系统（add-pet-growth）变更提案

## Why

W3 让宠物"记得住"，但它还不会"变"：陪伴闭环 §0.7（多轮互动后成长状态发生可解释变化并生成成长事件）与 §0.8 的"次日回来状态一致"都不成立。`pivot-to-ai-pet-mvp` D4 明确：成长数值必须由服务端确定性规则驱动，可测试、可回放、可解释，大模型只负责表达，不直接决定核心数值。

## What Changes

- 数据库：新增 `pet_growth_events`、`pet_daily_stats` 两张表（migration up/down + 真容器结构断言）
- 服务端：
  - `internal/model`：PetGrowthEvent / PetDailyStat 实体与 DTO、事件类型常量
  - `internal/repo/growth_repo`：事件写入（唯一键幂等）、按宠物倒序查询、每日统计 upsert
  - `internal/service/growth_service`：确定性成长规则（亲密度、等级、心情、连续互动天数），生成可解释的成长事件
  - `internal/service/conversation_service`：一条宠物回复落库后结算一次成长（以消息 id 为幂等键）
  - `internal/api`：新增 `GET /api/v1/pets/{id}/growth-events`
- Flutter App：
  - 成长事件模型与 repository、`GrowthController`、成长时间线页
  - 路由 `/growth/:petId`，主页「成长」入口由占位改为真实跳转；宠物主页状态卡随互动刷新
- 脚本：`scripts/smoke-pet-growth.ps1` 全栈冒烟（多轮互动 → 亲密度/等级变化 → 事件可查）

## Capabilities

### New Capabilities

- `pet-growth`：亲密度、等级、心情、连续互动的确定性成长规则与可解释成长事件

### Modified Capabilities

- `pet-conversation`：每轮对话成功后触发一次成长结算（幂等，不因重试重复加经验）
- `pet-profile`：宠物状态（level/intimacy/mood）由成长规则更新

## Impact

- **数据库**：`pet_growth_events`、`pet_daily_stats`（纯新增）
- **服务端代码**：新增 growth 域 model/repo/service/api；`conversation_service`、`router.go`、`main.go` 改动
- **Flutter 代码**：`data/model`、`data/repository`、`core/growth`、`features/growth`、`core/router`、`features/home`、`core/l10n`
- **新依赖**：无
- **测试**：成长规则边界（升级阈值、冷却、跨天连续、重复不加经验）、事件生成、越权；api 端点 happy + error；客户端 controller/页面测试；全栈冒烟
- **不做**：权益与付费（W5）、推送与次日提醒（W6）、模型驱动的成长叙事（模型只负责表达）
