package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
	"uuid"
)

type UserStorage interface {
	CreateUser(ctx context.Context, model db.UserCreateModel) (uuid.UUID, error)
	UpdateUser(ctx context.Context, model db.UserUpdateModel) error
	GetUserByID(ctx context.Context, ID uuid.UUID) (*db.UserInfoModel, error)
	GetUserByEmail(ctx context.Context, email string) (*db.UserInfoModel, error)
}

type UserService struct {
	storage UserStorage
	log     *slog.Logger
}

func NewUserService(storage UserStorage, log *slog.Logger) *UserService {
	return &UserService{
		storage: storage,
		log:     log,
	}
}

func (s *UserService) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	userInfo, err := s.storage.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	if userInfo == nil {
		return false, nil
	}

	return true, nil
}

func (s *UserService) Create(ctx context.Context, u dtos.UserCreateRequest) (uuid.UUID, error) {
	dbModel := db.UserCreateModel{
		Email: u.Email,
	}

	userID, err := s.storage.CreateUser(ctx, dbModel)
	if err != nil {
		return uuid.Nil(), err
	}

	return userID, nil
}

func (s *UserService) Update(ctx context.Context, u dtos.UserUpdateRequest) error {
	dbModel := db.UserUpdateModel{
		ID:          u.ID,
		Name:        u.Name,
		Address:     u.Address,
		PhoneNumber: u.PhoneNumber,
	}

	if err := s.storage.UpdateUser(ctx, dbModel); err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, ID uuid.UUID) (*dtos.UserInfoResponse, error) {
	userModel, err := s.storage.GetUserByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	response := &dtos.UserInfoResponse{
		ID:          userModel.ID,
		Name:        userModel.Name,
		Email:       *userModel.Email,
		PhoneNumber: userModel.PhoneNumber,
		Address:     userModel.Address,
		Role:        *userModel.Role,
	}

	return response, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*dtos.UserInfoResponse, error) {
	userModel, err := s.storage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	response := &dtos.UserInfoResponse{
		ID:          userModel.ID,
		Name:        userModel.Name,
		Email:       *userModel.Email,
		PhoneNumber: userModel.PhoneNumber,
		Address:     userModel.Address,
		Role:        *userModel.Role,
	}

	return response, nil
}
