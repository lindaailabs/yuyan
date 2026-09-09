package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

func TestHandlerGrowthEvents(t *testing.T) {
	e, token, bToken, petID := newPetChatEnv(t)

	// 对话触发成长结算。
	for i := 0; i < 2; i++ {
		_, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{
			"pet_id":  petID,
			"content": "今天也很开心",
		})
		if resp.Code != 0 {
			t.Fatalf("send %d: code=%d msg=%s", i, resp.Code, resp.Msg)
		}
	}

	_, listResp := doJSON(t, e.r, http.MethodGet,
		"/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/growth-events", token, nil)
	if listResp.Code != 0 {
		t.Fatalf("growth events: code=%d msg=%s", listResp.Code, listResp.Msg)
	}
	var items []model.GrowthEventItem
	if err := json.Unmarshal(listResp.Data, &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("成长事件为空")
	}
	if items[0].EventType != model.GrowthEventMessage {
		t.Errorf("最新事件类型 = %s, want %s", items[0].EventType, model.GrowthEventMessage)
	}
	if items[0].Reason == "" {
		t.Error("成长事件缺少可解释的原因")
	}

	// 宠物状态应随互动提升。
	_, stateResp := doJSON(t, e.r, http.MethodGet, "/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/state", token, nil)
	if stateResp.Code != 0 {
		t.Fatalf("state: %d", stateResp.Code)
	}
	var state model.PetState
	if err := json.Unmarshal(stateResp.Data, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Intimacy <= 0 {
		t.Errorf("亲密度未增长: %+v", state)
	}

	// 越权与未认证。
	if _, resp := doJSON(t, e.r, http.MethodGet,
		"/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/growth-events", bToken, nil); resp.Code != 2303 {
		t.Errorf("他人查看 code = %d, want 2303", resp.Code)
	}
	if _, resp := doJSON(t, e.r, http.MethodGet,
		"/api/v1/pets/"+strconv.FormatInt(petID, 10)+"/growth-events", "", nil); resp.Code != 1002 {
		t.Errorf("未认证 code = %d, want 1002", resp.Code)
	}
	if _, resp := doJSON(t, e.r, http.MethodGet, "/api/v1/pets/0/growth-events", token, nil); resp.Code != 1001 {
		t.Errorf("非法 pet id code = %d, want 1001", resp.Code)
	}
}
