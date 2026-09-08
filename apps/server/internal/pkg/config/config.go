// Package config 提供服务配置的加载与访问。
// 配置仅从环境变量读取，未设置时使用默认值；本包是唯一允许读取环境变量的地方（见 LLM_DEV_GUIDE.md §3.3）。
package config

import (
	"os"
	"regexp"
)

// Config 服务运行配置。
type Config struct {
	HTTPPort  string // HTTP 监听端口
	MySQLDSN  string // MySQL 连接串（须含 charset=utf8mb4）
	RedisAddr string // Redis 地址 host:port
	JWTSecret string // JWT HS256 签名密钥（JWT_SECRET 注入，禁硬编码 guide §9.3）
	LogLevel  string // 日志级别：debug/info/warn/error
	AppEnv    string // 运行环境：dev/prod
}

// Load 从环境变量加载配置，未设置的项使用默认值。
func Load() Config {
	return Config{
		HTTPPort:  getenv("HTTP_PORT", "8080"),
		MySQLDSN:  getenv("MYSQL_DSN", "root:yuyan123@tcp(127.0.0.1:3306)/yuyan?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr: getenv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret: getenv("JWT_SECRET", "dev-only-secret-change-me"),
		LogLevel:  getenv("LOG_LEVEL", "info"),
		AppEnv:    getenv("APP_ENV", "dev"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// dsnPasswordRe 匹配 DSN 中的密码部分（user:password@tcp(...)）。
var dsnPasswordRe = regexp.MustCompile(`:[^:@/]*@`)

// SafeString 返回脱敏后的配置摘要（DSN 密码替换为 ***），用于启动日志。
func (c Config) SafeString() string {
	return "port=" + c.HTTPPort +
		" mysql=" + dsnPasswordRe.ReplaceAllString(c.MySQLDSN, ":***@") +
		" redis=" + c.RedisAddr +
		" log=" + c.LogLevel +
		" env=" + c.AppEnv
}
