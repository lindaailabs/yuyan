// Package api 是 HTTP handler 层（Gin）：只做参数校验与调用 service，禁止直接访问 repo。
package api

import (
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
	JWT          *jwt.Manager
}

// NewRouter 装配 HTTP 路由与中间件。
// /api/v1/auth/* 匿名可访问；业务端点须经 AuthMiddleware。
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(TraceMiddleware())

	r.GET("/healthz", handleHealthz)

	v1 := r.Group("/api/v1")

	authHandler := NewAuthHandler(deps.Auth)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/sms-code", authHandler.SendSmsCode)
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
	petGroup := v1.Group("/pets", AuthMiddleware(deps.JWT))
	{
		petGroup.POST("", petHandler.Create)
		petGroup.GET("", petHandler.List)
		petGroup.GET("/:id", petHandler.Detail)
		petGroup.PUT("/:id", petHandler.Update)
		petGroup.GET("/:id/state", petHandler.State)
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
