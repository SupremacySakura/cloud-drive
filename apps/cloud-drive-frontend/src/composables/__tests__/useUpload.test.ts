import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { getListByFolderIDAndUserID, makeDirectory, uploadFile } from '../../services/apis/file'
import type { UploadTask } from '../../types/file'
import { useUpload } from '../useUpload'

vi.mock('../../services/apis/file', () => ({
  getListByFolderIDAndUserID: vi.fn(),
  makeDirectory: vi.fn(),
  uploadFile: vi.fn(),
}))

const mockedGetList = vi.mocked(getListByFolderIDAndUserID)
const mockedMakeDirectory = vi.mocked(makeDirectory)
const mockedUploadFile = vi.mocked(uploadFile)

let nextTaskId = 0

const createTask = (relativePath = ''): UploadTask => {
  nextTaskId += 1
  return {
    id: `task-${nextTaskId}`,
    file: new File([`file-${nextTaskId}`], `file-${nextTaskId}.txt`, {
      type: 'text/plain',
    }),
    targetFolderId: 0,
    relativePath,
    status: 'pending',
    percent: 0,
    message: '等待中',
  }
}

const deferred = <T>() => {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((done, fail) => {
    resolve = done
    reject = fail
  })
  return { promise, resolve, reject }
}

type UploadGate = {
  signal: AbortSignal
  progress: (value: number) => void
  finish: () => void
}

const installControlledUploads = () => {
  const gates: UploadGate[] = []
  let activeCount = 0
  let maxActiveCount = 0

  mockedUploadFile.mockImplementation((_file, _config, onProgress, signal) => {
    if (!signal) throw new Error('expected an upload AbortSignal')
    activeCount += 1
    maxActiveCount = Math.max(maxActiveCount, activeCount)
    onProgress(0)

    return new Promise<void>((resolve, reject) => {
      let settled = false
      const settle = (callback: () => void) => {
        if (settled) return
        settled = true
        activeCount -= 1
        callback()
      }
      signal.addEventListener('abort', () => settle(() => reject(new Error('aborted'))), {
        once: true,
      })
      gates.push({
        signal,
        progress: onProgress,
        finish: () => settle(resolve),
      })
    })
  })

  return {
    gates,
    getMaxActiveCount: () => maxActiveCount,
  }
}

