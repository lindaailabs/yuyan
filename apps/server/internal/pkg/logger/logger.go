// Package logger 初始化全局 slog 结构化日志（JSON handler，输出到 stdout）。
// 禁止使用 fmt.Println 输出日志（见 LLM_DEV_GUIDE.md §6.1）。
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init 按级别初始化全局 logger；未知级别回退为 info。
func Init(level string) {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	slog.SetDefault(slog.New(h))
}
