package dtos

type AuthCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	AuthCode string `json:"auth_code" validate:"required,auth_code"`
}

type SelfInfoResponse struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
