## 1. 服务端（apps/server）

- [x] 1.1 migration `20260908_add_users`（up/down SQL，guide §4 基线 + avatar_id）+ 真容器结构断言测试
- [x] 1.2 `internal/pkg/jwt`：HS256 签发/解析（claims uid/exp/typ），JWT_SECRET env 注入；单测（过期/篡改/typ 混用）
- [x] 1.3 `internal/repo/captcha_repo`：Set/Get/Delete（sms:code:{phone} TTL 5min）+ 限频（sms:freq:{phone} TTL 60s）；miniredis 单测
- [x] 1.4 `internal/repo/user_repo`：FindByPhone/CreateUser/GetByID/UpdateProfile/SearchByPhone；testcontainers 集成测试
- [x] 1.5 `internal/service/auth_service`：SendSmsCode（限频+base64Captcha 图形渲染）、Login（校验+一次性删除+自动注册+双 token 签发）、Refresh；单测（含错误路径全覆盖，覆盖率≥70%）
- [x] 1.6 `internal/service/user_service`：Profile/UpdateProfile（nickname 1~20、avatar_id 1~8 校验）/Search（脱敏）；单测
- [x] 1.7 `internal/api/middleware_auth.go`：Bearer 解析、typ=access 校验、uid 注入 context
- [x] 1.8 `internal/api` handler 与路由：5 组端点注册到 /api/v1，binding tag 校验；统一错误转换
- [x] 1.9 httptest：每端点 ≥1 happy + ≥1 error（1001 参数 / 2xxx 业务 / 1002 认证）
- [x] 1.10 make test-server 全绿 + golangci-lint 0 issue

## 2. Flutter App（apps/app）

- [x] 2.1 `core/auth/`：SecureTokenStorage（flutter_secure_storage）、AuthState（未登录/未引导/就绪）、authStateProvider（启动恢复 token+拉取 me）
- [x] 2.2 `data/remote/api_client.dart` 实现：Dio 封装 + 401 拦截器（refresh 一次、重放、失败登出）+ 统一包裹解析；单测（mock adapter）
- [x] 2.3 `data/repository/auth_repository.dart`：sendSmsCode/login/refresh/getMe/updateMe/search（实现 ApiClient 接口）
- [x] 2.4 8 个预置头像 assets（scripts/ 生成占位图）+ AvatarWidget（id→asset 映射）
- [x] 2.5 `features/auth/login_page`：手机号+验证码图片（点击刷新）+ 短信兼容态 + 提交登录
- [x] 2.6 `features/auth/onboarding_page`：8 头像选择 + 昵称输入 + 提交
- [x] 2.7 `features/profile/profile_page`：展示 + 编辑昵称/头像
- [x] 2.8 `features/contacts/search_page`：手机号搜索 + 结果卡片（加好友按钮禁用占位）
- [x] 2.9 路由三态守卫（go_router redirect 纯函数 + 单测）+ core/l10n 文案补全
- [x] 2.10 flutter test（token 管理/拦截器/状态机/三态路由，19 例全绿）+ flutter analyze 0 issue

## 3. 联调与收尾

- [x] 3.1 docker compose 全栈起 + API 级冒烟（重建 server 镜像后）：验证码下发→Redis 存码→登录自动注册→me（nickname 空）→PUT 引导→B 登录→搜索 A（脱敏 138****1111）→refresh 换新 token；错误路径（无 token/篡改 token→401，错码→2002，非法手机号→1001，限频→2001）全数符合；中文昵称 utf8mb4 往返正确（HEX=E8AFAD）。客户端「重启保持会话」由单测覆盖（token 持久化+refresh 流程）；Android 模拟器 UI 走查因本机无 Android SDK 仍为 TEST-GAP
- [x] 3.2 PR 材料：文件清单、新依赖声明（golang-jwt/v5、base64Captcha、flutter_secure_storage、dio、flutter_riverpod、go_router）、测试清单
- [ ] 3.3 里程碑提交（feat(account): W2 账号体系）+ push（网络允许时）
