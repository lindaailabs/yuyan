package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthz(t *testing.T) {
	r := NewRouter(RouterDeps{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body Response
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != 0 || body.Msg != "ok" {
		t.Errorf("body = %+v, want code 0 msg ok", body)
	}
}

func TestTraceIDEchoed(t *testing.T) {
	r := NewRouter(RouterDeps{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(TraceIDHeader, "trace-abc-123")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(TraceIDHeader); got != "trace-abc-123" {
		t.Errorf("trace id = %q, want trace-abc-123", got)
	}
}

func TestTraceIDGenerated(t *testing.T) {
	r := NewRouter(RouterDeps{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	got := w.Header().Get(TraceIDHeader)
	if got == "" {
		t.Error("trace id should be generated when absent")
	}
	if len(got) < 16 {
		t.Errorf("generated trace id too short: %q", got)
	}
}

func TestTraceIDFromContext(t *testing.T) {
	r := NewRouter(RouterDeps{})
	var seen string
	r.GET("/probe", func(c *gin.Context) {
		seen = TraceIDFrom(c.Request.Context())
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set(TraceIDHeader, "ctx-id-1")
	r.ServeHTTP(httptest.NewRecorder(), req)
	if seen != "ctx-id-1" {
		t.Errorf("context trace id = %q, want ctx-id-1", seen)
	}
}
