package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// gormOpen 打开 GORM 连接。
func gormOpen(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := server.MigrateUp(sqlDB); err != nil {
		return nil, err
	}
	return db, nil
}

// newAuthService 装配 AuthService（生产 main.go 同款路径）。
func newAuthService(rdb *redis.Client, gdb *gorm.DB, secret string) *AuthService {
	return NewAuthService(
		repo.NewCaptchaRepo(rdb),
		repo.NewUserRepo(gdb),
		jwt.NewManager(secret),
	)
}

// serviceEnv 测试环境：miniredis + MySQL 容器（含 migration）。
type serviceEnv struct {
	rdb  *redis.Client
	mr   *miniredis.Miniredis
	svc  *AuthService
	user *UserService
	ur   *repo.UserRepo
}

// newServiceEnv 搭建完整依赖（每个测试独立环境）。
func newServiceEnv(t *testing.T) *serviceEnv {
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
	gdb, err := gormOpen(dsn)
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}

	svc := newAuthService(rdb, gdb, "unit-test-secret")
	userRepo := repo.NewUserRepo(gdb)
	return &serviceEnv{
		rdb:  rdb,
		mr:   mr,
		svc:  svc,
		user: NewUserService(userRepo),
		ur:   userRepo,
	}
}

// createUser 直插测试用户（绕过验证码路径），返回 uid。
func (e *serviceEnv) createUser(t *testing.T, phone string) int64 {
	t.Helper()
	u, err := e.ur.CreateUser(context.Background(), phone)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u.ID
}

// storedCode 从 Redis 读取已下发的验证码（模拟用户看图/收短信输入）。
func (e *serviceEnv) storedCode(t *testing.T, phone string) string {
	t.Helper()
	code, err := e.rdb.Get(context.Background(), "sms:code:"+phone).Result()
	if err != nil {
		t.Fatalf("read stored code: %v", err)
	}
	return code
}

// sendCode 下发并取回验证码。
func (e *serviceEnv) sendCode(t *testing.T, phone string) string {
	t.Helper()
	if _, err := e.svc.SendSmsCode(context.Background(), phone); err != nil {
		t.Fatalf("SendSmsCode: %v", err)
	}
	return e.storedCode(t, phone)
}
