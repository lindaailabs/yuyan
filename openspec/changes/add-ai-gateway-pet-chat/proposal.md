# W2 文本对话与 AI Gateway（add-ai-gateway-pet-chat）变更提案

## Why

W1 交付了宠物档案与宠物主页，但用户还不能和宠物说话：陪伴闭环 §0.4（发送文本 → AI 生成回复 → App 展示发送中/成功/失败）与 §0.8（杀进程重开后历史与状态恢复）目前都不成立。

同时定位文档已把 AI 链路定为核心约束：`docs/LLM_DEV_GUIDE.md` v2.0 §5 要求「AI Gateway 是模型调用的唯一入口，业务 service 不得直接调用具体模型 SDK」，§10 要求「模型调用必须可 mock，测试不得依赖真实模型服务」，`openspec/changes/pivot-to-ai-pet-mvp` D3 同样明确 Gateway 是唯一入口。因此在实现对话之前，必须先把 Gateway 抽象与 mock provider 落地。

## What Changes

- 数据库：新增 `pet_conversations`、`pet_messages`、`ai_call_logs` 三张表（migration up/down + 真容器结构断言）
- 服务端：
  - `internal/pkg/ai`：Gateway 接口、默认 `mock` provider、Prompt v1 组装（上限控制与脱敏）
  - `internal/model`：会话/消息/调用日志实体与 DTO（role、status 常量）
  - `internal/repo`：conversation_repo（GetOrCreate / 更新 last_msg）、message_repo（幂等插入、游标分页、失败标记）、ai_call_log_repo
  - `internal/service/conversation_service`：发送编排（宠物归属校验 → 幂等 → Gateway 调用 → 落 assistant 消息 → 会话更新 → 用量落库）、历史游标分页、会话创建或获取；错误码 23xx 段
  - `internal/api/handler_conversation`：3 个 REST 端点 + binding 校验；`router.go` 注册；`main.go` 装配 Gateway
  - `internal/pkg/config`：`AI_PROVIDER`（默认 `mock`）、`AI_TIMEOUT_MS`、`AI_MOCK_FAIL_RATE`
- Flutter App：
  - drift 本地消息表（schemaVersion 2）+ DAO
  - `PetChatRepository`：发送（携带 `client_msg_id`）、本地优先 + 服务端增量合并去重
  - `ChatController`（StateNotifier）：加载 / 发送中 / 成功 / 失败重试状态机
  - 聊天页 `/chat`（气泡列表、输入区、空/错/加载态）、主页「和它聊天」入口、路由注册、l10n 文案
- 脚本：`scripts/smoke-pet-chat.ps1` 全栈冒烟（登录 → 建宠 → 建会话 → 20 轮对话 → 无重复 → 游标分页）

## Capabilities

### New Capabilities

- `ai-gateway`：模型调用唯一入口——Provider 抽象、默认 mock、Prompt v1 拼装与上限、超时重试与兜底、用量统计字段
- `pet-conversation`：用户与宠物的文本对话——会话创建或获取、消息持久化与幂等、游标分页历史、失败状态与重试

### Modified Capabilities

- （无——宠物档案、账号、好友/IM 遗留能力均不在本变更内修改；`pets` 表与本变更无交集）

## Impact

- **数据库**：3 张新表（utf8mb4/InnoDB，纯新增，无破坏性变更）
- **服务端代码**：新增 `internal/pkg/ai` 包与 conversation 域的 model/repo/service/api；`router.go` 与 `cmd/server/main.go` 装配改动；`config.go` 新增 AI 配置项
- **Flutter 代码**：`data/local`（drift 升版本）、`data/model`、`data/repository`、`core/chat`、`features/chat`、`core/router`、`features/home`、`core/providers`、`core/l10n`
- **新依赖**：无（Gateway 用标准库实现；drift、dio、riverpod 均已在依赖中）
- **测试**：Gateway 单测（成功/失败/超时/拼装上限）、service 单测（幂等/越权/兜底/分页/用量，全部 mock 无网络）、api 每端点 happy + error httptest、表结构容器断言、客户端 controller/DAO/widget 测试、全栈冒烟脚本
- **不做**：记忆（W3）、成长数值（W4）、权益额度（W5）、流式输出（protocol-v2）、真实模型接入、任何好友/IM 主线扩展
