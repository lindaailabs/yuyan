# W3 好友关系（contacts）变更提案

## Why

W2 已完成账号体系（登录/资料/搜索），但用户之间尚无法建立好友关系。好友关系是单聊的前置条件（W4 消息链路以好友会话为主场景），也是一期闭环验收 §0.2（A 搜索 B → 发好友申请 → B 同意）的核心环节。

## What Changes

- 服务端新增 `friendships` 表 migration（双向各一行，SMALLINT 状态常量，uk_user_friend 唯一键）
- 服务端新增 5 个 REST 端点（挂 `/api/v1`，均需 Bearer 认证）：
  - `POST /friends/requests`：发起申请（不能加自己、防重复申请、已是好友/反向申请待处理时拒绝）
  - `GET /friends/requests`：收到的待处理申请列表（含申请人资料）
  - `POST /friends/requests/{id}/accept`：同意申请（事务内写双向两行）
  - `POST /friends/requests/{id}/reject`：拒绝申请（状态标记，不物理删除）——**guide §5.3 原缺失此端点，本变更同步补录**
  - `GET /friends`：好友列表（含好友资料，手机号脱敏）
- Flutter App 新增：通讯录页（好友列表）、好友申请列表页（同意/拒绝）、搜索页「加好友」按钮启用（W2 占位转真实调用）、主页入口调整
- 错误码补充：2xxx 业务段新增好友域错误（已是好友/重复申请/反向待处理/申请不存在或已处理/用户不存在）
- **文档更新（先于代码）**：`docs/LLM_DEV_GUIDE.md` §4 friendships 状态常量补 `4=rejected`；§5.3 补 `POST /api/v1/friends/requests/{id}/reject` 端点

## Capabilities

### New Capabilities

- `contacts-service`: 服务端好友域——申请/同意/拒绝/列表的完整链路，含 friendships 表结构、状态机、事务原子性与错误码契约
- `app-contacts`: 客户端好友域——通讯录页、申请列表页、搜索页加好友入口、好友相关状态管理与文案

### Modified Capabilities

（无——现有 auth/user-profile 规格不受影响；guide 文档更新属于本变更交付物，不改既有规格行为）

## Impact

- **数据库**：新增 `friendships` 表（utf8mb4/InnoDB，无破坏性变更）
- **服务端代码**：`internal/model`（friendship 实体/DTO）、`internal/repo`（friendship_repo + 事务）、`internal/service`（contacts_service + 校验规则）、`internal/api`（handler/路由/错误转换）、`internal/pkg/errcode`（新增业务码）
- **Flutter 代码**：`data/repository`（contacts_repository）、`features/contacts`（通讯录/申请列表/搜索页改造）、`features/home`（入口）、`core/l10n`（文案）、`core/providers`（装配）
- **文档**：`docs/LLM_DEV_GUIDE.md` §4/§5.3 增量更新
- **新依赖**：无（服务端与客户端均复用现有依赖）
- **测试**：repo/service 层事务原子性与防重复单测（testcontainers）；api 层每端点 happy + error httptest；客户端 contacts_controller 单测
