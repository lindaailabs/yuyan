## ADDED Requirements

### Requirement: 好友申请数据模型
系统 SHALL 以 `friendships` 表（utf8mb4/InnoDB）承载好友关系：`id` BIGINT UNSIGNED 自增主键、`user_id`/`friend_id` BIGINT UNSIGNED、`status` SMALLINT 常量（1=pending 2=accepted 3=blocked 4=rejected）、`created_at` DATETIME，唯一键 `uk_user_friend(user_id, friend_id)`。关系 MUST 双向各存一行（A↔B 好友 = `(A,B)` 与 `(B,A)` 两行均为 accepted）。所有行 MUST 使用 status 标记状态流转，禁止物理删除。

#### Scenario: 表结构约束
- **WHEN** migration 在空库执行后
- **THEN** friendships 表存在 uk_user_friend 唯一键，字符集为 utf8mb4，且同向重复行插入失败

### Requirement: 发起好友申请
`POST /api/v1/friends/requests` SHALL 接收 `{user_id}`（Bearer 认证），执行防重复校验矩阵：目标用户不存在 MUST 返回 2101；目标为本人 MUST 返回 2102；已存在本人→目标的 pending 行 MUST 返回 2103；任一方向已 accepted MUST 返回 2104；目标→本人存在 pending 行 MUST 返回 2105；本人→目标行曾 rejected MUST 翻回 pending 并刷新 created_at（允许重新申请）；否则 MUST 插入 pending 行。成功响应 data MUST 含 `{id, user_id, friend_id, status, created_at}`。user_id 非法（缺省/非正数）返回 1001。

#### Scenario: 正常申请
- **WHEN** A 向已存在的 B 发起申请（无任何历史关系行）
- **THEN** 返回 code=0，产生 `(A,B,pending)` 一行，B 的申请列表可见

#### Scenario: 不能加自己
- **WHEN** user_id 等于当前登录用户
- **THEN** 返回 2102，不产生数据行

#### Scenario: 重复申请
- **WHEN** A 已有向 B 的 pending 申请，再次提交
- **THEN** 返回 2103，仍只有一行 pending

#### Scenario: 已是好友
- **WHEN** A、B 已是好友（双向 accepted），A 再次提交
- **THEN** 返回 2104

#### Scenario: 反向申请待处理
- **WHEN** B 已向 A 发出 pending 申请，A 又向 B 发起申请
- **THEN** 返回 2105

#### Scenario: 被拒后重新申请
- **WHEN** A→B 的申请被 B 拒绝后，A 再次发起
- **THEN** 原 rejected 行翻回 pending，返回 code=0

#### Scenario: 目标不存在
- **WHEN** user_id 指向不存在的用户
- **THEN** 返回 2101

### Requirement: 申请列表
`GET /api/v1/friends/requests`（Bearer 认证）SHALL 返回当前用户收到的全部 pending 申请，按 id 倒序；每项 MUST 含申请 id、申请人资料（id、nickname、avatar_id、脱敏手机号）、created_at。已处理（accepted/rejected）的申请 MUST NOT 出现。

#### Scenario: 收到申请
- **WHEN** B 登录查询，存在 A、C 两条发给 B 的 pending 申请
- **THEN** 返回 2 项，包含 A 与 C 的昵称/头像/脱敏手机号

#### Scenario: 无申请
- **WHEN** B 未收到任何 pending 申请
- **THEN** 返回空列表

### Requirement: 同意申请
`POST /api/v1/friends/requests/{id}/accept`（Bearer 认证）SHALL 在单数据库事务内：将目标申请行（MUST 校验 friend_id=当前用户且 status=pending，否则返回 2106）置为 accepted，并 upsert 反向行（当前用户→申请人）为 accepted。事务原子性 MUST 有测试证明（两行同现或同无）。重复同意 MUST 返回 2106。

#### Scenario: 同意成功
- **WHEN** B 同意 A 的 pending 申请
- **THEN** `(A,B)` 与 `(B,A)` 两行均为 accepted，双方 GET /friends 互见

#### Scenario: 二次同意
- **WHEN** 同一申请被再次调用 accept
- **THEN** 返回 2106，数据无变化

#### Scenario: 处理他人申请
- **WHEN** C 尝试 accept A→B 的申请
- **THEN** 返回 2106

### Requirement: 拒绝申请
`POST /api/v1/friends/requests/{id}/reject`（Bearer 认证）SHALL 将目标申请行（MUST 校验 friend_id=当前用户且 status=pending，否则返回 2106）置为 rejected，MUST NOT 物理删除，MUST NOT 产生反向行。拒绝后申请人可重新发起申请。

#### Scenario: 拒绝成功
- **WHEN** B 拒绝 A 的 pending 申请
- **THEN** 行状态为 rejected，B 的申请列表不再显示，双方好友列表互不可见

#### Scenario: 拒绝不存在的申请
- **WHEN** 申请 id 不存在或已被处理
- **THEN** 返回 2106

### Requirement: 好友列表
`GET /api/v1/friends`（Bearer 认证）SHALL 返回当前用户全部 accepted 好友，按结交时间（行 id）升序；每项 MUST 含好友资料（id、nickname、avatar_id、脱敏手机号）与结交时间。pending/rejected 行 MUST NOT 出现。

#### Scenario: 好友互见
- **WHEN** A、B 成为好友后任一方查询
- **THEN** 各自列表包含对方（资料完整、手机号脱敏）

#### Scenario: 非好友不可见
- **WHEN** A 查询好友列表而 C 与 A 仅有 pending 申请
- **THEN** 列表不含 C
