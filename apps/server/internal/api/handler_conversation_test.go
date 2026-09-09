package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// petChatEnv 准备两个用户与其宠物，返回各自的 token 与宠物 id。
func newPetChatEnv(t *testing.T) (e *handlerEnv, aToken, bToken string, aPetID int64) {
	t.Helper()
	e = newHandlerEnv(t)
	aToken, _ = loginByPhone(t, e, "13803000001")
	bToken, _ = loginByPhone(t, e, "13803000002")

	_, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pets", aToken, map[string]any{"name": "小燕"})
	if resp.Code != 0 {
		t.Fatalf("create pet: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var pet model.PetProfile
	if err := json.Unmarshal(resp.Data, &pet); err != nil {
		t.Fatalf("unmarshal pet: %v", err)
	}
	return e, aToken, bToken, pet.ID
}

func TestHandlerConversationHappyPath(t *testing.T) {
	e, token, _, petID := newPetChatEnv(t)

	_, convResp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-conversations", token, map[string]any{"pet_id": petID})
	if convResp.Code != 0 {
		t.Fatalf("create conversation: code=%d msg=%s", convResp.Code, convResp.Msg)
	}
	var conv model.ConversationItem
	if err := json.Unmarshal(convResp.Data, &conv); err != nil {
		t.Fatalf("unmarshal conversation: %v", err)
	}
	if conv.PetID != petID || conv.ID == 0 {
		t.Fatalf("会话数据异常: %+v", conv)
	}

	_, sendResp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{
		"pet_id": petID, "content": "你好呀",
	})
	if sendResp.Code != 0 {
		t.Fatalf("send message: code=%d msg=%s", sendResp.Code, sendResp.Msg)
	}
	var sent model.SendMessageResult
	if err := json.Unmarshal(sendResp.Data, &sent); err != nil {
		t.Fatalf("unmarshal send result: %v", err)
	}
	if sent.UserMessage.Role != model.MessageRoleUser || sent.AssistantMessage == nil {
		t.Fatalf("返回消息不完整: %+v", sent)
	}
	if sent.AssistantMessage.Role != model.MessageRoleAssistant || sent.AssistantMessage.Content == "" {
		t.Fatalf("宠物回复异常: %+v", sent.AssistantMessage)
	}
	if sent.Streaming {
		t.Error("本变更为非流式，streaming 应为 false")
	}
	if sent.Usage.Model == "" || sent.Usage.InputTokens <= 0 {
		t.Errorf("用量统计缺失: %+v", sent.Usage)
	}

	_, histResp := doJSON(t, e.r, http.MethodGet, "/api/v1/pet-messages?conv_id="+strconv.FormatInt(conv.ID, 10)+"&limit=20", token, nil)
	if histResp.Code != 0 {
		t.Fatalf("history: code=%d msg=%s", histResp.Code, histResp.Msg)
	}
	var page model.MessagePage
	if err := json.Unmarshal(histResp.Data, &page); err != nil {
		t.Fatalf("unmarshal page: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("历史消息 = %d, want 2", len(page.Items))
	}
	if page.HasMore {
		t.Error("仅 2 条消息不应 has_more")
	}
}

func TestHandlerConversationErrorPaths(t *testing.T) {
	e, token, bToken, petID := newPetChatEnv(t)

	// 未认证 → 1002
	if _, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-conversations", "", map[string]any{"pet_id": petID}); resp.Code != 1002 {
		t.Errorf("未认证 code = %d, want 1002", resp.Code)
	}
	// pet_id 非法 → 1001
	if _, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-conversations", token, map[string]any{"pet_id": 0}); resp.Code != 1001 {
		t.Errorf("pet_id=0 code = %d, want 1001", resp.Code)
	}
	// 宠物不存在 → 2303
	if _, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-conversations", token, map[string]any{"pet_id": 999999}); resp.Code != 2303 {
		t.Errorf("宠物不存在 code = %d, want 2303", resp.Code)
	}
	// 他人宠物 → 2303
	if _, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-conversations", bToken, map[string]any{"pet_id": petID}); resp.Code != 2303 {
		t.Errorf("他人宠物 code = %d, want 2303", resp.Code)
	}
	// content 缺失 → 1001
	if _, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{"pet_id": petID}); resp.Code != 1001 {
		t.Errorf("content 缺失 code = %d, want 1001", resp.Code)
	}
	// content 空白 → 2302
	if _, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, map[string]any{"pet_id": petID, "content": "   "}); resp.Code != 2302 {
		t.Errorf("content 空白 code = %d, want 2302", resp.Code)
	}
	// 历史缺 conv_id → 1001
	if _, resp := doJSON(t, e.r, http.MethodGet, "/api/v1/pet-messages", token, nil); resp.Code != 1001 {
		t.Errorf("缺少 conv_id code = %d, want 1001", resp.Code)
	}
}

func TestHandlerHistoryAccessControl(t *testing.T) {
	e, token, bToken, petID := newPetChatEnv(t)

	_, convResp := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-conversations", token, map[string]any{"pet_id": petID})
	if convResp.Code != 0 {
		t.Fatalf("create conversation: %d", convResp.Code)
	}
	var conv model.ConversationItem
	if err := json.Unmarshal(convResp.Data, &conv); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, resp := doJSON(t, e.r, http.MethodGet, "/api/v1/pet-messages?conv_id="+strconv.FormatInt(conv.ID, 10), bToken, nil); resp.Code != 2301 {
		t.Errorf("他人读取会话 code = %d, want 2301", resp.Code)
	}
}

func TestHandlerSendMessageIdempotent(t *testing.T) {
	e, token, _, petID := newPetChatEnv(t)

	body := map[string]any{
		"pet_id":        petID,
		"content":       "重复提交测试",
		"client_msg_id": "22222222-2222-2222-2222-222222222222",
	}
	_, first := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, body)
	if first.Code != 0 {
		t.Fatalf("first send: %d", first.Code)
	}
	_, second := doJSON(t, e.r, http.MethodPost, "/api/v1/pet-messages", token, body)
	if second.Code != 0 {
		t.Fatalf("replay send: %d", second.Code)
	}
	if first.Data == nil || string(first.Data) != string(second.Data) {
		t.Errorf("重放应返回相同结果:\n%s\n%s", first.Data, second.Data)
	}
}
