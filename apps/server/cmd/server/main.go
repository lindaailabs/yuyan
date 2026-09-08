// 入口：只做配置加载与依赖装配（LLM_DEV_GUIDE.md §3.3）。
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lindaailabs/yuyan/server"
	"github.com/lindaailabs/yuyan/server/internal/api"
	"github.com/lindaailabs/yuyan/server/internal/pkg/config"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/pkg/logger"
	"github.com/lindaailabs/yuyan/server/internal/repo"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

func main() {
	cfg := config.Load()
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
	authSvc := service.NewAuthService(repo.NewCaptchaRepo(rdb), repo.NewUserRepo(db), jwtMgr)
	userSvc := service.NewUserService(repo.NewUserRepo(db))

	r := api.NewRouter(api.RouterDeps{Auth: authSvc, User: userSvc, JWT: jwtMgr})
	slog.Info("http listening", "port", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		slog.Error("http server exited", "err", err)
		os.Exit(1)
	}
}
