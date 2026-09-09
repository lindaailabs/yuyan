// Package config 提供服务配置的加载与访问。
// 配置仅从环境变量读取，未设置时使用默认值；本包是唯一允许读取环境变量的地方（见 LLM_DEV_GUIDE.md §3.3）。
package config

import (
	"os"
	"regexp"
	"strconv"
)

// Config 服务运行配置。
type Config struct {
	HTTPPort  string // HTTP 监听端口
	MySQLDSN  string // MySQL 连接串（须含 charset=utf8mb4）
	RedisAddr string // Redis 地址 host:port
	JWTSecret string // JWT HS256 签名密钥（JWT_SECRET 注入，禁硬编码 guide §9.3）
	LogLevel  string // 日志级别：debug/info/warn/error
	AppEnv    string // 运行环境：dev/prod

	AIProvider     string  // AI Gateway provider：mock（默认）；未实现取值启动即失败
	AITimeoutMS    int     // 单次模型调用超时（毫秒）
	AIMockFailRate float64 // 仅 mock 生效：注入失败率，用于验证兜底路径
}

// Load 从环境变量加载配置，未设置的项使用默认值。
func Load() Config {
	return Config{
		HTTPPort:       getenv("HTTP_PORT", "8080"),
		MySQLDSN:       getenv("MYSQL_DSN", "root:yuyan123@tcp(127.0.0.1:3306)/yuyan?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:      getenv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:      getenv("JWT_SECRET", "dev-only-secret-change-me"),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		AppEnv:         getenv("APP_ENV", "dev"),
		AIProvider:     getenv("AI_PROVIDER", "mock"),
		AITimeoutMS:    getenvInt("AI_TIMEOUT_MS", 8000),
		AIMockFailRate: getenvFloat("AI_MOCK_FAIL_RATE", 0),
	}
}

// getenvInt 读取整型环境变量，非法值回退默认值。
func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// getenvFloat 读取浮点环境变量，非法值回退默认值。
func getenvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
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
		" env=" + c.AppEnv +
		" ai_provider=" + c.AIProvider +
		" ai_timeout_ms=" + strconv.Itoa(c.AITimeoutMS) +
		" ai_mock_fail_rate=" + strconv.FormatFloat(c.AIMockFailRate, 'f', 2, 64)
}
