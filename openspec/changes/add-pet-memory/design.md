# W3 记忆系统设计

## Context

W2 已交付 AI Gateway 与文本对话链路（`pet_conversations`/`pet_messages`/`ai_call_logs`），`CompletionRequest` 已预留 `Memories []string` 字段，但一直是空的。现在要把「长期记忆」这段上下文真正填上，并让用户可控。

权威约束：`docs/LLM_DEV_GUIDE.md` v2.0 §4（pet_memories 表结构、长期记忆必须可追溯来源、可软删除）、§5（长期记忆只召回少量相关事实；日志脱敏）、§12 红线（不得记录 prompt 与隐私全量内容）、§13（记忆检索一期用 MySQL + 规则筛选，量级上来再评估向量检索）；`pivot-to-ai-pet-mvp` D5（记忆必须有来源消息、类型、置信度、状态和更新时间；用户可查看与删除；删除后不得召回）；`specs/ai-pet-mvp` 的 `Memory Control`；`docs/MILESTONES.md` W3。

## Goals / Non-Goals

**Goals**
- 从用户消息中稳定抽取少量长期事实（偏好、称呼、画像、事件），可追溯来源与置信度
- 对话时按相关度召回少量记忆（默认 3 条）注入 Gateway 上下文
- 用户可查看记忆列表并删除；删除后不再召回
- 抽取与召回均为服务端确定性规则，可测试、可回放

**Non-Goals**
- 不用大模型做抽取（真实 provider 接入后再扩展，接口与表结构已预留 `memory_type`/`confidence`）
- 不引入向量数据库或全文检索引擎（guide §13）
- 不做记忆编辑、审核后台、记忆容量权益（W5）
- 不改动成长数值（W4）

## Decisions

### D1 抽取用确定性规则，而非模型
一期 mock provider 无法稳定产出结构化记忆；规则抽取可测试、可回放、零额外成本。规则表形态：
`{Type, Regexp, Confidence}` → 命中即生成 `content`，并计算 `content_hash = sha256(petID|type|content)` 用于去重。

示例规则（可扩展）：
- `我(喜欢|爱|最爱)(.+)` → type=preference, confidence=0.8
- `我(不喜欢|讨厌)(.+)` → type=preference, confidence=0.8
- `我(叫|是)(.+)` → type=profile, confidence=0.7
- `我(住在|在)(.+)工作?` → type=profile, confidence=0.6

### D2 去重与软删除用 upsert
`(pet_id, content_hash)` 唯一键：同一事实重复表达只更新 `confidence`/`source_msg_id`/`updated_at`，并把历史「已删除」记录复活为 active（用户再次表达即视为重新授权）。删除走 `status=3`，永不物理删除（guide §4 / 红线 §12）。

### D3 召回 = 规则打分 + 上限
- 取该宠物 active 记忆（上限 200 条），对当前输入做字符二元组（bigram）重合度打分
- 只保留得分 > 0 的条目，按「得分 → confidence → 最近使用时间」排序，取 TopN（默认 3）
- 召回后批量更新 `last_used_at`
- 相关性判定与排序全部在服务端，测试可断言顺序与数量

### D4 注入位置固定为上下文第 4 段
Gateway 已在 Prompt v1 的「# 长期记忆」段落渲染 `Memories`；本变更只需在 service 侧填值，不改 prompt 结构（保持 prompt 版本 `pet-chat-v1`）。

### D5 mock 必须能体现召回
否则「记得我说过什么」这条验收无法端到端验证。mock provider 在 `Memories` 非空时，回复中引用首条记忆内容（确定性模板），真实 provider 不受影响。

### D6 记忆形成对客户端可见
`SendMessageResult` 增加 `new_memories []MemoryItem`：聊天页可展示「记住了：…」轻量提示（MILESTONES W3「聊天页可展示轻量记忆形成提示」）。

### D7 错误码与脱敏
- 2401 记忆不存在或无权限
- 日志只记 user_id/pet_id/记忆数量/类型，不输出记忆正文与 prompt 内容（红线 §12）

## Risks / Trade-offs

- [规则抽取覆盖率有限] → 规则表集中一处、易扩展；真实 provider 接入后可叠加模型抽取，表结构无需变更
- [bigram 打分对长句噪声敏感] → 只取 Top3 且得分需 > 0；后续可换 TF-IDF 或向量检索（§13 已规划）
- [删除后又重新表达会复活记忆] → 视为用户重新授权，且在 UI 与文档中明示
- [召回更新 last_used_at 增加写放大] → 批量单条 UPDATE，条数受 TopN 限制

## Migration Plan

1. 新增 `pet_memories` migration（时间戳前缀，up/down 成对）
2. `docker compose up -d --build` 自动执行；回滚 = down 迁移 + 回退镜像
3. 里程碑提交：`feat(pet-memory): W3 记忆系统（抽取/召回/展示/删除）`

## Open Questions

（无——抽取与检索策略均为一期既定规则；模型抽取与向量检索明确后置）
