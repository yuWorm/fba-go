package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-admin/dto"
	adminmigration "github.com/yuWorm/fba-plugin-admin/migration"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGORMRepositoryPersistsCoreAdminRelations(t *testing.T) {
	repository := newGORMRepository(t)
	ctx := context.Background()

	admin, err := repository.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("GetUserByUsername(admin) error = %v", err)
	}
	if !admin.IsSuperuser || !admin.IsStaff {
		t.Fatalf("admin flags = superuser:%v staff:%v, want true true", admin.IsSuperuser, admin.IsStaff)
	}

	roles, err := repository.UserRoles(ctx, admin.ID)
	if err != nil {
		t.Fatalf("UserRoles(admin) error = %v", err)
	}
	if len(roles) != 1 || roles[0].Name != "admin" {
		t.Fatalf("admin roles = %+v, want admin role", roles)
	}

	menus, err := repository.RoleMenus(ctx, roles[0].ID)
	if err != nil {
		t.Fatalf("RoleMenus(admin) error = %v", err)
	}
	if len(menus) != 1 || menus[0].Name != "Dashboard" {
		t.Fatalf("admin role menus = %+v, want Dashboard", menus)
	}

	nickname := "GORM User"
	created, err := repository.CreateUser(ctx, dto.UserCreateParam{
		Username: "gorm_user",
		Password: "secret",
		Nickname: &nickname,
		DeptID:   1,
		Roles:    []int{2},
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	userRoles, err := repository.UserRoles(ctx, created.ID)
	if err != nil {
		t.Fatalf("UserRoles(created) error = %v", err)
	}
	if len(userRoles) != 1 || userRoles[0].ID != 2 {
		t.Fatalf("created user roles = %+v, want role 2", userRoles)
	}

	users, total, err := repository.ListUsers(ctx, repo.UserFilter{Username: "gorm"}, 1, 20)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if total != 1 || len(users) != 1 || users[0].Username != "gorm_user" {
		t.Fatalf("ListUsers(gorm) = total:%d items:%+v, want gorm_user", total, users)
	}
}

func TestGORMRepositoryPersistsSessions(t *testing.T) {
	repository := newGORMRepository(t)
	ctx := context.Background()
	expires := time.Now().Add(time.Hour).Truncate(time.Second)
	session := model.Session{
		ID:            1,
		SessionUUID:   "gorm-session",
		Username:      "admin",
		Nickname:      "Admin",
		IP:            "127.0.0.1",
		OS:            "test",
		Browser:       "test",
		Device:        "test",
		Status:        1,
		LastLoginTime: "2026-06-02 10:00:00",
		ExpireTime:    expires,
	}

	if err := repository.UpsertSession(ctx, session); err != nil {
		t.Fatalf("UpsertSession(insert) error = %v", err)
	}
	session.Nickname = "Updated Admin"
	if err := repository.UpsertSession(ctx, session); err != nil {
		t.Fatalf("UpsertSession(update) error = %v", err)
	}

	got, err := repository.GetSession(ctx, 1, "gorm-session")
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.Nickname != "Updated Admin" || !got.ExpireTime.Equal(expires) {
		t.Fatalf("session = %+v, want updated nickname and expire time %s", got, expires)
	}

	if err := repository.DeleteSession(ctx, 1, "gorm-session"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := repository.GetSession(ctx, 1, "gorm-session"); err != repo.ErrNotFound {
		t.Fatalf("GetSession(after delete) error = %v, want ErrNotFound", err)
	}
}

func newGORMRepository(t *testing.T) repo.Repository {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	provider := db.NewGORMProvider(gormDB, nil)
	migration := adminmigration.AutoMigrate(provider)
	if err := migration.Up(context.Background()); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repository := repo.NewGORMRepository(provider, repo.SeedData())
	if err := repository.Seed(context.Background()); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	return repository
}
