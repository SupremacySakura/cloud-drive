package vo

import (
	"cloud-drive-backend/internal/model"
	"time"
)

type InitUploadFileResp struct {
	TaskID         uint               `json:"task_id"`
	UploadedChunks []int              `json:"uploaded_chunks"`
	Status         model.UploadStatus `json:"status"`
}

type PickUpCodeListItem struct {
	ID           uint                   `json:"id"`
	Code         string                 `json:"code"`
	FileID       *uint                  `json:"file_id"`
	FolderID     *uint                  `json:"folder_id"`
	Name         string                 `json:"name"`
	Type         model.PickUpTargetType `json:"type"`
	Download     int                    `json:"download"`
	MaxDownload  int                    `json:"max_download"`
	ExpireTime   time.Time              `json:"expire_time"`
	CreatedAt    time.Time              `json:"created_at"`
	Status       model.PickUpCodeStatus `json:"status"`
}
