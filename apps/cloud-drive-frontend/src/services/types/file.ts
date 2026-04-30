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

export type MakeDirectoryRequest = {
  folder_id: number
  name: string
}

export type RenameFileRequest = {
  file_id: number
  folder_id: number
  name: string
}

export type MoveFileRequest = {
  file_id: number
  folder_id: number
  target_folder_id: number
}

export type DeleteFileRequest = {
  file_id: number
  folder_id: number
}

export type PickupCodeType = 'file' | 'folder'
export type PickupCodeStatus = 'Active' | 'Expired'

export interface CreatePickupCodeRequest {
  code: string
  file_id: number | null
  folder_id: number | null
  type: PickupCodeType
  max_downloads: number
  expire_time: string
}

export interface GetPickupCodeListRequest {
  page: number
  page_size: number
}

export interface PickupCodeItem {
  id: number
  code: string
  file_id: number | null
  folder_id: number | null
  name: string
  type: PickupCodeType
  download: number
  max_download: number
  expire_time: string
  created_at: string
  status: PickupCodeStatus
}

export interface DashboardFileStatItem {
  type: string
  count: number
  size: number
}

export interface DashboardRecentActivityItem {
  id: number
  name: string
  folder_name: string
  file_type: string
  size: number
  updated_at: string
}

export interface DashboardOverviewResponse {
  storage_used: number
  storage_total: number
  storage_left: number
  storage_used_percent: number
  file_stats: DashboardFileStatItem[]
  recent_activities: DashboardRecentActivityItem[]
}
