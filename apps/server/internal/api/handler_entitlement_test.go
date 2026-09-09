package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestEntitlementEndpoints(t *testing.T) {
	env := newHandlerEnv(t)
	access, _ := loginByPhone(t, env, "13700001001", "secret123")

	t.Run("me 默认免费权益", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/entitlements/me", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var v struct {
			Plan string `json:"plan"`
			Quota struct {
				DailyMessages int `json:"daily_messages"`
				DailyRemain   int `json:"daily_remain"`
			} `json:"quota"`
		}
		if err := json.Unmarshal(resp.Data, &v); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if v.Plan != "free" {
			t.Errorf("plan want free got %s", v.Plan)
		}
		if v.Quota.DailyRemain != 50 {
			t.Errorf("remain want 50 got %d", v.Quota.DailyRemain)
		}
	})

	t.Run("sandbox-purchase pro", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/entitlements/sandbox-purchase", access, map[string]string{"plan": "pro"})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var v struct {
			Plan  string `json:"plan"`
			Quota struct {
				DailyMessages int `json:"daily_messages"`
			} `json:"quota"`
		}
		_ = json.Unmarshal(resp.Data, &v)
		if v.Plan != "pro" {
			t.Errorf("plan want pro got %s", v.Plan)
		}
		if v.Quota.DailyMessages != 500 {
			t.Errorf("daily messages want 500 got %d", v.Quota.DailyMessages)
		}
	})

	t.Run("sandbox-purchase 非法 plan 1001", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/entitlements/sandbox-purchase", access, map[string]string{"plan": "unknown"})
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("status=%d code=%d want 400/1001", w.Code, resp.Code)
		}
	})

	t.Run("payments/callback 幂等发放", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/entitlements/payments/callback", access,
			map[string]any{"order_no": "cb-1", "plan": "pro"})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var v struct{ Plan string `json:"plan"` }
		_ = json.Unmarshal(resp.Data, &v)
		if v.Plan != "pro" {
			t.Errorf("plan want pro got %s", v.Plan)
		}
		// 重复回调：仍成功且幂等（不重复发放）。
		w2, resp2 := doJSON(t, env.r, http.MethodPost, "/api/v1/entitlements/payments/callback", access,
			map[string]any{"order_no": "cb-1", "plan": "pro"})
		if w2.Code != http.StatusOK || resp2.Code != 0 {
			t.Fatalf("dup status=%d code=%d", w2.Code, resp2.Code)
		}
	})

	t.Run("payments/callback 缺 order_no 1001", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/entitlements/payments/callback", access,
			map[string]any{"plan": "pro"})
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("status=%d code=%d want 400/1001", w.Code, resp.Code)
		}
	})

	t.Run("me 无 token 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/entitlements/me", "", nil)
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d want 401/1002", w.Code, resp.Code)
		}
	})
}

func TestAnalyticsEndpoints(t *testing.T) {
	env := newHandlerEnv(t)
	access, _ := loginByPhone(t, env, "13700001002", "secret123")

	t.Run("events 上报成功", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/events", access,
			map[string]any{"events": []map[string]any{{"name": "sub_view", "props": map[string]any{"level": "2"}}}})
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
		var d struct{ Accepted int `json:"accepted"` }
		if err := json.Unmarshal(resp.Data, &d); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if d.Accepted != 1 {
			t.Errorf("accepted want 1 got %d", d.Accepted)
		}
	})

	t.Run("events 空列表 1001", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/events", access, map[string]any{"events": []map[string]any{}})
		if w.Code != http.StatusBadRequest || resp.Code != 1001 {
			t.Errorf("status=%d code=%d want 400/1001", w.Code, resp.Code)
		}
	})

	t.Run("admin/events 可查（非生产）", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodGet, "/api/v1/admin/events?name=sub_view", access, nil)
		if w.Code != http.StatusOK || resp.Code != 0 {
			t.Fatalf("status=%d code=%d msg=%s", w.Code, resp.Code, resp.Msg)
		}
	})

	t.Run("events 无 token 1002", func(t *testing.T) {
		w, resp := doJSON(t, env.r, http.MethodPost, "/api/v1/events", "",
			map[string]any{"events": []map[string]any{{"name": "x"}}})
		if w.Code != http.StatusUnauthorized || resp.Code != 1002 {
			t.Errorf("status=%d code=%d want 401/1002", w.Code, resp.Code)
		}
	})
}
