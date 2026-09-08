## ADDED Requirements

### Requirement: 图形验证码下发
`POST /api/v1/auth/sms-code` SHALL 接收 `{phone}`，生成 6 位数字验证码存入 Redis（`sms:code:{phone}`，TTL 5min），并将验证码渲染为图形验证码，以 base64 图片（响应字段 `captcha_image`，可缺省）随统一包裹下发。同一手机号 60s 内重复请求 MUST 返回 2xxx 业务错误。phone MUST 通过中国大陆手机号格式校验（11 位、1 开头），非法返回 1001。

#### Scenario: 正常下发
- **WHEN** POST /auth/sms-code `{phone: "13800138000"}`（首次/超过限频窗口）
- **THEN** 返回 code=0，data.captcha_image 为非空 base64 图片字符串

#### Scenario: 限频
- **WHEN** 同一 phone 60s 内第二次请求
- **THEN** 返回 2xxx 业务错误码

#### Scenario: 手机号非法
- **WHEN** phone 为 "12345"
- **THEN** 返回 1001 参数错误

### Requirement: 验证码登录与自动注册
`POST /api/v1/auth/login` SHALL 接收 `{phone, code}`：校验 Redis 中验证码（不匹配/过期返回 2xxx 业务错误）；用户不存在时 MUST 自动注册（nickname=NULL、avatar_id=1）；成功后签发 access token（2h）与 refresh token（30d）并在响应 `data` 中返回 `{access_token, refresh_token, expires_in}`。验证码 MUST 一次性（成功后删除）。

#### Scenario: 新用户登录
- **WHEN** 未注册 phone 携带正确验证码登录
- **THEN** 自动创建用户，返回双 token，GET /users/me 显示 nickname=null

#### Scenario: 老用户登录
- **WHEN** 已注册 phone 携带正确验证码登录
- **THEN** 不创建新用户，返回双 token

#### Scenario: 验证码错误
- **WHEN** code 与 Redis 不符
- **THEN** 返回 2xxx 业务错误，不签发 token

#### Scenario: 验证码重放
- **WHEN** 同一验证码第二次登录
- **THEN** 返回 2xxx 业务错误（验证码已删除）

### Requirement: Token 刷新
`POST /api/v1/auth/refresh` SHALL 接收 `{refresh_token}`：校验签名与 typ=refresh，有效则签发新的 access/refresh 双 token；无效/过期 MUST 返回 1002。

#### Scenario: 刷新成功
- **WHEN** 携带未过期的 refresh token
- **THEN** 返回新双 token

#### Scenario: 用 access 刷新
- **WHEN** refresh_token 实为 access token（typ 不符）
- **THEN** 返回 1002

### Requirement: 认证中间件
除 sms-code/login/refresh 外的全部 API 端点 MUST 经过 auth 中间件：校验 `Authorization: Bearer <access>`（签名有效、未过期、typ=access），并将 uid 注入请求上下文；缺失/无效返回 1002 且不进入 handler。

#### Scenario: 无 token 访问受保护端点
- **WHEN** 不带 Authorization 访问 GET /users/me
- **THEN** 返回 1002

#### Scenario: refresh token 调业务端点
- **WHEN** Bearer 使用 typ=refresh 的 token 访问 GET /users/me
- **THEN** 返回 1002

#### Scenario: 合法 token
- **WHEN** 携带有效 access token
- **THEN** handler 可从上下文取到 uid，返回 200
