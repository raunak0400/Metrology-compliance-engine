package users

type ListUsersQuery struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	IsActive string `form:"isActive" binding:"omitempty,oneof=true false"`
	Search   string `form:"search" binding:"omitempty,max=120"`
}

type UpdateUserRequest struct {
	FullName string   `json:"fullName" binding:"omitempty,min=2,max=120"`
	Phone    string   `json:"phone" binding:"omitempty,min=6,max=20"`
	IsActive *bool    `json:"isActive"`
	RoleIDs  []string `json:"roleIds" binding:"omitempty,dive,uuid4"`
}
