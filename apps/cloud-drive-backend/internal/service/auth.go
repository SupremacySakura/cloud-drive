package service

import (
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/utils"
	"errors"
)

type AuthService interface {
	RegisterUser(user *model.UserModel) error
	ValidateUser(username string, password string) (user *model.UserModel, error error)
	GenerateToken(userID uint) (token string, error error)
	GetUserByID(userID uint) (user *model.UserModel, error error)
}

type authService struct {
	UserRepository *repository.UserRepository
}

func NewAuthService(userRepository *repository.UserRepository) AuthService {
	return &authService{
		UserRepository: userRepository,
	}
}

func (s *authService) RegisterUser(user *model.UserModel) error {
	return s.UserRepository.Create(user)
}

func (s *authService) ValidateUser(username string, password string) (user *model.UserModel, error error) {
	user, err := s.UserRepository.GetUserByName(username)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, errors.New("密码错误")
	}
	return user, nil
}

func (s *authService) GenerateToken(userID uint) (token string, error error) {
	return utils.GenerateToken(userID)
}

func (s *authService) GetUserByID(userID uint) (user *model.UserModel, error error) {
	return s.UserRepository.GetUserByID(userID)
}