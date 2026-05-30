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
