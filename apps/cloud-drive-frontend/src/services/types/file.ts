export type InitUploadFileRequest = {
    file_name: string
    file_size: number
    file_hash: string
    chunk_size: number
    total_chunks: number
    file_type: string
    folder_id: number
}

export type UploadStatus = 'uploading' | 'completed'

export type InitUploadFileResponse = {
    task_id: number
    uploaded_chunks: number[]
    status: UploadStatus
}

export type UploadChunkRequest = {
    task_id: number
    chunk_index: number
    chunk_data: string
    chunk_hash: string
}

export type MergeUploadedChunksRequest = {
    task_id: number
}

export type FileListItem = {
    id: number
    name: string
    type: 'folder' | 'file'
    file_type: string
    size: number
    updated_at: string
}