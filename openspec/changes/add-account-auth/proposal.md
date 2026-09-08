## Why

MILESTONES.md W2：多账号对话闭环的第一环——没有账号就没有"多账号"。需要实现从获取图形验证码、验证码登录（自动注册）、JWT 会话保持，到资料完善（昵称/预置头像）与用户搜索的完整账号链路，为 W3 好友关系与 W4 消息提供 user_id 主体。

## What Changes

- 新增 migration `20260908`：`users` 表（guide §4 基线：id BIGINT UNSIGNED AUTO_INCREMENT PK、phone VARCHAR(20) UNIQUE、nickname、avatar_id SMALLINT NOT NULL DEFAULT 1、created_at、updated_at）
- 服务端新增 5 组 REST 端点（guide §5.3 契约不变）：
  - `POST /api/v1/auth/sms-code`：6 位数字码存 Redis（`sms:code:{phone}`，5min TTL，限频 1/min），渲染为图形验证码 base64 图片随响应下发（不对接短信）
  - `POST /api/v1/auth/login`：校验验证码，用户不存在则自动注册（默认昵称+avatar_id=1），签发 JWT（access 2h + refresh 30d）
  - `POST /api/v1/auth/refresh`：refresh token 换发新 access/refresh
  - `GET/PUT /api/v1/users/me`：查询/更新个人资料（昵称、avatar_id 1~8 校验）
  - `GET /api/v1/users/search?q=`：按手机号精确搜索（仅返回已注册用户）
- 服务端新增 auth 中间件：JWT 校验，user_id 注入 context；除 sms-code/login/refresh 外全部端点受保护
- Flutter App 新增：登录页（手机号+图形验证码图片+输入，响应无图片字段时兼容"等待短信"两态）、首登引导页（选 8 个预置头像之一+设昵称）、我的资料页（查看/修改）、用户搜索页；token 持久化（flutter_secure_storage）与 Dio 拦截器（401 自动 refresh 重放）；8 个预置头像 assets
- 首登判定：`nickname IS NULL` 视为未完成引导（注册时默认昵称置 NULL，引导页提交后写入）

## Capabilities

### New Capabilities

- `auth-service`: 验证码（图形验证码 Redis 方案）、登录注册（自动注册）、JWT 签发校验刷新（含 auth 中间件）
- `user-profile`: 个人资料（昵称/预置头像）与用户搜索（手机号精确）
- `app-auth`: 客户端登录/引导/资料/搜索页面与 token 生命周期管理

### Modified Capabilities

（无——server-skeleton/app-skeleton/protocol-schema 的 W1 规格 不受影响；users 表属新增能力，不触碰骨架规格）

## Impact

- **代码**：apps/server（新增 auth/user 两个业务域的 service/repo/api 层与 migration）、apps/app（features/auth、features/profile、features/contacts-search、core 增 token 管理）、packages/protocol（不改动——REST 契约已含全部端点，无新帧）
- **新依赖声明（PR 评审用）**：
  - Go：golang-jwt/jwt/v5（W1 已计划）、base64Captcha（图形验证码渲染；若体积/依赖不可接受则用标准库 image 手绘，design 定案）、redis 经 W1 go-redis 复用
  - Flutter：flutter_secure_storage（token 持久化）、原有 dio/riverpod/go_router 复用
- **风险**：JWT 密钥管理（env 注入，禁硬编码 guide §9.3）；自动注册的手机号伪造风险（一期无短信，接受——验证码即登录凭证）
