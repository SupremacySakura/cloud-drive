import type { FileListItem } from '../services/types/file'

export enum FileType {
  Image = 'image',
  Video = 'video',
  Audio = 'audio',
  Document = 'document',
  Other = 'other',
}

export type UploadFileConfig = {
  file_type: FileType
  folder_id?: number
}

export type FileViewMode = 'list' | 'grid'
export type FileSortKey = 'name' | 'size' | 'modified'
export type SortDirection = 'asc' | 'desc'
export type FileFilterKey = 'all' | 'folder' | 'image' | 'video' | 'audio' | 'document' | 'other'
export type PreviewKind = 'image' | 'pdf' | 'video' | 'audio' | 'text' | 'unsupported'

export type BreadcrumbItem = {
  id: number
  name: string
}

export type FileItemKey = `${FileListItem['type']}:${number}`

export type FileDisplayItem = FileListItem & {
  key: FileItemKey
  icon: string
  iconBg: string
  iconFg: string
  typeLabel: string
  lastModifiedText: string
}

export type UploadTaskStatus =
  | 'pending'
  | 'hashing'
  | 'uploading'
  | 'merging'
  | 'success'
  | 'failed'
  | 'canceled'

export type UploadTask = {
  id: string
  file: File
  targetFolderId: number
  relativePath: string
  status: UploadTaskStatus
  percent: number
  message: string | null
}

export type ToastType = 'success' | 'error' | 'info'
