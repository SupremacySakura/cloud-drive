import { afterEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { useToast } from '../useToast'

describe('useToast', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('replaces the active message and disposes its timer with the scope', () => {
    vi.useFakeTimers()
    const scope = effectScope()
    const toast = scope.run(() => useToast(100))!

    toast.displayToast('first')
    toast.displayToast('second', 'success')
    expect(toast.toastMessage.value).toBe('second')
    expect(toast.toastType.value).toBe('success')

    vi.advanceTimersByTime(100)
    expect(toast.showToast.value).toBe(false)

    toast.displayToast('third')
    scope.stop()
    expect(toast.showToast.value).toBe(false)
  })
})
