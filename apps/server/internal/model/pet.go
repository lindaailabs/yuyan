package model

import "time"

const (
	DefaultPetSpecies = "swallow"
	DefaultPetMood    = "curious"
	DefaultPetLevel   = 1
)

// Pet AI 宠物档案（pets 表）。
type Pet struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;not null;index"`
	Name      string    `gorm:"column:name;size:20;not null"`
	Species   string    `gorm:"column:species;size:32;not null;default:swallow"`
	AvatarID  int16     `gorm:"column:avatar_id;not null;default:1"`
	Persona   *string   `gorm:"column:persona"`
	Level     int       `gorm:"column:level;not null;default:1"`
	Intimacy  int       `gorm:"column:intimacy;not null;default:0"`
	Mood      string    `gorm:"column:mood;size:32;not null;default:curious"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Pet) TableName() string { return "pets" }

// CreatePetInput POST /pets 请求体。
type CreatePetInput struct {
	Name     string  `json:"name" binding:"required"`
	Species  *string `json:"species"`
	AvatarID *int16  `json:"avatar_id"`
	Persona  *string `json:"persona"`
}

// UpdatePetInput PUT /pets/:id 请求体，nil 表示字段不更新。
type UpdatePetInput struct {
	Name     *string `json:"name"`
	Species  *string `json:"species"`
	AvatarID *int16  `json:"avatar_id"`
	Persona  *string `json:"persona"`
}

// PetProfile 宠物档案响应。
type PetProfile struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	Species   string  `json:"species"`
	AvatarID  int16   `json:"avatar_id"`
	Persona   *string `json:"persona"`
	Level     int     `json:"level"`
	Intimacy  int     `json:"intimacy"`
	Mood      string  `json:"mood"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

// PetState 宠物主页状态响应。
type PetState struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Species  string `json:"species"`
	AvatarID int16  `json:"avatar_id"`
	Level    int    `json:"level"`
	Intimacy int    `json:"intimacy"`
	Mood     string `json:"mood"`
}
