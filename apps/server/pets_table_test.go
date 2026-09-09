package server

import (
	"database/sql"
	"testing"
)

// TestPetsTableStructure 验证 pets 表结构与 AI 宠物档案基线一致。
func TestPetsTableStructure(t *testing.T) {
	db := newMySQLContainer(t)

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	rows, err := db.Query("DESCRIBE pets")
	if err != nil {
		t.Fatalf("DESCRIBE pets: %v", err)
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
		"name":       "varchar(20)",
		"species":    "varchar(32)",
		"avatar_id":  "smallint",
		"persona":    "text",
		"level":      "int",
		"intimacy":   "int",
		"mood":       "varchar(32)",
		"created_at": "datetime",
		"updated_at": "datetime",
	}
	for field, typ := range want {
		got, ok := cols[field]
		if !ok {
			t.Errorf("pets 表缺少字段 %s", field)
			continue
		}
		if got != typ {
			t.Errorf("字段 %s 类型 = %s, want %s", field, got, typ)
		}
	}

	if _, err := db.Exec("INSERT INTO users (phone) VALUES ('13800139999')"); err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	if _, err := db.Exec("INSERT INTO pets (user_id, name) VALUES (1, '小燕')"); err != nil {
		t.Fatalf("插入宠物失败: %v", err)
	}
	var species, mood string
	var avatarID, level, intimacy int
	if err := db.QueryRow("SELECT species, avatar_id, level, intimacy, mood FROM pets WHERE user_id=1").Scan(&species, &avatarID, &level, &intimacy, &mood); err != nil {
		t.Fatalf("查询默认值: %v", err)
	}
	if species != "swallow" || avatarID != 1 || level != 1 || intimacy != 0 || mood != "curious" {
		t.Errorf("pets 默认值异常: species=%s avatar=%d level=%d intimacy=%d mood=%s", species, avatarID, level, intimacy, mood)
	}
}
