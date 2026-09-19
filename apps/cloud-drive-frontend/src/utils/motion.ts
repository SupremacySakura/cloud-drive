import type { $Transition } from 'motion-v'

/**
 * 弹簧预设 — 对应 DESIGN.md §2.4（damping / response 映射到 bounce / duration）。
 * bounce=0 ≈ damping 1.0（无回弹）；bounce 0.2 ≈ damping 0.8（动量场景轻微回弹）。
 */
export const springSnappy: $Transition = { type: 'spring', bounce: 0, duration: 0.35 }
export const springBouncy: $Transition = { type: 'spring', bounce: 0.2, duration: 0.35 }

/** 纯淡入淡出：非手势入场/退场与 reduced-motion 降级 */
export const fadeTransition: $Transition = { duration: 0.2, ease: 'easeOut' }
