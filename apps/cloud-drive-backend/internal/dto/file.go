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
	Page     int  `form:"page" binding:"min=1"`
	PageSize int  `form:"page_size" binding:"min=1,max=100"`
}

type DownloadByPickUpCodeReq struct {
	Code string `form:"code" binding:"required"`
}
