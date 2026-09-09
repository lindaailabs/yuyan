package service

import (
	"context"
	"strings"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/ai"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// conversationEnv 对话域测试环境：MySQL 容器 + mock AI Gateway（无网络依赖）。
type conversationEnv struct {
	db   *gorm.DB
	svc  *ConversationService
	pets *PetService
	ur   *repo.UserRepo
}

func newConversationEnv(t *testing.T, mockFailRate float64) *conversationEnv {
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

	gateway, err := ai.New(ai.Config{Provider: ai.ProviderMock, TimeoutMS: 2000, MockFailRate: mockFailRate})
	if err != nil {
		t.Fatalf("ai gateway: %v", err)
	}

	petRepo := repo.NewPetRepo(gdb)
	svc := NewConversationService(
		repo.NewConversationRepo(gdb),
		repo.NewMessageRepo(gdb),
		repo.NewAICallLogRepo(gdb),
		petRepo,
		NewMemoryService(repo.NewMemoryRepo(gdb), petRepo),
		gateway,
	)
	return &conversationEnv{
		db:   gdb,
		svc:  svc,
		pets: NewPetService(petRepo),
		ur:   repo.NewUserRepo(gdb),
	}
}

func (e *conversationEnv) user(t *testing.T, phone string) int64 {
	t.Helper()
	u, err := e.ur.CreateUser(context.Background(), phone)
	if err != nil {
		t.Fatalf("create user %s: %v", phone, err)
	}
	return u.ID
}

func (e *conversationEnv) pet(t *testing.T, uid int64, name string) int64 {
	t.Helper()
	pet, err := e.pets.Create(context.Background(), uid, &model.CreatePetInput{Name: name})
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return pet.ID
}

// countLogs 统计 ai_call_logs 行数（用量落库断言用）。
func (e *conversationEnv) countLogs(t *testing.T) int64 {
	t.Helper()
	var n int64
	if err := e.db.Model(&model.AICallLog{}).Count(&n).Error; err != nil {
		t.Fatalf("count ai_call_logs: %v", err)
	}
	return n
}

func TestConversationGetOrCreateIsIdempotent(t *testing.T) {
	env := newConversationEnv(t, 0)
	ctx := context.Background()
	uid := env.user(t, "13802000001")
	petID := env.pet(t, uid, "小燕")

	first, err := env.svc.GetOrCreateConversation(ctx, uid, &model.CreateConversationInput{PetID: petID})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := env.svc.GetOrCreateConversation(ctx, uid, &model.CreateConversationInput{PetID: petID})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("重复创建会话: %d vs %d", first.ID, second.ID)
	}

	if _, err := env.svc.GetOrCreateConversation(ctx, uid, &model.CreateConversationInput{PetID: 999999}); err == nil {
		t.Error("不存在的宠物应返回错误")
	} else {
		wantCode(t, err, ErrPetNotAccessible.Code)
	}
}

func TestConversationSendMessageAndHistory(t *testing.T) {
	env := newConversationEnv(t, 0)
	ctx := context.Background()
	uid := env.user(t, "13802000002")
	petID := env.pet(t, uid, "小燕")

	conv, err := env.svc.GetOrCreateConversation(ctx, uid, &model.CreateConversationInput{PetID: petID})
	if err != nil {
		t.Fatalf("conversation: %v", err)
	}

	for i := 0; i < 5; i++ {
		if _, err := env.svc.SendMessage(ctx, uid, &model.SendMessageInput{
			PetID:   petID,
			Content: "第几条消息",
		}); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	// 5 轮 = 10 条消息；limit=4 分页应不重不漏。
	page1, err := env.svc.History(ctx, uid, conv.ID, 0, 4)
	if err != nil {
		t.Fatalf("history page1: %v", err)
	}
	if len(page1.Items) != 4 || !page1.HasMore {
		t.Fatalf("page1 = %d items hasMore=%v, want 4/true", len(page1.Items), page1.HasMore)
	}
	page2, err := env.svc.History(ctx, uid, conv.ID, page1.NextCursor, 4)
	if err != nil {
		t.Fatalf("history page2: %v", err)
	}
	if len(page2.Items) != 4 || !page2.HasMore {
		t.Fatalf("page2 = %d items hasMore=%v, want 4/true", len(page2.Items), page2.HasMore)
	}
	page3, err := env.svc.History(ctx, uid, conv.ID, page2.NextCursor, 4)
	if err != nil {
		t.Fatalf("history page3: %v", err)
	}
	if len(page3.Items) != 2 || page3.HasMore {
		t.Fatalf("page3 = %d items hasMore=%v, want 2/false", len(page3.Items), page3.HasMore)
	}

	seen := map[int64]bool{}
	for _, p := range []*model.MessagePage{page1, page2, page3} {
		for _, it := range p.Items {
			if seen[it.ID] {
				t.Errorf("消息 id %d 在多页重复出现", it.ID)
			}
			seen[it.ID] = true
		}
	}
	if len(seen) != 10 {
		t.Errorf("分页合计 %d 条, want 10", len(seen))
	}
	// 每条用户消息后都应有宠物回复。
	if env.countLogs(t) != 5 {
		t.Errorf("ai_call_logs = %d, want 5", env.countLogs(t))
	}
}

func TestConversationSendMessageIdempotent(t *testing.T) {
	env := newConversationEnv(t, 0)
	ctx := context.Background()
	uid := env.user(t, "13802000003")
	petID := env.pet(t, uid, "小燕")

	clientMsgID := "11111111-1111-1111-1111-111111111111"
	in := &model.SendMessageInput{PetID: petID, ClientMsgID: &clientMsgID, Content: "你好呀"}
	first, err := env.svc.SendMessage(ctx, uid, in)
	if err != nil {
		t.Fatalf("first send: %v", err)
	}
	second, err := env.svc.SendMessage(ctx, uid, in)
	if err != nil {
		t.Fatalf("replay send: %v", err)
	}
	if second.UserMessage.ID != first.UserMessage.ID {
		t.Errorf("重放产生了新的用户消息: %d vs %d", second.UserMessage.ID, first.UserMessage.ID)
	}
	if second.AssistantMessage == nil || second.AssistantMessage.ID != first.AssistantMessage.ID {
		t.Fatal("重放应返回首次的宠物回复")
	}

	page, err := env.svc.History(ctx, uid, first.ConversationID, 0, 50)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("消息总数 = %d, want 2（重复提交不应新增）", len(page.Items))
	}
}

