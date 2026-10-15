import { describe, expect, it } from 'vitest'
import type { OperationLogAction, OperationLogItem } from '../../services/types/log'
import { describeOperationLog } from '../operationLog'

const makeLog = (overrides: Partial<OperationLogItem> = {}): OperationLogItem => ({
  id: 1,
  action: 'upload',
  target_type: 'file',
  target_id: 88,
  target_name: 'hello.txt',
  folder_name: '工作',
  detail: '{}',
  created_at: '2026-10-14T12:00:00+08:00',
  ...overrides,
})

describe('describeOperationLog', () => {
  it('maps upload with folder highlight', () => {
    const d = describeOperationLog(makeLog({ detail: '{"size":5}' }))
    expect(d.icon).toBe('material-symbols:upload-outline-rounded')
    expect(d.colorClass).toBe('bg-primary-tint text-primary')
    expect(d.title).toBe('hello.txt')
    expect(d.subtitle).toBe('上传至')
    expect(d.highlight).toBe('工作')
  })

  it('maps upload_instant', () => {
    const d = describeOperationLog(makeLog({ action: 'upload_instant' }))
    expect(d.subtitle).toBe('秒传至')
    expect(d.colorClass).toBe('bg-info-tint text-info')
    expect(d.highlight).toBe('工作')
  })

  it('maps mkdir without highlight', () => {
    const d = describeOperationLog(
      makeLog({ action: 'mkdir', target_type: 'folder', target_name: '资料' }),
    )
    expect(d.subtitle).toBe('新建文件夹')
    expect(d.highlight).toBe('')
  })

  it('maps rename with old_name highlight', () => {
    const d = describeOperationLog(
      makeLog({ action: 'rename', target_name: 'new.txt', detail: '{"old_name":"old.txt"}' }),
    )
    expect(d.title).toBe('new.txt')
    expect(d.subtitle).toBe('重命名自')
    expect(d.highlight).toBe('old.txt')
  })

  it('maps rename without old_name gracefully', () => {
    const d = describeOperationLog(makeLog({ action: 'rename', detail: '{}' }))
    expect(d.subtitle).toBe('重命名')
    expect(d.highlight).toBe('')
  })

  it('maps move with detail.to and falls back to root', () => {
    const moved = describeOperationLog(
      makeLog({ action: 'move', detail: '{"from":"工作","to":"归档"}' }),
    )
    expect(moved.subtitle).toBe('移动至')
    expect(moved.highlight).toBe('归档')

    const movedToRoot = describeOperationLog(makeLog({ action: 'move', detail: '{"to":""}' }))
    expect(movedToRoot.highlight).toBe('根目录')

    const movedNoDetail = describeOperationLog(makeLog({ action: 'move' }))
    expect(movedNoDetail.highlight).toBe('根目录')
  })

  it('maps delete', () => {
    const d = describeOperationLog(makeLog({ action: 'delete' }))
    expect(d.colorClass).toBe('bg-pink-tint text-pink')
    expect(d.subtitle).toBe('已删除')
    expect(d.highlight).toBe('')
  })

  it('maps download and distinguishes packaged', () => {
    const single = describeOperationLog(makeLog({ action: 'download' }))
    expect(single.subtitle).toBe('下载')

    const packaged = describeOperationLog(
      makeLog({ action: 'download', target_type: 'folder', detail: '{"packaged":true}' }),
    )
    expect(packaged.subtitle).toBe('打包下载')
  })

  it('maps preview', () => {
    const d = describeOperationLog(makeLog({ action: 'preview' }))
    expect(d.subtitle).toBe('预览')
  })

  it('maps share link create and delete', () => {
    const created = describeOperationLog(makeLog({ action: 'share_link_create' }))
    expect(created.subtitle).toBe('创建分享链接')
    expect(created.colorClass).toBe('bg-info-tint text-info')

    const deleted = describeOperationLog(makeLog({ action: 'share_link_delete' }))
    expect(deleted.subtitle).toBe('删除分享链接')
    expect(deleted.colorClass).toBe('bg-pink-tint text-pink')
  })

  it('maps pickup code create and delete with code highlight', () => {
    const created = describeOperationLog(
      makeLog({ action: 'pickup_code_create', detail: '{"code":"ABC123"}' }),
    )
    expect(created.subtitle).toBe('创建取件码')
    expect(created.highlight).toBe('ABC123')

    const deleted = describeOperationLog(
      makeLog({ action: 'pickup_code_delete', detail: '{"code":"ABC123"}' }),
    )
    expect(deleted.subtitle).toBe('删除取件码')
    expect(deleted.highlight).toBe('ABC123')
  })

  it('falls back to placeholder title when target name is empty', () => {
    const fileLog = describeOperationLog(makeLog({ target_name: '' }))
    expect(fileLog.title).toBe('未知文件')

    const folderLog = describeOperationLog(makeLog({ target_name: '', target_type: 'folder' }))
    expect(folderLog.title).toBe('未知文件夹')
  })

  it('falls back to root folder label when folder_name is empty', () => {
    const d = describeOperationLog(makeLog({ folder_name: '' }))
    expect(d.highlight).toBe('根目录')
  })

  it('sanitizes control characters from names', () => {
    const d = describeOperationLog(makeLog({ target_name: 'a\x00b.txt', folder_name: '工\x7f作' }))
    expect(d.title).toBe('ab.txt')
    expect(d.highlight).toBe('工作')
  })

  it.each(['not-json', '123', '"str"', '[1,2]', 'null'])(
    'treats invalid detail %s as empty object',
    raw => {
      const d = describeOperationLog(
        makeLog({ action: 'rename', target_name: 'a.txt', detail: raw }),
      )
      expect(d.subtitle).toBe('重命名')
      expect(d.highlight).toBe('')
    },
  )

  it('falls back to unknown meta for unmapped actions', () => {
    const d = describeOperationLog(makeLog({ action: 'future_action' as OperationLogAction }))
    expect(d.icon).toBe('material-symbols:help-outline-rounded')
    expect(d.title).toBe('hello.txt')
    expect(d.subtitle).toBe('')
  })
})
