package service

import (
	"context"
	"fmt"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// 好友域业务错误码（2xxx 段，21xx 子段）。
var (
	ErrTargetNotFound   = errcode.New(2101, "用户不存在")
	ErrAddSelf          = errcode.New(2102, "不能添加自己为好友")
	ErrRequestDuplicate = errcode.New(2103, "已向对方发出申请，请等待处理")
	ErrAlreadyFriend    = errcode.New(2104, "你们已经是好友")
	ErrReversePending   = errcode.New(2105, "对方已向你发出申请，请先处理")
	ErrRequestInvalid   = errcode.New(2106, "申请不存在或已处理")
	ErrBlocked          = errcode.New(2107, "对方不可添加")
)

// ContactsService 好友关系业务逻辑：申请防重复矩阵与事务编排都在本层（guide §3.3）。
type ContactsService struct {
	friends *repo.FriendshipRepo
	users   *repo.UserRepo
}

// NewContactsService 构造。
func NewContactsService(friends *repo.FriendshipRepo, users *repo.UserRepo) *ContactsService {
	return &ContactsService{friends: friends, users: users}
}

// SendRequest 发起好友申请（防重复矩阵见 change design D2）。
func (s *ContactsService) SendRequest(ctx context.Context, me, targetID int64) (*model.FriendRequestResult, error) {
	if me == targetID {
		return nil, ErrAddSelf
	}
	// 目标存在性（非事务读取即可：插入以 uk 兜底，目标消失属不可达脏数据）。
	if _, err := s.users.FindByID(ctx, targetID); repo.IsNotFound(err) {
		return nil, ErrTargetNotFound
	} else if err != nil {
		return nil, fmt.Errorf("send request find target: %w", err)
	}

	var result *model.FriendRequestResult
	err := s.friends.Transaction(ctx, func(tx *repo.FriendshipRepo) error {
		forward, reverse, err := tx.FindPairForUpdate(ctx, me, targetID)
		if err != nil {
			return fmt.Errorf("send request lock pair: %w", err)
		}

		// 反向 pending：对方先一步向我发出申请。
		if reverse != nil && reverse.Status == model.FriendshipPending {
			return ErrReversePending
		}
		// 反向 accepted：好友已成立（防御分支，正常流正向亦应为 accepted）。
		if reverse != nil && reverse.Status == model.FriendshipAccepted {
			return ErrAlreadyFriend
		}

		switch {
		case forward == nil:
			row, err := tx.InsertPending(ctx, me, targetID)
			if err != nil {
				return err
			}
			result = toFriendRequestResult(row)
			return nil
		case forward.Status == model.FriendshipPending:
			return ErrRequestDuplicate
		case forward.Status == model.FriendshipAccepted:
			return ErrAlreadyFriend
		case forward.Status == model.FriendshipBlocked:
			return ErrBlocked
		default: // rejected：翻回 pending 允许重新申请（刷新 created_at）。
			if err := tx.ReactivatePending(ctx, forward.ID); err != nil {
				return fmt.Errorf("reactivate request: %w", err)
			}
			row, err := tx.FindByIDForUpdate(ctx, forward.ID)
			if err != nil {
				return err
			}
			result = toFriendRequestResult(row)
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListRequests 我收到的 pending 申请列表（含申请人脱敏资料）。
func (s *ContactsService) ListRequests(ctx context.Context, me int64) ([]model.FriendshipRequestItem, error) {
	rows, err := s.friends.ListPendingRequests(ctx, me)
	if err != nil {
		return nil, err
	}
	userIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}
	users, err := s.users.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.User, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}

	items := make([]model.FriendshipRequestItem, 0, len(rows))
	for _, row := range rows {
		u, ok := byID[row.UserID]
		if !ok {
			continue // 申请人资料缺失（脏数据）：跳过而非阻断列表
		}
		items = append(items, model.FriendshipRequestItem{
			ID:        row.ID,
			FromUser:  toUserSearchItem(&u),
			CreatedAt: row.CreatedAt.Unix(),
		})
	}
	return items, nil
}

// Accept 同意申请：单事务内原行 pending→accepted + upsert 反向 accepted 行（design D3）。
func (s *ContactsService) Accept(ctx context.Context, me, requestID int64) (*model.FriendOpResult, error) {
	var result *model.FriendOpResult
	err := s.friends.Transaction(ctx, func(tx *repo.FriendshipRepo) error {
		row, err := tx.FindByIDForUpdate(ctx, requestID)
		if repo.IsNotFound(err) {
			return ErrRequestInvalid
		}
		if err != nil {
			return fmt.Errorf("accept find request: %w", err)
		}
		// 只处理"发给我的"且仍为 pending 的申请。
		if row.FriendID != me || row.Status != model.FriendshipPending {
			return ErrRequestInvalid
		}

		if err := tx.UpdateStatus(ctx, row.ID, model.FriendshipPending, model.FriendshipAccepted); err != nil {
			return fmt.Errorf("accept update status: %w", err)
		}
		// 反向行（me → 申请人）：不存在则插入，存在（历史 rejected）则覆盖为 accepted。
		if err := tx.UpsertAccepted(ctx, me, row.UserID); err != nil {
			return err
		}
		result = &model.FriendOpResult{ID: row.ID, Status: model.FriendshipAccepted}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Reject 拒绝申请：原行 pending→rejected（不物理删除、不产生反向行）。
func (s *ContactsService) Reject(ctx context.Context, me, requestID int64) (*model.FriendOpResult, error) {
	var result *model.FriendOpResult
	err := s.friends.Transaction(ctx, func(tx *repo.FriendshipRepo) error {
		row, err := tx.FindByIDForUpdate(ctx, requestID)
		if repo.IsNotFound(err) {
			return ErrRequestInvalid
		}
		if err != nil {
			return fmt.Errorf("reject find request: %w", err)
		}
		if row.FriendID != me || row.Status != model.FriendshipPending {
			return ErrRequestInvalid
		}

		if err := tx.UpdateStatus(ctx, row.ID, model.FriendshipPending, model.FriendshipRejected); err != nil {
			return fmt.Errorf("reject update status: %w", err)
		}
		result = &model.FriendOpResult{ID: row.ID, Status: model.FriendshipRejected}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListFriends 我的好友列表（含好友脱敏资料，按结交时间升序）。
func (s *ContactsService) ListFriends(ctx context.Context, me int64) ([]model.FriendshipItem, error) {
	rows, err := s.friends.ListFriends(ctx, me)
	if err != nil {
		return nil, err
	}
	friendIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		friendIDs = append(friendIDs, row.FriendID)
	}
	users, err := s.users.FindByIDs(ctx, friendIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.User, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}

	items := make([]model.FriendshipItem, 0, len(rows))
	for _, row := range rows {
		u, ok := byID[row.FriendID]
		if !ok {
			continue // 好友资料缺失（脏数据）：跳过而非阻断列表
		}
		items = append(items, model.FriendshipItem{
			User:      toUserSearchItem(&u),
			CreatedAt: row.CreatedAt.Unix(),
		})
	}
	return items, nil
}

// toFriendRequestResult 实体 → 发起申请响应 DTO。
func toFriendRequestResult(f *model.Friendship) *model.FriendRequestResult {
	return &model.FriendRequestResult{
		ID:        f.ID,
		UserID:    f.UserID,
		FriendID:  f.FriendID,
		Status:    f.Status,
		CreatedAt: f.CreatedAt.Unix(),
	}
}

// toUserSearchItem 实体 → 脱敏资料 DTO（与搜索结果同形态）。
func toUserSearchItem(u *model.User) model.UserSearchItem {
	return model.UserSearchItem{
		ID:       u.ID,
		Nickname: u.Nickname,
		AvatarID: u.AvatarID,
		Phone:    maskPhone(u.Phone),
	}
}
