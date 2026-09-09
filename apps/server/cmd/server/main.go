// 入口：只做配置加载与依赖装配（LLM_DEV_GUIDE.md §3.3）。
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/lindaailabs/yuyan/server"
	"github.com/lindaailabs/yuyan/server/internal/api"
	"github.com/lindaailabs/yuyan/server/internal/pkg/ai"
	"github.com/lindaailabs/yuyan/server/internal/pkg/config"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/pkg/logger"
	"github.com/lindaailabs/yuyan/server/internal/repo"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

func main() {
	// 命令行参数指定环境/配置，免设环境变量（等价于 APP_ENV / CONFIG_FILE）。
	envFlag := flag.String("env", "", "运行环境，加载 config.<env>.yaml（如 test）；不填则取 APP_ENV 或默认 dev")
	configFlag := flag.String("config", "", "显式指定配置文件路径（如 config.test.yaml），优先级最高")
	flag.Parse()

	cfg := config.Load(config.WithEnv(*envFlag), config.WithConfigFile(*configFlag))
	logger.Init(cfg.LogLevel)
	slog.Info("starting yuyan server", "config", cfg.SafeString())

	// MySQL：连接失败直接退出（无库无法服务）。
	db, err := repo.NewDB(cfg.MySQLDSN)
	if err != nil {
		slog.Error("mysql connect failed", "err", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("mysql raw handle unavailable", "err", err)
		os.Exit(1)
	}
	if err := server.MigrateUp(sqlDB); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}
	slog.Info("migration up to date")

	// Redis：连接失败仅告警不退出（W2 起按业务决定失败策略）。
	rdb := repo.NewRedis(cfg.RedisAddr)
	if err := repo.PingRedis(context.Background(), rdb); err != nil {
		slog.Error("redis ping failed", "err", err)
	} else {
		slog.Info("redis ping ok")
	}

	// 业务装配：repo → service → api（guide §3.3 分层）。
	jwtMgr := jwt.NewManager(cfg.JWTSecret)
	authSvc := service.NewAuthService(repo.NewUserRepo(db), jwtMgr)
	userSvc := service.NewUserService(repo.NewUserRepo(db))
	contactsSvc := service.NewContactsService(repo.NewFriendshipRepo(db), repo.NewUserRepo(db))

	// 权益与数据看板（W5）：服务端是权益唯一事实源，沙盒仅非生产环境可用。
	sandboxEnabled := cfg.AppEnv != "prod"
	if !sandboxEnabled {
		slog.Warn("sandbox purchase disabled in prod env")
	}
	entSvc := service.NewEntitlementService(
		repo.NewEntitlementRepo(db),
		repo.NewUsageRepo(db),
		repo.NewPaymentOrderRepo(db),
		nil,
		sandboxEnabled,
	)
	analyticsSvc := service.NewAnalyticsService(repo.NewEventRepo(db), nil)

	// AI Gateway：模型调用唯一入口（guide §5）；provider 未配置时启动即失败，不静默降级。
	gateway, err := ai.New(ai.Config{
		Provider:     cfg.AIProvider,
		TimeoutMS:    cfg.AITimeoutMS,
		MockFailRate: cfg.AIMockFailRate,
		BaseURL:      cfg.AIBaseURL,
		APIKey:       cfg.AIAPIKey,
		Model:        cfg.AIModel,
	})
	if err != nil {
		slog.Error("ai gateway init failed", "err", err)
		os.Exit(1)
	}
	petRepo := repo.NewPetRepo(db)
	petSvc := service.NewPetService(petRepo, analyticsSvc)
	memSvc := service.NewMemoryService(repo.NewMemoryRepo(db), petRepo, analyticsSvc)
	growthSvc := service.NewGrowthService(repo.NewGrowthRepo(db), petRepo, nil)
	convSvc := service.NewConversationService(
		repo.NewConversationRepo(db),
		repo.NewMessageRepo(db),
		repo.NewAICallLogRepo(db),
		petRepo,
		memSvc,
		growthSvc,
		gateway,
		entSvc,
		analyticsSvc,
	)

	r := api.NewRouter(api.RouterDeps{
		Auth:         authSvc,
		User:         userSvc,
		Contacts:     contactsSvc,
		Pet:          petSvc,
		Conversation: convSvc,
		Memory:       memSvc,
		Growth:       growthSvc,
		Entitlement:  entSvc,
		Analytics:    analyticsSvc,
		AppEnv:       cfg.AppEnv,
		JWT:          jwtMgr,
	})
	slog.Info("http listening", "port", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		slog.Error("http server exited", "err", err)
		os.Exit(1)
	}
}
