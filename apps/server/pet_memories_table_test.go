package server

import "testing"

// TestPetMemoriesTableStructure 验证 pet_memories 表与 guide §4 基线一致。
func TestPetMemoriesTableStructure(t *testing.T) {
	db := newMySQLContainer(t)

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	assertColumns(t, db, "pet_memories", map[string]string{
		"id":            "bigint unsigned",
		"user_id":       "bigint unsigned",
		"pet_id":        "bigint unsigned",
		"memory_type":   "varchar(32)",
		"content":       "text",
		"content_hash":  "char(64)",
		"source_msg_id": "bigint unsigned",
		"confidence":    "decimal(4,3)",
		"last_used_at":  "datetime",
		"status":        "smallint",
		"created_at":    "datetime",
		"updated_at":    "datetime",
	})
	if !hasIndex(t, db, "pet_memories", "uk_pet_hash") {
		t.Error("pet_memories 缺少唯一键 uk_pet_hash（记忆去重依赖）")
	}
	if !hasIndex(t, db, "pet_memories", "idx_pet_status") {
		t.Error("pet_memories 缺少索引 idx_pet_status（active 召回依赖）")
	}

	if _, err := db.Exec("INSERT INTO users (phone) VALUES ('13800137777')"); err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	if _, err := db.Exec("INSERT INTO pets (user_id, name) VALUES (1, '小燕')"); err != nil {
		t.Fatalf("插入宠物失败: %v", err)
	}
	const insertMem = `INSERT INTO pet_memories (user_id, pet_id, memory_type, content, content_hash, status)
	                   VALUES (1, 1, 'preference', '用户喜欢蓝色', 'hash-0001', 1)`
	if _, err := db.Exec(insertMem); err != nil {
		t.Fatalf("首次插入记忆失败: %v", err)
	}
	if _, err := db.Exec(insertMem); err == nil {
		t.Error("同一 (pet_id, content_hash) 重复插入应当失败（uk_pet_hash 未生效）")
	}

	var status int
	if err := db.QueryRow("SELECT status FROM pet_memories WHERE pet_id=1").Scan(&status); err != nil {
		t.Fatalf("查询默认状态: %v", err)
	}
	if status != 1 {
		t.Errorf("默认 status = %d, want 1", status)
	}
}
