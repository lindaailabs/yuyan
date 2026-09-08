package repo

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

// startMySQL 启动 MySQL 8 容器（testcontainers），带 users 表 migration。
// 注册到 t.Cleanup 自动销毁。
func startMySQL(t *testing.T) *mysql.MySQLContainer {
	t.Helper()
	ctx := context.Background()

	container, err := mysql.Run(ctx, "mysql:8.0",
		mysql.WithDatabase("yuyan"),
		mysql.WithUsername("yuyan"),
		mysql.WithPassword("yuyan123"),
	)
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})
	return container
}
