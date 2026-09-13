package service

import (
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
)

type UserStorage interface {
	CreateUser(model db.UserCreateModel) error
}

type UserService struct {
	storage UserStorage
}

func NewUserService(storage UserStorage) *UserService {
	return &UserService{
		storage: storage,
	}
}

func (s *UserService) Create(u dtos.UserCreateRequest) error {
	dbbModel := db.UserCreateModel{
		Email: u.Email,
	}

	if err := s.storage.CreateUser(dbbModel); err != nil {
		return err
	}

	return nil
}
