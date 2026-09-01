package roles

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=40"`
	Description string   `json:"description" binding:"omitempty,max=200"`
	Permissions []string `json:"permissions"`
}