func TestConversationSendMessageValidation(t *testing.T) {
	env := newConversationEnv(t, 0)
	ctx := context.Background()
	uid := env.user(t, "13802000004")
	petID := env.pet(t, uid, "小燕")

	if _, err := env.svc.SendMessage(ctx, uid, &model.SendMessageInput{PetID: petID, Content: "   "}); err == nil {
		t.Error("空内容应报错")
	} else {
		wantCode(t, err, ErrMessageInvalid.Code)
	}
	if _, err := env.svc.SendMessage(ctx, uid, &model.SendMessageInput{PetID: petID, Content: strings.Repeat("长", 2001)}); err == nil {
		t.Error("超长内容应报错")
	} else {
		wantCode(t, err, ErrMessageInvalid.Code)
	}

	other := env.user(t, "13802000005")
	if _, err := env.svc.SendMessage(ctx, other, &model.SendMessageInput{PetID: petID, Content: "你好"}); err == nil {
		t.Error("他人的宠物应报错")
	} else {
		wantCode(t, err, ErrPetNotAccessible.Code)
	}
}

func TestConversationAIFailureFallback(t *testing.T) {
	env := newConversationEnv(t, 1) // mock 必失败
	ctx := context.Background()
	uid := env.user(t, "13802000006")
	petID := env.pet(t, uid, "小燕")

	res, err := env.svc.SendMessage(ctx, uid, &model.SendMessageInput{PetID: petID, Content: "你好"})
	if err != nil {
		t.Fatalf("AI 失败不应让整个请求失败（应有兜底）: %v", err)
	}
	if res.AssistantMessage.Status != model.MessageStatusFailed {
		t.Errorf("宠物消息状态 = %d, want failed(%d)", res.AssistantMessage.Status, model.MessageStatusFailed)
	}
	if res.AssistantMessage.Content == "" {
		t.Error("兜底文案不应为空")
	}
	if res.UserMessage.Status != model.MessageStatusOK {
		t.Error("用户消息应保持成功")
	}
	if res.Usage.ErrCode == 0 {
		t.Error("失败时 usage.err_code 不应为 0")
	}

	var logs []model.AICallLog
	if err := env.db.Where("user_id = ?", uid).Find(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("ai_call_logs = %d, want 1", len(logs))
	}
	if logs[0].ErrCode == 0 {
		t.Error("失败调用日志 err_code 不应为 0")
	}
	if logs[0].Model == "" {
		t.Error("调用日志应记录模型名")
	}
}

func TestConversationHistoryAccessControl(t *testing.T) {
	env := newConversationEnv(t, 0)
	ctx := context.Background()
	uid := env.user(t, "13802000007")
	petID := env.pet(t, uid, "小燕")
	conv, err := env.svc.GetOrCreateConversation(ctx, uid, &model.CreateConversationInput{PetID: petID})
	if err != nil {
		t.Fatalf("conversation: %v", err)
	}

	other := env.user(t, "13802000008")
	if _, err := env.svc.History(ctx, other, conv.ID, 0, 20); err == nil {
		t.Error("他人读取会话应报错")
	} else {
		wantCode(t, err, ErrConversationNotFound.Code)
	}
	if _, err := env.svc.History(ctx, uid, 0, 0, 20); err == nil {
		t.Error("conv_id=0 应报参数错误")
	}
}
