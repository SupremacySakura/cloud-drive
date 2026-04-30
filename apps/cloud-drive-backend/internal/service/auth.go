package service

import (
	"fmt"

	"cloud-drive-backend/internal/errors"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/utils"
)

type AuthService interface {
	RegisterUser(user *model.UserModel) error
	ValidateUser(username string, password string) (user *model.UserModel, error error)
	GenerateToken(userID uint) (token string, error error)
	GetUserByID(userID uint) (user *model.UserModel, error error)
}

type UserRepoInterface interface {
	Create(user *model.UserModel) error
	GetUserByName(username string) (*model.UserModel, error)
	GetUserByID(userID uint) (*model.UserModel, error)
}

type authService struct {
	UserRepo UserRepoInterface
}

func NewAuthService(userRepository *repository.UserRepository) AuthService {
	return &authService{
		UserRepo: userRepository,
	}
}

func (s *authService) RegisterUser(user *model.UserModel) error {
	err := s.UserRepo.Create(user)
	if err != nil {
		return fmt.Errorf("注册用户失败: %w", err)
	}
	return nil
}

func (s *authService) ValidateUser(username string, password string) (user *model.UserModel, error error) {
	user, err := s.UserRepo.GetUserByName(username)
	if err != nil {
		return nil, errors.Wrap(err, "用户不存在")
	}
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, errors.ErrInvalidPassword
	}
	return user, nil
}

func (s *authService) GenerateToken(userID uint) (token string, error error) {
	return utils.GenerateToken(userID)
}

func (s *authService) GetUserByID(userID uint) (user *model.UserModel, error error) {
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}
