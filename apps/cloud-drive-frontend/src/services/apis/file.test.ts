import { beforeEach, describe, expect, it, vi } from 'vitest'
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
