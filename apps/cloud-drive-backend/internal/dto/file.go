package dto

import (
	"cloud-drive-backend/internal/model"
	"time"
)

type InitUploadFileReq struct {
	FileName    string `json:"file_name" binding:"required"`
	FileSize    uint64 `json:"file_size" binding:"required,min=1"`
	FileHash    string `json:"file_hash" binding:"required"`
	ChunkSize   int    `json:"chunk_size" binding:"required,min=1"`
	TotalChunks int    `json:"total_chunks" binding:"required,min=1"`
	FileType    string `json:"file_type" binding:"required"`
	FolderID    uint   `json:"folder_id"`
}

type UploadChunkReq struct {
	TaskID     uint   `form:"task_id" binding:"required,min=1"`
	ChunkIndex int    `form:"chunk_index" binding:"min=0"`
	ChunkHash  string `form:"chunk_hash"`
}

type MergeUploadedChunksReq struct {
	TaskID uint `json:"task_id" binding:"required,min=1"`
}

type FileListItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // file / folder
	FileType  string    `json:"file_type,omitempty"`
	Size      uint64    `json:"size,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetListByFolderIDAndUserIDReq struct {
	FolderID uint `form:"folder_id"`
	Page     int  `form:"page" binding:"min=1"`
	PageSize int  `form:"page_size" binding:"min=1,max=100"`
}

type GetListCountByFolderIDAndUserIDReq struct {
	FolderID uint `form:"folder_id"`
}

type MakeDirectoryReq struct {
	FolderID uint   `json:"folder_id"`
	Name     string `json:"name"`
}

type CreatePickUpCodeReq struct {
	Code         string                 `json:"code"`
	FileID       *uint                  `json:"file_id"`
	FolderID     *uint                  `json:"folder_id"`
	Type         model.PickUpTargetType `json:"type"`
	MaxDownloads int                    `json:"max_downloads"`
	ExpireTime   time.Time              `json:"expire_time"`
}

type GetPickUpCodeListByUserIDAndPageReq struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

type DeletePickUpCodeReq struct {
	ID uint `form:"id" binding:"required,min=1"`
}

type DownloadByPickUpCodeReq struct {
	Code string `form:"code" binding:"required"`
}

type PreviewFileReq struct {
	FileID uint `form:"file_id" binding:"required,min=1"`
}

type DownloadFileReq struct {
	FileID   uint `form:"file_id" binding:"omitempty,min=1"`
	FolderID uint `form:"folder_id" binding:"omitempty,min=1"`
}

type RenameFileReq struct {
	FileID   uint   `json:"file_id" binding:"omitempty,min=1"`
	FolderID uint   `json:"folder_id" binding:"omitempty,min=1"`
	Name     string `json:"name" binding:"required"`
}

type MoveFileReq struct {
	FileID         uint `json:"file_id" binding:"omitempty,min=1"`
	FolderID       uint `json:"folder_id" binding:"omitempty,min=1"`
	TargetFolderID uint `json:"target_folder_id"`
}

type DeleteFileReq struct {
	FileID   uint `json:"file_id" binding:"omitempty,min=1"`
	FolderID uint `json:"folder_id" binding:"omitempty,min=1"`
}

type CreatePublicShareLinkReq struct {
	FileID uint `json:"file_id" binding:"required,min=1"`
}

type GetPublicShareLinkReq struct {
	FileID uint `form:"file_id" binding:"required,min=1"`
}

type DeletePublicShareLinkReq struct {
	FileID uint `form:"file_id" binding:"required,min=1"`
}

type OpenPublicShareReq struct {
	Token string `form:"token" binding:"required"`
}

type DashboardFileStatItem struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
	Size  uint64 `json:"size"`
}

type DashboardRecentActivityItem struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	FolderName string    `json:"folder_name"`
	FileType   string    `json:"file_type"`
	Size       uint64    `json:"size"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DashboardOverviewResp struct {
	StorageUsed        uint64                        `json:"storage_used"`
	StorageTotal       uint64                        `json:"storage_total"`
	StorageLeft        uint64                        `json:"storage_left"`
	StorageUsedPercent int                           `json:"storage_used_percent"`
	FileStats          []DashboardFileStatItem       `json:"file_stats"`
	RecentActivities   []DashboardRecentActivityItem `json:"recent_activities"`
}
