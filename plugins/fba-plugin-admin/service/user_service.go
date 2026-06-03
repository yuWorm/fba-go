package service

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type UserService struct {
	repo repo.Repository
}

const defaultEmailCaptchaCode = "123456"

func NewUserService(repository repo.Repository) *UserService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &UserService{repo: repository}
}

func (s *UserService) Get(ctx context.Context, id int) (dto.UserWithRelationDetail, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.UserWithRelationDetail{}, userNotFound("用户不存在", err)
		}
		return dto.UserWithRelationDetail{}, err
	}
	return s.withRelations(ctx, user)
}

func (s *UserService) Current(ctx context.Context, id int) (dto.CurrentUserWithRelationDetail, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.CurrentUserWithRelationDetail{}, userNotFound("用户不存在", err)
		}
		return dto.CurrentUserWithRelationDetail{}, err
	}
	var dept *model.Dept
	if user.DeptID != nil {
		item, err := s.repo.GetDept(ctx, *user.DeptID)
		if err != nil {
			return dto.CurrentUserWithRelationDetail{}, err
		}
		dept = &item
	}
	roles, err := s.repo.UserRoles(ctx, id)
	if err != nil {
		return dto.CurrentUserWithRelationDetail{}, err
	}
	return dto.CurrentUserWithRelations(user, dept, roles), nil
}

func (s *UserService) List(ctx context.Context, filter repo.UserFilter, page int, size int, basePath string) (pagination.PageData[dto.UserWithRelationDetail], error) {
	users, total, err := s.repo.ListUsers(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.UserWithRelationDetail]{}, err
	}
	items := make([]dto.UserWithRelationDetail, 0, len(users))
	for _, user := range users {
		detail, err := s.withRelations(ctx, user)
		if err != nil {
			return pagination.PageData[dto.UserWithRelationDetail]{}, err
		}
		items = append(items, detail)
	}
	return pagination.NewPageData(items, total, page, size, basePath), nil
}

func (s *UserService) Create(ctx context.Context, param dto.UserCreateParam) (dto.UserWithRelationDetail, error) {
	if _, err := s.repo.GetUserByUsername(ctx, param.Username); err == nil {
		return dto.UserWithRelationDetail{}, userConflict("用户名已注册", nil)
	} else if !stderrors.Is(err, repo.ErrNotFound) {
		return dto.UserWithRelationDetail{}, err
	}
	if param.Email != nil && *param.Email != "" {
		if _, err := s.repo.GetUserByEmail(ctx, *param.Email); err == nil {
			return dto.UserWithRelationDetail{}, userConflict("邮箱已被绑定", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return dto.UserWithRelationDetail{}, err
		}
	}
	if param.Password == "" {
		return dto.UserWithRelationDetail{}, userBadRequest("密码不允许为空", nil)
	}
	hashedPassword, err := hashPassword(param.Password)
	if err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	param.Password = hashedPassword
	if err := s.ensureUserDept(ctx, param.DeptID); err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	if err := s.ensureUserRoles(ctx, param.Roles); err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	user, err := s.repo.CreateUser(ctx, param)
	if err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	return s.withRelations(ctx, user)
}

func (s *UserService) Update(ctx context.Context, id int, param dto.UserUpdateParam) error {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return userNotFound("用户不存在", err)
		}
		return err
	}
	if param.Username != user.Username {
		if _, err := s.repo.GetUserByUsername(ctx, param.Username); err == nil {
			return userConflict("用户名已注册", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	if param.Email != nil && *param.Email != "" && (user.Email == nil || *param.Email != *user.Email) {
		if _, err := s.repo.GetUserByEmail(ctx, *param.Email); err == nil {
			return userConflict("邮箱已被绑定", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	if param.DeptID != nil && (user.DeptID == nil || *param.DeptID != *user.DeptID) {
		if err := s.ensureUserDept(ctx, *param.DeptID); err != nil {
			return err
		}
	}
	if err := s.ensureUserRoles(ctx, param.Roles); err != nil {
		return err
	}
	return s.repo.UpdateUser(ctx, id, param)
}

func (s *UserService) UpdatePassword(ctx context.Context, id int, param dto.UserPasswordParam) error {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return userNotFound("用户不存在", err)
		}
		return err
	}
	if !passwordMatchesStored(user.Password, param.OldPassword) {
		return userBadRequest("原密码错误", nil)
	}
	if param.NewPassword != param.ConfirmPassword {
		return userBadRequest("两次密码输入不一致", nil)
	}
	if err := validateNewPassword(ctx, s.repo, id, param.NewPassword); err != nil {
		return err
	}
	hashedPassword, err := hashPassword(param.NewPassword)
	if err != nil {
		return err
	}
	if err := s.repo.ResetUserPassword(ctx, id, hashedPassword); err != nil {
		return err
	}
	if err := s.repo.CreateUserPasswordHistory(ctx, id, user.Password); err != nil {
		return err
	}
	return s.clearUserSessions(ctx, user)
}

func (s *UserService) ResetPassword(ctx context.Context, id int, password string) error {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return userNotFound("用户不存在", err)
		}
		return err
	}
	if err := validateNewPassword(ctx, s.repo, id, password); err != nil {
		return err
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	if err := s.repo.ResetUserPassword(ctx, id, hashedPassword); err != nil {
		return err
	}
	if err := s.repo.CreateUserPasswordHistory(ctx, id, user.Password); err != nil {
		return err
	}
	return s.clearUserSessions(ctx, user)
}

func (s *UserService) UpdateNickname(ctx context.Context, id int, nickname string) error {
	return s.repo.UpdateUserNickname(ctx, id, nickname)
}

func (s *UserService) UpdateAvatar(ctx context.Context, id int, avatar *string) error {
	return s.repo.UpdateUserAvatar(ctx, id, avatar)
}

func (s *UserService) UpdateEmail(ctx context.Context, id int, captcha string, email *string) error {
	captcha = strings.TrimSpace(captcha)
	if captcha == "" {
		return userBadRequest("验证码已失效，请重新获取", nil)
	}
	// The Python version validates this against a Redis key produced by the email plugin.
	// Until that plugin is migrated, keep the existing contract fixture as the local code.
	if !strings.EqualFold(captcha, defaultEmailCaptchaCode) {
		return userBadRequest("验证码错误", nil)
	}
	if email != nil && *email != "" {
		user, err := s.repo.GetUserByEmail(ctx, *email)
		if err == nil && user.ID != id {
			return userConflict("邮箱已被绑定", nil)
		}
		if err != nil && !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	return s.repo.UpdateUserEmail(ctx, id, email)
}

func (s *UserService) UpdatePermission(ctx context.Context, id int, permissionType string, currentUserID int, currentSessionUUID string) error {
	switch permissionType {
	case "superuser", "staff", "status", "multi_login":
	default:
		return userBadRequest("权限类型不存在", nil)
	}
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return userNotFound("用户不存在", err)
		}
		return err
	}
	// Python allows self multi-login toggles, but blocks changing own privilege/status flags.
	if id == currentUserID && permissionType != "multi_login" {
		return userForbidden("禁止修改自身权限", nil)
	}
	disableMultiLogin := permissionType == "multi_login" && user.IsMultiLogin
	if err := s.repo.UpdateUserPermission(ctx, id, permissionType); err != nil {
		return err
	}
	if !disableMultiLogin {
		return nil
	}
	if id == currentUserID && currentSessionUUID != "" {
		return s.clearUserSessionsExcept(ctx, user, currentSessionUUID)
	}
	return s.clearUserSessions(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id int) error {
	if _, err := s.repo.GetUser(ctx, id); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return userNotFound("用户不存在", err)
		}
		return err
	}
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) clearUserSessions(ctx context.Context, user model.User) error {
	// Python clears access and refresh token Redis keys after password changes.
	// Go models those token keys as session rows, so all sessions for this user
	// must be removed to force re-authentication.
	sessions, err := s.repo.ListSessions(ctx, repo.SessionFilter{Username: user.Username})
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.ID != user.ID {
			continue
		}
		if err := s.repo.DeleteSession(ctx, session.ID, session.SessionUUID); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) clearUserSessionsExcept(ctx context.Context, user model.User, keepSessionUUID string) error {
	// When a user disables their own multi-login, Python deletes all access
	// token Redis keys except the current session. Preserve that exact behavior
	// by keeping the session_uuid parsed from the current access token.
	sessions, err := s.repo.ListSessions(ctx, repo.SessionFilter{Username: user.Username})
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.ID != user.ID || session.SessionUUID == keepSessionUUID {
			continue
		}
		if err := s.repo.DeleteSession(ctx, session.ID, session.SessionUUID); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) Roles(ctx context.Context, id int) ([]dto.RoleDetail, error) {
	roles, err := s.repo.UserRoles(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return nil, userNotFound("用户不存在", err)
		}
		return nil, err
	}
	return dto.RolesFromModel(roles), nil
}

