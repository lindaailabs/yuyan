# W4 成长系统设计

## Context

W1~W3 已交付宠物档案、文本对话与记忆。宠物目前只有静态属性（level=1、intimacy=0、mood=curious），互动不会带来任何变化。W4 让"互动 → 成长"这条线成立，并且必须**服务端确定性**：guide §4 关键约定 4、§12 红线 7、`pivot-to-ai-pet-mvp` D4 都禁止让模型决定核心数值。

权威依据：`docs/LLM_DEV_GUIDE.md` §4（pet_growth_events 表结构）、§5（上下文含成长状态）、§12 红线；`docs/MILESTONES.md` W4。

## Goals / Non-Goals

**Goals**
- 每轮有效互动产生确定的亲密度增长，并据此升级、变化心情、累计连续互动天数
- 每次状态变化生成**可解释**的成长事件（类型、delta、原因、来源消息）
- 重复处理同一条消息不重复加经验（可回放、可重入）
- 事件可按宠物查询，供 App 展示成长时间线

**Non-Goals**
- 不让模型输出决定任何数值（模型只负责把事件"讲出来"）
- 不做权益/付费、推送、次日提醒（W5/W6）
- 不做复杂成就系统、排行榜、装扮解锁

## Decisions

### D1 结算时机与幂等键
在**宠物回复落库之后**结算一次成长，幂等键 = `(pet_id, source_msg_id, event_type)` 唯一键。
- 唯一键冲突 → 视为已结算，直接返回当前状态（重复请求/重试不会重复加经验）
- AI 失败也计一次互动（用户确实陪伴了宠物），但事件 `reason` 标注 `ai_failed`

### D2 规则集中且纯函数化
| 维度 | 规则 |
|---|---|
| 亲密度 | 每条宠物消息 +2（常量 `intimacyPerMessage`） |
| 等级 | `level = 1 + intimacy / 20`（`intimacyPerLevel`），升级写 `level_up` 事件 |
| 连续互动 | 按 Asia/Shanghai 自然日：跨天首次互动 `streak+1`；同日不重复累计；断天归 1 |
| 心情 | 依据距上次互动时长：≤1h `happy`；≤24h `curious`；≤72h `sleepy`；>72h `lonely`；首次互动 `curious` |
| 里程碑 | 连续互动达到 3/7/30 天写 `streak_milestone` 事件 |

规则集中在 `growth_service.go` 常量区，纯函数可单测，不依赖网络与模型。

### D3 状态写入顺序
1. 插入 `event_type=message` 事件（幂等闸门，冲突即返回）
2. 累加亲密度 → 计算等级 → 必要时更新并写 `level_up`
3. upsert 每日统计 → 必要时写 `streak_milestone`
4. 计算心情 → 变化时更新并写 `mood_change`

每一步都是短事务/单行写；任一后续步骤失败只告警，不回滚已确认的互动（避免经验丢失）。

### D4 事件可解释
事件 `delta` 为 JSON（如 `{"intimacy":2}`），`reason` 为人类可读说明（如「完成一次对话，亲密度 +2」），并带 `source_msg_id` 便于回溯。

### D5 错误码与边界
- 宠物越权复用 `ErrPetNotAccessible`（2303）
- 事件查询无权限同 2303；参数非法 1001
- 单次结算的事件数量受控（最多 3 类：message / level_up / mood_change，另加可能的 streak_milestone）

## Risks / Trade-offs

- [每日统计与事件双写不一致] → 事件是事实源，统计可重建；一期无分布式事务需求
- [时区] → 固定 Asia/Shanghai 计算自然日（与 MySQL 容器 `default-time-zone=+08:00` 一致）
- [心情随真实时间变化] → 读取时按「现在 - 上次互动」计算，无需定时任务；次日回来自然呈现变化
- [重复结算] → 唯一键兜底，单测覆盖重放

## Migration Plan

1. 新增两张表 migration（时间戳前缀，up/down 成对）
2. `docker compose up -d --build` 自动执行；回滚 = down 迁移 + 回退镜像
3. 里程碑提交：`feat(pet-growth): W4 成长系统（确定性规则 + 成长事件）`

## Open Questions

（无——规则与阈值集中在常量区，后续调参不需要结构变更）
