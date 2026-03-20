package dto

type InitUploadFileReq struct {
	FileName    string `json:"file_name"`
	FileSize    uint64 `json:"file_size"`
	FileHash    string `json:"file_hash"`
	ChunkSize   int    `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
	FileType    string `json:"file_type"`
	FolderID    uint   `json:"folder_id"`
}
