## ADDED Requirements

### Requirement: 会话创建或获取

`POST /api/v1/pet-conversations` SHALL 为「当前用户 + 指定宠物」创建或返回既有会话；宠物不属于当前用户时 SHALL 返回业务错误（2303）。

#### Scenario: 首次对话自动建会话

- **WHEN** 用户首次对某只自己的宠物发起对话
- **THEN** 系统 SHALL 创建会话并返回会话 id
- **AND** 重复调用 SHALL 返回同一会话（不产生重复会话）

### Requirement: 发送消息与 AI 回复

`POST /api/v1/pet-messages` SHALL 持久化用户消息，经 AI Gateway 生成宠物回复，并返回双方消息；请求 MUST 携带 `client_msg_id`（UUID）用于幂等。

#### Scenario: 正常一轮对话

- **WHEN** 用户发送合法文本
- **THEN** 系统 SHALL 持久化一条 role=user 消息与一条 role=assistant 消息
- **AND** 响应 SHALL 返回两条消息（含 id、role、content、status、created_at）与用量字段

#### Scenario: 重复提交幂等

- **WHEN** 同一 `client_msg_id` 被重复提交（重试或并发）
- **THEN** 系统 SHALL 返回首次结果
- **AND** 会话中 SHALL NOT 产生第二条相同的用户消息或第二条回复

#### Scenario: 内容非法

- **WHEN** 消息内容为空或超过长度上限
- **THEN** 系统 SHALL 返回 2302，且不写入任何消息行

#### Scenario: 越权访问

- **WHEN** 目标宠物不属于当前登录用户
- **THEN** 系统 SHALL 返回 2303，且不写入任何消息行

### Requirement: AI 失败的兜底

AI 调用失败或超时时，系统 SHALL 保留用户消息，并写入一条 `status=failed` 的宠物消息与可展示的兜底文案，同时在 `ai_call_logs` 记录错误码。

#### Scenario: 模型不可用

- **WHEN** Gateway 返回错误
- **THEN** 用户消息 SHALL 保持成功状态
- **AND** 宠物消息 SHALL 标记为失败并携带兜底文案
- **AND** 客户端 SHALL 可携带同一 `client_msg_id` 重试

### Requirement: 历史消息游标分页

`GET /api/v1/pet-messages?conv_id=&cursor=&limit=20` SHALL 按消息 id 升序返回游标分页结果，禁止 offset 或时间戳排序；越权访问他人会话 SHALL 返回 2301。

#### Scenario: 翻页不重不漏

- **WHEN** 会话存在 25 条消息且 limit=20
- **THEN** 首次返回 20 条与 `has_more=true`、`next_cursor` 为末条 id
- **AND** 第二页返回剩余 5 条且与第一页无交集

#### Scenario: 空会话

- **WHEN** 会话尚无消息
- **THEN** 返回空列表与 `has_more=false`

### Requirement: 本地缓存与恢复

App SHALL 在本地缓存会话消息：进入聊天先渲染本地数据，再按游标增量拉取并去重合并；杀进程重开后 SHALL 立即恢复历史记录。

#### Scenario: 重启恢复

- **WHEN** 用户杀进程后重新进入聊天页
- **THEN** 页面 SHALL 立即展示本地缓存的消息
- **AND** 随后 SHALL 与服务端增量同步且不产生重复气泡
