-- 用量计数器：按 (用户, 动作, 自然日) 记录已用量，额度扣减走条件 UPDATE 保证并发安全。
CREATE TABLE usage_counters (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  action VARCHAR(32) NOT NULL,
  period CHAR(10) NOT NULL,
  used INT NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_action_period (user_id, action, period)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
