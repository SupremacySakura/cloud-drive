import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { getOperationLogs } from '../../services/apis/log'
import { OPERATION_LOG_PAGE_SIZE, useOperationLogs } from '../useOperationLogs'
import type { OperationLogItem } from '../../services/types/log'

vi.mock('../../services/apis/log', () => ({
  getOperationLogs: vi.fn(),
}))

const getOperationLogsMock = vi.mocked(getOperationLogs)

const tick = () => new Promise(resolve => setTimeout(resolve, 0))

const deferred = <T>() => {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(done => {
    resolve = done
  })
  return { promise, resolve }
}

const makeLog = (id: number): OperationLogItem => ({
  id,
  action: 'upload',
  target_type: 'file',
  target_id: id,
  target_name: `file-${id}.txt`,
  folder_name: '',
  detail: '{}',
  created_at: '2026-10-14T12:00:00+08:00',
})

const runComposable = () => {
  const scope = effectScope()
  const composable = scope.run(() => useOperationLogs())!
  return { scope, ...composable }
}

describe('useOperationLogs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads the first page and exposes total', async () => {
    getOperationLogsMock.mockResolvedValue({ list: [makeLog(1)], total: 1 })
    const { scope, logs, total, isLoading, hasMore, load } = runComposable()

    const loadPromise = load()
    expect(isLoading.value).toBe(true)
    await loadPromise

    expect(getOperationLogsMock).toHaveBeenCalledWith(
      1,
      OPERATION_LOG_PAGE_SIZE,
      expect.any(AbortSignal),
    )
    expect(logs.value).toHaveLength(1)
    expect(total.value).toBe(1)
    expect(isLoading.value).toBe(false)
    expect(hasMore.value).toBe(false)
    scope.stop()
  })

  it('falls back to an empty page when the request fails', async () => {
    getOperationLogsMock.mockRejectedValue(new Error('network down'))
    const { scope, logs, total, isLoading, load } = runComposable()

    await load()

    expect(logs.value).toEqual([])
    expect(total.value).toBe(0)
    expect(isLoading.value).toBe(false)
    scope.stop()
  })

  it('aborts the previous load so a stale response never overwrites newer data', async () => {
    const stale = deferred<{ list: OperationLogItem[]; total: number }>()
    getOperationLogsMock.mockImplementationOnce(() => stale.promise)
    getOperationLogsMock.mockResolvedValueOnce({ list: [makeLog(2)], total: 1 })
    const { scope, logs, total, isLoading, load } = runComposable()

    const stalePromise = load()
    await load()

    stale.resolve({ list: [makeLog(1)], total: 99 })
    await stalePromise
    await tick()

    expect(logs.value.map(log => log.id)).toEqual([2])
    expect(total.value).toBe(1)
    expect(isLoading.value).toBe(false)
    scope.stop()
  })

  it('appends the next page on loadMore and derives hasMore from total', async () => {
    getOperationLogsMock
      .mockResolvedValueOnce({
        list: Array.from({ length: OPERATION_LOG_PAGE_SIZE }, (_, i) => makeLog(i + 1)),
        total: OPERATION_LOG_PAGE_SIZE + 1,
      })
      .mockResolvedValueOnce({
        list: [makeLog(OPERATION_LOG_PAGE_SIZE + 1)],
        total: OPERATION_LOG_PAGE_SIZE + 1,
      })
    const { scope, logs, hasMore, loadMore, isLoadingMore, load } = runComposable()

    await load()
    expect(hasMore.value).toBe(true)

    const loadingMore = loadMore()
    expect(isLoadingMore.value).toBe(true)
    await loadingMore

    expect(getOperationLogsMock).toHaveBeenLastCalledWith(2, OPERATION_LOG_PAGE_SIZE)
    expect(logs.value).toHaveLength(OPERATION_LOG_PAGE_SIZE + 1)
    expect(hasMore.value).toBe(false)
    expect(isLoadingMore.value).toBe(false)
    scope.stop()
  })

  it('ignores loadMore when everything is loaded or already loading more', async () => {
    getOperationLogsMock.mockResolvedValue({ list: [makeLog(1)], total: 1 })
    const { scope, loadMore, load } = runComposable()

    await load()
    await loadMore()

    expect(getOperationLogsMock).toHaveBeenCalledTimes(1)
    scope.stop()
  })

  it('reset clears state and aborts the in-flight load', async () => {
    const pending = deferred<{ list: OperationLogItem[]; total: number }>()
    getOperationLogsMock.mockImplementation(() => pending.promise)
    const { scope, logs, total, isLoading, reset, load } = runComposable()

    const loadPromise = load()
    await tick()
    reset()

    expect(logs.value).toEqual([])
    expect(total.value).toBe(0)
    expect(isLoading.value).toBe(false)

    pending.resolve({ list: [makeLog(1)], total: 1 })
    await loadPromise
    await tick()

    expect(logs.value).toEqual([])
    expect(isLoading.value).toBe(false)
    scope.stop()
  })

  it('reset during an in-flight loadMore discards the stale page', async () => {
    const stale = deferred<{ list: OperationLogItem[]; total: number }>()
    getOperationLogsMock
      .mockResolvedValueOnce({
        list: Array.from({ length: OPERATION_LOG_PAGE_SIZE }, (_, i) => makeLog(i + 1)),
        total: OPERATION_LOG_PAGE_SIZE * 2,
      })
      .mockImplementationOnce(() => stale.promise)
    const { scope, logs, total, reset, load, loadMore } = runComposable()

    await load()
    const morePromise = loadMore()
    await tick()
    reset()

    stale.resolve({
      list: [makeLog(OPERATION_LOG_PAGE_SIZE + 1)],
      total: OPERATION_LOG_PAGE_SIZE * 2,
    })
    await morePromise
    await tick()

    expect(logs.value).toEqual([])
    expect(total.value).toBe(0)
    scope.stop()
  })

  it('load during an in-flight loadMore aborts the stale page', async () => {
    const stale = deferred<{ list: OperationLogItem[]; total: number }>()
    getOperationLogsMock
      .mockResolvedValueOnce({
        list: Array.from({ length: OPERATION_LOG_PAGE_SIZE }, (_, i) => makeLog(i + 1)),
        total: OPERATION_LOG_PAGE_SIZE * 2,
      })
      .mockImplementationOnce(() => stale.promise)
      .mockResolvedValueOnce({ list: [makeLog(100)], total: 1 })
    const { scope, logs, total, load, loadMore } = runComposable()

    await load()
    const morePromise = loadMore()
    await tick()
    await load()

    stale.resolve({
      list: [makeLog(OPERATION_LOG_PAGE_SIZE + 1)],
      total: OPERATION_LOG_PAGE_SIZE * 2,
    })
    await morePromise
    await tick()

    expect(logs.value.map(log => log.id)).toEqual([100])
    expect(total.value).toBe(1)
    scope.stop()
  })
})
