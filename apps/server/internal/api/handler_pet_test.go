package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPetEndpoints(t *testing.T) {
	env := newHandlerEnv(t)
	access, _ := loginByPhone(t, env, "13700000101", "secret123")
	otherAccess, _ := loginByPhone(t, env, "13700000102", "secret123")

	t.Run("create/list/detail/state 正常", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/pets", access, map[string]any{"name": "小燕", "avatar_id": 12})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("create status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var created struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Species  string `json:"species"`
			AvatarID int16  `json:"avatar_id"`
			Level    int    `json:"level"`
			Mood     string `json:"mood"`
		}
		if err := json.Unmarshal(resp.Data, &created); err != nil {
			t.Fatalf("unmarshal create: %v", err)
		}
		if created.ID == 0 || created.Name != "小燕" || created.Species != "swallow" || created.AvatarID != 12 || created.Level != 1 || created.Mood != "curious" {
			t.Fatalf("create data 异常: %+v", created)
		}

		w, resp = doJSON(t, env.r, http.MethodGet, "/api/v1/pets", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("list status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var items []json.RawMessage
		if err := json.Unmarshal(resp.Data, &items); err != nil {
			t.Fatalf("unmarshal list: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("list len=%d, want 1", len(items))
		}

		w, resp = doJSON(t, env.r, http.MethodGet, "/api/v1/pets/"+itoa(created.ID)+"/state", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("state status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var state struct {
			Name     string `json:"name"`
			Intimacy int    `json:"intimacy"`
		}
		if err := json.Unmarshal(resp.Data, &state); err != nil {
			t.Fatalf("unmarshal state: %v", err)
		}
		if state.Name != "小燕" || state.Intimacy != 0 {
			t.Errorf("state 异常: %+v", state)
		}
	})

	t.Run("create 参数错误", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/pets", access, map[string]any{"name": ""})
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("status=%d code=%d, want 400/1001", w.Code, resp.Code)
		}

		w, resp = doJSON(t, env.r, http.MethodPost, "/api/v1/pets", access, map[string]any{"name": "小燕", "avatar_id": 99})
		if w.Code != http.StatusOK || resp.Code != 2203 {
			t.Errorf("status=%d code=%d, want 200/2203", w.Code, resp.Code)
		}
	})

	t.Run("无 token 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/pets", "", nil)
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d, want 401/1002", w.Code, resp.Code)
		}
	})

	t.Run("越权访问返回不存在", func(t *testing.T) {
		_, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/pets", access, map[string]any{"name": "阿语"})
		var created struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(resp.Data, &created); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/pets/"+itoa(created.ID), otherAccess, nil)
		if w.Code != http.StatusOK || resp.Code != 2205 {
			t.Errorf("status=%d code=%d, want 200/2205", w.Code, resp.Code)
		}
	})
}
