import { describe, expect, it } from 'vitest'
import type { FileListItem } from '../../services/types/file'
import {
  buildVisibleFileItems,
  fileItemKey,
  toFileTarget,
  uploadTaskMeta,
} from '../file-management'

const item = (overrides: Partial<FileListItem>): FileListItem => ({
  id: 1,
  name: 'alpha.txt',
  type: 'file',
  file_type: 'document',
  size: 10,
  updated_at: '2026-01-01T00:00:00Z',
  ...overrides,
})

describe('file-management helpers', () => {
  it('uses entity type as part of the stable item key', () => {
    expect(fileItemKey(item({ id: 1, type: 'file' }))).toBe('file:1')
    expect(fileItemKey(item({ id: 1, type: 'folder' }))).toBe('folder:1')
  })

  it('maps entity targets without allowing file/folder id collisions', () => {
    expect(toFileTarget(item({ id: 7, type: 'file' }))).toEqual({ file_id: 7, folder_id: 0 })
    expect(toFileTarget(item({ id: 7, type: 'folder' }))).toEqual({ file_id: 0, folder_id: 7 })
  })

  it('decorates, searches, filters and sorts one page of entries', () => {
    const items = [
      item({ id: 1, name: 'zeta.png', file_type: 'image', size: 2 }),
      item({ id: 2, name: 'alpha.png', file_type: 'image', size: 1 }),
      item({ id: 3, name: 'notes.txt', file_type: 'document', size: 3 }),
    ]

    const result = buildVisibleFileItems(items, {
      query: '.PNG',
      filter: 'image',
      sortKey: 'size',
      sortDirection: 'asc',
    })

    expect(result.map(entry => entry.key)).toEqual(['file:2', 'file:1'])
  })

  it('groups unknown backend file types under other', () => {
    const result = buildVisibleFileItems([item({ file_type: 'archive' })], {
      query: '',
      filter: 'other',
      sortKey: 'name',
      sortDirection: 'asc',
    })
    expect(result).toHaveLength(1)
  })

  it('exposes upload task presentation from one status mapping', () => {
    expect(uploadTaskMeta('uploading').active).toBe(true)
    expect(uploadTaskMeta('failed')).toMatchObject({ active: false, label: '失败' })
  })
})
