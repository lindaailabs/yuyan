## 1. 文档与变更

- [x] 1.1 新增 `add-pet-memory` change：proposal / design / tasks / specs（pet-memory）
- [x] 1.2 更新 docs/MILESTONES.md：W3 交付状态与 TEST-GAP 留档

## 2. 服务端（apps/server）

- [x] 2.1 migration `20260911090000_add_pet_memories`（up/down，`uk_pet_hash`、`idx_pet_status`）
- [x] 2.2 `internal/model/memory.go`：PetMemory 实体、MemoryItem DTO、类型与状态常量（1=active 2=archived 3=deleted）
- [x] 2.3 `internal/repo/memory_repo`：upsert（内容哈希去重 + 复活）、active 列表、归属查询、软删除、last_used_at 更新
- [x] 2.4 `internal/service/memory_service`：规则抽取（偏好/称呼/画像/事件，含疑问句过滤）、Jaccard 打分召回（TopN=3）、列表、软删除；单测覆盖规则/去重/排序/删除后不召回/越权/last_used_at
- [x] 2.5 对话链路接入：发送后抽取、Gateway 前召回注入；`SendMessageResult` 增加 `new_memories`
- [x] 2.6 `internal/pkg/ai`：mock provider 命中记忆时把记忆体现到回复；prompt 段落与版本保持不变
- [x] 2.7 `internal/api/handler_memory` + 路由注册 + `main.go` 装配；错误码 2401
- [x] 2.8 测试：service 单测 + api 两端点 happy/error httptest；`go test ./...` 全绿 + `go vet` 0 issue（golangci-lint 本机未安装）

## 3. Flutter App（apps/app）

- [x] 3.1 `data/model/pet_memory.dart` 与 `data/repository/pet_memory_repository.dart`（列表、删除；ApiClient 新增 delete）
- [x] 3.2 `core/memory/memory_controller.dart`：加载、删除、失败重试与局部移除
- [x] 3.3 `features/memory/memory_page.dart`：列表 + 删除确认 + 空/错/加载态
- [x] 3.4 路由 `/memories/:petId` + 主页「记忆」入口接真实跳转 + 聊天页新记忆提示 + l10n 文案 + providers 装配
- [x] 3.5 测试：controller 5 例、记忆页 widget 3 例；`flutter test` 55 例全绿 + `flutter analyze` 0 issue

## 4. 联调与收尾

- [x] 4.1 `scripts/smoke-pet-memory.ps1`：说偏好 → 形成记忆（含来源与置信度）→ 后续召回并体现 → 删除后列表不可见且不再召回 → 错误码 2401/1001/1002；W2 对话冒烟回归通过
- [ ] 4.2 里程碑提交 + push
