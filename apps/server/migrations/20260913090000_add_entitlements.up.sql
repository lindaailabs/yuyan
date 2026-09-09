-- 订阅权益（LLM_DEV_GUIDE.md §4、§7）：服务端是权益唯一事实源，客户端只展示。
-- quota_json 保存额度配置（每日 AI 对话条数、记忆容量、高级模型开关等）。
CREATE TABLE entitlements (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  plan VARCHAR(32) NOT NULL DEFAULT 'free',
  status SMALLINT NOT NULL DEFAULT 1,
  quota_json JSON NULL,
  renew_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user (user_id),
  KEY idx_user_status (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
