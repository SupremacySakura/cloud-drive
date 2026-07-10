import type { FileListItem } from '../services/types/file'
import {
  FileType,
  type FileDisplayItem,
  type FileFilterKey,
  type FileItemKey,
  type FileSortKey,
  type SortDirection,
  type UploadTaskStatus,
} from '../types/file'
import { formatTime, iconForListItem, sanitizeFileName, typeLabelForListItem } from './file'

export const fileItemKey = (item: Pick<FileListItem, 'id' | 'type'>): FileItemKey => {
  return `${item.type}:${item.id}`
}

export const decorateFileItem = (item: FileListItem): FileDisplayItem => {
  const icon = iconForListItem(item)
  return {
    ...item,
    key: fileItemKey(item),
    name: sanitizeFileName(item.name),
    icon: icon.icon,
    iconBg: icon.bg,
    iconFg: icon.fg,
    typeLabel: typeLabelForListItem(item),
    lastModifiedText: formatTime(item.updated_at),
  }
}

const matchesFilter = (item: FileDisplayItem, filter: FileFilterKey) => {
  if (filter === 'all') return true
  if (filter === 'folder') return item.type === 'folder'
  if (item.type !== 'file') return false
  if (filter === 'other') {
    return ![FileType.Image, FileType.Video, FileType.Audio, FileType.Document].includes(
      item.file_type as FileType,
    )
  }
  return item.file_type === filter
}

export const buildVisibleFileItems = (
  items: FileListItem[],
  options: {
    query: string
    filter: FileFilterKey
    sortKey: FileSortKey
    sortDirection: SortDirection
  },
): FileDisplayItem[] => {
  const query = options.query.trim().toLocaleLowerCase()
  const direction = options.sortDirection === 'asc' ? 1 : -1
  const visible = items
    .map(decorateFileItem)
    .filter(item => !query || item.name.toLocaleLowerCase().includes(query))
    .filter(item => matchesFilter(item, options.filter))

  return visible.sort((a, b) => {
    if (options.sortKey === 'name') return direction * a.name.localeCompare(b.name)
    if (options.sortKey === 'size') return direction * ((a.size ?? 0) - (b.size ?? 0))
    return direction * a.updated_at.localeCompare(b.updated_at)
  })
}

export const toFileTarget = (item: Pick<FileListItem, 'id' | 'type'>) => ({
  file_id: item.type === 'file' ? item.id : 0,
  folder_id: item.type === 'folder' ? item.id : 0,
})

export const ownerInitials = (name: string) => {
  const trimmed = name.trim()
  if (!trimmed) return '?'
  if (trimmed.toLocaleLowerCase() === 'me') return 'ME'
  const parts = trimmed.split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return `${parts[0]?.[0] ?? ''}${parts[1]?.[0] ?? ''}`.toUpperCase()
  return trimmed.slice(0, 2).toUpperCase()
}

export const uploadTaskMeta = (status: UploadTaskStatus) => {
  const active = ['pending', 'hashing', 'uploading', 'merging'].includes(status)
  if (status === 'success')
    return { active, label: '成功', color: 'text-green-500', bar: 'bg-green-500' }
  if (status === 'failed')
    return { active, label: '失败', color: 'text-red-500', bar: 'bg-red-500' }
  if (status === 'canceled')
    return { active, label: '取消', color: 'text-slate-400', bar: 'bg-slate-400' }
  return { active, label: null, color: 'text-primary', bar: 'bg-primary' }
}
