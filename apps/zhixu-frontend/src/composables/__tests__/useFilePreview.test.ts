import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import {
  createPublicShareLink,
  deletePublicShareLink,
  getPublicShareLink,
  previewFileById,
} from '../../services/apis/file'
import type { FileDisplayItem } from '../../types/file'
import { useFilePreview } from '../useFilePreview'

vi.mock('../../services/apis/file', () => ({
  createPublicShareLink: vi.fn(),
  deletePublicShareLink: vi.fn(),
  getPublicShareLink: vi.fn(),
  previewFileById: vi.fn(),
}))

const createShareMock = vi.mocked(createPublicShareLink)
const deleteShareMock = vi.mocked(deletePublicShareLink)
const getShareMock = vi.mocked(getPublicShareLink)
const previewMock = vi.mocked(previewFileById)

const deferred = <T>() => {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((done, fail) => {
    resolve = done
    reject = fail
  })
  return { promise, resolve, reject }
}

const fileItem = (id: number, name: string): FileDisplayItem => ({
  id,
  name,
  type: 'file',
  file_type: 'document',
  size: 10,
  updated_at: '2026-01-01T00:00:00Z',
  key: `file:${id}`,
  icon: 'icon',
  iconBg: 'bg',
  iconFg: 'fg',
  typeLabel: 'Document',
  lastModifiedText: 'now',
})

