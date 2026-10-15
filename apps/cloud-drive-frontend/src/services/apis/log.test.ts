import { beforeEach, describe, expect, it, vi } from 'vitest'
import request from '../request'
import { getOperationLogs } from './log'

vi.mock('../request', () => ({
  default: {
    get: vi.fn(),
  },
}))

const mockedRequest = vi.mocked(request)

describe('log api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns list and total on success', async () => {
    const payload = {
      list: [
        {
          id: 10,
          action: 'upload',
          target_type: 'file',
          target_id: 88,
          target_name: 'hello.txt',
          folder_name: '',
          detail: '{"size":5}',
          created_at: '2026-10-14T12:00:00+08:00',
        },
      ],
      total: 1,
    }
    mockedRequest.get.mockResolvedValueOnce({ data: { code: 0, msg: 'ok', data: payload } })

    const result = await getOperationLogs(2, 20)

    expect(result).toEqual(payload)
    expect(mockedRequest.get).toHaveBeenCalledWith('/api/file/logs', {
      params: { page: 2, page_size: 20 },
      signal: undefined,
    })
  })

  it('returns empty page on business error', async () => {
    mockedRequest.get.mockResolvedValueOnce({
      data: { code: 1004, msg: '服务器错误', data: null },
    })

    const result = await getOperationLogs(1, 20)

    expect(result).toEqual({ list: [], total: 0 })
  })

  it('returns empty page when data is missing', async () => {
    mockedRequest.get.mockResolvedValueOnce({ data: { code: 0, msg: 'ok', data: null } })

    const result = await getOperationLogs(1, 20)

    expect(result).toEqual({ list: [], total: 0 })
  })

  it('forwards abort signal', async () => {
    mockedRequest.get.mockResolvedValueOnce({
      data: { code: 0, msg: 'ok', data: { list: [], total: 0 } },
    })
    const controller = new AbortController()

    await getOperationLogs(1, 20, controller.signal)

    expect(mockedRequest.get).toHaveBeenCalledWith('/api/file/logs', {
      params: { page: 1, page_size: 20 },
      signal: controller.signal,
    })
  })
})
