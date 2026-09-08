package service

import (
	"context"
	"errors"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// contactsEnv 好友域测试环境：仅 MySQL 容器（无 Redis 依赖）。
type contactsEnv struct {
	svc *ContactsService
	fr  *repo.FriendshipRepo
	ur  *repo.UserRepo
}

func newContactsEnv(t *testing.T) *contactsEnv {
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

	fr := repo.NewFriendshipRepo(gdb)
	ur := repo.NewUserRepo(gdb)
	return &contactsEnv{svc: NewContactsService(fr, ur), fr: fr, ur: ur}
}

// user 创建测试用户。
func (e *contactsEnv) user(t *testing.T, phone string) int64 {
	t.Helper()
	u, err := e.ur.CreateUser(context.Background(), phone)
	if err != nil {
		t.Fatalf("create user %s: %v", phone, err)
	}
	return u.ID
}

// wantCode 断言 err 业务码。
func wantCode(t *testing.T, err error, code errcode.Code) {
	t.Helper()
	var ec *errcode.Error
	if !errors.As(err, &ec) {
		t.Fatalf("err = %v, want errcode %d", err, code)
	}
	if ec.Code != code {
		t.Fatalf("err code = %d (%s), want %d", ec.Code, ec.Msg, code)
	}
}

func TestContactsSendRequestMatrix(t *testing.T) {
	env := newContactsEnv(t)
	ctx := context.Background()

	a := env.user(t, "13800000001")
	b := env.user(t, "13800000002")

	// 正常申请。
	res, err := env.svc.SendRequest(ctx, a, b)
	if err != nil {
		t.Fatalf("首次申请: %v", err)
	}
	if res.Status != model.FriendshipPending || res.UserID != a || res.FriendID != b {
		t.Errorf("首次申请结果异常: %+v", res)
	}

	// 不能加自己。
	_, err = env.svc.SendRequest(ctx, a, a)
	wantCode(t, err, ErrAddSelf.Code)

	// 重复申请。
	_, err = env.svc.SendRequest(ctx, a, b)
	wantCode(t, err, ErrRequestDuplicate.Code)

	// 目标不存在。
	_, err = env.svc.SendRequest(ctx, a, 99999)
	wantCode(t, err, ErrTargetNotFound.Code)

	// 反向 pending：B 向 A 申请。
	_, err = env.svc.SendRequest(ctx, b, a)
	wantCode(t, err, ErrReversePending.Code)
}

func TestContactsAcceptAtomicity(t *testing.T) {
	env := newContactsEnv(t)
	ctx := context.Background()

	a := env.user(t, "13800000001")
	b := env.user(t, "13800000002")

	res, err := env.svc.SendRequest(ctx, a, b)
	if err != nil {
		t.Fatalf("申请: %v", err)
	}

	// 同意：双向两行 accepted（原子性断言——直接查库验证两行状态）。
	op, err := env.svc.Accept(ctx, b, res.ID)
	if err != nil {
		t.Fatalf("同意: %v", err)
	}
	if op.Status != model.FriendshipAccepted {
		t.Errorf("同意结果状态 = %d, want accepted", op.Status)
	}
	err = env.fr.Transaction(ctx, func(tx *repo.FriendshipRepo) error {
		forward, reverse, err := tx.FindPairForUpdate(ctx, a, b)
		if err != nil {
			return err
		}
		if forward == nil || forward.Status != model.FriendshipAccepted {
			t.Errorf("正向行 = %+v, want accepted", forward)
		}
		if reverse == nil || reverse.Status != model.FriendshipAccepted {
			t.Errorf("反向行 = %+v, want accepted", reverse)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("落库断言: %v", err)
	}

	// 二次同意：2106。
	_, err = env.svc.Accept(ctx, b, res.ID)
	wantCode(t, err, ErrRequestInvalid.Code)

	// 他人（C）处理该申请：2106。
	c := env.user(t, "13800000003")
	_, err = env.svc.Accept(ctx, c, res.ID)
	wantCode(t, err, ErrRequestInvalid.Code)

	// 已是好友再申请：2104。
	_, err = env.svc.SendRequest(ctx, a, b)
	wantCode(t, err, ErrAlreadyFriend.Code)

	// 好友互见。
	fa, err := env.svc.ListFriends(ctx, a)
	if err != nil || len(fa) != 1 || fa[0].User.ID != b {
		t.Fatalf("A 好友列表 = %+v err=%v, want [B]", fa, err)
	}
	fb, err := env.svc.ListFriends(ctx, b)
	if err != nil || len(fb) != 1 || fb[0].User.ID != a {
		t.Fatalf("B 好友列表 = %+v err=%v, want [A]", fb, err)
	}
}

func TestContactsRejectAndReapply(t *testing.T) {
	env := newContactsEnv(t)
	ctx := context.Background()

	a := env.user(t, "13800000001")
	b := env.user(t, "13800000002")

	res, err := env.svc.SendRequest(ctx, a, b)
	if err != nil {
		t.Fatalf("申请: %v", err)
	}

	// 拒绝：行保留且 rejected，列表消失，好友列表为空。
	op, err := env.svc.Reject(ctx, b, res.ID)
	if err != nil {
		t.Fatalf("拒绝: %v", err)
	}
	if op.Status != model.FriendshipRejected {
		t.Errorf("拒绝结果状态 = %d, want rejected", op.Status)
	}
	if items, _ := env.svc.ListRequests(ctx, b); len(items) != 0 {
		t.Errorf("拒绝后申请列表 = %d 项, want 0", len(items))
	}
	if friends, _ := env.svc.ListFriends(ctx, a); len(friends) != 0 {
		t.Errorf("拒绝后好友列表 = %d 项, want 0", len(friends))
	}

	// 二次拒绝：2106。
	_, err = env.svc.Reject(ctx, b, res.ID)
	wantCode(t, err, ErrRequestInvalid.Code)

	// 重新申请：rejected 翻回 pending（同一行复用）。
	res2, err := env.svc.SendRequest(ctx, a, b)
	if err != nil {
		t.Fatalf("重新申请: %v", err)
	}
	if res2.ID != res.ID || res2.Status != model.FriendshipPending {
		t.Errorf("重新申请应复用原行: got id=%d status=%d, want id=%d pending", res2.ID, res2.Status, res.ID)
	}

	// 重新申请后可同意。
	if _, err := env.svc.Accept(ctx, b, res2.ID); err != nil {
		t.Fatalf("同意重新申请: %v", err)
	}
	if friends, _ := env.svc.ListFriends(ctx, a); len(friends) != 1 {
		t.Errorf("同意后 A 好友列表 = %d 项, want 1", len(friends))
	}
}

func TestContactsListRequestsAndFriends(t *testing.T) {
	env := newContactsEnv(t)
	ctx := context.Background()

	a := env.user(t, "13800000001")
	b := env.user(t, "13800000002")
	c := env.user(t, "13800000003")

	// C、A 先后向 B 申请 → B 列表按 id 倒序（新的在前）。
	resC, _ := env.svc.SendRequest(ctx, c, b)
	resA, _ := env.svc.SendRequest(ctx, a, b)
	items, err := env.svc.ListRequests(ctx, b)
	if err != nil {
		t.Fatalf("申请列表: %v", err)
	}
	if len(items) != 2 || items[0].ID != resA.ID || items[1].ID != resC.ID {
		t.Fatalf("申请列表顺序异常: %+v", items)
	}
	// 申请人资料 + 手机号脱敏。
	if items[0].FromUser.ID != a || items[0].FromUser.Phone != "138****0001" {
		t.Errorf("申请人资料异常: %+v", items[0].FromUser)
	}

	// B 同意两份申请 → 好友列表按结交时间升序（先同意的在前）。
	if _, err := env.svc.Accept(ctx, b, resC.ID); err != nil {
		t.Fatalf("同意 C: %v", err)
	}
	if _, err := env.svc.Accept(ctx, b, resA.ID); err != nil {
		t.Fatalf("同意 A: %v", err)
	}
	friends, err := env.svc.ListFriends(ctx, b)
	if err != nil {
		t.Fatalf("好友列表: %v", err)
	}
	if len(friends) != 2 || friends[0].User.ID != c || friends[1].User.ID != a {
		t.Fatalf("好友列表顺序异常: %+v", friends)
	}
	if friends[0].User.Phone != "138****0003" {
		t.Errorf("好友手机号未脱敏: %s", friends[0].User.Phone)
	}

	// 无申请用户：空列表。
	if items, _ := env.svc.ListRequests(ctx, a); len(items) != 0 {
		t.Errorf("A 申请列表应为空, got %d 项", len(items))
	}
}
