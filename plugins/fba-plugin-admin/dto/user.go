package dto

import "github.com/yuWorm/fba-plugin-admin/model"

type UserCreateParam struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	DeptID   int     `json:"dept_id"`
	Roles    []int   `json:"roles"`
}

type UserUpdateParam struct {
	DeptID   *int    `json:"dept_id"`
	Username string  `json:"username"`
	Nickname string  `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Roles    []int   `json:"roles"`
}

type UserPasswordParam struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type UserResetPasswordParam struct {
	Password string `json:"password"`
}

type UserNicknameParam struct {
	Nickname string `json:"nickname"`
}

type UserAvatarParam struct {
	Avatar *string `json:"avatar"`
}

type UserEmailParam struct {
	Captcha string  `json:"captcha"`
	Email   *string `json:"email"`
}

type UserDetail struct {
	DeptID        *int    `json:"dept_id"`
	Username      string  `json:"username"`
	Nickname      string  `json:"nickname"`
	Avatar        *string `json:"avatar"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	ID            int     `json:"id"`
	UUID          string  `json:"uuid"`
	Status        int     `json:"status"`
	IsSuperuser   bool    `json:"is_superuser"`
	IsStaff       bool    `json:"is_staff"`
	IsMultiLogin  bool    `json:"is_multi_login"`
	JoinTime      string  `json:"join_time"`
	LastLoginTime *string `json:"last_login_time"`
}

type UserWithRelationDetail struct {
	UserDetail
	Dept  *DeptDetail              `json:"dept"`
	Roles []RoleWithRelationDetail `json:"roles"`
}

type CurrentUserWithRelationDetail struct {
	UserDetail
	Dept  *string  `json:"dept"`
	Roles []string `json:"roles"`
}

func UserFromModel(item model.User) UserDetail {
	return UserDetail{
		ID:            item.ID,
		UUID:          item.UUID,
		DeptID:        item.DeptID,
		Username:      item.Username,
		Nickname:      item.Nickname,
		Avatar:        item.Avatar,
		Email:         item.Email,
		Phone:         item.Phone,
		Status:        item.Status,
		IsSuperuser:   item.IsSuperuser,
		IsStaff:       item.IsStaff,
		IsMultiLogin:  item.IsMultiLogin,
		JoinTime:      formatTime(item.JoinTime),
		LastLoginTime: formatTimePtr(item.LastLoginTime),
	}
}

func UserWithRelations(item model.User, dept *model.Dept, roles []RoleWithRelationDetail) UserWithRelationDetail {
	var deptDetail *DeptDetail
	if dept != nil {
		detail := DeptFromModel(*dept)
		deptDetail = &detail
	}
	return UserWithRelationDetail{
		UserDetail: UserFromModel(item),
		Dept:       deptDetail,
		Roles:      roles,
	}
}

func CurrentUserWithRelations(item model.User, dept *model.Dept, roles []model.Role) CurrentUserWithRelationDetail {
	var deptName *string
	if dept != nil {
		deptName = &dept.Name
	}
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	return CurrentUserWithRelationDetail{
		UserDetail: UserFromModel(item),
		Dept:       deptName,
		Roles:      roleNames,
	}
}
