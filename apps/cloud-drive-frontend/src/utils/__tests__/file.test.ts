import { describe, it, expect } from 'vitest'
import {
  detectFileType,
  iconForFile,
  iconForListItem,
  typeLabelForListItem,
  formatTime,
  sanitizeFileName,
} from '../file'
import { FileType } from '../../types/file'

describe('detectFileType', () => {
  it('should detect image files', () => {
    const imageFile = new File([''], 'test.jpg', { type: 'image/jpeg' })
    expect(detectFileType(imageFile)).toBe(FileType.Image)
  })

  it('should detect video files', () => {
    const videoFile = new File([''], 'test.mp4', { type: 'video/mp4' })
    expect(detectFileType(videoFile)).toBe(FileType.Video)
  })

  it('should detect audio files', () => {
    const audioFile = new File([''], 'test.mp3', { type: 'audio/mpeg' })
    expect(detectFileType(audioFile)).toBe(FileType.Audio)
  })

  it('should detect PDF files as documents', () => {
    const pdfFile = new File([''], 'test.pdf', { type: 'application/pdf' })
    expect(detectFileType(pdfFile)).toBe(FileType.Document)
  })

  it('should detect Word files as documents', () => {
    const docFile = new File([''], 'test.docx', {
      type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    })
    expect(detectFileType(docFile)).toBe(FileType.Document)
  })

  it('should detect text files as documents', () => {
    const txtFile = new File([''], 'test.txt', { type: 'text/plain' })
    expect(detectFileType(txtFile)).toBe(FileType.Document)
  })

  it('should detect unknown files as other', () => {
    const unknownFile = new File([''], 'test.xyz', { type: 'application/x-unknown' })
    expect(detectFileType(unknownFile)).toBe(FileType.Other)
  })

  it('should handle files without type', () => {
    const noTypeFile = new File([''], 'test')
    expect(detectFileType(noTypeFile)).toBe(FileType.Other)
  })
})

describe('iconForFile', () => {
  it('should return video icon for video files', () => {
    const videoFile = new File([''], 'test.mp4', { type: 'video/mp4' })
    const result = iconForFile(videoFile)
    expect(result.icon).toBe('material-symbols:movie')
    expect(result.bg).toBe('bg-blue-100')
    expect(result.fg).toBe('text-blue-600')
  })

  it('should return image icon for image files', () => {
    const imageFile = new File([''], 'test.png', { type: 'image/png' })
    const result = iconForFile(imageFile)
    expect(result.icon).toBe('material-symbols:image')
    expect(result.bg).toBe('bg-red-100')
    expect(result.fg).toBe('text-red-600')
  })

  it('should return music icon for audio files', () => {
    const audioFile = new File([''], 'test.mp3', { type: 'audio/mpeg' })
    const result = iconForFile(audioFile)
    expect(result.icon).toBe('material-symbols:music-note')
    expect(result.bg).toBe('bg-purple-100')
    expect(result.fg).toBe('text-purple-600')
  })

  it('should return document icon for document files', () => {
    const pdfFile = new File([''], 'test.pdf', { type: 'application/pdf' })
    const result = iconForFile(pdfFile)
    expect(result.icon).toBe('material-symbols:description')
    expect(result.bg).toBe('bg-primary/10')
    expect(result.fg).toBe('text-primary')
  })

  it('should return default icon for other files', () => {
    const otherFile = new File([''], 'test.xyz', { type: 'application/x-unknown' })
    const result = iconForFile(otherFile)
    expect(result.icon).toBe('material-symbols:description')
  })
})

