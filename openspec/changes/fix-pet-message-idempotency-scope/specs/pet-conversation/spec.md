## MODIFIED Requirements

### Requirement: 发送消息与 AI 回复

`POST /api/v1/pet-messages` SHALL persist a user message, generate a pet reply through the AI Gateway, and return both messages. `client_msg_id` idempotency SHALL be scoped to the current authenticated user and current conversation.

#### Scenario: 重复提交幂等

- **WHEN** the same authenticated user resubmits the same `client_msg_id` for the same pet conversation
- **THEN** the system SHALL return the first result
- **AND** the conversation SHALL NOT contain a second duplicate user message or reply

#### Scenario: 跨用户 client_msg_id 隔离

- **WHEN** two different authenticated users submit the same `client_msg_id` in their own pet conversations
- **THEN** each user's message SHALL be accepted or replayed only within that user's own conversation
- **AND** neither response SHALL include messages from the other user
