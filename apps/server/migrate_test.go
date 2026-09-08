package server

import (
	"context"
	"database/sql"
	"testing"

	// MySQL 驱动注册（sql.Open 使用）。
	_ "github.com/go-sql-driver/mysql"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

// TestMigrateUpIdempotent 在真实 MySQL 8 容器上验证：
// 1) 空库首次执行成功且 schema_migrations 记录版本；
// 2) 重复执行不报错（幂等）。
func TestMigrateUpIdempotent(t *testing.T) {
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

	dsn := container.MustConnectionString(ctx) + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open sql: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := MigrateUp(db); err != nil {
		t.Fatalf("first MigrateUp: %v", err)
	}
	if err := MigrateUp(db); err != nil {
		t.Fatalf("second MigrateUp should be no-op, got: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("schema_migrations rows = %d, want 1", count)
	}
}
