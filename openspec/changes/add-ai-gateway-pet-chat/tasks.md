## 1. 文档与变更

- [x] 1.1 新增 `add-ai-gateway-pet-chat` change：proposal / design / tasks / specs（ai-gateway、pet-conversation）
- [x] 1.2 更新 docs/MILESTONES.md：W2 交付状态与 TEST-GAP 留档

## 2. 服务端（apps/server）

- [x] 2.1 migration `20260910100000_add_pet_conversations`（up/down，`uk_user_pet`、`idx_user_updated`）
- [x] 2.2 migration `20260910110000_add_pet_messages`（up/down，`uk_client_msg_id`、`idx_conv_id_id`、`created_at BIGINT`）
- [x] 2.3 migration `20260910120000_add_ai_call_logs`（up/down，model/tokens/latency/err_code/cache_hit）
- [x] 2.4 表结构真容器断言测试（索引走 information_schema，校验 uk_client_msg_id 幂等约束）
- [x] 2.5 `internal/pkg/ai`：Gateway 接口 + mock provider + Prompt v1（分层 + 轮次/字符双上限）+ 单测（默认 mock、未实现 provider 明确失败、重试、超时不重试、失败 err_code、拼装上限、确定性）
- [x] 2.6 `internal/pkg/config`：`AI_PROVIDER`（默认 mock）、`AI_TIMEOUT_MS`、`AI_MOCK_FAIL_RATE`，SafeString 脱敏
- [x] 2.7 `internal/model`：PetConversation / PetMessage / AICallLog 实体与 DTO、role/status 常量
- [x] 2.8 `internal/repo`：conversation_repo（GetOrCreate/更新 last_msg）、message_repo（幂等插入/游标分页/失败标记）、ai_call_log_repo（含 FindByMsgID）
- [x] 2.9 `internal/service/conversation_service`：发送编排（归属校验→幂等→Gateway→落库→用量）、历史游标分页、会话创建或获取；23xx 错误码
- [x] 2.10 `internal/api/handler_conversation` + 路由注册 + `main.go` 装配 Gateway（provider 未配置启动即失败）
- [x] 2.11 测试：service 单测（会话幂等/分页不重不漏/内容校验/越权/AI 失败兜底与日志/历史越权）+ api 每端点 happy/error httptest（1001/1002/2301/2302/2303、幂等重放）
- [x] 2.12 `go test ./...` 全绿（含 testcontainers）+ `go vet ./...` 0 issue；golangci-lint 本机未安装，以 go vet 等价校验

## 3. Flutter App（apps/app）

- [x] 3.1 drift 本地消息表 + DAO（schemaVersion 2，含迁移）+ build_runner 生成
- [x] 3.2 `data/model/pet_message.dart`：消息模型、分页结果、发送结果、用量
- [x] 3.3 `data/repository/pet_chat_repository.dart`：发送（带 client_msg_id）、本地优先 + 增量合并去重、失败态回填
- [x] 3.4 `core/chat/chat_controller.dart`：加载 / 发送中 / 成功 / 失败重试状态机（重试复用同一 client_msg_id）
- [x] 3.5 `features/chat/chat_page.dart`：气泡列表 + 输入区 + 空/错/加载态 + 继续加载（暖色陪伴风 UI）
- [x] 3.6 路由 `/chat/:petId` 注册 + 主页「和它聊天」入口 + l10n 文案 + providers 装配
- [x] 3.7 测试：controller 状态机（6 例）、DAO 内存库（4 例）、聊天页关键 widget（3 例）；`flutter test` 47 例全绿 + `flutter analyze` 0 issue

## 4. 联调与收尾

- [x] 4.1 `scripts/smoke-pet-chat.ps1` 全栈冒烟：登录→建宠→会话幂等→20 轮对话无重复→重放幂等→游标分页不重不漏（42 条）→用量字段→错误码 2301/2302/2303/1001/1002，全部通过
- [x] 4.2 里程碑提交 + push
