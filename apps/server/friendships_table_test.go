package server

import (
	"database/sql"
	"testing"
)

// TestFriendshipsTableStructure 验证 friendships 表结构与 guide §4 基线一致
// （openspec contacts-service 规格：uk_user_friend 唯一键 + 双向双行模型）。
func TestFriendshipsTableStructure(t *testing.T) {
	db := newMySQLContainer(t)

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	rows, err := db.Query("DESCRIBE friendships")
	if err != nil {
		t.Fatalf("DESCRIBE friendships: %v", err)
	}
	defer func() { _ = rows.Close() }()

	cols := map[string]string{}
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
		"user_id":    "bigint unsigned",
		"friend_id":  "bigint unsigned",
		"status":     "smallint",
		"created_at": "datetime",
	}
	for field, typ := range want {
		got, ok := cols[field]
		if !ok {
			t.Errorf("friendships 表缺少字段 %s", field)
			continue
		}
		if got != typ {
			t.Errorf("字段 %s 类型 = %s, want %s", field, got, typ)
		}
	}

	// 前置用户数据（外键语义由服务层保证，表内仅存 id）。
	if _, err := db.Exec("INSERT INTO users (phone) VALUES ('13800000001'), ('13800000002')"); err != nil {
		t.Fatalf("插入用户: %v", err)
	}

	// 唯一键：同向重复插入必须失败；双向各一行必须允许。
	if _, err := db.Exec("INSERT INTO friendships (user_id, friend_id) VALUES (1, 2)"); err != nil {
		t.Fatalf("首次插入失败: %v", err)
	}
	if _, err := db.Exec("INSERT INTO friendships (user_id, friend_id) VALUES (1, 2)"); err == nil {
		t.Error("同向重复插入应失败（uk_user_friend 缺失）")
	}
	if _, err := db.Exec("INSERT INTO friendships (user_id, friend_id) VALUES (2, 1)"); err != nil {
		t.Errorf("反向行插入应成功（双向双行模型）: %v", err)
	}

	// 默认值：status=1（pending）。
	var status int
	if err := db.QueryRow("SELECT status FROM friendships WHERE user_id=1 AND friend_id=2").Scan(&status); err != nil {
		t.Fatalf("查询默认值: %v", err)
	}
	if status != 1 {
		t.Errorf("status 默认值 = %d, want 1", status)
	}
}
