package service

import (
	"context"
	"strings"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

type petEnv struct {
	svc *PetService
	ur  *repo.UserRepo
}

func newPetEnv(t *testing.T) *petEnv {
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
	return &petEnv{svc: NewPetService(repo.NewPetRepo(gdb), nil), ur: repo.NewUserRepo(gdb)}
}

func (e *petEnv) user(t *testing.T, phone string) int64 {
	t.Helper()
	u, err := e.ur.CreateUser(context.Background(), phone)
	if err != nil {
		t.Fatalf("create user %s: %v", phone, err)
	}
	return u.ID
}

func TestPetServiceCreateListUpdateState(t *testing.T) {
	env := newPetEnv(t)
	ctx := context.Background()
	uid := env.user(t, "13801000001")
	persona := `{"style":"warm"}`
	avatar := int16(3)

	pet, err := env.svc.Create(ctx, uid, &model.CreatePetInput{
		Name:     "  小燕  ",
		AvatarID: &avatar,
		Persona:  &persona,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pet.ID == 0 || pet.UserID != uid || pet.Name != "小燕" || pet.Species != model.DefaultPetSpecies || pet.AvatarID != 3 {
		t.Fatalf("创建结果异常: %+v", pet)
	}
	if pet.Level != 1 || pet.Intimacy != 0 || pet.Mood != model.DefaultPetMood {
		t.Errorf("默认状态异常: %+v", pet)
	}

	items, err := env.svc.List(ctx, uid)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 || items[0].ID != pet.ID {
		t.Fatalf("List = %+v, want created pet", items)
	}

	newName := "阿语"
	newSpecies := "cat"
	newAvatar := int16(5)
	updated, err := env.svc.Update(ctx, uid, pet.ID, &model.UpdatePetInput{Name: &newName, Species: &newSpecies, AvatarID: &newAvatar})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "阿语" || updated.Species != "cat" || updated.AvatarID != 5 {
		t.Errorf("更新结果异常: %+v", updated)
	}

	state, err := env.svc.State(ctx, uid, pet.ID)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if state.Name != "阿语" || state.Level != 1 || state.Intimacy != 0 || state.Mood != model.DefaultPetMood {
		t.Errorf("状态异常: %+v", state)
	}
}

func TestPetServiceValidationAndScope(t *testing.T) {
	env := newPetEnv(t)
	ctx := context.Background()
	uid := env.user(t, "13801000002")
	other := env.user(t, "13801000003")

	_, err := env.svc.Create(ctx, uid, &model.CreatePetInput{Name: ""})
	wantCode(t, err, ErrPetNameInvalid.Code)

	badAvatar := int16(9)
	_, err = env.svc.Create(ctx, uid, &model.CreatePetInput{Name: "小燕", AvatarID: &badAvatar})
	wantCode(t, err, ErrPetAvatarInvalid.Code)

	longPersona := strings.Repeat("设", maxPetPersonaLen+1)
	_, err = env.svc.Create(ctx, uid, &model.CreatePetInput{Name: "小燕", Persona: &longPersona})
	wantCode(t, err, ErrPetPersonaInvalid.Code)

	pet, err := env.svc.Create(ctx, uid, &model.CreatePetInput{Name: "小燕"})
	if err != nil {
		t.Fatalf("Create valid: %v", err)
	}
	_, err = env.svc.Detail(ctx, other, pet.ID)
	wantCode(t, err, ErrPetNotFound.Code)

	newName := "偷看"
	_, err = env.svc.Update(ctx, other, pet.ID, &model.UpdatePetInput{Name: &newName})
	wantCode(t, err, ErrPetNotFound.Code)
}
