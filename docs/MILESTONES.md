# 一期里程碑拆解（AI 宠物方向）

> 依据: [LLM_DEV_GUIDE.md](LLM_DEV_GUIDE.md)
> 定位: AI 宠物应用，一只有记忆、会成长、能对话的虚拟宠物
> 主线验收: 陪伴闭环，而不是人与人 IM 闭环

## 0. 闭环验收场景

一期完成的唯一标准：一个真实用户在一台手机或模拟器上完成以下流程。

1. 用户注册登录，完成昵称和头像设置。
2. 用户创建或领养一只宠物，设置名字、外观和初始性格。
3. 用户进入宠物主页，能看到宠物状态、心情、等级和亲密度。
4. 用户发送文本消息，服务端通过 AI Gateway 生成回复，App 能展示发送中、流式/返回中、成功和失败状态。
5. 宠物从对话中抽取至少一条长期记忆，用户能在记忆页查看并删除。
6. 后续对话能召回相关记忆，并体现在宠物回复中。
7. 多轮互动后，宠物成长状态发生可解释变化，生成成长事件。
8. 用户杀进程重开后，宠物档案、历史消息、记忆和成长状态完整恢复。
9. 服务端记录每轮 AI 调用的模型、token、耗时、错误码和成本统计字段。
10. 权益模型可表达免费/订阅用户的额度差异，真实支付接入可后置。

## W0 定位切换与遗留资产盘点

目标：把项目从早期 IM 路线切到 AI 宠物路线，避免继续扩展旧主线。

- 更新 README、LLM_DEV_GUIDE、MILESTONES、OpenSpec context。
- 标记好友/通讯录/人与人单聊为非一期主线。
- 盘点可复用资产：账号、JWT、验证码、Dio、Riverpod、drift、WebSocket、消息分页、错误码、测试框架。
- 新建 OpenSpec change：`pivot-to-ai-pet-mvp`，覆盖 pet、memory、ai-gateway、entitlement、analytics。
- 决定旧 friendship 代码策略：保留但不扩展，或后续迁移为隐藏模块。

验收：后续任务模板均以 AI 宠物为主线；没有新任务继续要求实现好友闭环。

## W1 账号与宠物档案

目标：用户能登录并拥有一只宠物。

服务端：

- 复用现有 auth/user 能力。
- migration：`pets` 表，必要时补 `users.avatar_id` 兼容字段。
- REST：`POST /api/v1/pets`、`GET /api/v1/pets/{id}`、`PUT /api/v1/pets/{id}`、`GET /api/v1/pets/{id}/state`。
- service：创建宠物时初始化等级、亲密度、心情、persona。
- 测试：创建宠物、重复创建策略、越权访问、状态默认值。

App：

- 宠物创建/领养页。
- 宠物主页壳：头像/形象、名字、心情、等级、亲密度。
- provider/repository 分层，不在 UI 直接调 Dio。

验收：新用户登录后可创建宠物；重启 App 后宠物信息恢复。

## W2 文本对话与 AI Gateway

目标：用户能与宠物进行稳定文本对话，模型调用通过统一网关。

服务端：

- migration：`pet_conversations`、`pet_messages`、`ai_call_logs`。
- AI Gateway 接口：`CompletePetReply(ctx, input) -> output`，真实模型与 mock 实现解耦。
- Prompt v1：系统规则、宠物 persona、成长状态、最近对话。
- REST：`POST /api/v1/pet-conversations`、`GET /api/v1/pet-messages`、`POST /api/v1/pet-messages`。
- 可靠性：`client_msg_id` 幂等；消息 ID 游标分页；失败写状态和错误码。
- 统计：模型、token、延迟、错误、缓存命中字段。

App：

- 宠物聊天页。
- 本地 drift 缓存消息。
- 发送状态机：pending -> sent/failed，失败可重试同 `client_msg_id`。
- 打开聊天先读本地，再拉服务端增量。

验收：连续 20 轮文本对话无重复消息；模型 mock 测试可稳定通过；真实模型调用可通过配置开关启用。

状态（W2 已交付，见 `openspec/changes/add-ai-gateway-pet-chat`）：

