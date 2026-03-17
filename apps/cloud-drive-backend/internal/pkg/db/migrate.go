package db

import "cloud-drive-backend/internal/model"

// Migrate 数据库迁移
func Migrate() error {
	return  DB.AutoMigrate(&model.UserModel{})
}
