## 1. Specification

- [x] 1.1 Add pet-conversation requirement clarifying user-scoped `client_msg_id` idempotency.

## 2. Server

- [x] 2.1 Add migration replacing global `uk_client_msg_id` with `(user_id, client_msg_id)`.
- [x] 2.2 Scope replay lookup by current user and conversation.
- [x] 2.3 Add regression test for cross-user `client_msg_id` reuse.

## 3. Verification

- [x] 3.1 Run OpenSpec validation, Go tests, Flutter tests, and Web build.

