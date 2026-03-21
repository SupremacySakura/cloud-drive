package vo

import "cloud-drive-backend/internal/model"

type InitUploadFileResp struct {
	TaskID         uint               `json:"task_id"`
	UploadedChunks []int              `json:"uploaded_chunks"`
	Status         model.UploadStatus `json:"status"`
}
