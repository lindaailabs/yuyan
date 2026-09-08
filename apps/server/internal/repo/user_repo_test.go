package repo

import (
	"context"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server"
	"github.com/lindaailabs/yuyan/server/internal/model"
)

// newTestGORM 启动 MySQL 容器 + migration，返回 GORM 连接（与 main.go 相同的建库路径）。
func newTestGORM(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	container := startMySQL(t)
	dsn := container.MustConnectionString(ctx) + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("raw handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := server.MigrateUp(sqlDB); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}
	return db
}

func newTestUserRepo(t *testing.T) *UserRepo {
	return NewUserRepo(newTestGORM(t))
}

func TestUserFindByPhone(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()

	// 不存在。
	if _, err := repo.FindByPhone(ctx, "13900139000"); !IsNotFound(err) {
		t.Fatalf("不存在时应返回 ErrNotFound, got %v", err)
	}

	// 创建后可查。
	created, err := repo.CreateUser(ctx, "13900139000")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("创建后 id 应非零")
	}
	if created.Nickname != nil {
		t.Errorf("新用户 nickname 应为 NULL, got %q", *created.Nickname)
	}
	if created.AvatarID != 1 {
		t.Errorf("新用户 avatar_id = %d, want 1", created.AvatarID)
	}

	found, err := repo.FindByPhone(ctx, "13900139000")
	if err != nil {
		t.Fatalf("FindByPhone: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("FindByPhone id = %d, want %d", found.ID, created.ID)
	}
}

func TestUserCreateDuplicatePhone(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()

	if _, err := repo.CreateUser(ctx, "13911112222"); err != nil {
		t.Fatalf("首次创建: %v", err)
	}
	if _, err := repo.CreateUser(ctx, "13911112222"); err == nil {
		t.Fatal("重复 phone 创建应失败（唯一索引）")
	}
}

func TestUserFindByID(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()

	created, _ := repo.CreateUser(ctx, "13922223333")
	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Phone != "13922223333" {
		t.Errorf("FindByID phone = %s", found.Phone)
	}

	if _, err := repo.FindByID(ctx, 999999); !IsNotFound(err) {
		t.Errorf("不存在 id 应返回 ErrNotFound, got %v", err)
	}
}

func TestUserUpdateProfile(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()

	created, _ := repo.CreateUser(ctx, "13933334444")
	name := "语燕用户"
	avatar := int16(5)

	// 全量更新。
	if err := repo.UpdateProfile(ctx, created.ID, &model.UpdateProfileInput{Nickname: &name, AvatarID: &avatar}); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	got, _ := repo.FindByID(ctx, created.ID)
	if *got.Nickname != name || got.AvatarID != 5 {
		t.Errorf("更新后 = %+v", got)
	}

	// 部分更新：仅改头像，昵称不变。
	avatar8 := int16(8)
	if err := repo.UpdateProfile(ctx, created.ID, &model.UpdateProfileInput{AvatarID: &avatar8}); err != nil {
		t.Fatalf("部分更新: %v", err)
	}
	got, _ = repo.FindByID(ctx, created.ID)
	if got.AvatarID != 8 {
		t.Errorf("avatar = %d, want 8", got.AvatarID)
	}
	if *got.Nickname != name {
		t.Errorf("昵称应保持 %q, got %q", name, *got.Nickname)
	}

	// 空更新：无操作不报错。
	if err := repo.UpdateProfile(ctx, created.ID, &model.UpdateProfileInput{}); err != nil {
		t.Errorf("空更新应无操作: %v", err)
	}

	// 不存在的用户。
	if err := repo.UpdateProfile(ctx, 999999, &model.UpdateProfileInput{Nickname: &name}); !IsNotFound(err) {
		t.Errorf("不存在用户应 ErrNotFound, got %v", err)
	}
}
