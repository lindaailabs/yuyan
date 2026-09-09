package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// friendRequestPayload POST /friends/requests 请求体。
type friendRequestPayload struct {
	UserID int64 `json:"user_id"`
}

// loginAs 便捷封装：登录并完成首登引导（nickname 非空），返回 token 与 uid。
func loginAs(t *testing.T, e *handlerEnv, phone, nickname string) string {
	t.Helper()
	access, _ := loginByPhone(t, e, phone, "secret123")
	if nickname != "" {
		_, resp := doJSON(t, e.r, http.MethodPut, "/api/v1/users/me", access, map[string]any{"nickname": nickname})
		if resp.Code != 0 {
			t.Fatalf("引导 %s: code=%d msg=%s", phone, resp.Code, resp.Msg)
		}
	}
	return access
}

// meID 取当前用户 id。
func meID(t *testing.T, e *handlerEnv, token string) int64 {
	t.Helper()
	_, resp := doJSON(t, e.r, http.MethodGet, "/api/v1/users/me", token, nil)
	if resp.Code != 0 {
		t.Fatalf("me: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var me struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &me); err != nil {
		t.Fatalf("unmarshal me: %v", err)
	}
	return me.ID
}

func TestFriendsRequestEndpoints(t *testing.T) {
	e := newHandlerEnv(t)
	tokenA := loginAs(t, e, "13800000001", "用户A")
	tokenB := loginAs(t, e, "13800000002", "用户B")
	uidA, uidB := meID(t, e, tokenA), meID(t, e, tokenB)

	// happy：A 向 B 发申请。
	_, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenA, friendRequestPayload{UserID: uidB})
	if resp.Code != 0 {
		t.Fatalf("发起申请: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var created struct {
		ID     int64 `json:"id"`
		Status int16 `json:"status"`
	}
	if err := json.Unmarshal(resp.Data, &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if created.Status != 1 {
		t.Errorf("申请状态 = %d, want 1(pending)", created.Status)
	}

	// error：无 token（1002）。
	w, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", "", friendRequestPayload{UserID: uidB})
	if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
		t.Errorf("无 token: http=%d code=%d, want 401/1002", w.Code, resp.Code)
	}

	// error：参数缺失（1001）。
	w, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenA, map[string]any{})
	if w.Code != http.StatusBadRequest || resp.Code != 1001 {
		t.Errorf("缺 user_id: http=%d code=%d, want 400/1001", w.Code, resp.Code)
	}

	// error：加自己（2102）。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenA, friendRequestPayload{UserID: uidA})
	if resp.Code != 2102 {
		t.Errorf("加自己: code=%d, want 2102", resp.Code)
	}

	// happy：B 的申请列表可见 A（脱敏手机号）。
	_, resp = doJSON(t, e.r, http.MethodGet, "/api/v1/friends/requests", tokenB, nil)
	if resp.Code != 0 {
		t.Fatalf("申请列表: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var items []struct {
		ID       int64 `json:"id"`
		FromUser struct {
			ID    int64  `json:"id"`
			Phone string `json:"phone"`
		} `json:"from_user"`
	}
	if err := json.Unmarshal(resp.Data, &items); err != nil {
		t.Fatalf("unmarshal items: %v", err)
	}
	if len(items) != 1 || items[0].FromUser.ID != uidA || items[0].FromUser.Phone != "138****0001" {
		t.Fatalf("申请列表内容异常: %+v", items)
	}
	if items[0].ID != created.ID {
		t.Errorf("申请 id = %d, want %d", items[0].ID, created.ID)
	}

	// happy：B 同意 → 双方好友列表互见。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests/"+itoa(created.ID)+"/accept", tokenB, nil)
	if resp.Code != 0 {
		t.Fatalf("同意: code=%d msg=%s", resp.Code, resp.Msg)
	}
	_, resp = doJSON(t, e.r, http.MethodGet, "/api/v1/friends", tokenA, nil)
	if resp.Code != 0 {
		t.Fatalf("A 好友列表: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var friends []struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(resp.Data, &friends); err != nil {
		t.Fatalf("unmarshal friends: %v", err)
	}
	if len(friends) != 1 || friends[0].User.ID != uidB {
		t.Fatalf("A 好友列表 = %+v, want [B]", friends)
	}

	// error：重复同意（2106）。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests/"+itoa(created.ID)+"/accept", tokenB, nil)
	if resp.Code != 2106 {
		t.Errorf("二次同意: code=%d, want 2106", resp.Code)
	}

	// error：已是好友再申请（2104）。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenA, friendRequestPayload{UserID: uidB})
	if resp.Code != 2104 {
		t.Errorf("已是好友: code=%d, want 2104", resp.Code)
	}

	// error：accept 非法 id（1001）。
	w, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests/abc/accept", tokenB, nil)
	if w.Code != http.StatusBadRequest || resp.Code != 1001 {
		t.Errorf("非法 id: http=%d code=%d, want 400/1001", w.Code, resp.Code)
	}
}

func TestFriendsRejectFlow(t *testing.T) {
	e := newHandlerEnv(t)
	tokenA := loginAs(t, e, "13800000001", "用户A")
	tokenB := loginAs(t, e, "13800000002", "用户B")
	uidA, uidB := meID(t, e, tokenA), meID(t, e, tokenB)

	// A 申请 → B 拒绝。
	_, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenA, friendRequestPayload{UserID: uidB})
	if resp.Code != 0 {
		t.Fatalf("发起申请: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	_ = json.Unmarshal(resp.Data, &created)

	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests/"+itoa(created.ID)+"/reject", tokenB, nil)
	if resp.Code != 0 {
		t.Fatalf("拒绝: code=%d msg=%s", resp.Code, resp.Msg)
	}

	// 拒绝后：申请列表为空、好友列表为空。
	_, resp = doJSON(t, e.r, http.MethodGet, "/api/v1/friends/requests", tokenB, nil)
	if resp.Code != 0 || string(resp.Data) == "null" {
		t.Fatalf("拒绝后申请列表: code=%d data=%s", resp.Code, resp.Data)
	}
	var items []json.RawMessage
	_ = json.Unmarshal(resp.Data, &items)
	if len(items) != 0 {
		t.Errorf("拒绝后申请列表 = %d 项, want 0", len(items))
	}

	// 二次拒绝：2106。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests/"+itoa(created.ID)+"/reject", tokenB, nil)
	if resp.Code != 2106 {
		t.Errorf("二次拒绝: code=%d, want 2106", resp.Code)
	}

	// 重新申请成功（rejected 翻回 pending）。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenA, friendRequestPayload{UserID: uidB})
	if resp.Code != 0 {
		t.Fatalf("重新申请: code=%d msg=%s", resp.Code, resp.Msg)
	}
	// 反向 pending：B 向 A 申请 → 2105。
	_, resp = doJSON(t, e.r, http.MethodPost, "/api/v1/friends/requests", tokenB, friendRequestPayload{UserID: uidA})
	if resp.Code != 2105 {
		t.Errorf("反向 pending: code=%d, want 2105", resp.Code)
	}
}

// itoa 路径参数数字转字符串（测试内可读性）。
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
