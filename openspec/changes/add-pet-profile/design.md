## Context

`pivot-to-ai-pet-mvp` 已将一期主线改为 AI 宠物陪伴闭环。本变更落第一块业务能力：宠物档案。现有 auth/user 能力继续复用，宠物归属于当前登录用户。

## Goals / Non-Goals

Goals:

- 用户登录后可以创建一只宠物。
- 用户可以查询自己的宠物列表、宠物详情和宠物状态。
- 用户可以修改宠物名字、头像/外观和基础 persona。
- App 首页展示宠物创建入口或宠物状态。
- 单测覆盖 service、api、repository/controller 的核心路径。

Non-Goals:

- 不接入 AI 模型。
- 不实现聊天、记忆、成长事件规则。
- 不实现多个宠物并行养成的复杂切换体验；服务端允许多条数据，App 一期默认展示第一只。
- 不删除旧 contacts/friendship 代码。

## Decisions

### D1. pets 表允许一个用户多只宠物，App 一期展示第一只

表结构不做 `user_id` 唯一约束，避免未来扩展多宠物时迁移成本过高。App 一期只取列表第一只作为当前宠物。

### D2. persona 用 JSON 字符串存储

一期不引入复杂 schema 或 JSON 查询。服务端保存基础 persona 字符串，如 `{"style":"warm","traits":["curious"]}`；后续 AI Gateway 再细化 prompt 结构。

### D3. 状态字段先落在 pets 表

`level`、`intimacy`、`mood` 先存在 pets 表，W4 成长系统再引入 `pet_growth_events` 和更完整规则。创建宠物默认 `level=1`、`intimacy=0`、`mood=curious`。

### D4. 鉴权按 user_id 隔离

所有宠物读写都必须带当前 uid 条件。查不到或不属于当前用户统一返回宠物不存在。

## API

- `POST /api/v1/pets`
- `GET /api/v1/pets`
- `GET /api/v1/pets/:id`
- `PUT /api/v1/pets/:id`
- `GET /api/v1/pets/:id/state`

## Validation

- name：1~20 字符。
- species：1~32 字符，默认 `swallow`。
- avatar_id：1~8。
- persona：可选，最长 2000 字符。
