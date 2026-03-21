package dto

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
