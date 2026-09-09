-- 用户与宠物的消息（LLM_DEV_GUIDE.md §4）。
-- id 兼作时序与增量拉取游标，禁止 offset/时间戳分页（红线 §12）。
-- client_msg_id 唯一键是幂等的最终兜底：重复提交直接返回首次结果。
CREATE TABLE pet_messages (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  conv_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  pet_id BIGINT UNSIGNED NOT NULL,
  role VARCHAR(16) NOT NULL,
  content TEXT NOT NULL,
  client_msg_id CHAR(36) NULL,
  model VARCHAR(64) NOT NULL DEFAULT '',
  input_tokens INT NOT NULL DEFAULT 0,
  output_tokens INT NOT NULL DEFAULT 0,
  latency_ms INT NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  error_code INT NOT NULL DEFAULT 0,
  created_at BIGINT NOT NULL,
  UNIQUE KEY uk_client_msg_id (client_msg_id),
  KEY idx_conv_id_id (conv_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
