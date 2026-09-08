package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestAuthEndpoints 认证三端点：sms-code / login / refresh（每端点 happy + error）。
func TestAuthEndpoints(t *testing.T) {
	env := newHandlerEnv(t)

	t.Run("sms-code 正常下发", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/sms-code", "", map[string]string{"phone": "13700000001"})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var data struct {
			CaptchaImage string `json:"captcha_image"`
		}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(data.CaptchaImage) < 100 {
			t.Errorf("captcha_image 应为有效 base64 图片, got %d 字节", len(data.CaptchaImage))
		}
	})

	t.Run("sms-code 限频 2001", func(t *testing.T) {
		doJSON(t, env.r, http.MethodPost, "/api/v1/auth/sms-code", "", map[string]string{"phone": "13700000002"})
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/sms-code", "", map[string]string{"phone": "13700000002"})
		if w.Code != http.StatusOK || resp.Code != 2001 {
			t.Errorf("status=%d code=%d, want 200/2001", w.Code, resp.Code)
		}
	})

	t.Run("sms-code 参数错误 1001", func(t *testing.T) {
		for _, body := range []map[string]string{{}, {"phone": "12345"}} {
			w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/sms-code", "", body)
			if w.Code != http.StatusBadRequest || resp.Code != 1001 {
				t.Errorf("body=%v: status=%d code=%d, want 400/1001", body, w.Code, resp.Code)
			}
		}
	})

	access, refresh := loginByPhone(t, env, "13700000003")

	t.Run("login 正常（自动注册+双 token）", func(t *testing.T) {
		if access == "" || refresh == "" {
			t.Fatalf("tokens 为空: access=%q refresh=%q", access, refresh)
		}
	})

	t.Run("login 验证码错误 2002", func(t *testing.T) {
		doJSON(t, env.r, http.MethodPost, "/api/v1/auth/sms-code", "", map[string]string{"phone": "13700000004"})
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"phone": "13700000004", "code": "000000"})
		if w.Code != http.StatusOK || resp.Code != 2002 {
			t.Errorf("status=%d code=%d, want 200/2002", w.Code, resp.Code)
		}
	})

	t.Run("login 缺字段 1001", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"phone": "13700000005"})
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("status=%d code=%d, want 400/1001", w.Code, resp.Code)
		}
	})

	t.Run("refresh 正常", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{"refresh_token": refresh})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
		}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if data.AccessToken == "" || data.RefreshToken == "" || data.ExpiresIn != 7200 {
			t.Errorf("刷新结果异常: %+v", data)
		}
	})

	t.Run("refresh 用 access 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{"refresh_token": access})
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d, want 401/1002", w.Code, resp.Code)
		}
	})

	t.Run("refresh 缺字段 1001", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{})
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("status=%d code=%d, want 400/1001", w.Code, resp.Code)
		}
	})
}

// TestUserEndpoints 用户三端点：me / me 更新 / search（每端点 happy + error）。
func TestUserEndpoints(t *testing.T) {
	env := newHandlerEnv(t)

	access, refresh := loginByPhone(t, env, "13700000011")

	t.Run("me 正常（首登形态 nickname=null）", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/me", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var me struct {
			ID       int64   `json:"id"`
			Phone    string  `json:"phone"`
			Nickname *string `json:"nickname"`
			AvatarID int16   `json:"avatar_id"`
		}
		if err := json.Unmarshal(resp.Data, &me); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if me.ID == 0 || me.Phone != "13700000011" || me.Nickname != nil || me.AvatarID != 1 {
			t.Errorf("me 字段异常: %+v", me)
		}
	})

	t.Run("me 无 token 1002/401", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/me", "", nil)
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d, want 401/1002", w.Code, resp.Code)
		}
	})

	t.Run("me 用 refresh token 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/me", refresh, nil)
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d, want 401/1002", w.Code, resp.Code)
		}
	})

	t.Run("更新资料正常", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPut, "/api/v1/users/me", access, map[string]any{"nickname": "语燕用户", "avatar_id": 3})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		// 再查 me：新值生效。
		_, meResp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/me", access, nil)
		var me struct {
			Nickname *string `json:"nickname"`
			AvatarID int16   `json:"avatar_id"`
		}
		if err := json.Unmarshal(meResp.Data, &me); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if me.Nickname == nil || *me.Nickname != "语燕用户" || me.AvatarID != 3 {
			t.Errorf("更新未生效: %+v", me)
		}
	})

	t.Run("更新头像越界 2xxx", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPut, "/api/v1/users/me", access, map[string]any{"avatar_id": 9})
		if w.Code != http.StatusOK || resp.Code < 2000 || resp.Code >= 3000 {
			t.Errorf("status=%d code=%d, want 200/2xxx", w.Code, resp.Code)
		}
	})

	t.Run("更新昵称超长 2xxx", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPut, "/api/v1/users/me", access, map[string]any{"nickname": "这是一个超过二十个字符的超长昵称测试用例哟"})
		if w.Code != http.StatusOK || resp.Code < 2000 || resp.Code >= 3000 {
			t.Errorf("status=%d code=%d, want 200/2xxx", w.Code, resp.Code)
		}
	})

	t.Run("更新无 token 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPut, "/api/v1/users/me", "", map[string]any{"nickname": "匿名"})
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d, want 401/1002", w.Code, resp.Code)
		}
	})

	t.Run("搜索命中（脱敏）", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/search?q=13700000011", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var items []struct {
			Phone    string  `json:"phone"`
			Nickname *string `json:"nickname"`
		}
		if err := json.Unmarshal(resp.Data, &items); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("应命中 1 条, got %d", len(items))
		}
		if items[0].Phone != "137****0011" || items[0].Nickname == nil || *items[0].Nickname != "语燕用户" {
			t.Errorf("搜索结果异常: %+v", items[0])
		}
	})

	t.Run("搜索未注册空列表", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/search?q=13799990000", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var items []json.RawMessage
		if err := json.Unmarshal(resp.Data, &items); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("未注册应空列表, got %d 条", len(items))
		}
	})

	t.Run("搜索缺参/非法 1001", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/search", access, nil) // 无 q
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("无 q: status=%d code=%d, want 400/1001", w.Code, resp.Code)
		}
		w, resp = doJSON(t, env.r, http.MethodGet, "/api/v1/users/search?q=12345", access, nil)
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("非法 q: status=%d code=%d, want 400/1001", w.Code, resp.Code)
		}
	})

	t.Run("搜索无 token 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/users/search?q=13700000011", "", nil)
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d, want 401/1002", w.Code, resp.Code)
		}
	})
}
