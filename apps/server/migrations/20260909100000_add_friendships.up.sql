-- W3 好友关系：friendships 表（guide §4 基线 + rejected 状态补录）。
-- 双向各存一行；status SMALLINT 常量：1=pending 2=accepted 3=blocked 4=rejected。
-- idx_friend_status 支撑「我收到的申请」与「我的好友」高频查询。
CREATE TABLE friendships (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  friend_id BIGINT UNSIGNED NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_friend (user_id, friend_id),
  KEY idx_friend_status (friend_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
