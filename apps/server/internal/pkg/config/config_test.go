package config

import (
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
