# W3 好友关系（contacts）设计

## Context

W2 交付了账号体系（验证码登录、资料、搜索），用户之间还没有任何关系数据。W3 建立「申请-同意/拒绝」闭环与通讯录，为 W4 单聊会话提供 peer 基础。

权威约束（docs/LLM_DEV_GUIDE.md）：§4 friendships 双向各一行 + uk_user_friend；§5.3 REST 全集；「所有表禁止物理删除，需要删除语义时用 status 标记」；SMALLINT 状态常量代替 ENUM；分层 api → service → repo。

**文档缺口（本变更先补文档再编码）**：MILESTONES W3 要求申请页「同意/拒绝」，但 guide §5.3 缺 reject 端点、§4 状态常量缺 rejected。本设计新增 `POST /friends/requests/{id}/reject` 与状态 `4=rejected`，并同步更新 guide §4/§5.3。

## Goals / Non-Goals

**Goals**
- A 能向 B 发好友申请，B 能同意/拒绝，双方 `GET /friends` 可见彼此
- 防重复申请、不能加自己、已是好友/反向申请待处理时返回明确业务错误码
- accept 在单事务内写双向两行（原子性有测试证明）
- App 提供通讯录页、申请列表页、搜索页加好友入口

**Non-Goals（一期不做）**
- 拉黑（status=3 常量保留，无入口不可达）
- 好友申请实时推送/角标（WS 属 W4+；一期进页拉取）
- 删除好友、备注名、好友分组
- 客户端 drift 本地缓存好友（服务端为唯一事实源，多端一致天然满足 §0.7）

## Decisions

### D1 数据模型：guide 既定的双向双行
- 申请 A→B 只产生一行 `(user_id=A, friend_id=B, status=1)`
- 同意后事务内：`(A,B)` 置 accepted + upsert `(B,A)` 为 accepted
- 备选方案「单行 min(user_id,friend_id)/max(user_id,friend_id)」被否：查询需 CASE 判方向，可读性差（团队 Java 背景，guide 明确要「Java 工程师能读懂」）

```sql
CREATE TABLE friendships (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  friend_id BIGINT UNSIGNED NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,  -- 1=pending 2=accepted 3=blocked 4=rejected
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_friend (user_id, friend_id),
  KEY idx_friend_status (friend_id, status)   -- 「我收到的申请」与「我的好友」查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### D2 状态机与申请防重复（service 层校验矩阵）

```
POST /friends/requests (me → target)
  target 不存在            → 2101 用户不存在
  target == me             → 2102 不能添加自己
  (me,target) pending      → 2103 申请已存在，等待对方处理
  (me,target) accepted     → 2104 你们已经是好友
  (me,target) rejected     → 翻回 pending（允许重新申请，刷新 created_at）
  (me,target) blocked      → 2107 对方不可添加（一期不可达，防御保留）
  (target,me) pending      → 2105 对方已向你发出申请，请先处理
  (target,me) accepted     → 2104（双向不一致的防御分支）
  否则                      → INSERT (me,target,pending)
```

并发防重复：uk_user_friend 唯一键是最终兜底；事务内按 uk SELECT ... FOR UPDATE 两行后再决策，行不存在时靠唯一键冲突转 2103。

错误码段（复用 errcode 包模式）：2101 ErrTargetNotFound / 2102 ErrAddSelf / 2103 ErrRequestDuplicate / 2104 ErrAlreadyFriend / 2105 ErrReversePending / 2106 ErrRequestInvalid（accept/reject 目标不存在或非 pending）/ 2107 ErrBlocked。

### D3 accept / reject 事务（repo 层显式事务）
```
accept(requestID, me):
  BEGIN
  SELECT * FROM friendships WHERE id=requestID FOR UPDATE
    校验 friend_id==me 且 status==1，否则 2106
  UPDATE friendships SET status=2 WHERE id=requestID          -- (A,B)
  INSERT INTO friendships(user_id=me, friend_id=对方, status=2)
    ON DUPLICATE KEY UPDATE status=2                          -- (B,A) upsert
  COMMIT

