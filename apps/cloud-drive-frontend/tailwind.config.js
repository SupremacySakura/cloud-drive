/**
 * @deprecated 设计令牌已迁移至 Tailwind v4 CSS-first 配置。
 *
 * 自 2026-09 起，本文件不再被引用（src/style.css 已移除 @config），
 * 仅为历史参考保留，后续版本将删除。
 *
 * 旧令牌迁移对照（新代码一律使用 src/style.css @theme 中的令牌）：
 * - theme.extend.colors.primary          → --color-primary
 * - theme.extend.colors.background-light → --color-canvas（deprecated 兼容令牌）
 * - theme.extend.colors.background-dark  → --color-background-dark（暗色令牌后续单独定义）
 * - theme.extend.fontFamily.display      → --font-sans（deprecated 兼容令牌 --font-display）
 * - theme.extend.borderRadius            → 与 v4 默认值一致，不再单独声明
 *
 * 设计规范唯一权威来源：仓库根目录 DESIGN.md。
 */
export default {}
