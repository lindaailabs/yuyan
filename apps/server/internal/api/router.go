// Package api 是 HTTP handler 层（Gin）：只做参数校验与调用 service，禁止直接访问 repo。
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// RouterDeps 路由装配依赖（main 组装后注入）。
type RouterDeps struct {
	Auth         *service.AuthService
	User         *service.UserService
	Contacts     *service.ContactsService
	Pet          *service.PetService
	Conversation *service.ConversationService
	Memory       *service.MemoryService
	Growth       *service.GrowthService
	Entitlement  *service.EntitlementService
	Analytics     *service.AnalyticsService
	AppEnv        string
	JWT          *jwt.Manager
}

// NewRouter 装配 HTTP 路由与中间件。
// /api/v1/auth/* 匿名可访问；业务端点须经 AuthMiddleware。
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(TraceMiddleware())

	// 本地 Web 调试（Chrome）跨域：仅非生产环境放开，生产环境不注册以免暴露 CORS。
	if deps.AppEnv != "prod" {
		r.Use(CORSMiddleware())
	}

	r.GET("/healthz", handleHealthz)

	v1 := r.Group("/api/v1")

	authHandler := NewAuthHandler(deps.Auth)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	userHandler := NewUserHandler(deps.User)
	userGroup := v1.Group("/users", AuthMiddleware(deps.JWT))
	{
		userGroup.GET("/me", userHandler.Me)
		userGroup.PUT("/me", userHandler.UpdateMe)
		userGroup.GET("/search", userHandler.Search)
	}

	petHandler := NewPetHandler(deps.Pet)
	memHandler := NewMemoryHandler(deps.Memory)
	growthHandler := NewGrowthHandler(deps.Growth)
	petGroup := v1.Group("/pets", AuthMiddleware(deps.JWT))
	{
		petGroup.POST("", petHandler.Create)
		petGroup.GET("", petHandler.List)
		petGroup.GET("/:id", petHandler.Detail)
		petGroup.PUT("/:id", petHandler.Update)
		petGroup.GET("/:id/state", petHandler.State)
		petGroup.GET("/:id/memories", memHandler.List)
		petGroup.GET("/:id/growth-events", growthHandler.List)
	}
	memGroup := v1.Group("/pet-memories", AuthMiddleware(deps.JWT))
	{
		memGroup.DELETE("/:id", memHandler.Delete)
	}

	convHandler := NewConversationHandler(deps.Conversation)
	convGroup := v1.Group("/pet-conversations", AuthMiddleware(deps.JWT))
	{
		convGroup.POST("", convHandler.CreateConversation)
	}
	msgGroup := v1.Group("/pet-messages", AuthMiddleware(deps.JWT))
	{
		msgGroup.GET("", convHandler.History)
		msgGroup.POST("", convHandler.SendMessage)
	}

	entHandler := NewEntitlementHandler(deps.Entitlement)
	analyticsHandler := NewAnalyticsHandler(deps.Analytics)

	entGroup := v1.Group("/entitlements", AuthMiddleware(deps.JWT))
	{
		entGroup.GET("/me", entHandler.Me)
		entGroup.POST("/sandbox-purchase", entHandler.SandboxPurchase)
		entGroup.POST("/payments/callback", entHandler.PaymentCallback)
	}

	eventsGroup := v1.Group("/events", AuthMiddleware(deps.JWT))
	{
		eventsGroup.POST("", analyticsHandler.Report)
	}

	// 内部数据出口仅非生产环境注册（生产环境不暴露 /admin，避免数据外泄）。
	if deps.AppEnv != "prod" {
		adminGroup := v1.Group("/admin")
		{
			adminGroup.GET("/events", analyticsHandler.List)
		}
	}

	contactsHandler := NewContactsHandler(deps.Contacts)
	friendsGroup := v1.Group("/friends", AuthMiddleware(deps.JWT))
	{
		friendsGroup.POST("/requests", contactsHandler.SendRequest)
		friendsGroup.GET("/requests", contactsHandler.ListRequests)
		friendsGroup.POST("/requests/:id/accept", contactsHandler.Accept)
		friendsGroup.POST("/requests/:id/reject", contactsHandler.Reject)
		friendsGroup.GET("", contactsHandler.ListFriends)
	}
	return r
}

// CORSMiddleware 本地调试跨域放行：回显请求 Origin，允许带凭证的预检。
// 仅非生产环境注册（见 NewRouter），生产环境不启用。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header(
			"Access-Control-Allow-Headers",
			"Origin, Content-Type, Accept, Authorization, Access-Control-Request-Private-Network",
		)
		// Chrome 私有网络访问(PNA)：从内网页面(如 192.168.x.x)请求内网服务时，
		// Chrome 会带 Access-Control-Request-Private-Network 预检，服务端须显式允许，
		// 否则即使 Origin 匹配也会被拦（手机/同网段联调才会遇到，localhost 不触发）。
		if c.GetHeader("Access-Control-Request-Private-Network") == "true" {
			c.Header("Access-Control-Allow-Private-Network", "true")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
