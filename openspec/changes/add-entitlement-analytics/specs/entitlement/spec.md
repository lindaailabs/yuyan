## ADDED Requirements

### Requirement: 服务端权益事实源

系统 SHALL 在服务端维护用户权益（plan、状态、额度），客户端 SHALL 只展示权益状态，不得作为最终判断来源。

#### Scenario: 查询权益与剩余额度

- **WHEN** 用户查询自己的权益
- **THEN** 系统 SHALL 返回 plan、状态、每日额度上限、当日已用量与剩余额度

#### Scenario: 默认免费权益

- **WHEN** 新用户首次查询且无权益记录
- **THEN** 系统 SHALL 自动建立免费权益（默认每日 AI 对话额度）

### Requirement: 消耗型额度校验与扣减

AI 对话等消耗型能力 SHALL 由服务端校验并扣减额度；扣减 SHALL 原子执行，并发下不得超出上限。

#### Scenario: 额度内对话

- **WHEN** 当日用量未达上限
- **THEN** 对话 SHALL 正常进行，且已用量 +1

#### Scenario: 额度耗尽

- **WHEN** 当日用量已达上限
- **THEN** 系统 SHALL 返回 2501
- **AND** SHALL NOT 写入任何消息
- **AND** SHALL NOT 发起 AI 调用

#### Scenario: 并发不超卖

- **WHEN** 多个请求同时扣减额度
- **THEN** 成功次数 SHALL NOT 超过剩余额度

### Requirement: 支付回调幂等

支付回调 SHALL 以订单号建立唯一键：同一订单重复到达 SHALL NOT 重复发放权益。

#### Scenario: 重复回调

- **WHEN** 同一订单号的回调重复提交
- **THEN** 权益 SHALL 只生效一次
- **AND** 响应 SHALL 返回当前权益（幂等成功）

### Requirement: 沙盒支付门禁

沙盒开通 SHALL 仅在非生产环境可用；生产环境 SHALL 返回明确错误，且不得硬编码任何商户密钥。

#### Scenario: 生产环境调用沙盒

- **WHEN** 在生产环境调用沙盒开通
- **THEN** 系统 SHALL 返回 2504
- **AND** SHALL NOT 变更权益
