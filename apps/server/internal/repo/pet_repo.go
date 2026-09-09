package repo

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// PetRepo AI 宠物档案数据访问。
type PetRepo struct {
	db *gorm.DB
}

func NewPetRepo(db *gorm.DB) *PetRepo {
	return &PetRepo{db: db}
}

func (r *PetRepo) Create(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Create(pet).Error; err != nil {
		return fmt.Errorf("create pet: %w", err)
	}
	return nil
}

func (r *PetRepo) ListByUser(ctx context.Context, userID int64) ([]model.Pet, error) {
	var pets []model.Pet
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&pets).Error; err != nil {
		return nil, fmt.Errorf("list pets: %w", err)
	}
	return pets, nil
}

func (r *PetRepo) FindByUserAndID(ctx context.Context, userID, petID int64) (*model.Pet, error) {
	var pet model.Pet
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, petID).First(&pet).Error; err != nil {
		return nil, err
	}
	return &pet, nil
}

func (r *PetRepo) Update(ctx context.Context, userID, petID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&model.Pet{}).Where("user_id = ? AND id = ?", userID, petID).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update pet: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
