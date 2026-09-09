# W3 记忆系统（add-pet-memory）变更提案

## Why

W2 打通了「用户 ↔ 宠物」的文本对话，但宠物仍然"转头就忘"：陪伴闭环 §0.5（从对话抽取至少一条长期记忆，用户可查看并删除）与 §0.6（后续对话召回相关记忆并体现在回复中）都不成立。记忆是 AI 宠物与普通聊天壳的核心差异点，`pivot-to-ai-pet-mvp` 明确 W3/W4 不宜后置。

## What Changes

- 数据库：新增 `pet_memories` 表（migration up/down + 真容器结构断言）
- 服务端：
  - `internal/model`：PetMemory 实体、MemoryItem DTO、记忆类型与状态常量（1=active 2=archived 3=deleted）
  - `internal/repo/memory_repo`：upsert 写入（按内容哈希去重）、active 列表、归属查询、软删除、召回时间更新
  - `internal/service/memory_service`：规则抽取（偏好/称呼/画像/事件）、按相关度的少量召回（TopN）、用户列表、软删除
  - `internal/service/conversation_service`：发送时抽取记忆，调用 Gateway 前召回并注入 `Memories`
  - `internal/pkg/ai`：mock provider 在命中记忆时把记忆内容体现到回复中（便于验证召回链路）
  - `internal/api`：新增 `GET /api/v1/pets/{id}/memories`、`DELETE /api/v1/pet-memories/{id}`；错误码 24xx
- Flutter App：
  - 记忆模型与 repository、`MemoryController`（加载/删除/失败重试）、记忆页（列表 + 删除确认 + 空/错/加载态）
  - 路由 `/memories/:petId`，主页「记忆」入口由占位改为真实跳转
  - 聊天页在形成新记忆时展示轻量提示（`new_memories`）
- 脚本：`scripts/smoke-pet-memory.ps1` 全栈冒烟（说偏好 → 后续召回 → 删除后不召回）

## Capabilities

### New Capabilities

- `pet-memory`：长期记忆的抽取、召回、展示与用户可控删除

### Modified Capabilities

- `pet-conversation`：发送流程新增「抽取 → 召回 → 注入上下文」环节，发送结果携带 `new_memories`
- `ai-gateway`：上下文的「长期记忆」段落由真实记忆填充；mock provider 在命中记忆时体现到回复

## Impact

- **数据库**：新增 `pet_memories`（utf8mb4/InnoDB，纯新增）
- **服务端代码**：新增 memory 域的 model/repo/service/api；`conversation_service` 与 `ai/mock_provider` 改动；`router.go`/`main.go` 装配
- **Flutter 代码**：`data/model`、`data/repository`、`core/memory`、`features/memory`、`core/router`、`features/home`、`features/chat`、`core/l10n`
- **新依赖**：无（抽取与召回均为服务端确定性规则，符合 guide §13「一期用 MySQL + 规则筛选」）
- **测试**：抽取规则与去重、召回排序与上限、删除后不召回、越权；api 两端点 happy + error；客户端 controller/页面测试；全栈冒烟
- **不做**：向量检索、模型驱动的自动抽取（真实 provider 接入后扩展）、记忆编辑与审核后台、记忆容量权益（W5）
