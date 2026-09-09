-- AI 调用日志（LLM_DEV_GUIDE.md §5/§10、MILESTONES W2）。
-- 每次调用（成功或失败）落一条：模型、token、耗时、错误码、缓存命中，供成本与稳定性观测。
CREATE TABLE ai_call_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  pet_id BIGINT UNSIGNED NOT NULL,
  conv_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  msg_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  provider VARCHAR(32) NOT NULL DEFAULT '',
  model VARCHAR(64) NOT NULL DEFAULT '',
  prompt_version VARCHAR(32) NOT NULL DEFAULT '',
  input_tokens INT NOT NULL DEFAULT 0,
  output_tokens INT NOT NULL DEFAULT 0,
  latency_ms INT NOT NULL DEFAULT 0,
  err_code INT NOT NULL DEFAULT 0,
  cache_hit TINYINT NOT NULL DEFAULT 0,
  created_at BIGINT NOT NULL,
  KEY idx_user_created (user_id, created_at),
  KEY idx_pet_created (pet_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
