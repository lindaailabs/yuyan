package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

var (
	ErrPetNameInvalid    = errcode.New(2201, "宠物名字须为 1~20 个字符")
	ErrPetSpeciesInvalid = errcode.New(2202, "宠物物种须为 1~32 个字符")
	ErrPetAvatarInvalid  = errcode.New(2203, "宠物外观不存在")
	ErrPetPersonaInvalid = errcode.New(2204, "宠物设定过长")
	ErrPetNotFound       = errcode.New(2205, "宠物不存在")
)

const maxPetPersonaLen = 2000

// PetService 宠物档案业务逻辑。
type PetService struct {
	pets *repo.PetRepo
}

func NewPetService(pets *repo.PetRepo) *PetService {
	return &PetService{pets: pets}
}

func (s *PetService) Create(ctx context.Context, uid int64, in *model.CreatePetInput) (*model.PetProfile, error) {
	if in == nil {
		return nil, errcode.New(errcode.ErrInvalidParam, "请求体不能为空")
	}
	name, err := validatePetName(in.Name)
	if err != nil {
		return nil, err
	}
	species := model.DefaultPetSpecies
	if in.Species != nil {
		species, err = validatePetSpecies(*in.Species)
		if err != nil {
			return nil, err
		}
	}
	avatarID := int16(1)
	if in.AvatarID != nil {
		avatarID = *in.AvatarID
	}
	if err := validatePetAvatar(avatarID); err != nil {
		return nil, err
	}
	persona, err := validatePetPersona(in.Persona)
	if err != nil {
		return nil, err
	}

	pet := &model.Pet{
		UserID:   uid,
		Name:     name,
		Species:  species,
		AvatarID: avatarID,
		Persona:  persona,
		Level:    model.DefaultPetLevel,
		Mood:     model.DefaultPetMood,
	}
	if err := s.pets.Create(ctx, pet); err != nil {
		return nil, fmt.Errorf("create pet: %w", err)
	}
	return toPetProfile(pet), nil
}

func (s *PetService) List(ctx context.Context, uid int64) ([]model.PetProfile, error) {
	pets, err := s.pets.ListByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list pets: %w", err)
	}
	items := make([]model.PetProfile, 0, len(pets))
	for i := range pets {
		items = append(items, *toPetProfile(&pets[i]))
	}
	return items, nil
}

func (s *PetService) Detail(ctx context.Context, uid, petID int64) (*model.PetProfile, error) {
	pet, err := s.pets.FindByUserAndID(ctx, uid, petID)
	if repo.IsNotFound(err) {
		return nil, ErrPetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find pet: %w", err)
	}
	return toPetProfile(pet), nil
}

func (s *PetService) Update(ctx context.Context, uid, petID int64, in *model.UpdatePetInput) (*model.PetProfile, error) {
	if in == nil {
		return nil, errcode.New(errcode.ErrInvalidParam, "请求体不能为空")
	}
	updates := map[string]any{}
	if in.Name != nil {
		name, err := validatePetName(*in.Name)
		if err != nil {
			return nil, err
		}
		updates["name"] = name
	}
	if in.Species != nil {
		species, err := validatePetSpecies(*in.Species)
		if err != nil {
			return nil, err
		}
		updates["species"] = species
	}
	if in.AvatarID != nil {
		if err := validatePetAvatar(*in.AvatarID); err != nil {
			return nil, err
		}
		updates["avatar_id"] = *in.AvatarID
	}
	if in.Persona != nil {
		persona, err := validatePetPersona(in.Persona)
		if err != nil {
			return nil, err
		}
		updates["persona"] = persona
	}

	if err := s.pets.Update(ctx, uid, petID, updates); err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrPetNotFound
		}
		return nil, fmt.Errorf("update pet: %w", err)
	}
	return s.Detail(ctx, uid, petID)
}

func (s *PetService) State(ctx context.Context, uid, petID int64) (*model.PetState, error) {
	pet, err := s.pets.FindByUserAndID(ctx, uid, petID)
	if repo.IsNotFound(err) {
		return nil, ErrPetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find pet state: %w", err)
	}
	return &model.PetState{
		ID:       pet.ID,
		Name:     pet.Name,
		Species:  pet.Species,
		AvatarID: pet.AvatarID,
		Level:    pet.Level,
		Intimacy: pet.Intimacy,
		Mood:     pet.Mood,
	}, nil
}

func validatePetName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if n := utf8.RuneCountInString(trimmed); n < 1 || n > 20 {
		return "", ErrPetNameInvalid
	}
	return trimmed, nil
}

func validatePetSpecies(species string) (string, error) {
	trimmed := strings.TrimSpace(species)
	if n := utf8.RuneCountInString(trimmed); n < 1 || n > 32 {
		return "", ErrPetSpeciesInvalid
	}
	return trimmed, nil
}

func validatePetAvatar(id int16) error {
	if id < 1 || id > maxAvatarID {
		return ErrPetAvatarInvalid
	}
	return nil
}

func validatePetPersona(persona *string) (*string, error) {
	if persona == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*persona)
	if utf8.RuneCountInString(trimmed) > maxPetPersonaLen {
		return nil, ErrPetPersonaInvalid
	}
	return &trimmed, nil
}

func toPetProfile(p *model.Pet) *model.PetProfile {
	return &model.PetProfile{
		ID:        p.ID,
		UserID:    p.UserID,
		Name:      p.Name,
		Species:   p.Species,
		AvatarID:  p.AvatarID,
		Persona:   p.Persona,
		Level:     p.Level,
		Intimacy:  p.Intimacy,
		Mood:      p.Mood,
		CreatedAt: p.CreatedAt.Unix(),
		UpdatedAt: p.UpdatedAt.Unix(),
	}
}
