package repository

import (
	"cloud-drive-backend/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.UserModel) error {
	return r.DB.Create(user).Error
}

// 根据用户名查询用户
func (r *UserRepository) GetUserByName(username string) (*model.UserModel, error) {
	var user model.UserModel
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据用户ID查询用户
func (r *UserRepository) GetUserByID(userID uint) (*model.UserModel, error) {
	var user model.UserModel
	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
