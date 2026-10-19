import { describe, it, expect } from 'vitest'
import { formatBytes } from '../file'

describe('formatBytes', () => {
  it('should format 0 bytes', () => {
    expect(formatBytes(0)).toBe('0 B')
  })

  it('should format bytes less than 1KB', () => {
    expect(formatBytes(512)).toBe('512 B')
  })

  it('should format exactly 1KB', () => {
    expect(formatBytes(1024)).toBe('1.0 KB')
  })

  it('should format kilobytes', () => {
    expect(formatBytes(1536)).toBe('1.5 KB')
  })

  it('should format megabytes', () => {
    expect(formatBytes(1048576)).toBe('1.0 MB')
  })

  it('should format gigabytes', () => {
    expect(formatBytes(1073741824)).toBe('1.0 GB')
  })

  it('should handle non-finite values', () => {
    expect(formatBytes(NaN)).toBe('0 B')
    expect(formatBytes(Infinity)).toBe('0 B')
  })

  it('should handle negative values', () => {
    expect(formatBytes(-100)).toBe('0 B')
  })
})
