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
	for _, name := range []string{"Dashboard", "System", "PluginConfig", "PluginDict", "PluginNotice", "Scheduler", "AddConfig", "AddDictType", "AddNotice", "AddScheduler"} {
		if !hasMenuName(menus, name) {
			t.Fatalf("admin role menus = %+v, want menu %s", menus, name)
		}
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

func hasMenuName(menus []model.Menu, name string) bool {
	for _, menu := range menus {
		if menu.Name == name {
			return true
		}
	}
	return false
}

func TestGORMRepositoryPersistsSessions(t *testing.T) {
	repository := newGORMRepository(t)
	ctx := context.Background()
	expires := time.Now().Add(time.Hour).Truncate(time.Second)
	session := model.Session{
		ID:            1,
		SessionUUID:   "gorm-session",
		AccessToken:   "initial-access-token",
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
	session.AccessToken = "updated-access-token"
	if err := repository.UpsertSession(ctx, session); err != nil {
		t.Fatalf("UpsertSession(update) error = %v", err)
	}

	got, err := repository.GetSession(ctx, 1, "gorm-session")
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.Nickname != "Updated Admin" || got.AccessToken != "updated-access-token" || !got.ExpireTime.Equal(expires) {
		t.Fatalf("session = %+v, want updated nickname, access token and expire time %s", got, expires)
	}

	if err := repository.DeleteSession(ctx, 1, "gorm-session"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := repository.GetSession(ctx, 1, "gorm-session"); err != repo.ErrNotFound {
		t.Fatalf("GetSession(after delete) error = %v, want ErrNotFound", err)
	}
}

func TestGORMRepositoryPersistsPlugins(t *testing.T) {
	provider := newGORMProvider(t)
	repository := seedGORMRepository(t, provider)
	ctx := context.Background()

	plugins, err := repository.AllPlugins(ctx)
	if err != nil {
		t.Fatalf("AllPlugins() error = %v", err)
	}
	if len(plugins) < 3 {
		t.Fatalf("plugins = %+v, want seeded built-in plugins", plugins)
	}

	if err := repository.TogglePluginStatus(ctx, "dict"); err != nil {
		t.Fatalf("TogglePluginStatus(dict) error = %v", err)
	}
	nextRepository := repo.NewGORMRepository(provider, repo.SeedData())
	dictPlugin, err := nextRepository.GetPlugin(ctx, "dict")
	if err != nil {
		t.Fatalf("GetPlugin(dict) after toggle error = %v", err)
	}
	if dictPlugin.Enabled {
		t.Fatalf("dict Enabled = true, want persisted false")
	}
	changed, err := nextRepository.PluginsChanged(ctx)
	if err != nil {
		t.Fatalf("PluginsChanged() error = %v", err)
	}
	if !changed {
		t.Fatal("PluginsChanged() = false, want true after persisted toggle")
	}

	installed, err := repository.InstallPlugin(ctx, dto.PluginInstallParam{Type: "git", Name: "external"})
	if err != nil {
		t.Fatalf("InstallPlugin() error = %v", err)
	}
	if installed.ID != "external" || !installed.Enabled {
		t.Fatalf("installed plugin = %+v, want enabled external", installed)
	}
	if err := repository.UninstallPlugin(ctx, "external"); err != nil {
		t.Fatalf("UninstallPlugin(external) error = %v", err)
	}
	if _, err := nextRepository.GetPlugin(ctx, "external"); err != repo.ErrNotFound {
		t.Fatalf("GetPlugin(external after uninstall) error = %v, want ErrNotFound", err)
	}
}

func TestGORMRepositoryPersistsLogs(t *testing.T) {
	provider := newGORMProvider(t)
	repository := seedGORMRepository(t, provider)
	ctx := context.Background()

	loginLogs, total, err := repository.ListLoginLogs(ctx, repo.LogFilter{Username: "admin", Status: intPtr(1), IP: "127"}, 1, 20)
	if err != nil {
		t.Fatalf("ListLoginLogs() error = %v", err)
	}
	if total != 1 || len(loginLogs) != 1 || loginLogs[0].Username != "admin" {
		t.Fatalf("login logs = total:%d items:%+v, want admin fixture", total, loginLogs)
	}
	deleted, err := repository.DeleteLoginLogs(ctx, []int{loginLogs[0].ID})
	if err != nil {
		t.Fatalf("DeleteLoginLogs() error = %v", err)
	}
	if deleted != 1 {
		t.Fatalf("DeleteLoginLogs() deleted = %d, want 1", deleted)
	}
	nextRepository := repo.NewGORMRepository(provider, repo.SeedData())
	_, total, err = nextRepository.ListLoginLogs(ctx, repo.LogFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("ListLoginLogs(after delete) error = %v", err)
	}
	if total != 0 {
		t.Fatalf("login log total after delete = %d, want 0", total)
	}
	if err := nextRepository.CreateLoginLog(ctx, model.LoginLog{
		UserUUID:  "created-login-user",
		Username:  "created_login",
		Status:    0,
		IP:        "10.0.0.8",
		Msg:       "用户名或密码有误",
		LoginTime: time.Now(),
	}); err != nil {
		t.Fatalf("CreateLoginLog() error = %v", err)
	}
	createdLoginLogs, total, err := nextRepository.ListLoginLogs(ctx, repo.LogFilter{Username: "created_login", Status: intPtr(0), IP: "10.0"}, 1, 20)
	if err != nil {
		t.Fatalf("ListLoginLogs(created) error = %v", err)
	}
	if total != 1 || len(createdLoginLogs) != 1 || createdLoginLogs[0].Msg != "用户名或密码有误" {
		t.Fatalf("created login logs = total:%d items:%+v, want persisted failure log", total, createdLoginLogs)
	}
	if err := nextRepository.DeleteAllLoginLogs(ctx); err != nil {
		t.Fatalf("DeleteAllLoginLogs() error = %v", err)
	}

	operaLogs, total, err := nextRepository.ListOperaLogs(ctx, repo.LogFilter{Username: "admin", Status: intPtr(1), IP: "127"}, 1, 20)
	if err != nil {
		t.Fatalf("ListOperaLogs() error = %v", err)
	}
	if total != 1 || len(operaLogs) != 1 || operaLogs[0].Args["page"] != "1" {
		t.Fatalf("opera logs = total:%d items:%+v, want admin fixture with args", total, operaLogs)
	}
	if err := nextRepository.DeleteAllOperaLogs(ctx); err != nil {
		t.Fatalf("DeleteAllOperaLogs() error = %v", err)
	}
	finalRepository := repo.NewGORMRepository(provider, repo.SeedData())
	_, total, err = finalRepository.ListOperaLogs(ctx, repo.LogFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("ListOperaLogs(after clear) error = %v", err)
	}
	if total != 0 {
		t.Fatalf("opera log total after clear = %d, want 0", total)
	}
}

func newGORMRepository(t *testing.T) *repo.GORMRepository {
	return seedGORMRepository(t, newGORMProvider(t))
}

func newGORMProvider(t *testing.T) db.Provider {
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
	return provider
}

func seedGORMRepository(t *testing.T, provider db.Provider) *repo.GORMRepository {
	t.Helper()
	repository := repo.NewGORMRepository(provider, repo.SeedData())
	if err := repository.Seed(context.Background()); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	return repository
}

func intPtr(value int) *int {
	return &value
}
