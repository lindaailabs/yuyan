package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// 埋点约束（guide §8 + §12 脱敏红线）。
const (
	maxEventNameLen  = 64
	maxPropsLen      = 2048
	maxReportBatch   = 50
	maxQueryLimit    = 200
)

// 敏感键：命中即丢弃该字段，绝不落库（手机号/token/密码等）。
var sensitiveKeys = []string{
	"phone", "mobile", "token", "access_token", "refresh_token",
	"password", "secret", "prompt", "id_card",
}

// AnalyticsService 事件埋点：落库、批量上报与内部查询。
type AnalyticsService struct {
	events *repo.EventRepo
	now    func() time.Time
}

func NewAnalyticsService(events *repo.EventRepo, now func() time.Time) *AnalyticsService {
	return &AnalyticsService{events: events, now: now}
}

func (s *AnalyticsService) timeNow() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// Track 记录一条服务端事件（失败仅告警，不阻断业务）。
func (s *AnalyticsService) Track(ctx context.Context, userID int64, name string, props map[string]any) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxEventNameLen {
		return
	}
	row := model.EventLog{
		UserID:    userID,
		Name:      name,
		Props:     marshalProps(sanitizeProps(props)),
		CreatedAt: s.timeNow().Unix(),
	}
	if err := s.events.InsertEvents(ctx, []model.EventLog{row}); err != nil {
		slog.Warn("track event failed", "err", err, "name", name, "user_id", userID)
	}
}

// Report 客户端批量上报：校验条数、事件名与属性长度，并做脱敏。
func (s *AnalyticsService) Report(ctx context.Context, userID int64, in *model.EventReportInput) (int, error) {
	if in == nil || len(in.Events) == 0 {
		return 0, errcode.New(errcode.ErrInvalidParam, "事件列表不能为空")
	}
	if len(in.Events) > maxReportBatch {
		return 0, errcode.New(errcode.ErrInvalidParam, fmt.Sprintf("单次上报最多 %d 条", maxReportBatch))
	}

	now := s.timeNow().Unix()
	rows := make([]model.EventLog, 0, len(in.Events))
	for _, e := range in.Events {
		name := strings.TrimSpace(e.Name)
		if name == "" || utf8.RuneCountInString(name) > maxEventNameLen {
			return 0, errcode.New(errcode.ErrInvalidParam, "事件名非法或过长")
		}
		props := sanitizeProps(e.Props)
		if p := marshalProps(props); p != nil && utf8.RuneCountInString(*p) > maxPropsLen {
			return 0, errcode.New(errcode.ErrInvalidParam, "事件属性过长")
		} else {
			rows = append(rows, model.EventLog{
				UserID:    userID,
				Name:      name,
				Props:     p,
				CreatedAt: now,
			})
			continue
		}
	}
	if err := s.events.InsertEvents(ctx, rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}

// List 内部查询出口（非生产环境使用）。
func (s *AnalyticsService) List(ctx context.Context, name string, limit int) ([]model.EventItem, error) {
	if limit <= 0 || limit > maxQueryLimit {
		limit = 50
	}
	rows, err := s.events.ListEvents(ctx, strings.TrimSpace(name), limit)
	if err != nil {
		return nil, err
	}
	items := make([]model.EventItem, 0, len(rows))
	for i := range rows {
		item := model.EventItem{
			ID:        rows[i].ID,
			UserID:    rows[i].UserID,
			Name:      rows[i].Name,
			CreatedAt: rows[i].CreatedAt,
		}
		if rows[i].Props != nil {
			var props map[string]any
			if err := json.Unmarshal([]byte(*rows[i].Props), &props); err == nil {
				item.Props = props
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// sanitizeProps 丢弃敏感键（防手机号/token/prompt 落库）。
func sanitizeProps(props map[string]any) map[string]any {
	if len(props) == 0 {
		return nil
	}
	cleaned := make(map[string]any, len(props))
	for k, v := range props {
		lower := strings.ToLower(k)
		sensitive := false
		for _, s := range sensitiveKeys {
			if strings.Contains(lower, s) {
				sensitive = true
				break
			}
		}
		if sensitive {
			continue
		}
		cleaned[k] = v
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func marshalProps(props map[string]any) *string {
	if len(props) == 0 {
		return nil
	}
	b, err := json.Marshal(props)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
