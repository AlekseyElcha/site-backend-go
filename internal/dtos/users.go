package dtos

import "uuid"

//type UserInfoRequest struct {
//	ID uuid.UUID `json:"id"`
//}

type UserInfoResponse struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        *string   `db:"name" json:"name"`
	Email       string    `db:"email" json:"email"`
	PhoneNumber *string   `db:"phone_number" json:"phone_number"`
	Address     *string   `db:"address" json:"address"`
	Role        string    `db:"role" json:"role"`
}

type UserCreateRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type UserUpdateRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
	// Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
	Name        *string `json:"name,omitempty" validate:"omitempty"`
	Address     *string `json:"address,omitempty" validate:"omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty" validate:"omitempty"`
}
