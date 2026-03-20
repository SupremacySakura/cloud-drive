package model

import "time"

type UserModel struct {
	ID           uint      `gorm:"primaryKey" json:"id"`            // 用户ID
	Username     string    `gorm:"unique;not null" json:"username"` // 用户名
	Email        string    `gorm:"unique" json:"email"`             // 邮箱
	PasswordHash string    `gorm:"not null" json:"-"`               // 密码 绝对不能返回
	AvatarURL    string    `json:"avatar_url"`                      // 头像URL
	StorageLimit uint64    `json:"storage_limit"`                   // 总空间（字节）
	StorageUsed  uint64    `json:"storage_used"`                    // 已使用空间（字节）
	Status       int       `json:"status"`                          // 1正常 0禁用
	CreatedAt    time.Time `json:"created_at"`                      // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`                      // 更新时间
}
