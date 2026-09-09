-- 登录改为手机号+密码（移除一期图形验证码）：users 增加 password_hash。
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NULL AFTER avatar_id;
