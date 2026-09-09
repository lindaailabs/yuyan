package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server"
	"github.com/lindaailabs/yuyan/server/internal/pkg/ai"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/repo"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// handlerEnv API 层测试环境：miniredis + MySQL 容器 + 真实 service 装配的路由。
type handlerEnv struct {
	r   *gin.Engine
	mr  *miniredis.Miniredis
	rdb *redis.Client
}

// newHandlerEnv 搭建真实依赖栈（repo → service → api），HTTP 层用 httptest 打。
func newHandlerEnv(t *testing.T) *handlerEnv {
	t.Helper()
	ctx := context.Background()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	container, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("yuyan"),
		tcmysql.WithUsername("yuyan"),
		tcmysql.WithPassword("yuyan123"),
	)
	if err != nil {
		t.Fatalf("start mysql: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn := container.MustConnectionString(ctx) + "?charset=utf8mb4&parseTime=True&loc=Local"
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("gorm raw handle: %v", err)
	}
	if err := server.MigrateUp(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	jwtMgr := jwt.NewManager("handler-test-secret")
	authSvc := service.NewAuthService(repo.NewUserRepo(gdb), jwtMgr)
	userSvc := service.NewUserService(repo.NewUserRepo(gdb))
	contactsSvc := service.NewContactsService(repo.NewFriendshipRepo(gdb), repo.NewUserRepo(gdb))
	analyticsSvc := service.NewAnalyticsService(repo.NewEventRepo(gdb), nil)
	entSvc := service.NewEntitlementService(repo.NewEntitlementRepo(gdb), repo.NewUsageRepo(gdb), repo.NewPaymentOrderRepo(gdb), nil, true)
	petSvc := service.NewPetService(repo.NewPetRepo(gdb), analyticsSvc)

	// AI Gateway 用 mock provider：API 测试同样不依赖真实模型（guide §10）。
	gateway, err := ai.New(ai.Config{Provider: ai.ProviderMock, TimeoutMS: 2000})
	if err != nil {
		t.Fatalf("ai gateway: %v", err)
	}
	memSvc := service.NewMemoryService(repo.NewMemoryRepo(gdb), repo.NewPetRepo(gdb), analyticsSvc)
	growthSvc := service.NewGrowthService(repo.NewGrowthRepo(gdb), repo.NewPetRepo(gdb), nil)
	convSvc := service.NewConversationService(
		repo.NewConversationRepo(gdb),
		repo.NewMessageRepo(gdb),
		repo.NewAICallLogRepo(gdb),
		repo.NewPetRepo(gdb),
		memSvc,
		growthSvc,
		gateway,
		entSvc,
		analyticsSvc,
	)

	return &handlerEnv{
		r: NewRouter(RouterDeps{
			Auth:         authSvc,
			User:         userSvc,
			Contacts:     contactsSvc,
			Pet:          petSvc,
			Conversation: convSvc,
			Memory:       memSvc,
			Growth:       growthSvc,
			Entitlement:  entSvc,
			Analytics:    analyticsSvc,
			AppEnv:       "dev",
			JWT:          jwtMgr,
		}),
		mr:  mr,
		rdb: rdb,
	}
}

// envelope 统一响应包裹（测试解析用）。
type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// doJSON 发送 JSON/Query 请求并解析统一包裹；token 非空时携带 Bearer。
func doJSON(t *testing.T, r *gin.Engine, method, path, token string, body any) (*httptest.ResponseRecorder, envelope) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var e envelope
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatalf("unmarshal response %s: %v (raw=%s)", path, err, w.Body.String())
	}
	return w, e
}

// loginByPhone 走完整 HTTP 流程：注册（手机号+密码）→ 登录，返回双 token。
func loginByPhone(t *testing.T, e *handlerEnv, phone, password string) (access, refresh string) {
	t.Helper()

	_, resp := doJSON(t, e.r, http.MethodPost, "/api/v1/auth/register", "", map[string]string{"phone": phone, "password": password})
	if resp.Code != 0 {
		t.Fatalf("register: code=%d msg=%s", resp.Code, resp.Msg)
	}

	_, lresp := doJSON(t, e.r, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"phone": phone, "password": password})
	if lresp.Code != 0 {
		t.Fatalf("login: code=%d msg=%s", lresp.Code, lresp.Msg)
	}
	var data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(lresp.Data, &data); err != nil {
		t.Fatalf("unmarshal login data: %v", err)
	}
	return data.AccessToken, data.RefreshToken
}
