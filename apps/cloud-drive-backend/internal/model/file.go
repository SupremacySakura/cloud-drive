package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type IntSlice []int

func (s IntSlice) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]int(s))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *IntSlice) Scan(value interface{}) error {
	if s == nil {
		return fmt.Errorf("IntSlice: Scan on nil receiver")
	}

	switch v := value.(type) {
	case nil:
		*s = IntSlice{}
		return nil
	case []byte:
		if len(v) == 0 {
			*s = IntSlice{}
			return nil
		}
		var out []int
		if err := json.Unmarshal(v, &out); err != nil {
			return err
		}
		*s = IntSlice(out)
		return nil
	case string:
		if v == "" {
			*s = IntSlice{}
			return nil
		}
		var out []int
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			return err
		}
		*s = IntSlice(out)
		return nil
	default:
		return fmt.Errorf("IntSlice: unsupported Scan type %T", value)
	}
}

type FileModel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`      // 文件ID
	Name      string    `gorm:"not null" json:"name"`      // 文件名
	FileHash  string    `gorm:"not null" json:"file_hash"` // 文件哈希
	Type      string    `gorm:"not null" json:"type"`      // 文件类型: image/video/document/other
	Size      uint64    `gorm:"not null" json:"size"`      // 文件大小（字节）
	FolderID  uint      `json:"folder_id"`                 // 所属文件夹ID
	UserID    uint      `gorm:"not null" json:"user_id"`   // 所属用户ID
	CreatedAt time.Time `json:"created_at"`                // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                // 更新时间
	DeletedAt time.Time `gorm:"index" json:"deleted_at"`   // 删除时间（软删除）
}

type FolderModel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`    // 文件夹ID
	Name      string    `gorm:"not null" json:"name"`    // 文件夹名
	ParentID  uint      `json:"parent_id"`               // 父文件夹ID（0表示根文件夹）
	UserID    uint      `gorm:"not null" json:"user_id"` // 所属用户ID
	CreatedAt time.Time `json:"created_at"`              // 创建时间
	UpdatedAt time.Time `json:"updated_at"`              // 更新时间
	DeletedAt time.Time `gorm:"index" json:"deleted_at"` // 删除时间（软删除）
}

type UploadTask struct {
	ID             uint      `gorm:"primaryKey" json:"id"`                      // uploadId
	FileHash       string    `gorm:"not null" json:"file_hash"`                 // 文件hash（秒传关键）
	FileName       string    `gorm:"not null" json:"file_name"`                 // 文件名
	FileSize       uint64    `gorm:"not null" json:"file_size"`                 // 文件大小（字节）
	ChunkSize      int       `gorm:"not null" json:"chunk_size"`                // 分片大小（字节）
	TotalChunks    int       `gorm:"not null" json:"total_chunks"`              // 总分片数
	UploadedChunks IntSlice  `gorm:"type:json;not null" json:"uploaded_chunks"` // 已上传分片（已上传分片索引数组）
	FileType       string    `gorm:"not null" json:"file_type"`                 // 文件类型
	FolderID       uint      `json:"folder_id"`                                 // 所属文件夹ID
	UserID         uint      `gorm:"not null" json:"user_id"`                   // 所属用户ID
	Status         string    `gorm:"not null" json:"status"`                    // uploading / completed
	CreatedAt      time.Time `gorm:"not null" json:"created_at"`                // 创建时间
	UpdatedAt      time.Time `gorm:"not null" json:"updated_at"`                // 更新时间
}