describe('useUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    nextTaskId = 0
    mockedGetList.mockResolvedValue([])
    mockedMakeDirectory.mockResolvedValue(1)
    mockedUploadFile.mockImplementation(async (_file, _config, onProgress) => {
      onProgress(0)
      onProgress(50)
      onProgress(100)
    })
  })

  it('shares one three-worker limit and one drain summary across concurrent batches', async () => {
    const refresh = vi.fn(async () => undefined)
    const notify = vi.fn()
    const controlled = installControlledUploads()
    const upload = useUpload({ refresh, notify })
    const firstBatch = [createTask(), createTask(), createTask()]
    const secondBatch = [createTask(), createTask(), createTask()]

    const firstCompletion = upload.startUploadTasks(firstBatch)
    const secondCompletion = upload.startUploadTasks(secondBatch)

    await vi.waitFor(() => expect(mockedUploadFile).toHaveBeenCalledTimes(3))
    expect(upload.isUploading.value).toBe(true)
    expect(upload.activeUploadCount.value).toBe(3)
    expect(firstBatch.every(task => task.status === 'hashing')).toBe(true)
    expect(controlled.getMaxActiveCount()).toBe(3)

    controlled.gates[0].progress(50)
    expect(firstBatch[0].status).toBe('uploading')
    controlled.gates[0].progress(100)
    expect(firstBatch[0].status).toBe('merging')

    for (const gate of controlled.gates.slice(0, 3)) gate.finish()
    await vi.waitFor(() => expect(mockedUploadFile).toHaveBeenCalledTimes(6))
    expect(controlled.getMaxActiveCount()).toBe(3)

    for (const gate of controlled.gates.slice(3, 6)) gate.finish()
    await Promise.all([firstCompletion, secondCompletion])

    expect(firstBatch.concat(secondBatch).every(task => task.status === 'success')).toBe(true)
    expect(upload.activeUploadCount.value).toBe(0)
    expect(upload.isUploading.value).toBe(false)
    expect(refresh).toHaveBeenCalledTimes(1)
    expect(notify).toHaveBeenCalledTimes(1)
    expect(notify).toHaveBeenCalledWith('成功上传 6 个文件', 'success')
  })

  it('cancels pending and active tasks, while clearAll aborts every active signal and drains', async () => {
    const refresh = vi.fn(async () => undefined)
    const notify = vi.fn()
    const controlled = installControlledUploads()
    const upload = useUpload({ refresh, notify })
    const tasks = [createTask(), createTask(), createTask(), createTask()]

    const completion = upload.startUploadTasks(tasks)
    await vi.waitFor(() => expect(mockedUploadFile).toHaveBeenCalledTimes(3))

    upload.cancelTask(tasks[3])
    expect(tasks[3].status).toBe('canceled')
    upload.cancelTask(tasks[0])
    expect(controlled.gates[0].signal.aborted).toBe(true)

    upload.clearAllTasks()
    expect(controlled.gates.every(gate => gate.signal.aborted)).toBe(true)
    expect(upload.uploadTasks.value).toEqual([])

    await completion
    await upload.waitForIdle()

    expect(mockedUploadFile).toHaveBeenCalledTimes(3)
    expect(tasks.every(task => task.status === 'canceled')).toBe(true)
    expect(upload.isUploading.value).toBe(false)
    expect(refresh).toHaveBeenCalledTimes(1)
    expect(notify).toHaveBeenCalledTimes(1)
  })

  it('shares a Promise<number> folder lock, propagates failure, and clears the lock for retry', async () => {
    const folderCreation = deferred<number>()
    mockedMakeDirectory.mockReturnValueOnce(folderCreation.promise)
    const upload = useUpload()
    const firstTask = createTask('shared')
    const secondTask = createTask('shared')

    const completion = upload.startUploadTasks([firstTask, secondTask])
    await vi.waitFor(() => expect(mockedMakeDirectory).toHaveBeenCalledTimes(1))
    folderCreation.reject(new Error('directory unavailable'))
    await completion

    expect(firstTask.status).toBe('failed')
    expect(secondTask.status).toBe('failed')
    expect(firstTask.message).toBe('directory unavailable')
    expect(mockedUploadFile).not.toHaveBeenCalled()
    expect(mockedGetList).not.toHaveBeenCalled()

    mockedMakeDirectory.mockResolvedValueOnce(42)
    await upload.retryTask(firstTask)

    expect(mockedMakeDirectory).toHaveBeenCalledTimes(2)
    expect(mockedUploadFile).toHaveBeenCalledTimes(1)
    expect(mockedUploadFile.mock.calls[0]?.[1].folder_id).toBe(42)
    expect(firstTask.status).toBe('success')
  })

  it('puts a failed upload back through the queue when retrying', async () => {
    mockedUploadFile
      .mockRejectedValueOnce(new Error('network unavailable'))
      .mockImplementationOnce(async (_file, _config, onProgress) => {
        onProgress(0)
        onProgress(40)
        onProgress(100)
      })
    const refresh = vi.fn(async () => undefined)
    const upload = useUpload({ refresh })
    const task = createTask()

    await upload.startUploadTasks([task])
    expect(task.status).toBe('failed')
    expect(task.message).toBe('network unavailable')

    await upload.retryTask(task)

    expect(mockedUploadFile).toHaveBeenCalledTimes(2)
    expect(task.status).toBe('success')
    expect(task.percent).toBe(100)
    expect(refresh).toHaveBeenCalledTimes(2)
  })

  it('accepts File arrays and a base folder id for file and folder input', async () => {
    const upload = useUpload()
    const regularFile = new File(['plain'], 'plain.txt', { type: 'text/plain' })
    const nestedFile = new File(['nested'], 'nested.txt', { type: 'text/plain' })
    Object.defineProperty(nestedFile, 'webkitRelativePath', {
      value: 'photos/2026/nested.txt',
    })

    const fileTasks = upload.handleFileInput([regularFile], 7)
    const folderTasks = upload.handleFolderInput([nestedFile], 9)

    expect(fileTasks).toHaveLength(1)
    expect(fileTasks[0]).toMatchObject({ targetFolderId: 7, relativePath: '' })
    expect(folderTasks).toHaveLength(1)
    expect(folderTasks[0]).toMatchObject({
      targetFolderId: 9,
      relativePath: 'photos/2026',
    })

    await upload.waitForIdle()
    expect(fileTasks[0].status).toBe('success')
    expect(folderTasks[0].status).toBe('success')
  })

  it('aborts active uploads when the owning Vue scope is disposed', async () => {
    const refresh = vi.fn(async () => undefined)
    const notify = vi.fn()
    const controlled = installControlledUploads()
    const scope = effectScope()
    const upload = scope.run(() => useUpload({ refresh, notify }))
    if (!upload) throw new Error('expected an upload composable')
    const task = createTask()

    const completion = upload.startUploadTasks([task])
    await vi.waitFor(() => expect(mockedUploadFile).toHaveBeenCalledTimes(1))
    scope.stop()

    expect(controlled.gates[0].signal.aborted).toBe(true)
    await completion
    await vi.waitFor(() => expect(task.status).toBe('canceled'))
    expect(upload.activeUploadCount.value).toBe(0)
    expect(upload.isUploading.value).toBe(false)
    expect(refresh).not.toHaveBeenCalled()
    expect(notify).not.toHaveBeenCalled()
  })
})
