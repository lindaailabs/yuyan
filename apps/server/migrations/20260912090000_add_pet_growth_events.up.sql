-- 成长事件（LLM_DEV_GUIDE.md §4）：每次状态变化留一条可解释记录。
-- uk_pet_source_type 是成长结算的幂等闸门：同一条消息重复结算不会产生第二条同类事件。
CREATE TABLE pet_growth_events (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  pet_id BIGINT UNSIGNED NOT NULL,
  event_type VARCHAR(32) NOT NULL,
  delta JSON NULL,
  reason VARCHAR(255) NOT NULL DEFAULT '',
  source_msg_id BIGINT UNSIGNED NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_pet_source_type (pet_id, source_msg_id, event_type),
  KEY idx_pet_created (pet_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
