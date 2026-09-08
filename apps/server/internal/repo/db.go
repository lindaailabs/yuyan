// Package repo 是数据访问层：GORM CRUD 与原生 SQL 的唯一发生地。
// 调用方向约束：api/ws → service → repo（LLM_DEV_GUIDE.md §3.3）。
package repo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB 建立 GORM（MySQL）连接。
func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// NewRedis 构造 Redis 客户端。
func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

// PingRedis 对 Redis 执行一次 PING（3s 超时），返回错误（可为 nil）。
func PingRedis(ctx context.Context, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}