describe('useFilePreview', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(URL, 'createObjectURL').mockImplementation(blob =>
      blob instanceof Blob ? `blob:${blob.size}` : 'blob:media-source',
    )
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('keeps the newest preview and share response when requests resolve out of order', async () => {
    const firstPreview = deferred<Awaited<ReturnType<typeof previewFileById>>>()
    const secondPreview = deferred<Awaited<ReturnType<typeof previewFileById>>>()
    const firstShare = deferred<Awaited<ReturnType<typeof getPublicShareLink>>>()
    const secondShare = deferred<Awaited<ReturnType<typeof getPublicShareLink>>>()
    previewMock
      .mockImplementationOnce(() => firstPreview.promise)
      .mockImplementationOnce(() => secondPreview.promise)
    getShareMock
      .mockImplementationOnce(() => firstShare.promise)
      .mockImplementationOnce(() => secondShare.promise)

    const scope = effectScope()
    const preview = scope.run(() => useFilePreview())!
    const firstRun = preview.openPreviewModal(fileItem(1, 'old.pdf'))
    const firstPreviewSignal = previewMock.mock.calls[0]?.[1]
    const firstShareSignal = getShareMock.mock.calls[0]?.[1]
    const secondRun = preview.openPreviewModal(fileItem(2, 'new.pdf'))

    const newBlob = new Blob(['new'], { type: 'application/pdf' })
    secondPreview.resolve({
      blob: newBlob,
      contentType: 'application/pdf',
      fileName: 'new.pdf',
      fileSize: newBlob.size,
    })
    secondShare.resolve({ exists: true, token: 'new', url: 'https://share/new' })
    await secondRun
    await Promise.resolve()

    const oldBlob = new Blob(['old'], { type: 'application/pdf' })
    firstPreview.resolve({
      blob: oldBlob,
      contentType: 'application/pdf',
      fileName: 'old.pdf',
      fileSize: oldBlob.size,
    })
    firstShare.resolve({ exists: true, token: 'old', url: 'https://share/old' })
    await firstRun
    await Promise.resolve()

    expect(firstPreviewSignal?.aborted).toBe(true)
    expect(firstShareSignal?.aborted).toBe(true)
    expect(preview.previewingFile.value?.id).toBe(2)
    expect(preview.previewBlob.value?.size).toBe(newBlob.size)
    expect(preview.previewBlob.value?.type).toBe(newBlob.type)
    expect(preview.previewKind.value).toBe('pdf')
    expect(preview.publicShareLink.value).toBe('https://share/new')
    expect(preview.previewLoading.value).toBe(false)
    expect(preview.isLoadingShareLink.value).toBe(false)
    expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
    scope.stop()
  })

  it('aborts in-flight work on close and does not create a late object URL', async () => {
    const pendingPreview = deferred<Awaited<ReturnType<typeof previewFileById>>>()
    const pendingShare = deferred<Awaited<ReturnType<typeof getPublicShareLink>>>()
    previewMock.mockImplementationOnce(() => pendingPreview.promise)
    getShareMock.mockImplementationOnce(() => pendingShare.promise)

    const scope = effectScope()
    const preview = scope.run(() => useFilePreview())!
    const running = preview.openPreviewModal(fileItem(1, 'late.txt'))
    const previewSignal = previewMock.mock.calls[0]?.[1]
    const shareSignal = getShareMock.mock.calls[0]?.[1]

    preview.closePreviewModal()
    const blob = new Blob(['late'], { type: 'text/plain' })
    pendingPreview.resolve({
      blob,
      contentType: 'text/plain',
      fileName: 'late.txt',
      fileSize: blob.size,
    })
    pendingShare.resolve({ exists: true, token: 'late', url: 'https://share/late' })
    await running
    await Promise.resolve()

    expect(previewSignal?.aborted).toBe(true)
    expect(shareSignal?.aborted).toBe(true)
    expect(preview.isPreviewModalOpen.value).toBe(false)
    expect(preview.previewBlob.value).toBeNull()
    expect(preview.publicShareLink.value).toBe('')
    expect(URL.createObjectURL).not.toHaveBeenCalled()
    scope.stop()
  })

  it('releases the previous URL on file switch and the current URL on scope disposal', async () => {
    getShareMock.mockResolvedValue({ exists: false, token: '', url: '' })
    const firstBlob = new Blob(['one'], { type: 'image/png' })
    const secondBlob = new Blob(['second'], { type: 'image/png' })
    previewMock
      .mockResolvedValueOnce({
        blob: firstBlob,
        contentType: 'image/png',
        fileName: 'one.png',
        fileSize: firstBlob.size,
      })
      .mockResolvedValueOnce({
        blob: secondBlob,
        contentType: 'image/png',
        fileName: 'two.png',
        fileSize: secondBlob.size,
      })

    const scope = effectScope()
    const preview = scope.run(() => useFilePreview())!
    await preview.openPreviewModal(fileItem(1, 'one.png'))
    await preview.openPreviewModal(fileItem(2, 'two.png'))

    expect(URL.revokeObjectURL).toHaveBeenCalledWith(`blob:${firstBlob.size}`)
    scope.stop()
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(`blob:${secondBlob.size}`)
  })

  it('preserves an existing public link when deletion fails and reports the error', async () => {
    const notify = vi.fn()
    const blob = new Blob(['file'], { type: 'application/pdf' })
    previewMock.mockResolvedValue({
      blob,
      contentType: 'application/pdf',
      fileName: 'file.pdf',
      fileSize: blob.size,
    })
    getShareMock.mockResolvedValue({
      exists: true,
      token: 'token',
      url: 'https://share/existing',
    })
    deleteShareMock.mockRejectedValue(new Error('删除被拒绝'))
    createShareMock.mockResolvedValue({ token: 'unused', url: 'https://share/unused' })

    const scope = effectScope()
    const preview = scope.run(() => useFilePreview({ notify }))!
    await preview.openPreviewModal(fileItem(1, 'file.pdf'))
    await Promise.resolve()
    expect(preview.publicShareLink.value).toBe('https://share/existing')

    await preview.removePublicShareLink()

    expect(preview.publicShareLink.value).toBe('https://share/existing')
    expect(preview.shareError.value).toBe('删除被拒绝')
    expect(notify).toHaveBeenCalledWith('删除被拒绝', 'error')
    scope.stop()
  })
})
