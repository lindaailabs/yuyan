package server

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

// newMySQLContainer 测试夹具：单容器复用（TestMain 级），这里按测试隔离简化为每测试一容器。
func newMySQLContainer(t *testing.T) *sql.DB {
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

	dsn := container.MustConnectionString(ctx) + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open sql: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestUsersTableStructure 验证 users 表结构与 guide §4 基线一致（openspec user-profile 规格）。
func TestUsersTableStructure(t *testing.T) {
	db := newMySQLContainer(t)

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	rows, err := db.Query("DESCRIBE users")
	if err != nil {
		t.Fatalf("DESCRIBE users: %v", err)
	}
	defer func() { _ = rows.Close() }()

	cols := map[string]string{} // 字段名 -> 类型
	for rows.Next() {
		var field, colType, null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			t.Fatalf("scan row: %v", err)
		}
		cols[field] = colType
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}

	want := map[string]string{
		"id":         "bigint unsigned",
		"phone":      "varchar(20)",
		"nickname":   "varchar(20)",
		"avatar_id":  "smallint",
		"created_at": "datetime",
		"updated_at": "datetime",
	}
	for field, typ := range want {
		got, ok := cols[field]
		if !ok {
			t.Errorf("users 表缺少字段 %s", field)
			continue
		}
		if got != typ {
			t.Errorf("字段 %s 类型 = %s, want %s", field, got, typ)
		}
	}

	// phone 唯一约束：重复插入必须失败。
	if _, err := db.Exec("INSERT INTO users (phone) VALUES ('13800138000')"); err != nil {
		t.Fatalf("首次插入失败: %v", err)
	}
	if _, err := db.Exec("INSERT INTO users (phone) VALUES ('13800138000')"); err == nil {
		t.Error("phone 重复插入应失败（唯一约束缺失）")
	}

	// 默认值：avatar_id=1。
	var avatarID int
	var nickname sql.NullString
	if err := db.QueryRow("SELECT avatar_id, nickname FROM users WHERE phone='13800138000'").Scan(&avatarID, &nickname); err != nil {
		t.Fatalf("查询默认值: %v", err)
	}
	if avatarID != 1 {
		t.Errorf("avatar_id 默认值 = %d, want 1", avatarID)
	}
	if nickname.Valid {
		t.Errorf("nickname 默认应为 NULL, got %q", nickname.String)
	}
}