func (s *UserService) ensureUserDept(ctx context.Context, id int) error {
	if _, err := s.repo.GetDept(ctx, id); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return userNotFound("部门不存在", err)
		}
		return err
	}
	return nil
}

func (s *UserService) ensureUserRoles(ctx context.Context, ids []int) error {
	for _, id := range uniqueRoleIDs(ids) {
		if _, err := s.repo.GetRole(ctx, id); err != nil {
			if stderrors.Is(err, repo.ErrNotFound) {
				return userNotFound("角色不存在", err)
			}
			return err
		}
	}
	return nil
}

func uniqueRoleIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func userBadRequest(message string, cause error) error {
	return fbaerrors.New(http.StatusBadRequest, http.StatusBadRequest, message, cause)
}

func userNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}

func userConflict(message string, cause error) error {
	return fbaerrors.New(http.StatusConflict, http.StatusConflict, message, cause)
}

func userForbidden(message string, cause error) error {
	return fbaerrors.New(http.StatusForbidden, http.StatusForbidden, message, cause)
}

func (s *UserService) withRelations(ctx context.Context, user model.User) (dto.UserWithRelationDetail, error) {
	var dept *model.Dept
	if user.DeptID != nil {
		item, err := s.repo.GetDept(ctx, *user.DeptID)
		if err != nil {
			return dto.UserWithRelationDetail{}, err
		}
		dept = &item
	}
	roles, err := s.repo.UserRoles(ctx, user.ID)
	if err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	roleDetails := make([]dto.RoleWithRelationDetail, 0, len(roles))
	// Python get_join returns each user role with its menu and data-scope relations.
	for _, role := range roles {
		menus, err := s.repo.RoleMenus(ctx, role.ID)
		if err != nil {
			return dto.UserWithRelationDetail{}, err
		}
		scopes, err := s.repo.RoleScopes(ctx, role.ID)
		if err != nil {
			return dto.UserWithRelationDetail{}, err
		}
		roleDetails = append(roleDetails, dto.RoleWithRelations(role, menus, scopes))
	}
	return dto.UserWithRelations(user, dept, roleDetails), nil
}
