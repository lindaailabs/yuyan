## 1. 文档与数据库

- [x] 1.1 更新 docs/LLM_DEV_GUIDE.md：§4 friendships 状态常量补 `4=rejected`；§5.3 补 `POST /api/v1/friends/requests/{id}/reject` 端点（文档先行约束）
- [x] 1.2 migration `20260909100000_add_friendships`（up 建表含 uk_user_friend/idx_friend_status，down 删表）+ 真容器结构断言测试（复用 W2 模式）

## 2. 服务端（apps/server）

- [x] 2.1 `internal/model`：Friendship 实体（status 常量 1/2/3/4）+ DTO（申请项含申请人脱敏资料、好友项含好友脱敏资料）
- [x] 2.2 `internal/repo/friendship_repo`：CreateOrUpdateRequest（FOR UPDATE 读两行+决策落库）、Accept（事务：校验+置 accepted+upsert 反向行）、Reject（事务：校验+置 rejected）、ListPendingRequests（JOIN users，id DESC）、ListFriends（JOIN users，id ASC）；testcontainers 集成测试（uk 冲突、JOIN 结果）
- [x] 2.3 `internal/service/contacts_service`：SendRequest 防重复矩阵（2101~2107）、Accept/Reject（2106）、ListRequests/ListFriends（脱敏复用 user_service 逻辑）；单测覆盖全部分支 + accept 双向原子性断言 + reject 后重新申请
- [x] 2.4 `internal/api` handler 与路由：5 端点注册 /api/v1（Bearer 保护），binding 校验 + 统一错误转换；错误码 21xx 入 errcode 分段
- [x] 2.5 httptest：每端点 ≥1 happy + ≥1 error（1001 参数 / 21xx 业务 / 1002 未认证）
- [x] 2.6 make test-server 全绿 + golangci-lint 0 issue（本机未装 golangci-lint，以 `go vet ./...` 等价校验，0 issue）

## 3. Flutter App（apps/app）

- [x] 3.1 `data/repository/contacts_repository`：sendRequest/listRequests/accept/reject/listFriends（走 ApiClient）+ DTO model（FriendRequestItem/FriendItem）
- [x] 3.2 `core/providers` + ContactsController（StateNotifier）：加载好友/申请列表、accept 局部更新（requests 移除+friends 插入）、reject 移除、sendRequest 错误透传（ApiException.msg 上抛 UI）
- [x] 3.3 通讯录页 `/contacts`：好友卡片列表（AvatarWidget+昵称+脱敏手机号）、空态引导、失败重试
- [x] 3.4 申请列表页 `/requests`：申请卡片 + 同意/拒绝按钮（防抖禁用、失败 snackbar）
- [x] 3.5 搜索页「加好友」启用：成功→「已申请」禁用态；业务错误 snackbar 展示服务端 msg
- [x] 3.6 主页入口（通讯录/申请图标，申请角标）+ 路由注册 + l10n 文案补全
- [x] 3.7 ContactsController 单测（fake ApiClient：加载/accept/reject/sendRequest 全路径）+ flutter test 全绿（30 例）+ analyze 0 issue

## 4. 联调与收尾

- [x] 4.1 docker compose 全栈 API 冒烟（`scripts/smoke-contacts.ps1`，每次运行取随机号段可重复执行）：A→B 申请→B 列表可见→同意→双向好友→防重复矩阵错误码逐项验证（2101/2102/2103/2104/2105/2106 + 1001/1002）→拒绝路径→重新申请→中文昵称 utf8mb4 往返，全部通过；TEST-GAP：本机无 Android SDK，App 未做真机/模拟器 UI 走查（与 W2 一致，留待 W6 E2E 补齐）
- [x] 4.2 里程碑提交 `feat(contacts): W3 好友关系（申请/同意/拒绝/通讯录）` + push
