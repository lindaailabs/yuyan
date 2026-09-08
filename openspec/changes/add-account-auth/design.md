## Context

W1 骨架已就绪（分层、errcode、trace_id、migration 机制、docker-compose、Flutter 壳）。W2 在其上落第一个业务域：账号。产品决策（docs/MILESTONES.md 已决）：图形验证码代替短信、8 个预置头像、登录即注册。团队约束：Java 转型 Go，写法朴素优先（guide §6.1）。

## Goals / Non-Goals

**Goals:**
- 验证码→登录/注册→JWT→资料→搜索 全链路可用，双端贯通
- 首登引导（昵称+头像）闭环，二次登录跳过引导
- service 层单测覆盖率 ≥70%，每端点 happy+error httptest（guide §7）

**Non-Goals:**
- 真实短信对接（生产切换只改 sms-code 内部）
- 好友/会话/消息（W3/W4）；头像上传（无对象存储）
- refresh token 吊销黑名单（一期单机无高安全诉求，30d 自然过期）

## Decisions

### D1. 图形验证码：base64Captcha
- **选择**：`github.com/mojocn/base64Captcha`（生成 6 位数字验证码图片，base64 输出）
- **理由**：单依赖、零配置、社区事实标准；验证码本身即登录凭证，数字内容与图片同源下发（一期不对接短信时无需分离"码"与"图"）
- **存储**：Redis `sms:code:{phone}` = 6 位数字（明文足够，TTL 5min）；限频键 `sms:freq:{phone}` TTL 60s（存在即拒绝）
- **注意**：生产接短信时，响应不再携带图片、验证码改为经短信下发——响应结构预留 `captcha_image`（base64，可缺省）字段，客户端两态兼容（MILESTONES W2 既定）

### D2. JWT：golang-jwt/jwt/v5，HS256 + env 密钥
- **选择**：HS256 对称签名；密钥 `JWT_SECRET` 环境变量注入（compose 传 dev 默认值，禁代码硬编码 guide §9.3）
- **claims**：`uid`、`exp`；refresh token 另加 `typ=refresh` 区分
- **有效期**：access 2h、refresh 30d（guide §5.3）
- **middleware**：`Authorization: Bearer <token>` 解析，uid 注入 gin context；`typ=refresh` 的 token 不能调业务端点（显式校验 typ）
- **备选**：RS256 非对称（否——一期单体自签自验，无多服务验证诉求）

### D3. 自动注册与首登判定
- login 时 phone 不存在→INSERT（nickname=NULL、avatar_id=1）
- 首登判定：`GET /users/me` 返回 `nickname` 为 null → App 跳引导页；引导页提交 `PUT /users/me`（nickname+avatar_id）后进入主页
- 昵称规则：1~20 字符（中文按字符计）；avatar_id ∈ 1~8，非法值 2xxx 业务错误
- **选 NULL 而非默认昵称**：避免"未引导用户"与"主动改回默认名用户"不可区分

### D4. users 表 migration（guide §4 基线 + avatar_id 决策）
```sql
-- 20260908_add_users.up.sql
CREATE TABLE users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  phone VARCHAR(20) NOT NULL UNIQUE,
  nickname VARCHAR(20) NULL,
  avatar_id SMALLINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```
- created_at/updated_at 用 DATETIME（DB 维护），message 时间戳用 BIGINT 是 §4 特例（时序游标），users 表不涉及时序游标——遵循 §4 原始定义

### D5. 分层落位（guide §3.3）
- `internal/repo`：user_repo（FindPhone/CreateUser/GetByID/UpdateProfile/SearchByPhone）
- `internal/service`：auth_service（SendSmsCode/VerifyLogin/RefreshToken）、user_service（Profile/UpdateProfile/Search）
- `internal/api`：auth_handler、user_handler、middleware_auth；路由按 `/api/v1` 前缀注册
- 验证码/限频/Redis 访问封装在 `repo/captcha_repo`（service 层唯一入口）

### D6. App 侧结构
- `features/auth/`：login_page（手机号+验证码图片+输入，image 缺省时提示等待短信）、onboarding_page（8 头像选择+昵称输入）
- `features/profile/`：profile_page（查看/编辑昵称头像）
- `features/contacts/`：search_page（手机号搜索，结果展示昵称+头像，"加好友"W3 再启用）
- `core/auth/`：token 存储（flutter_secure_storage）、Dio 拦截器（401→refresh→重放→失败登出）、authStateProvider（未登录/已登录未引导/已就绪三态驱动路由）
- 路由守卫：go_router redirect 按三态分流（login→onboarding→home）
- 预置头像：assets/avatars/avatar_1.png~avatar_8.png（W2 先用纯色+数字占位图，正式插画后续替换；生成脚本入 scripts/）

### D7. 测试策略（guide §7）
- service 单测：sqlmock GORM 较繁琐，改用 testcontainers MySQL（W1 已验证可行）+ 真实 Redis 容器；captcha_repo 用 miniredis（避免测试依赖外部队列）
- api httptest：每端点 ≥1 happy + ≥1 error（参数缺失/验证码错误/token 过期）
- Flutter：token 管理与 authStateProvider 单测（mock secure storage）；页面 widget 测试仅登录页 happy path

## Risks / Trade-offs

- [base64Captcha 引入 gob 编码依赖，体积 +?] → 可接受；若评审反对，降级为标准库 image 手绘（接口不变）
- [自动注册+无短信=任何人可占号] → 一期接受（MILESTONES 风险已过审）；二期接短信+设备指纹
- [refresh 无黑名单] → 一期接受；token 泄露面=单设备本地安全
- [testcontainers 每测试拉容器慢] → 用 TestMain 复用单容器 + TRUNCATE 清理

## Migration Plan

- 纯新增 migration（users 表），golang-migrate 启动自动执行；回滚=down.sql DROP TABLE
- 无存量数据（W1 基线空库）

## Open Questions

（无——图形验证码、预置头像、登录即注册均已在 MILESTONES.md 定案；JWT HS256/env 密钥为常规默认。）