- `pet_conversations` / `pet_messages` / `ai_call_logs` 三张表与 migration。
- `internal/pkg/ai`：Gateway 接口 + 默认 mock provider（`AI_PROVIDER`）+ Prompt v1（分层与双上限）+ 超时重试与兜底。
- REST：`POST /pet-conversations`、`POST /pet-messages`、`GET /pet-messages`（游标分页）。
- App：聊天页 + drift 本地消息缓存（本地优先 + 增量同步）+ 发送中/失败重试状态机。
- 全栈冒烟 `scripts/smoke-pet-chat.ps1`：20 轮无重复、重放幂等、游标分页不重不漏、错误码 2301~2303/1001/1002。
- TEST-GAP：本地库当前为进程内内存库（未接入 path_provider），跨进程恢复由服务端历史拉取兜底；真机/模拟器 UI 走查因本机无 Android SDK 未执行。

## W3 记忆系统

目标：宠物能记住重要信息，并让用户可控。

服务端：

- migration：`pet_memories`。
- 记忆抽取：从对话中识别用户偏好、称呼、重要事件、宠物关系进展。
- 记忆召回：按当前输入、记忆类型、最近使用时间和置信度筛选。
- 用户控制：`GET /api/v1/pets/{id}/memories`、`DELETE /api/v1/pet-memories/{id}`。
- 日志脱敏：prompt 和记忆日志不得泄漏 token、完整手机号等敏感信息。

App：

- 记忆页：列表、删除、空态。
- 聊天页可展示轻量记忆形成提示。

验收：用户说“我喜欢蓝色”，后续问“你记得我喜欢什么颜色吗”能回答；用户删除该记忆后不能再召回。

## W4 成长系统与每日回访

目标：宠物不是聊天壳，而是会随互动成长。

服务端：

- migration：`pet_growth_events`，必要时补 `pet_daily_stats`。
- 成长规则：亲密度、等级、心情、连续互动、冷却时间。
- 成长事件：升级、关系阶段变化、心情变化、纪念日。
- 规则必须确定性，可测试；大模型只负责表达，不直接决定核心数值。

App：

- 宠物主页展示成长状态。
- 成长事件时间线。
- 次日回访时展示自然的状态变化。

验收：多轮互动触发成长事件；重复请求不会重复加经验；服务端测试覆盖边界。

## W5 权益、支付地基与数据看板

目标：产品具备商业化和运营观测的最小地基。

服务端：

- migration：`entitlements`、`usage_counters`、`event_logs`。
- 权益：免费额度、订阅额度、记忆容量、语音/高级模型开关预留。
- 支付：预留 Apple IAP / Google Play Billing / 微信 / 支付宝收据或回调接口，不在一期硬接生产支付也要保证模型正确。
- 埋点：激活、注册、创建宠物、首轮对话、记忆生成、次日回访、订阅页曝光。
- 看板出口：基础 SQL 或 REST 管理接口，先满足内部观察。

App：

- 权益状态展示。
- 订阅页壳和额度耗尽提示。
- 埋点上报客户端事件。

验收：服务端能按用户返回权益和剩余额度；AI 对话能记录和检查额度；关键事件可查询。

## W6 联调、体验打磨与语音 POC

目标：把陪伴闭环跑顺，并验证下一阶段语音风险。

- 全链路 E2E：按 §0 场景完整走查。
- 性能：并发对话、超时、重试、限流、成本统计压测。
- 体验：宠物主页、聊天页、记忆页、成长事件页视觉和状态完整。
- 语音 POC：只做实验性短语音 ASR/TTS 或实时语音 demo，不作为一期生产验收。
- 文档：README、接口契约、环境变量、测试说明补齐。

验收：陪伴闭环全绿；测试全绿；AI Gateway mock 和真实配置路径都可跑；语音风险清单明确。

## 风险与建议

| 风险 | 影响 | 建议 |
|---|---|---|
| 继续沿 IM 主线开发 | 资源被好友/群聊消耗，宠物差异化不足 | 冻结好友需求，改做宠物会话 |
| 过早做实时语音 | 成本、延迟、权限、弱网体验风险高 | 先文本，再短语音，最后实时语音 |
| 记忆不可控 | 用户不信任，隐私风险高 | 记忆可查看、可删除、可追溯来源 |
| 成长全靠模型生成 | 状态漂移，付费权益难控制 | 服务端规则决定数值，模型负责表达 |
| AI 成本不可见 | 上线后毛利失控 | W2 起记录 token、模型、延迟、错误 |
| 支付只做客户端 | 容易被绕过，跨端状态不一致 | 服务端 entitlement 为唯一事实源 |
