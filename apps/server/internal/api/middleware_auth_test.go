package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
)

// newAuthTestRouter 构造带认证中间件的探针路由，返回路由与 handler 是否被触达的采集变量。
func newAuthTestRouter(jwtMgr *jwt.Manager) (*gin.Engine, *int64) {
	var seenUID int64
	r := gin.New()
	r.GET("/protected", AuthMiddleware(jwtMgr), func(c *gin.Context) {
		seenUID = UIDFrom(c)
		OK(c, gin.H{"uid": seenUID})
	})
	return r, &seenUID
}

func TestAuthMiddleware(t *testing.T) {
	jwtMgr := jwt.NewManager("mw-test-secret")
	pair, err := jwtMgr.Issue(42)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	cases := []struct {
		name     string
		header   string
		wantHit  bool
		wantUID  int64
		wantCode int
	}{
		{"无 Authorization", "", false, 0, 1002},
		{"非 Bearer 前缀", pair.AccessToken, false, 0, 1002},
		{"Bearer 空串", "Bearer ", false, 0, 1002},
		{"Bearer 垃圾串", "Bearer garbage", false, 0, 1002},
		{"Bearer refresh token", "Bearer " + pair.RefreshToken, false, 0, 1002},
		{"合法 access token", "Bearer " + pair.AccessToken, true, 42, 0},
	}
	for _, c := range cases {
		r, seenUID := newAuthTestRouter(jwtMgr)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		if c.header != "" {
			req.Header.Set("Authorization", c.header)
		}
		r.ServeHTTP(w, req)

		var body Response
		_ = json.Unmarshal(w.Body.Bytes(), &body)

		if !c.wantHit {
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s: status = %d, want 401", c.name, w.Code)
			}
			if body.Code != c.wantCode {
				t.Errorf("%s: code = %d, want %d", c.name, body.Code, c.wantCode)
			}
			if *seenUID != 0 {
				t.Errorf("%s: handler 不应被执行", c.name)
			}
			continue
		}
		if w.Code != http.StatusOK || body.Code != 0 {
			t.Errorf("%s: status=%d code=%d, want 200/0", c.name, w.Code, body.Code)
		}
		if *seenUID != c.wantUID {
			t.Errorf("%s: handler uid = %d, want %d", c.name, *seenUID, c.wantUID)
		}
	}
}

// 篡改签名：token 被改后应拒绝。
func TestAuthMiddlewareTampered(t *testing.T) {
	jwtMgr := jwt.NewManager("mw-test-secret")
	pair, err := jwtMgr.Issue(7)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	tampered := pair.AccessToken + "x"

	r, _ := newAuthTestRouter(jwtMgr)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// 用另一密钥签发的 token：应拒绝。
func TestAuthMiddlewareWrongSecret(t *testing.T) {
	other := jwt.NewManager("other-secret")
	pair, err := other.Issue(7)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	r, _ := newAuthTestRouter(jwt.NewManager("mw-test-secret"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}