reject(requestID, me):
  BEGIN
  SELECT ... FOR UPDATE + 同样校验
  UPDATE friendships SET status=4 WHERE id=requestID
  COMMIT
```
- `ON DUPLICATE KEY UPDATE` 是 MySQL 标准用法，处理「反向行已存在（历史 rejected）」的复用；备选「先查再插」在并发下有竞态，需额外处理 duplicate key，反而更绕
- FOR UPDATE 防两台设备同时同意/拒绝同一申请；二次操作因 status≠1 返回 2106

### D4 查询与响应
- `GET /friends/requests`：`WHERE friend_id=me AND status=1 ORDER BY id DESC`，JOIN users 取申请人昵称/头像/脱敏手机号。一期不含「我发出的申请」
- `GET /friends`：`WHERE user_id=me AND status=2 ORDER BY friendships.id ASC`（结交时间），JOIN users 同上
- 手机号脱敏复用 user_service 既有逻辑（`138****1111`）
- REST 统一包裹 `{code,msg,data}` 不变；accept/reject 的 data 返回 `{id, status}`

### D5 客户端结构
- `ContactsRepository`（data/repository）：sendRequest/listRequests/accept/reject/listFriends，走既有 ApiClient
- `ContactsController`（StateNotifier<ContactsState>）：state 含 friends 列表、pendingRequests 列表、各操作 loading；accept/reject 成功后局部更新状态（从 requests 移除、accept 时插入 friends）
- 页面：`/contacts`（通讯录）、`/requests`（申请列表，同意/拒绝按钮）；HomePage 加两个入口 icon；SearchPage 结果卡片「加好友」启用：成功置「已申请」禁用态，业务错误（已是好友等）snackbar 提示
- 文案全量入 core/l10n/zh.dart
- 不做本地持久化：好友数据每次进页拉取（W5 会话化后再评估缓存）

### D6 测试策略（对齐 guide §7）
- service 层（testcontainers MySQL，无 Redis 依赖）：防重复矩阵逐分支、accept 后双向两行均 accepted（原子性断言）、二次 accept→2106、reject 后重新申请成功、好友/申请列表内容与脱敏断言
- api 层 httptest：5 端点各 ≥1 happy + ≥1 error（1001 参数 / 21xx 业务 / 1002 未认证）
- Flutter：ContactsController 单测（fake ApiClient 复用 W2 模式：列表加载、accept 局部更新、reject 移除、sendRequest 错误透传）

## Risks / Trade-offs

- [两设备并发同意同一申请] → FOR UPDATE 行锁 + status 条件更新影响行数校验，二次操作返回 2106
- [双向行可能短暂不一致（历史脏数据）] → accept 用 upsert 保证最终一致；不提供删除接口，status 全生命周期可追溯
- [重新申请翻回 pending 的 created_at 刷新] → 显式 SET created_at=NOW()，申请列表按 id DESC 排序仍准确
- [一期无申请实时推送] → 用户感知延迟可接受（进申请页拉取）；W4 WS 上线后可挂 `conv.unread` 类推送，契约不变

## Migration Plan

1. 更新 docs/LLM_DEV_GUIDE.md §4/§5.3（先于代码合入）
2. migration `20260909100000_add_friendships.{up,down}.sql`（up 建表 / down DROP TABLE，纯新增无破坏）
3. docker compose 重建 server 自动执行 migration；回滚 = down 迁移 + 回退镜像
4. 里程碑提交：`feat(contacts): W3 好友关系（申请/同意/拒绝/通讯录）`

## Open Questions

（无——reject 端点与状态常量虽为 guide 补录，但 MILESTONES W3 已明确要求「同意/拒绝」，属文档缺口修复而非范围扩张）
