package model

import "time"

type UserModel struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Username     string    `gorm:"unique;not null" json:"username"`
    Email        string    `gorm:"unique" json:"email"`
    PasswordHash string    `gorm:"not null" json:"-"` // ❗绝对不能返回
    AvatarURL    string    `json:"avatar_url"`
    StorageLimit uint64    `json:"storage_limit"` // 总空间（字节）
    StorageUsed  uint64    `json:"storage_used"`  // 已使用
    Status       int       `json:"status"`        // 1正常 0禁用
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}