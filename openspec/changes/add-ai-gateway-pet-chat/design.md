# W2 文本对话与 AI Gateway 设计

## Context

W1 已交付宠物档案与宠物主页（`pets` 表 + 创建/列表/详情/更新/状态）。下一步是陪伴闭环的主干：用户发一句话，宠物回一句话，且重启后记录还在、失败可重试。

权威约束（`docs/LLM_DEV_GUIDE.md` v2.0）：§4 数据模型基线（消息 ID 兼游标、禁止 offset、`client_msg_id` 幂等、AI 用量必记）；§5 AI 链路（Gateway 唯一入口、上下文分层与上限、统计与脱敏、重试超时兜底）；§9 编码规范（分层、slog 脱敏、聊天页需发送中/失败重试/本地缓存恢复）；§10 测试（AI 调用必须可 mock）；§12 红线（不扩展好友 IM、不用 offset 分页、模型不决定核心数值、日志不含隐私）。

补充来源：`pivot-to-ai-pet-mvp` D3（Gateway 唯一入口）、D4（成长数值由服务端规则决定）；`pivot-to-ai-pet-mvp/specs/ai-pet-mvp` 的 `AI Gateway Boundary` 与 `Companion Loop`；`docs/MILESTONES.md` W2。

## Goals / Non-Goals

**Goals**
- 用户能对属于自己的宠物发送文本，服务端经 AI Gateway 生成宠物回复并持久化双方消息
- 同一 `client_msg_id` 重复提交不产生重复消息（幂等）
- 历史消息按游标分页（禁 offset），翻页不重不漏
- AI 失败/超时有兜底文案与失败状态，客户端可原样重试
- 每次调用落 `ai_call_logs`（模型、token、耗时、错误码、缓存命中）
- 客户端本地缓存优先，杀进程重开先渲染本地再增量同步

**Non-Goals（本变更明确不做）**
- 记忆抽取与召回（W3）、成长数值变更（W4）、权益与额度校验（W5）
- 流式输出：非流式返回完整回复，DTO 预留 `streaming` 标记，流式留待 `protocol-v2` 变更
- 真实模型供应商接入：只留 `AI_PROVIDER` 开关，未实现值返回明确错误
- 好友/通讯录/人与人聊天等旧 IM 主线的任何扩展（现有代码不动）
- 语音、图片、文件消息

## Decisions

### D1 Gateway 落位 `internal/pkg/ai`（基础设施）而非 service 域
Gateway 负责模型适配、超时重试、脱敏与统计，是基础设施；业务 service 只依赖其接口。这样既满足 §5「业务 service 不得直连供应商 SDK」，也避免 service 之间互相依赖。

```go
type Gateway interface {
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResult, error)
}
```

### D2 默认 mock，真实模型只留开关
- `AI_PROVIDER` 默认 `mock`（guide §10：测试不得依赖真实模型服务）
- mock 为确定性回复（含宠物名与输入回指，便于断言），并可注入失败/超时以验证兜底路径
- 切到未实现 provider 时返回明确错误（`ErrProviderNotConfigured`），**不静默降级**为 mock

### D3 AI 调用不进事务
- 事务 A（短）：幂等插入 user 消息 + 会话 GetOrCreate/更新
- 事务外：Gateway 调用（带超时 + 一次重试）
- 事务 B（短）：写 assistant 消息（含 model/tokens/latency）+ 更新会话 last_msg + 写 `ai_call_logs`
- 理由：避免模型 RT（可达数秒）长时间占用 DB 连接与行锁

### D4 幂等靠唯一键，而非先查后插
`uk_client_msg_id` 冲突即视为重放，直接返回首次结果；并发安全，无需额外锁。

### D5 严格游标分页
`WHERE conv_id = ? AND id > cursor ORDER BY id ASC LIMIT n`，响应返回 `items` + `next_cursor` + `has_more`；禁止 offset 与时间戳排序（§12 红线）。

### D6 Prompt v1 分层与上限
系统规则 → persona → 成长状态 → 长期记忆（本变更为空） → 最近 N 轮（**最近 10 轮且累计 ≤ 2000 字符**）→ 当前输入。日志只记各段长度与是否裁剪，**不输出 prompt 正文**（§12 红线）。

### D7 失败可解释、可重试
- Gateway 失败：user 消息保留（`status=1`），写 assistant 消息 `status=failed` + 兜底文案，并在 `ai_call_logs` 记录错误码
- 内容非法/超长/越权：直接返回业务错误码，不产生消息行
- 客户端失败气泡可点击重试，复用同一 `client_msg_id`

### D8 客户端本地优先
drift 升到 schemaVersion 2，新增本地消息表；进入聊天先读本地渲染，再按 `cursor` 增量拉取并按 `client_msg_id`/服务端 id 去重合并。

### D9 错误码段
宠物对话域使用 **23xx**：`2301` 会话不存在或无权限、`2302` 消息内容非法、`2303` 宠物不存在或无权限、`2304` AI 服务未配置、`2305` AI 调用失败（有兜底文案）。

## Risks / Trade-offs

- [mock 与真实模型表现差异] → 接口与 DTO 按真实模型字段设计（model/tokens/latency/cache_hit），换 provider 不改契约；真实接入另开变更
- [非流式首字延迟体感] → 客户端以「宠物正在思考」占位；流式在 `protocol-v2` 变更补齐
- [本地缓存与服务端不一致] → 以服务端 id 与 `client_msg_id` 为去重键，本地只做渲染加速，最终以服务端为准
- [长对话上下文膨胀] → Prompt 双上限（轮次 + 字符）裁剪；摘要与记忆召回留 W3
- [失败消息与成功消息混排] → 消息行带 `status`，客户端按状态渲染重试入口

## Migration Plan

1. 新增三张表 migration（时间戳前缀，up/down 成对）
2. `docker compose up -d --build` 重启 server 自动执行 migration；回滚 = down 迁移 + 回退镜像
3. 里程碑提交：`feat(pet-chat): W2 文本对话与 AI Gateway（mock provider + 消息持久化）`

## Open Questions

（无——流式、真实模型、记忆、成长、权益均已明确后置，且有对应后续变更入口）
