-- 行为事件（LLM_DEV_GUIDE.md §8）：最小可用数据闭环，禁止写入手机号/token/prompt 全文。
CREATE TABLE event_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  props JSON NULL,
  created_at BIGINT NOT NULL,
  KEY idx_name_created (name, created_at),
  KEY idx_user_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
