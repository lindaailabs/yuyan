package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	c := Load()
	if c.HTTPPort != "8080" {
		t.Errorf("default HTTPPort = %q, want 8080", c.HTTPPort)
	}
	if !strings.Contains(c.MySQLDSN, "charset=utf8mb4") {
		t.Errorf("default MySQLDSN should contain charset=utf8mb4, got %q", c.MySQLDSN)
	}
	if c.RedisAddr != "127.0.0.1:6379" {
		t.Errorf("default RedisAddr = %q, want 127.0.0.1:6379", c.RedisAddr)
	}
	if c.LogLevel != "info" {
		t.Errorf("default LogLevel = %q, want info", c.LogLevel)
	}
	if c.AppEnv != "dev" {
		t.Errorf("default AppEnv = %q, want dev", c.AppEnv)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("MYSQL_DSN", "user:secret@tcp(db:3306)/yuyan?charset=utf8mb4&parseTime=True")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("APP_ENV", "prod")

	c := Load()
	if c.HTTPPort != "9090" || c.RedisAddr != "redis:6379" || c.LogLevel != "debug" || c.AppEnv != "prod" {
		t.Errorf("overrides not applied: %+v", c)
	}
	if !strings.Contains(c.MySQLDSN, "db:3306") {
		t.Errorf("MySQLDSN override not applied: %q", c.MySQLDSN)
	}
}

func TestSafeStringMasksPassword(t *testing.T) {
	c := Config{MySQLDSN: "user:secret@tcp(127.0.0.1:3306)/yuyan?charset=utf8mb4"}
	s := c.SafeString()
	if strings.Contains(s, "secret") {
		t.Errorf("SafeString leaks password: %s", s)
	}
	if !strings.Contains(s, "user:***@") {
		t.Errorf("SafeString should mask password, got %s", s)
	}
}

func TestLoadFromFilePerEnv(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.test.yaml"
	yaml := `
app_env: "test"
mysql_dsn: "root:test@tcp(127.0.0.1:3306)/yuyan"
redis_addr: "127.0.0.1:6379"
ai_provider: "openai"
ai_model: "${AI_MODEL}"
`
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	// 用 CONFIG_FILE 显式加载本环境文件；${VAR} 展开。
	t.Setenv("CONFIG_FILE", path)
	t.Setenv("AI_MODEL", "gpt-4o")
	c := Load()
	if c.AppEnv != "test" || c.MySQLDSN != "root:test@tcp(127.0.0.1:3306)/yuyan" || c.RedisAddr != "127.0.0.1:6379" {
		t.Errorf("test 文件未生效: %+v", c)
	}
	if c.AIProvider != "openai" || c.AIModel != "gpt-4o" {
		t.Errorf("ai 字段未生效: %+v", c)
	}

	// CONFIG_FILE 优先于 APP_ENV 推导路径：即便 APP_ENV=prod，仍加载本文件（字段取自本文件）。
	// 注意：APP_ENV 作为环境变量仍会覆盖文件内的 app_env 字段，故 AppEnv 应为 prod。
	t.Setenv("APP_ENV", "prod")
	c = Load()
	if c.MySQLDSN != "root:test@tcp(127.0.0.1:3306)/yuyan" {
		t.Errorf("CONFIG_FILE 未优先加载本文件: %+v", c)
	}
	if c.AppEnv != "prod" {
		t.Errorf("APP_ENV 环境变量应覆盖文件 app_env，got %q", c.AppEnv)
	}

	// 环境变量覆盖配置文件字段（最高优先级）。
	t.Setenv("REDIS_ADDR", "redis:6379")
	c = Load()
	if c.RedisAddr != "redis:6379" {
		t.Errorf("环境变量覆盖未生效: %q", c.RedisAddr)
	}
}

// TestLoadFallbackWhenFileMissing 验证文件缺失时回退到内置默认值，不致命。
func TestLoadFallbackWhenFileMissing(t *testing.T) {
	t.Setenv("CONFIG_FILE", t.TempDir()+"/nope.yaml")
	c := Load()
	if c.HTTPPort != "8080" || c.AppEnv != "dev" || c.AIProvider != "mock" {
		t.Errorf("缺失文件未回退默认值: %+v", c)
	}
}
