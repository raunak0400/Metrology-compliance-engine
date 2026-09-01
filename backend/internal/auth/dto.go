package auth

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	FullName string `json:"fullName" binding:"required,min=2,max=120"`
	Phone    string `json:"phone" binding:"omitempty,min=6,max=20"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required,min=10"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required,min=10"`
}
