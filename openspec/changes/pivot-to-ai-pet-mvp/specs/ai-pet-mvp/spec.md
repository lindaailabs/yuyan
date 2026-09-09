# ai-pet-mvp Specification

## ADDED Requirements

### Requirement: Companion Loop

系统 SHALL 支持用户创建或领养一只 AI 宠物，并完成文本对话、记忆、成长状态恢复的陪伴闭环。

#### Scenario: First companion session

- WHEN 用户完成登录并创建宠物
- AND 用户发送第一条文本消息
- THEN 服务端 SHALL 持久化用户消息
- AND AI Gateway SHALL 生成宠物回复或返回可展示的兜底错误
- AND App SHALL 展示宠物回复与当前宠物状态

#### Scenario: App restart recovery

- WHEN 用户杀进程并重启 App
- THEN App SHALL 恢复宠物档案、最近消息、长期记忆摘要和成长状态

### Requirement: AI Gateway Boundary

系统 SHALL 通过统一 AI Gateway 调用模型，业务 service SHALL NOT 直接调用具体模型供应商 SDK。

#### Scenario: Mocked model test

- WHEN 测试环境运行宠物对话 service 测试
- THEN AI Gateway SHALL 支持 mock 实现
- AND 测试 SHALL NOT 依赖真实模型网络请求

#### Scenario: Usage logging

- WHEN 一次 AI 回复完成或失败
- THEN 系统 SHALL 记录模型、输入 token、输出 token、耗时、错误码和缓存命中信息

### Requirement: Memory Control

系统 SHALL 支持长期记忆的抽取、召回、展示和删除。

#### Scenario: Memory recall

- WHEN 用户在历史对话中表达稳定偏好
- AND 后续对话提到相关主题
- THEN AI Gateway 上下文 SHALL 能召回少量相关记忆

#### Scenario: Memory deletion

- WHEN 用户删除一条长期记忆
- THEN 该记忆 SHALL 标记为删除状态
- AND 后续上下文拼装 SHALL NOT 召回该记忆

### Requirement: Deterministic Growth

系统 SHALL 使用服务端确定性规则更新宠物成长状态。

#### Scenario: Growth event

- WHEN 用户互动满足成长规则
- THEN 系统 SHALL 更新亲密度、等级或心情
- AND SHALL 生成可解释的成长事件

#### Scenario: No model-owned counters

- WHEN 大模型回复包含成长相关表达
- THEN 系统 SHALL NOT 直接采用模型输出作为核心数值变更

### Requirement: Entitlement Source of Truth

系统 SHALL 以服务端权益模型作为免费/订阅额度的唯一事实源。

#### Scenario: Quota check

- WHEN 用户发送需要消耗额度的 AI 对话
- THEN 服务端 SHALL 校验 entitlement 和 quota
- AND 客户端 SHALL NOT 单独决定是否允许使用付费能力
