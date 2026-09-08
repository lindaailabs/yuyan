## ADDED Requirements

### Requirement: WS 帧 JSON Schema 全集
`packages/protocol`（Go module `github.com/lindaailabs/yuyan/protocol`）SHALL 为 §5.2 全部 7 个命令字（conn.auth、conn.heartbeat、msg.send、msg.push、msg.ack、msg.pull、conv.unread）提供 JSON Schema 文件（`schema/ws/`），字段语义与 LLM_DEV_GUIDE.md §5.1/§5.2 严格一致：cmd 必填字符串、seq 必填整数、data 对象、client_msg_id 仅 msg.send 必填（其余帧禁用或可选，以 guide 为准）。

#### Scenario: Schema 文件齐全
- **WHEN** 列举 schema/ws/ 目录
- **THEN** 恰好存在 7 个命令字的 .json Schema 文件且可被 JSON 解析

### Requirement: REST 统一响应包裹 Schema
`schema/rest/response.json` SHALL 定义统一响应包裹 `{code, msg, data}`：code 必填整数、msg 必填字符串、data 可为任意 JSON（含 null）。

#### Scenario: 包裹 Schema 校验
- **WHEN** 用该 Schema 校验 `{"code":0,"msg":"ok","data":null}`
- **THEN** 校验通过；校验 `{"msg":"x"}`（缺 code）失败

### Requirement: 校验用例（每帧类型）
包内 Go 测试 SHALL 对每个帧类型覆盖：至少 1 个合法样例通过校验、至少 1 个非法样例（缺必填字段/类型错误）被拒绝。

#### Scenario: 合法帧通过
- **WHEN** 运行 `go test ./...`（packages/protocol）
- **THEN** 全部合法样例通过对应 Schema 校验

#### Scenario: 非法帧拒绝
- **WHEN** 运行校验用例中非法样例（如 msg.send 缺 client_msg_id）
- **THEN** Schema 校验返回错误，测试断言通过

### Requirement: Schema 对双端可引用
协议包 SHALL 通过 go:embed 暴露 Schema 文件系统，并作为独立 Go module 可被 apps/server（go.work/replace）引用；`scripts/gen-protocol.*` 提供占位实现（仅执行 Schema 校验测试），为二期代码生成预留入口。

#### Scenario: 服务端引用协议包
- **WHEN** apps/server 通过 workspace/replace 引用 protocol 并访问其 embed FS
- **THEN** 编译通过，可读取任意命令字 Schema 内容
