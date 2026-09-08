-- W2 账号体系：users 表（guide §4 基线 + avatar_id 预置头像决策）。
-- nickname 可为 NULL：NULL = 未完成首登引导（区别于主动清空的场景，见 change design D3）。
CREATE TABLE users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  phone VARCHAR(20) NOT NULL UNIQUE,
  nickname VARCHAR(20) NULL,
  avatar_id SMALLINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
