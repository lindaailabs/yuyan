-- 用户与宠物会话（LLM_DEV_GUIDE.md §4）。
-- 一人一宠物一条会话：uk_user_pet 保证幂等，创建或获取走同一条路径。
CREATE TABLE pet_conversations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  pet_id BIGINT UNSIGNED NOT NULL,
  last_msg_id BIGINT UNSIGNED NULL,
  last_msg_preview VARCHAR(255) NOT NULL DEFAULT '',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_pet (user_id, pet_id),
  KEY idx_user_updated (user_id, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
