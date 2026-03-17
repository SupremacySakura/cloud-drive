package service

import (
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
)

type AuthService struct {
	UserRepository *repository.UserRepository
}

func (s AuthService) RegisterUser(user *model.UserModel) error {
	return s.UserRepository.Create(user)
}
