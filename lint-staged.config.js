/**
 * lint-staged 配置文件
 * 在 pre-commit hook 中自动运行
 */

export default {
  // 前端 TS/TSX/Vue 文件：运行 ESLint 自动修复
  'apps/cloud-drive-frontend/**/*.{ts,tsx,vue}': [
    'eslint --fix',
  ],
};
