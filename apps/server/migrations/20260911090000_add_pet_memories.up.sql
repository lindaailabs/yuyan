-- 长期记忆（LLM_DEV_GUIDE.md §4）。
-- status：1=active 2=archived 3=deleted，禁止物理删除（guide §12 红线）。
-- uk_pet_hash 保证同一事实不重复入库：重复表达走 upsert 更新置信度与来源。
CREATE TABLE pet_memories (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  pet_id BIGINT UNSIGNED NOT NULL,
  memory_type VARCHAR(32) NOT NULL,
  content TEXT NOT NULL,
  content_hash CHAR(64) NOT NULL,
  source_msg_id BIGINT UNSIGNED NULL,
  confidence DECIMAL(4,3) NOT NULL DEFAULT 0.500,
  last_used_at DATETIME NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_pet_hash (pet_id, content_hash),
  KEY idx_pet_status (pet_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
