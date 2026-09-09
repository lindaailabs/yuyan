## ADDED Requirements

### Requirement: 模型调用唯一入口

系统 SHALL 通过统一 AI Gateway 调用模型；业务 service SHALL NOT 直接调用具体模型供应商 SDK。Gateway SHALL 位于基础设施层（`internal/pkg/ai`），service 仅依赖其接口。

#### Scenario: service 只依赖接口

- **WHEN** 对话 service 需要生成宠物回复
- **THEN** 它 SHALL 仅通过 `Gateway.Complete(ctx, req)` 获取结果
- **AND** 代码中 SHALL NOT 出现任何供应商 SDK 的直接调用

### Requirement: 默认 mock provider

`AI_PROVIDER` 默认值 SHALL 为 `mock`；测试 SHALL NOT 依赖真实模型网络请求。

#### Scenario: 默认使用 mock

- **WHEN** 服务在未配置 `AI_PROVIDER` 的环境启动
- **THEN** Gateway SHALL 使用 mock provider 生成确定性回复
- **AND** 回复 SHALL 包含可断言的模型名与 token/耗时统计

#### Scenario: 切换未实现的 provider

- **WHEN** `AI_PROVIDER` 被设置为未实现的取值
- **THEN** Gateway SHALL 返回明确错误（2304）
- **AND** SHALL NOT 静默降级为 mock

#### Scenario: 注入失败与超时

- **WHEN** mock provider 被配置为失败或超时（如 `AI_MOCK_FAIL_RATE`）
- **THEN** Gateway SHALL 返回错误且不阻塞调用方
- **AND** 调用方 SHALL 能据此写入失败状态与用量错误码

### Requirement: Prompt 分层与上限

Gateway SHALL 按「系统规则 → 宠物 persona → 成长状态 → 长期记忆 → 最近 N 轮 → 当前输入」拼装上下文，并对最近对话做轮次与字符双上限裁剪；SHALL NOT 把全部历史消息塞进 prompt。

#### Scenario: 长对话裁剪

- **WHEN** 会话已有远超上限的历史消息
- **THEN** 上下文 SHALL 只保留最近 N 轮且累计字符不超过上限
- **AND** SHALL 标记发生裁剪以便日志观测

### Requirement: 用量与日志脱敏

每次调用完成或失败，系统 SHALL 记录模型、输入 token、输出 token、耗时、错误码与缓存命中信息；日志 SHALL NOT 输出 prompt 正文、完整手机号或 token。

#### Scenario: 调用成功留痕

- **WHEN** 一次 AI 回复完成
- **THEN** 系统 SHALL 在 `ai_call_logs` 写入一条记录（含 model/input_tokens/output_tokens/latency_ms/err_code/cache_hit）

#### Scenario: 调用失败留痕

- **WHEN** 一次 AI 调用失败或超时
- **THEN** 系统 SHALL 同样写入 `ai_call_logs`，`err_code` 非 0
- **AND** 日志 SHALL 只记录脱敏后的结构信息
