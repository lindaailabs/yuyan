package service

import (
	"context"
	"errors"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

func newEntitlementEnv(t *testing.T) (*EntitlementService, *repo.UserRepo) {
	t.Helper()
	ctx := context.Background()
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
	svc := NewEntitlementService(
		repo.NewEntitlementRepo(gdb),
		repo.NewUsageRepo(gdb),
		repo.NewPaymentOrderRepo(gdb),
		nil,
		true,
	)
	return svc, repo.NewUserRepo(gdb)
}

func TestEntitlementGetCreatesFree(t *testing.T) {
	svc, ur := newEntitlementEnv(t)
	u, err := ur.CreateUser(context.Background(), "13700000001")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	view, err := svc.Get(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if view.Plan != model.PlanFree {
		t.Errorf("plan want free got %s", view.Plan)
	}
	if view.Quota.DailyMessages != FreeDailyMessages {
		t.Errorf("daily messages want %d got %d", FreeDailyMessages, view.Quota.DailyMessages)
	}
	if view.Quota.DailyRemain != FreeDailyMessages {
		t.Errorf("daily remain want %d got %d", FreeDailyMessages, view.Quota.DailyRemain)
	}
}

func TestEntitlementConsumeExhaustion(t *testing.T) {
	svc, ur := newEntitlementEnv(t)
	u, _ := ur.CreateUser(context.Background(), "13700000002")
	for i := 0; i < FreeDailyMessages; i++ {
		if err := svc.ConsumeAIMessage(context.Background(), u.ID); err != nil {
			t.Fatalf("consume #%d failed: %v", i, err)
		}
	}
	if err := svc.ConsumeAIMessage(context.Background(), u.ID); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("expect ErrQuotaExceeded, got %v", err)
	}
	view, err := svc.Get(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if view.Quota.DailyUsed != FreeDailyMessages {
		t.Errorf("used want %d got %d", FreeDailyMessages, view.Quota.DailyUsed)
	}
	if view.Quota.DailyRemain != 0 {
		t.Errorf("remain want 0 got %d", view.Quota.DailyRemain)
	}
}

func TestEntitlementSandboxPurchase(t *testing.T) {
	svc, ur := newEntitlementEnv(t)
	u, _ := ur.CreateUser(context.Background(), "13700000003")
	view, err := svc.SandboxPurchase(context.Background(), u.ID, model.PlanPro)
	if err != nil {
		t.Fatalf("sandbox: %v", err)
	}
	if view.Plan != model.PlanPro {
		t.Errorf("plan want pro got %s", view.Plan)
	}
	if view.Quota.DailyMessages != ProDailyMessages {
		t.Errorf("daily messages want %d got %d", ProDailyMessages, view.Quota.DailyMessages)
	}

	// 生产环境沙盒禁用（2504）：SandboxPurchase 在触碰 repo 前先拦截。
	prodSvc := NewEntitlementService(nil, nil, nil, nil, false)
	if _, err := prodSvc.SandboxPurchase(context.Background(), u.ID, model.PlanPro); !errors.Is(err, ErrSandboxDisabled) {
		t.Fatalf("expect ErrSandboxDisabled, got %v", err)
	}
}

func TestEntitlementPaymentCallbackIdempotent(t *testing.T) {
	svc, ur := newEntitlementEnv(t)
	u, _ := ur.CreateUser(context.Background(), "13700000004")
	success := true
	view, err := svc.HandlePaymentCallback(context.Background(), u.ID,
		&model.PaymentCallbackInput{OrderNo: "o1", Plan: model.PlanPro, Success: &success})
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	if view.Plan != model.PlanPro {
		t.Errorf("plan want pro got %s", view.Plan)
	}
	// 重复回调：仍成功且幂等（不重复发放）。
	view2, err := svc.HandlePaymentCallback(context.Background(), u.ID,
		&model.PaymentCallbackInput{OrderNo: "o1", Plan: model.PlanPro, Success: &success})
	if err != nil {
		t.Fatalf("callback dup: %v", err)
	}
	if view2.Plan != model.PlanPro {
		t.Errorf("dup plan want pro got %s", view2.Plan)
	}
	// 失败回调：返回当前权益（不发放）。
	fail := false
	view3, err := svc.HandlePaymentCallback(context.Background(), u.ID,
		&model.PaymentCallbackInput{OrderNo: "o2", Plan: model.PlanPro, Success: &fail})
	if err != nil {
		t.Fatalf("callback fail: %v", err)
	}
	if view3.Plan != model.PlanPro {
		t.Errorf("failed callback plan want pro got %s", view3.Plan)
	}
}
