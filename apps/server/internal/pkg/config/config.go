// Package config 提供服务配置的加载与访问。
// 配置来源优先级（高→低）：同名环境变量 > 配置文件字段 > 内置默认值。
// 配置文件按 APP_ENV 选择独立文件 config.{env}.yaml（如 config.test.yaml / config.prod.yaml）；
// 也可用 CONFIG_FILE 显式指定路径。生产密钥用 ${ENV_VAR} 引用环境变量，不落明文。
// 本包是唯一允许读取环境变量与配置文件的地方（见 LLM_DEV_GUIDE.md §3.3）。
package config

import (
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 服务运行配置。
type Config struct {
	HTTPPort  string // HTTP 监听端口
	MySQLDSN  string // MySQL 连接串（须含 charset=utf8mb4）
	RedisAddr string // Redis 地址 host:port
	JWTSecret string // JWT HS256 签名密钥（JWT_SECRET 注入，禁硬编码 guide §9.3）
	LogLevel  string // 日志级别：debug/info/warn/error
	AppEnv    string // 运行环境：dev/test/prod

	AIProvider     string  // AI Gateway provider：mock（默认）/ openai（OpenAI 兼容）
	AITimeoutMS    int     // 单次模型调用超时（毫秒）
	AIMockFailRate float64 // 仅 mock 生效：注入失败率，用于验证兜底路径

	// 真实模型（provider=openai）配置；缺失 key 时启动即失败，不静默降级。
	AIBaseURL string // OpenAI 兼容端点，如 https://api.openai.com/v1
	AIAPIKey  string // 模型服务 API Key
	AIModel   string // 模型名，如 gpt-3.5-turbo / moonshot-v1-8k
}

// fileConfig 配置文件为扁平键值（值支持 ${ENV_VAR} 占位展开）。
type fileConfig map[string]any

// Option 调整 Load 行为（主要来自命令行参数，避免必须设置环境变量）。
type Option func(*loadOption)

type loadOption struct {
	env        string // 显式环境（命令行 -env），覆盖 APP_ENV 推导
	configFile string // 显式文件（命令行 -config），覆盖 CONFIG_FILE
}

// WithEnv 指定运行环境（对应 config.<env>.yaml），优先级高于 APP_ENV 环境变量。
func WithEnv(env string) Option { return func(o *loadOption) { o.env = env } }

// WithConfigFile 显式指定配置文件路径，优先级高于 CONFIG_FILE 环境变量。
func WithConfigFile(path string) Option { return func(o *loadOption) { o.configFile = path } }

// Load 从「环境变量 > 配置文件 > 默认值」加载配置。
// 配置文件路径优先级：WithConfigFile 命令行参数 > CONFIG_FILE 环境变量 > config.{env}.yaml > config.yaml（回退默认）。
// env 优先级：WithEnv 命令行参数 > APP_ENV 环境变量 > 默认 dev。
func Load(opts ...Option) Config {
	o := &loadOption{}
	for _, opt := range opts {
		opt(o)
	}

	env := o.env
	if env == "" {
		env = getenv("APP_ENV", "dev")
	}

	path := o.configFile
	if path == "" {
		path = getenv("CONFIG_FILE", "")
	}
	if path == "" {
		path = "config." + env + ".yaml"
	}
	file := loadFile(path)
	// 按 env 命名的文件缺失时，回退到 config.yaml（默认 dev 配置）。
	if len(file) == 0 && path != "config.yaml" {
		file = loadFile("config.yaml")
	}

	l := &loader{file: file}
	return Config{
		HTTPPort:       l.string("http_port", "8080"),
		MySQLDSN:       l.string("mysql_dsn", "root:yuyan123@tcp(127.0.0.1:3306)/yuyan?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:      l.string("redis_addr", "127.0.0.1:6379"),
		JWTSecret:      l.string("jwt_secret", "dev-only-secret-change-me"),
		LogLevel:       l.string("log_level", "info"),
		AppEnv:         l.string("app_env", "dev"),
		AIProvider:     l.string("ai_provider", "mock"),
		AITimeoutMS:    l.int("ai_timeout_ms", 8000),
		AIMockFailRate: l.float("ai_mock_fail_rate", 0),
		AIBaseURL:      l.string("ai_base_url", "https://api.openai.com/v1"),
		AIAPIKey:       l.string("ai_api_key", ""),
		AIModel:        l.string("ai_model", "gpt-3.5-turbo"),
	}
}

// loader 按优先级解析单个配置字段。
type loader struct {
	file fileConfig
}

// raw 返回字段原始字符串值（已展开 ${ENV}），命中优先级：环境变量 > 配置文件。
// 未命中返回 ("", false)。（nil map 索引在 Go 中安全，返回零值 + false。）
func (l *loader) raw(key string) (string, bool) {
	if v := os.Getenv(strings.ToUpper(key)); v != "" { // 环境变量优先（同名大写）
		return v, true
	}
	if v, ok := l.file[key]; ok {
		return expandEnv(toString(v)), true
	}
	return "", false
}

func (l *loader) string(key, def string) string {
	if v, ok := l.raw(key); ok {
		return v
	}
	return def
}

func (l *loader) int(key string, def int) int {
	if v, ok := l.raw(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func (l *loader) float(key string, def float64) float64 {
	if v, ok := l.raw(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// loadFile 读取并解析扁平配置文件；文件缺失或解析失败均回退到环境变量/默认值（不致命）。
func loadFile(path string) fileConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		slog.Warn("config file parse failed, fallback to env/default", "path", path, "err", err)
		return nil
	}
	return fc
}

// expandEnv 展开 ${VAR} / $VAR 为对应环境变量值；未设置则替换为空串。
func expandEnv(s string) string {
	return os.Expand(s, os.Getenv)
}

// toString 把 yaml 反序列化的 any 转成字符串（兼容字符串/数字/布尔字面量）。
func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	case nil:
		return ""
	default:
		return ""
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
		" env=" + c.AppEnv +
		" ai_provider=" + c.AIProvider +
		" ai_timeout_ms=" + strconv.Itoa(c.AITimeoutMS) +
		" ai_mock_fail_rate=" + strconv.FormatFloat(c.AIMockFailRate, 'f', 2, 64) +
		" ai_model=" + c.AIModel
}
