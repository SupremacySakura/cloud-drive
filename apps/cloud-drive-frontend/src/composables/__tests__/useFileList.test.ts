import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, nextTick } from 'vue'
import {
  getListByFolderIDAndUserID,
  getListCountByFolderIDAndUserID,
} from '../../services/apis/file'
import { useFileList } from '../useFileList'

vi.mock('../../services/apis/file', () => ({
  getListByFolderIDAndUserID: vi.fn(),
  getListCountByFolderIDAndUserID: vi.fn(),
}))

const listMock = vi.mocked(getListByFolderIDAndUserID)
const countMock = vi.mocked(getListCountByFolderIDAndUserID)

const deferred = <T>() => {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(done => {
    resolve = done
  })
  return { promise, resolve }
}

describe('useFileList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps file and folder selections separate when numeric ids collide', async () => {
    countMock.mockResolvedValue(2)
    listMock.mockResolvedValue([
      { id: 1, name: 'folder', type: 'folder', file_type: '', size: 0, updated_at: '' },
      { id: 1, name: 'file', type: 'file', file_type: 'other', size: 1, updated_at: '' },
    ])
    const scope = effectScope()
    const browser = scope.run(() => useFileList())!
    await browser.fetchFolder(0)

    browser.toggleOne(browser.sortedFiles.value[0]!, true)
    browser.toggleOne(browser.sortedFiles.value[1]!, true)

    expect(browser.selectedIds.value).toEqual(new Set(['file:1', 'folder:1']))
    expect(browser.allSelected.value).toBe(true)
    scope.stop()
  })

  it('does not let an older request clear the loading state or overwrite newer data', async () => {
    const firstCount = deferred<number>()
    countMock.mockImplementationOnce(() => firstCount.promise).mockResolvedValueOnce(1)
    listMock.mockResolvedValueOnce([
      { id: 2, name: 'new', type: 'file', file_type: 'other', size: 1, updated_at: '' },
    ])
    const scope = effectScope()
    const browser = scope.run(() => useFileList())!

    const first = browser.fetchFolder(1)
    const second = browser.fetchFolder(2)
    await second
    firstCount.resolve(99)
    await first
    await nextTick()

    expect(browser.isLoading.value).toBe(false)
    expect(browser.totalCount.value).toBe(1)
    expect(browser.rawItems.value[0]?.name).toBe('new')
    scope.stop()
  })

  it('clamps the current page before requesting a now shorter folder', async () => {
    countMock.mockResolvedValue(1)
    listMock.mockResolvedValue([])
    const scope = effectScope()
    const browser = scope.run(() => useFileList())!
    browser.page.value = 5

    await browser.fetchFolder(0)

    expect(browser.page.value).toBe(1)
    expect(listMock).toHaveBeenCalledWith(0, 1, 10, expect.any(AbortSignal))
    scope.stop()
  })
})
