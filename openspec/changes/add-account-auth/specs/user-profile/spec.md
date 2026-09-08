## ADDED Requirements

### Requirement: 个人资料查询
`GET /api/v1/users/me` SHALL 返回当前登录用户资料：`{id, phone, nickname, avatar_id, created_at}`；nickname 为 null 表示未完成首登引导。

#### Scenario: 查询资料
- **WHEN** 携带有效 access token 请求
- **THEN** 返回 uid 对应用户的完整资料（统一包裹）

### Requirement: 个人资料更新
`PUT /api/v1/users/me` SHALL 接收 `{nickname?, avatar_id?}`：nickname 存在时须为 1~20 字符；avatar_id 须为 1~8；非法值返回 2xxx 业务错误；成功后返回更新后的资料。仅更新请求中出现的字段。

#### Scenario: 更新昵称与头像
- **WHEN** `{nickname: "语燕用户", avatar_id: 3}`
- **THEN** 资料更新成功，再次查询返回新值

#### Scenario: 头像越界
- **WHEN** `{avatar_id: 9}`
- **THEN** 返回 2xxx 业务错误，资料不变

#### Scenario: 昵称超长
- **WHEN** `{nickname: "超过二十个字符的昵称……"}`
- **THEN** 返回 2xxx 业务错误

### Requirement: 用户搜索
`GET /api/v1/users/search?q=` SHALL 按手机号精确搜索：q 为空或非法返回 1001；未注册返回空列表（不暴露存在性之外的信息）；命中返回 `[{id, nickname, avatar_id, phone}]`（脱敏：phone 中间 4 位以 **** 替换显示字段）。

#### Scenario: 命中
- **WHEN** q 为已注册手机号
- **THEN** 返回该用户（昵称可为 null）

#### Scenario: 未命中
- **WHEN** q 为未注册手机号
- **THEN** 返回空列表，code=0

#### Scenario: 参数缺失
- **WHEN** 不带 q
- **THEN** 返回 1001

### Requirement: users 表结构
migration SHALL 按 LLM_DEV_GUIDE.md §4 基线创建 users 表：id BIGINT UNSIGNED AUTO_INCREMENT PK、phone VARCHAR(20) UNIQUE NOT NULL、nickname VARCHAR(20) NULL、avatar_id SMALLINT NOT NULL DEFAULT 1、created_at/updated_at DATETIME；utf8mb4。migration 文件一经提交 MUST NOT 修改删除。

#### Scenario: 表结构就位
- **WHEN** migration 执行后 DESCRIBE users
- **THEN** 字段/类型/唯一约束与上述一致
