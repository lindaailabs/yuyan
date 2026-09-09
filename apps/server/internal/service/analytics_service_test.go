package service

import (
	"context"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

func newAnalyticsEnv(t *testing.T) *AnalyticsService {
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
	return NewAnalyticsService(repo.NewEventRepo(gdb), nil)
}

func TestAnalyticsReportSanitizesSensitive(t *testing.T) {
	svc := newAnalyticsEnv(t)
	ctx := context.Background()
	_, err := svc.Report(ctx, 1, &model.EventReportInput{Events: []model.EventInput{
		{Name: "sub_view", Props: map[string]any{"phone": "13700000001", "level": "2"}},
	}})
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	items, err := svc.List(ctx, "sub_view", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 event got %d", len(items))
	}
	if _, ok := items[0].Props["phone"]; ok {
		t.Errorf("sensitive key 'phone' must be dropped, got props=%v", items[0].Props)
	}
	if items[0].Props["level"] != "2" {
		t.Errorf("non-sensitive key should remain, got %v", items[0].Props)
	}
}

func TestAnalyticsReportBatchLimits(t *testing.T) {
	svc := newAnalyticsEnv(t)
	ctx := context.Background()

	// 空列表：参数错误。
	if _, err := svc.Report(ctx, 1, &model.EventReportInput{Events: []model.EventInput{}}); err == nil {
		t.Errorf("empty events should error")
	}
	// 超过 50 条：参数错误。
	many := make([]model.EventInput, 51)
	for i := range many {
		many[i] = model.EventInput{Name: "x"}
	}
	if _, err := svc.Report(ctx, 1, &model.EventReportInput{Events: many}); err == nil {
		t.Errorf("over 50 events should error")
	}
	// 正常单条：成功。
	n, err := svc.Report(ctx, 1, &model.EventReportInput{Events: []model.EventInput{{Name: "x"}}})
	if err != nil {
		t.Fatalf("single report: %v", err)
	}
	if n != 1 {
		t.Errorf("accepted want 1 got %d", n)
	}
}
