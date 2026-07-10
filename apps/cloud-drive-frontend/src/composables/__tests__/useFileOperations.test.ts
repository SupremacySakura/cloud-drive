import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, ref } from 'vue'
import {
  deleteFile,
  downloadById,
  getListByFolderIDAndUserID,
  makeDirectory,
  moveFile,
  renameFile,
} from '../../services/apis/file'
import type { FileListItem } from '../../services/types/file'
import type { FileDisplayItem } from '../../types/file'
import { useFileOperations } from '../useFileOperations'

vi.mock('../../services/apis/file', () => ({
  deleteFile: vi.fn(),
  downloadById: vi.fn(),
  getListByFolderIDAndUserID: vi.fn(),
  makeDirectory: vi.fn(),
  moveFile: vi.fn(),
  renameFile: vi.fn(),
}))

const deleteMock = vi.mocked(deleteFile)
const downloadMock = vi.mocked(downloadById)
const getListMock = vi.mocked(getListByFolderIDAndUserID)
const makeDirectoryMock = vi.mocked(makeDirectory)
const moveMock = vi.mocked(moveFile)
const renameMock = vi.mocked(renameFile)

const deferred = <T>() => {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(done => {
    resolve = done
  })
  return { promise, resolve }
}

const displayItem = (
  id: number,
  name: string,
  type: FileDisplayItem['type'] = 'file',
): FileDisplayItem => ({
  id,
  name,
  type,
  file_type: type === 'file' ? 'document' : '',
  size: type === 'file' ? 10 : 0,
  updated_at: '2026-01-01T00:00:00Z',
  key: `${type}:${id}`,
  icon: 'icon',
  iconBg: 'bg',
  iconFg: 'fg',
  typeLabel: type === 'file' ? 'Document' : 'Folder',
  lastModifiedText: 'now',
})

const folderItem = (id: number, name: string): FileListItem => ({
  id,
  name,
  type: 'folder',
  file_type: '',
  size: 0,
  updated_at: '2026-01-01T00:00:00Z',
})

const setup = () => {
  const currentFolderId = ref(9)
  const breadcrumbs = ref([
    { id: 0, name: 'root' },
    { id: 7, name: 'breadcrumb' },
  ])
  const refresh = vi.fn().mockResolvedValue(undefined)
  const notify = vi.fn()
  const scope = effectScope()
  const operations = scope.run(() =>
    useFileOperations({ currentFolderId, breadcrumbs, refresh, notify }),
  )!
  return { breadcrumbs, currentFolderId, notify, operations, refresh, scope }
}

