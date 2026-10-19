import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import request from '../request'
import { FileType } from '../../types/file'
import { getPublicShareLink, uploadFile } from './file'

vi.mock('../request', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    delete: vi.fn(),
  },
}))

vi.mock('../../utils/hash', () => ({
  calculateHash: vi.fn(async () => 'hash'),
}))

const mockedRequest = vi.mocked(request)

describe('file api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.history.pushState({}, '', '/home')
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('throws when upload init returns a business error', async () => {
    mockedRequest.post.mockResolvedValueOnce({
      data: {
        code: 1004,
        msg: '初始化失败',
        data: null,
      },
    })
    const onProgress = vi.fn()

    await expect(
      uploadFile(
        new File(['hello'], 'hello.txt', { type: 'text/plain' }),
        { file_type: FileType.Document, folder_id: 0 },
        onProgress,
      ),
    ).rejects.toThrow('初始化失败')
    expect(onProgress).toHaveBeenLastCalledWith(0)
  })

  it('retries an idempotent merge while another request is still merging', async () => {
    vi.useFakeTimers()
    mockedRequest.post
      .mockResolvedValueOnce({
        data: {
          code: 0,
          msg: 'success',
          data: { task_id: 11, uploaded_chunks: [], status: 'uploading' },
        },
      })
      .mockResolvedValueOnce({ data: { code: 0, msg: 'success', data: { reused: false } } })
      .mockResolvedValueOnce({
        status: 202,
        data: { code: 0, msg: 'success', data: { status: 'merging', file_id: null } },
      })
      .mockResolvedValueOnce({
        status: 200,
        data: { code: 0, msg: 'success', data: { status: 'completed', file_id: 23 } },
      })

    const onProgress = vi.fn()
    const uploading = uploadFile(
      new File(['hello'], 'hello.txt', { type: 'text/plain' }),
      { file_type: FileType.Document, folder_id: 0 },
      onProgress,
    )

    await vi.advanceTimersByTimeAsync(1_000)
    await uploading

    const mergeCalls = mockedRequest.post.mock.calls.filter(([url]) => url === '/api/file/merge')
    expect(mergeCalls).toHaveLength(2)
    expect(onProgress).toHaveBeenLastCalledWith(100)
  })

  it('continues polling when init returns a task that is already merging', async () => {
    vi.useFakeTimers()
    mockedRequest.post
      .mockResolvedValueOnce({
        data: {
          code: 0,
          msg: 'success',
          data: { task_id: 11, uploaded_chunks: [0], status: 'merging' },
        },
      })
      .mockResolvedValueOnce({
        status: 202,
        data: { code: 0, msg: 'success', data: { status: 'merging', file_id: null } },
      })
      .mockResolvedValueOnce({
        status: 200,
        data: { code: 0, msg: 'success', data: { status: 'completed', file_id: 23 } },
      })

    const onProgress = vi.fn()
    const uploading = uploadFile(
      new File(['hello'], 'hello.txt', { type: 'text/plain' }),
      { file_type: FileType.Document, folder_id: 0 },
      onProgress,
    )

    await vi.advanceTimersByTimeAsync(1_000)
    await uploading

    const chunkCalls = mockedRequest.post.mock.calls.filter(([url]) => url === '/api/file/chunk')
    const mergeCalls = mockedRequest.post.mock.calls.filter(([url]) => url === '/api/file/merge')
    expect(chunkCalls).toHaveLength(0)
    expect(mergeCalls).toHaveLength(2)
    expect(onProgress).toHaveBeenLastCalledWith(100)
  })

  it('disables the global transport timeout for a synchronous merge request', async () => {
    mockedRequest.post
      .mockResolvedValueOnce({
        data: {
          code: 0,
          msg: 'success',
          data: { task_id: 11, uploaded_chunks: [0], status: 'merging' },
        },
      })
      .mockResolvedValueOnce({
        status: 200,
        data: { code: 0, msg: 'success', data: { status: 'completed', file_id: 23 } },
      })

    await uploadFile(
      new File(['hello'], 'hello.txt', { type: 'text/plain' }),
      { file_type: FileType.Document, folder_id: 0 },
      vi.fn(),
    )

    const mergeCall = mockedRequest.post.mock.calls.find(([url]) => url === '/api/file/merge')
    expect(mergeCall?.[2]).toMatchObject({ timeout: 0 })
  })

  it('aborts a merge request after the total merge wait limit', async () => {
    vi.useFakeTimers()
    let hasMergeSignal = false
    mockedRequest.post
      .mockResolvedValueOnce({
        data: {
          code: 0,
          msg: 'success',
          data: { task_id: 11, uploaded_chunks: [0], status: 'merging' },
        },
      })
      .mockImplementationOnce((_url, _data, config) => {
        hasMergeSignal = config?.signal !== undefined
        return new Promise((_resolve, reject) => {
          config?.signal?.addEventListener?.(
            'abort',
            () => reject(new Error('transport aborted')),
            { once: true },
          )
        })
      })

    const uploading = uploadFile(
      new File(['hello'], 'hello.txt', { type: 'text/plain' }),
      { file_type: FileType.Document, folder_id: 0 },
      vi.fn(),
    )
    const outcome = uploading.then(
      () => 'resolved',
      error => `rejected:${error instanceof Error ? error.message : String(error)}`,
    )

    await vi.advanceTimersByTimeAsync(0)
    expect(hasMergeSignal).toBe(true)
    await vi.advanceTimersByTimeAsync(10 * 60 * 1_000)

    expect(await Promise.race([outcome, Promise.resolve('pending')])).toBe(
      'rejected:文件合并超时，请稍后重试',
    )
  })

  it('cancels immediately while waiting to retry a merge', async () => {
    vi.useFakeTimers()
    const controller = new AbortController()
    mockedRequest.post
      .mockResolvedValueOnce({
        data: {
          code: 0,
          msg: 'success',
          data: { task_id: 11, uploaded_chunks: [0], status: 'merging' },
        },
      })
      .mockResolvedValueOnce({
        status: 202,
        data: { code: 0, msg: 'success', data: { status: 'merging', file_id: null } },
      })

    const uploading = uploadFile(
      new File(['hello'], 'hello.txt', { type: 'text/plain' }),
      { file_type: FileType.Document, folder_id: 0 },
      vi.fn(),
      controller.signal,
    )

    await vi.advanceTimersByTimeAsync(0)
    controller.abort()

    await expect(uploading).rejects.toThrow('上传已取消')
    const mergeCalls = mockedRequest.post.mock.calls.filter(([url]) => url === '/api/file/merge')
    expect(mergeCalls).toHaveLength(1)
  })

  it('normalizes existing public share links to the frontend api route', async () => {
    mockedRequest.get.mockResolvedValueOnce({
      data: {
        code: 0,
        msg: 'success',
        data: {
          exists: true,
          token: 'abc 123',
          url: 'http://backend:9000/file/share/open?token=abc%20123',
        },
      },
    })

    const data = await getPublicShareLink(7)

    expect(data.url).toBe('http://localhost:3000/api/file/share/open?token=abc%20123')
  })
})