describe('iconForListItem', () => {
  it('should return folder icon for folders', () => {
    const folder = { type: 'folder' as const, file_type: '' }
    const result = iconForListItem(folder)
    expect(result.icon).toBe('material-symbols:folder')
    expect(result.bg).toBe('bg-orange-100')
    expect(result.fg).toBe('text-orange-600')
  })

  it('should return video icon for video type', () => {
    const video = { type: 'file' as const, file_type: FileType.Video }
    const result = iconForListItem(video)
    expect(result.icon).toBe('material-symbols:movie')
  })

  it('should return image icon for image type', () => {
    const image = { type: 'file' as const, file_type: FileType.Image }
    const result = iconForListItem(image)
    expect(result.icon).toBe('material-symbols:image')
  })

  it('should return music icon for audio type', () => {
    const audio = { type: 'file' as const, file_type: FileType.Audio }
    const result = iconForListItem(audio)
    expect(result.icon).toBe('material-symbols:music-note')
  })

  it('should return document icon for document type', () => {
    const doc = { type: 'file' as const, file_type: FileType.Document }
    const result = iconForListItem(doc)
    expect(result.icon).toBe('material-symbols:description')
  })

  it('should return default icon for other type', () => {
    const other = { type: 'file' as const, file_type: FileType.Other }
    const result = iconForListItem(other)
    expect(result.icon).toBe('material-symbols:description')
  })
})

describe('typeLabelForListItem', () => {
  it('should return Folder for folders', () => {
    const folder = { type: 'folder' as const, file_type: '' }
    expect(typeLabelForListItem(folder)).toBe('Folder')
  })

  it('should return Image for image type', () => {
    const image = { type: 'file' as const, file_type: FileType.Image }
    expect(typeLabelForListItem(image)).toBe('Image')
  })

  it('should return Video for video type', () => {
    const video = { type: 'file' as const, file_type: FileType.Video }
    expect(typeLabelForListItem(video)).toBe('Video')
  })

  it('should return Audio for audio type', () => {
    const audio = { type: 'file' as const, file_type: FileType.Audio }
    expect(typeLabelForListItem(audio)).toBe('Audio')
  })

  it('should return Document for document type', () => {
    const doc = { type: 'file' as const, file_type: FileType.Document }
    expect(typeLabelForListItem(doc)).toBe('Document')
  })

  it('should return file_type value for other types', () => {
    const other = { type: 'file' as const, file_type: 'custom' }
    expect(typeLabelForListItem(other)).toBe('custom')
  })

  it('should return File for undefined file_type', () => {
    const undefinedType = { type: 'file' as const, file_type: undefined as any }
    expect(typeLabelForListItem(undefinedType)).toBe('File')
  })
})

describe('formatTime', () => {
  it('should return dash for empty value', () => {
    expect(formatTime('')).toBe('-')
  })

  it('should return dash for null/undefined', () => {
    expect(formatTime(null as any)).toBe('-')
    expect(formatTime(undefined as any)).toBe('-')
  })

  it('should format valid ISO date', () => {
    const isoDate = '2024-01-15T10:30:00Z'
    const result = formatTime(isoDate)
    expect(result).not.toBe('-')
    expect(result).not.toBe(isoDate)
    // 验证返回的是有效的日期字符串
    expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/)
  })

  it('should return original value for invalid date', () => {
    const invalidDate = 'not-a-date'
    expect(formatTime(invalidDate)).toBe(invalidDate)
  })
})

describe('sanitizeFileName', () => {
  it('should remove control characters', () => {
    const name = 'test\x00file\x1fname\x7f.txt'
    expect(sanitizeFileName(name)).toBe('testfilename.txt')
  })

  it('should return empty string for empty input', () => {
    expect(sanitizeFileName('')).toBe('')
  })

  it('should return empty string for null/undefined', () => {
    expect(sanitizeFileName(null as any)).toBe('')
    expect(sanitizeFileName(undefined as any)).toBe('')
  })

  it('should keep valid characters intact', () => {
    const name = 'normal-file_name.txt'
    expect(sanitizeFileName(name)).toBe(name)
  })

  it('should handle special Unicode characters', () => {
    const name = '文件📄名称.txt'
    expect(sanitizeFileName(name)).toBe(name)
  })
})