describe('useFileOperations', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    deleteMock.mockResolvedValue(undefined)
    downloadMock.mockResolvedValue({ fileName: 'download', fileSize: 1, contentType: 'text/plain' })
    getListMock.mockResolvedValue([])
    makeDirectoryMock.mockResolvedValue(1)
    moveMock.mockResolvedValue(undefined)
    renameMock.mockResolvedValue(undefined)
  })

  it('uses typed file targets and never renames a breadcrumb for a file id collision', async () => {
    const { breadcrumbs, operations, refresh, scope } = setup()
    const file = displayItem(7, 'old.txt', 'file')

    operations.openRenameModal(file)
    operations.renameName.value = 'new.txt'
    await operations.handleRename()

    expect(renameMock).toHaveBeenCalledWith(
      { file_id: 7, folder_id: 0, name: 'new.txt' },
      expect.any(AbortSignal),
    )
    expect(breadcrumbs.value[1]?.name).toBe('breadcrumb')
    expect(refresh).toHaveBeenCalledWith(9)

    const folder = displayItem(7, 'breadcrumb', 'folder')
    operations.openRenameModal(folder)
    operations.renameName.value = 'renamed-folder'
    await operations.handleRename()

    expect(renameMock).toHaveBeenLastCalledWith(
      { file_id: 0, folder_id: 7, name: 'renamed-folder' },
      expect.any(AbortSignal),
    )
    expect(breadcrumbs.value[1]?.name).toBe('renamed-folder')
    scope.stop()
  })

  it('ignores a stale move-browser response and keeps loading owned by the newest request', async () => {
    const firstList = deferred<FileListItem[]>()
    const secondList = deferred<FileListItem[]>()
    getListMock
      .mockImplementationOnce(() => firstList.promise)
      .mockImplementationOnce(() => secondList.promise)
    const { operations, scope } = setup()

    const opening = operations.openMoveModal(displayItem(3, 'file.txt'))
    const firstSignal = getListMock.mock.calls[0]?.[3]
    const newest = operations.loadMoveBrowserFolders(2)
    secondList.resolve([folderItem(2, 'new-folder')])
    await newest

    expect(operations.isMoveBrowserLoading.value).toBe(false)
    expect(operations.moveBrowserFolders.value.map(item => item.name)).toEqual(['new-folder'])

    firstList.resolve([folderItem(1, 'stale-folder')])
    await opening

    expect(firstSignal?.aborted).toBe(true)
    expect(operations.moveBrowserFolders.value.map(item => item.name)).toEqual(['new-folder'])
    scope.stop()
  })

  it('keeps delete confirmation open after failure and closes it only after success', async () => {
    const { notify, operations, refresh, scope } = setup()
    const file = displayItem(5, 'protected.txt')
    deleteMock.mockRejectedValueOnce(new Error('无权限'))

    operations.handleDeleteFromMenu(file)
    const failed = await operations.confirmDeleteFromModal()

    expect(failed).toBe(false)
    expect(operations.isDeleteConfirmModalOpen.value).toBe(true)
    expect(operations.deleteConfirmTarget.value).toStrictEqual(file)
    expect(operations.deleteError.value).toBe('无权限')
    expect(notify).toHaveBeenCalledWith('无权限', 'error')
    expect(refresh).not.toHaveBeenCalled()

    deleteMock.mockResolvedValueOnce(undefined)
    const succeeded = await operations.confirmDeleteFromModal()

    expect(succeeded).toBe(true)
    expect(deleteMock).toHaveBeenLastCalledWith(
      { file_id: 5, folder_id: 0 },
      expect.any(AbortSignal),
    )
    expect(operations.isDeleteConfirmModalOpen.value).toBe(false)
    expect(operations.deleteConfirmTarget.value).toBeNull()
    expect(refresh).toHaveBeenCalledWith(9)
    scope.stop()
  })

  it('rejects moving a folder into itself and sends a folder-only target for valid moves', async () => {
    const { notify, operations, refresh, scope } = setup()
    const folder = displayItem(4, 'folder', 'folder')

    const rejected = await operations.handleMove(folder, 4)

    expect(rejected).toBe(false)
    expect(moveMock).not.toHaveBeenCalled()
    expect(notify).toHaveBeenCalledWith('不能移动到自身目录', 'error')

    const moved = await operations.handleMove(folder, 8)

    expect(moved).toBe(true)
    expect(moveMock).toHaveBeenCalledWith(
      { file_id: 0, folder_id: 4, target_folder_id: 8 },
      expect.any(AbortSignal),
    )
    expect(refresh).toHaveBeenCalledWith(9)
    scope.stop()
  })

  it('aborts active folder browsing and downloads when its scope is disposed', async () => {
    const pendingList = deferred<FileListItem[]>()
    const pendingDownload = deferred<Awaited<ReturnType<typeof downloadById>>>()
    getListMock.mockImplementationOnce(() => pendingList.promise)
    downloadMock.mockImplementationOnce(() => pendingDownload.promise)
    const { operations, scope } = setup()

    const opening = operations.openMoveModal(displayItem(1, 'file.txt'))
    const downloading = operations.handleDownloadFromMenu(displayItem(2, 'other.txt'))
    const listSignal = getListMock.mock.calls[0]?.[3]
    const downloadSignal = downloadMock.mock.calls[0]?.[3]

    scope.stop()

    expect(listSignal?.aborted).toBe(true)
    expect(downloadSignal?.aborted).toBe(true)
    pendingList.resolve([])
    pendingDownload.resolve({ fileName: 'other.txt', fileSize: 1, contentType: 'text/plain' })
    await opening
    await downloading
  })

  it('creates a folder in the captured current folder and refreshes the current view', async () => {
    const { currentFolderId, operations, refresh, scope } = setup()
    operations.openCreateFolderModal()
    operations.newFolderName.value = '  reports  '

    const created = await operations.handleCreateFolder()

    expect(created).toBe(true)
    expect(makeDirectoryMock).toHaveBeenCalledWith(
      { folder_id: 9, name: 'reports' },
      expect.any(AbortSignal),
    )
    expect(refresh).toHaveBeenCalledWith(9)
    expect(operations.isCreateFolderModalOpen.value).toBe(false)

    currentFolderId.value = 12
    operations.openCreateFolderModal()
    operations.newFolderName.value = 'next'
    await operations.handleCreateFolder()
    expect(makeDirectoryMock).toHaveBeenLastCalledWith(
      { folder_id: 12, name: 'next' },
      expect.any(AbortSignal),
    )
    scope.stop()
  })
})
