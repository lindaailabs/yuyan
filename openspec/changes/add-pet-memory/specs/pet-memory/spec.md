## ADDED Requirements

### Requirement: 长期记忆抽取

系统 SHALL 在用户消息进入对话时，按服务端确定性规则抽取长期事实（偏好、称呼、画像、事件），并持久化到 `pet_memories`；每条记忆 MUST 记录来源消息 id、类型、置信度与状态。

#### Scenario: 抽取偏好

- **WHEN** 用户说「我喜欢蓝色」
- **THEN** 系统 SHALL 写入一条 `memory_type=preference` 的记忆，内容包含「蓝色」
- **AND** SHALL 记录来源消息 id 与置信度

#### Scenario: 同一事实去重

- **WHEN** 用户再次表达同一事实
- **THEN** 系统 SHALL NOT 产生重复记忆，仅更新置信度、来源与更新时间

### Requirement: 记忆召回

对话时系统 SHALL 按当前输入与记忆的相关度召回少量（默认 3 条）active 记忆，注入 AI Gateway 上下文的长期记忆段落；被召回的记忆 SHALL 更新最近使用时间。

#### Scenario: 相关记忆被召回

- **WHEN** 用户曾说「我喜欢蓝色」，之后问「你记得我喜欢什么颜色吗」
- **THEN** 该记忆 SHALL 被召回并进入上下文
- **AND** 宠物回复 SHALL 体现该记忆内容

#### Scenario: 不相关记忆不进入上下文

- **WHEN** 当前输入与某条记忆无字词重合
- **THEN** 该记忆 SHALL NOT 被召回

### Requirement: 记忆可查看

`GET /api/v1/pets/{id}/memories` SHALL 返回当前用户该宠物的 active 记忆列表（内容、类型、置信度、创建与最近使用时间）；宠物不属于当前用户时 SHALL 返回业务错误。

#### Scenario: 查看记忆

- **WHEN** 用户查看自己宠物的记忆
- **THEN** 返回 active 记忆列表，已删除（status=3）的记忆 SHALL NOT 出现

### Requirement: 记忆可删除

`DELETE /api/v1/pet-memories/{id}` SHALL 把记忆标记为删除状态（MUST NOT 物理删除）；删除后 SHALL NOT 再被召回，也不再出现在列表中。

#### Scenario: 删除后不再召回

- **WHEN** 用户删除「我喜欢蓝色」这条记忆
- **THEN** 该记忆状态为 deleted
- **AND** 后续对话上下文 SHALL NOT 包含它
- **AND** 记忆列表 SHALL NOT 展示它

#### Scenario: 越权删除

- **WHEN** 用户尝试删除不属于自己的记忆
- **THEN** 系统 SHALL 返回 2401，且不改变数据

### Requirement: 记忆形成可见

发送消息的响应 SHALL 携带本次新形成的记忆（`new_memories`），App SHALL 在聊天页展示轻量提示。

#### Scenario: 聊天页提示

- **WHEN** 本轮对话抽取出新记忆
- **THEN** 响应 SHALL 包含该记忆
- **AND** 聊天页 SHALL 展示「记住了：…」提示
