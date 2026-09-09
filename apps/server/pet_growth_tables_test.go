package server

import "testing"

// TestPetGrowthTablesStructure 验证成长事件与每日统计表结构。
func TestPetGrowthTablesStructure(t *testing.T) {
	db := newMySQLContainer(t)

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	assertColumns(t, db, "pet_growth_events", map[string]string{
		"id":            "bigint unsigned",
		"user_id":       "bigint unsigned",
		"pet_id":        "bigint unsigned",
		"event_type":    "varchar(32)",
		"delta":         "json",
		"reason":        "varchar(255)",
		"source_msg_id": "bigint unsigned",
		"created_at":    "datetime",
	})
	if !hasIndex(t, db, "pet_growth_events", "uk_pet_source_type") {
		t.Error("pet_growth_events 缺少唯一键 uk_pet_source_type（成长结算幂等依赖）")
	}
	if !hasIndex(t, db, "pet_growth_events", "idx_pet_created") {
		t.Error("pet_growth_events 缺少索引 idx_pet_created")
	}

	assertColumns(t, db, "pet_daily_stats", map[string]string{
		"id":             "bigint unsigned",
		"pet_id":         "bigint unsigned",
		"last_date":      "char(10)",
		"streak_days":    "int",
		"total_messages": "int",
		"updated_at":     "datetime",
	})
	if !hasIndex(t, db, "pet_daily_stats", "uk_pet") {
		t.Error("pet_daily_stats 缺少唯一键 uk_pet")
	}

	// 幂等闸门：同一 (pet_id, source_msg_id, event_type) 只能有一条。
	const insertEvent = `INSERT INTO pet_growth_events (user_id, pet_id, event_type, reason, source_msg_id)
	                     VALUES (1, 1, 'message', '完成一次对话，亲密度 +2', 11)`
	if _, err := db.Exec(insertEvent); err != nil {
		t.Fatalf("首次插入成长事件失败: %v", err)
	}
	if _, err := db.Exec(insertEvent); err == nil {
		t.Error("同一来源消息的同类事件重复插入应当失败（uk_pet_source_type 未生效）")
	}
}
