package server

import (
	"database/sql"
	"testing"
)

// describeColumns 读取表结构：字段名 → 类型。
func describeColumns(t *testing.T, db *sql.DB, table string) map[string]string {
	t.Helper()
	rows, err := db.Query("DESCRIBE " + table)
	if err != nil {
		t.Fatalf("DESCRIBE %s: %v", table, err)
	}
	defer func() { _ = rows.Close() }()

	cols := map[string]string{}
	for rows.Next() {
		var field, colType, null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			t.Fatalf("scan %s column: %v", table, err)
		}
		cols[field] = colType
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("%s rows err: %v", table, err)
	}
	return cols
}

// assertColumns 断言表必须包含给定字段与类型。
func assertColumns(t *testing.T, db *sql.DB, table string, want map[string]string) {
	t.Helper()
	cols := describeColumns(t, db, table)
	for field, typ := range want {
		got, ok := cols[field]
		if !ok {
			t.Errorf("%s 表缺少字段 %s", table, field)
			continue
		}
		if got != typ {
			t.Errorf("%s 字段 %s 类型 = %s, want %s", table, field, got, typ)
		}
	}
}

// hasIndex 断言表存在指定索引（走 information_schema，避免 SHOW INDEX 列数差异）。
func hasIndex(t *testing.T, db *sql.DB, table, index string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.statistics "+
			"WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		table, index).Scan(&n)
	if err != nil {
		t.Fatalf("query index %s.%s: %v", table, index, err)
	}
	return n > 0
}

// TestPetChatTablesStructure 验证会话/消息/调用日志三张表与 guide §4 基线一致。
func TestPetChatTablesStructure(t *testing.T) {
	db := newMySQLContainer(t)

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	assertColumns(t, db, "pet_conversations", map[string]string{
		"id":               "bigint unsigned",
		"user_id":          "bigint unsigned",
		"pet_id":           "bigint unsigned",
		"last_msg_id":      "bigint unsigned",
		"last_msg_preview": "varchar(255)",
		"updated_at":       "datetime",
	})
	if !hasIndex(t, db, "pet_conversations", "uk_user_pet") {
		t.Error("pet_conversations 缺少唯一键 uk_user_pet（一人一宠物一条会话）")
	}
	if !hasIndex(t, db, "pet_conversations", "idx_user_updated") {
		t.Error("pet_conversations 缺少索引 idx_user_updated")
	}

	assertColumns(t, db, "pet_messages", map[string]string{
		"id":            "bigint unsigned",
		"conv_id":       "bigint unsigned",
		"user_id":       "bigint unsigned",
		"pet_id":        "bigint unsigned",
		"role":          "varchar(16)",
		"content":       "text",
		"client_msg_id": "char(36)",
		"model":         "varchar(64)",
		"input_tokens":  "int",
		"output_tokens": "int",
		"latency_ms":    "int",
		"status":        "smallint",
		"error_code":    "int",
		"created_at":    "bigint",
	})
	if !hasIndex(t, db, "pet_messages", "uk_user_client_msg_id") {
		t.Error("pet_messages 缺少唯一键 uk_user_client_msg_id（用户维度幂等依赖）")
	}
	if !hasIndex(t, db, "pet_messages", "idx_conv_id_id") {
		t.Error("pet_messages 缺少索引 idx_conv_id_id（游标分页依赖）")
	}

	assertColumns(t, db, "ai_call_logs", map[string]string{
		"id":             "bigint unsigned",
		"user_id":        "bigint unsigned",
		"pet_id":         "bigint unsigned",
		"conv_id":        "bigint unsigned",
		"msg_id":         "bigint unsigned",
		"provider":       "varchar(32)",
		"model":          "varchar(64)",
		"prompt_version": "varchar(32)",
		"input_tokens":   "int",
		"output_tokens":  "int",
		"latency_ms":     "int",
		"err_code":       "int",
		"cache_hit":      "tinyint",
		"created_at":     "bigint",
	})

	// 幂等兜底：同一用户的 client_msg_id 二次插入必须失败；不同用户可复用同一 UUID。
	if _, err := db.Exec("INSERT INTO users (phone) VALUES ('13800138888'), ('13800139999')"); err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	if _, err := db.Exec("INSERT INTO pets (user_id, name) VALUES (1, '小燕'), (2, '小羽')"); err != nil {
		t.Fatalf("插入宠物失败: %v", err)
	}
	if _, err := db.Exec("INSERT INTO pet_conversations (user_id, pet_id) VALUES (1, 1), (2, 2)"); err != nil {
		t.Fatalf("插入会话失败: %v", err)
	}
	const clientMsgID = "11111111-2222-3333-4444-555555555555"
	const insertMsgUser1 = "INSERT INTO pet_messages (conv_id, user_id, pet_id, role, content, client_msg_id, created_at) VALUES (1, 1, 1, 'user', '你好', '" + clientMsgID + "', 1788900000)"
	const insertMsgUser2 = "INSERT INTO pet_messages (conv_id, user_id, pet_id, role, content, client_msg_id, created_at) VALUES (2, 2, 2, 'user', '你好', '" + clientMsgID + "', 1788900000)"
	if _, err := db.Exec(insertMsgUser1); err != nil {
		t.Fatalf("首次插入消息失败: %v", err)
	}
	if _, err := db.Exec(insertMsgUser1); err == nil {
		t.Error("同一用户相同 client_msg_id 重复插入应当失败（uk_user_client_msg_id 未生效）")
	}
	if _, err := db.Exec(insertMsgUser2); err != nil {
		t.Fatalf("不同用户复用 client_msg_id 应当成功: %v", err)
	}
}
