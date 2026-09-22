package db

import (
	"context"
	"uuid"
)

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*UserInfoModel, error) {
	query := `
		SELECT 
			id
			, name
			, email
			, phone_number
			, address
			, role
		FROM users
		WHERE email = $1
	`

	var userModel UserInfoModel

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&userModel.ID,
		&userModel.Name,
		&userModel.Email,
		&userModel.PhoneNumber,
		&userModel.Address,
		&userModel.Role,
	)

	if err != nil {
		return nil, err
	}

	return &userModel, nil
}

func (s *Storage) GetUserByID(ctx context.Context, ID uuid.UUID) (*UserInfoModel, error) {
	query := `
		SELECT 
			id
			, name
			, email
			, phone_number
			, address
			, role
		FROM users
		WHERE id = $1
	`

	var userModel UserInfoModel

	err := s.db.QueryRowContext(ctx, query, ID).Scan(
		&userModel.ID,
		&userModel.Name,
		&userModel.Email,
		&userModel.PhoneNumber,
		&userModel.Address,
		&userModel.Role,
	)

	if err != nil {
		panic(err) // переписать
	}

	return &userModel, nil
}

func (s *Storage) GetUserRoleByID(ctx context.Context, ID uuid.UUID) (string, error) {
	query := `
		SELECT
			role
		FROM users
		WHERE id = $1
	`
	var role string
	err := s.db.QueryRowContext(ctx, query, ID).Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}

func (s *Storage) UpdateUser(ctx context.Context, model UserUpdateModel) error {
	query := `
		UPDATE users
		SET 
		    name = COALESCE($2, name)
			, address = COALESCE($3, address)
			, phone_number = COALESCE($4, phone_number)
		WHERE id = $1
	`

	if _, err := s.db.ExecContext(ctx, query, model.ID, model.Name, model.Address, model.PhoneNumber); err != nil {
		return err
	}

	return nil
}

func (s *Storage) CreateUser(ctx context.Context, model UserCreateModel) (uuid.UUID, error) {
	query := `
		INSERT INTO users (email)
		VALUES ($1)
		RETURNING id;`

	var userID uuid.UUID

	err := s.db.QueryRow(query, model.Email).Scan(&userID)
	if err != nil {
		return uuid.Nil(), err
	}

	return userID, nil
}
