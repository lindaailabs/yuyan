-- 支付订单（LLM_DEV_GUIDE.md §7）：order_no 唯一键保证回调幂等，重复回调不重复发放权益。
-- 一期不接真实渠道，也不存储任何商户密钥。
CREATE TABLE payment_orders (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  platform VARCHAR(32) NOT NULL DEFAULT 'sandbox',
  plan VARCHAR(32) NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_user_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
