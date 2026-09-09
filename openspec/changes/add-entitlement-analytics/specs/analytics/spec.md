## ADDED Requirements

### Requirement: 关键事件落库

系统 SHALL 记录关键行为事件（激活/注册、创建宠物、首轮对话、记忆生成与删除、成长升级、额度拒绝、订阅页曝光与点击等），事件 SHALL 含名称、属性与时间。

#### Scenario: 服务端事件

- **WHEN** 用户完成创建宠物或首轮对话等行为
- **THEN** 系统 SHALL 写入对应事件记录

### Requirement: 客户端批量上报

系统 SHALL 提供批量上报端点，供 App 上报端上事件（如订阅页曝光与点击）；上报 SHALL 限制单次条数与属性长度，并拒绝包含敏感信息的字段。

#### Scenario: 批量上报成功

- **WHEN** 客户端一次上报若干条合法事件
- **THEN** 系统 SHALL 全部落库

#### Scenario: 超限或非法

- **WHEN** 单次上报条数超过上限或事件名非法
- **THEN** 系统 SHALL 返回参数错误，且不落库

### Requirement: 事件脱敏

事件属性 SHALL NOT 包含完整手机号、access/refresh token 或 prompt 全文（guide §12 红线）。

#### Scenario: 敏感字段过滤

- **WHEN** 上报属性中出现敏感键（如 phone、token、password）
- **THEN** 系统 SHALL 丢弃该键或整体拒绝该条事件

### Requirement: 内部查询出口

内部观察 SHALL 能按事件名查询最近事件；该出口 SHALL 仅在非生产环境启用。

#### Scenario: 按名称查询

- **WHEN** 在非生产环境按事件名查询
- **THEN** 返回按时间倒序的事件列表
