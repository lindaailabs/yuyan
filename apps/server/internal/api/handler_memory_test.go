package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

func TestHandlerMemoryListAndDelete(t *testing.T) {
	e, token, _, petID := newPetChatEnv(t)

	// 先通过对话形成一条记忆。
	_, sendResp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{
		"pet_id":  petID,
		"content": "我喜欢蓝色",
	})
	if sendResp.Code != 0 {
		t.Fatalf("send: code=%d msg=%s", sendResp.Code, sendResp.Msg)
	}
	var sent model.SendMessageResult
	if err := json.Unmarshal(sendResp.Data, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(sent.NewMemories) == 0 {
		t.Fatal("本轮应形成新记忆（new_memories 为空）")
	}
	if !strings.Contains(sent.NewMemories[0].Content, "蓝色") {
		t.Errorf("记忆内容异常: %+v", sent.NewMemories[0])
	}

	_, listResp := doJSON(t, e.r, http.MethodGet, "/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/memories", token, nil)
	if listResp.Code != 0 {
		t.Fatalf("list memories: code=%d msg=%s", listResp.Code, listResp.Msg)
	}
	var items []model.MemoryItem
	if err := json.Unmarshal(listResp.Data, &items); err != nil {
		t.Fatalf("unmarshal memories: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("记忆条数 = %d, want 1", len(items))
	}
	memoryID := items[0].ID

	// 后续对话应召回该记忆并体现在回复中。
	_, recallResp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{
		"pet_id":  petID,
		"content": "你记得我喜欢什么颜色吗",
	})
	if recallResp.Code != 0 {
		t.Fatalf("recall send: %d", recallResp.Code)
	}
	var recalled model.SendMessageResult
	if err := json.Unmarshal(recallResp.Data, &recalled); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if recalled.AssistantMessage == nil || !strings.Contains(recalled.AssistantMessage.Content, "蓝色") {
		t.Errorf("宠物回复应体现召回的记忆: %+v", recalled.AssistantMessage)
	}

	// 删除后不可见、不再召回。
	_, delResp := doJSON(t, e.r, http.MethodDelete, "/api/v1/pet-memories/"+strconv.FormatInt(memoryID, 10), token, nil)
	if delResp.Code != 0 {
		t.Fatalf("delete memory: code=%d msg=%s", delResp.Code, delResp.Msg)
	}

	_, afterList := doJSON(t, e.r, http.MethodGet, "/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/memories", token, nil)
	if afterList.Code != 0 {
		t.Fatalf("list after delete: %d", afterList.Code)
	}
	var left []model.MemoryItem
	if err := json.Unmarshal(afterList.Data, &left); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(left) != 0 {
		t.Errorf("删除后仍有 %d 条记忆", len(left))
	}

	_, afterRecall := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{
		"pet_id":  petID,
		"content": "我喜欢什么颜色呢",
	})
	var after model.SendMessageResult
	if err := json.Unmarshal(afterRecall.Data, &after); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if after.AssistantMessage != nil && strings.Contains(after.AssistantMessage.Content, "蓝色") {
		t.Error("删除后不应再召回该记忆")
	}
}

func TestHandlerMemoryErrorPaths(t *testing.T) {
	e, token, bToken, petID := newPetChatEnv(t)

	// 未认证 → 1002
	if _, resp := doJSON(t, e.r, http.MethodGet, "/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/memories", "", nil); resp.Code != 1002 {
		t.Errorf("未认证 code = %d, want 1002", resp.Code)
	}
	// 他人宠物 → 2303
	if _, resp := doJSON(t, e.r, http.MethodGet, "/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/memories", bToken, nil); resp.Code != 2303 {
		t.Errorf("他人宠物 code = %d, want 2303", resp.Code)
	}
	// 删除不存在的记忆 → 2401
	if _, resp := doJSON(t, e.r, http.MethodDelete, "/api/v1/pet-memories/999999", token, nil); resp.Code != 2401 {
		t.Errorf("删除不存在记忆 code = %d, want 2401", resp.Code)
	}
	// 非法 id → 1001
	if _, resp := doJSON(t, e.r, http.MethodDelete, "/api/v1/pet-memories/0", token, nil); resp.Code != 1001 {
		t.Errorf("非法 id code = %d, want 1001", resp.Code)
	}
}
