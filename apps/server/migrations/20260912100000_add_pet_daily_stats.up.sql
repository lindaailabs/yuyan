-- 宠物每日统计：连续互动天数与累计消息数（W4 成长规则的事实源之一）。
-- 事件表为权威事实源，本表可由事件重建。
CREATE TABLE pet_daily_stats (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  pet_id BIGINT UNSIGNED NOT NULL,
  last_date CHAR(10) NOT NULL,
  streak_days INT NOT NULL DEFAULT 0,
  total_messages INT NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_pet (pet_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
