package repository

import (
	"cloud-drive-backend/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

// Create 创建用户
func (r *UserRepository) Create(user *model.UserModel) error {
	return r.DB.Create(user).Error
}

func InitUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}
