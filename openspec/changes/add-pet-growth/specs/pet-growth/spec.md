## ADDED Requirements

### Requirement: 互动产生确定性成长

系统 SHALL 在每轮有效互动后，按服务端确定性规则增加亲密度；等级 SHALL 由亲密度阈值推导；连续互动天数 SHALL 按自然日累计。大模型 SHALL NOT 决定这些数值。

#### Scenario: 亲密度随对话增长

- **WHEN** 用户与宠物完成一轮对话
- **THEN** 宠物亲密度 SHALL 按固定步长增加
- **AND** SHALL 生成一条可解释的成长事件（含增量与原因）

#### Scenario: 升级

- **WHEN** 亲密度达到下一级阈值
- **THEN** 宠物等级 SHALL 提升
- **AND** SHALL 生成 `level_up` 事件，说明升级前后等级

#### Scenario: 连续互动

- **WHEN** 用户在连续两个自然日都与宠物互动
- **THEN** 连续互动天数 SHALL 递增
- **AND** 达到里程碑（3/7/30 天）时 SHALL 生成 `streak_milestone` 事件

### Requirement: 重复处理不重复加经验

成长结算 SHALL 以来源消息为幂等键：同一条消息重复结算 SHALL NOT 再次增加亲密度。

#### Scenario: 重放同一条消息

- **WHEN** 同一条宠物消息被重复结算（重试或重放）
- **THEN** 亲密度 SHALL 保持不变
- **AND** SHALL NOT 产生第二条同类型成长事件

### Requirement: 心情随时间变化

宠物心情 SHALL 由距上次互动的时长按确定性规则推导，并在变化时生成 `mood_change` 事件。

#### Scenario: 长时间未互动

- **WHEN** 距上次互动超过 72 小时
- **THEN** 宠物心情 SHALL 变为表示思念的状态
- **AND** SHALL 生成可解释的心情变化事件

### Requirement: 成长事件可查询

`GET /api/v1/pets/{id}/growth-events` SHALL 按时间倒序返回该宠物的成长事件（类型、增量、原因、时间）；宠物不属于当前用户 SHALL 返回业务错误。

#### Scenario: 查看成长时间线

- **WHEN** 用户查看自己宠物的成长事件
- **THEN** 返回按时间倒序的事件列表
- **AND** 每条事件 SHALL 含类型、增量与人类可读的原因
