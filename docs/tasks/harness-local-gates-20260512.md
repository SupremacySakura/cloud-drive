---
task_id: ad-hoc-harness-review-20260512
status: completed
created: 2026-05-12
last_reviewed: 2026-05-12
applicable_to: "本项目本地 harness、提交前检查、构建发布前验证"
---

# 本地 Harness 门禁修复

## 结果

- 修复 `FileManagement.vue` 模板事件表达式导致的前端生产构建失败。
- 将根目录 `pnpm build` 调整为真实前端构建入口，避免裸 `tsc -b` 对 Vue SFC 解析失败。
- 将根目录 `pnpm lint` 覆盖范围扩展到 `.vue` 文件。
- 为 `pre-commit` 增加失败即退出，并改用本地 `pnpm exec lint-staged`。
- 用户明确不需要 CI，因此未添加 GitHub Actions CI；同时移除了 README 中指向不存在 workflow 的 CI badge。

## 教训

- 只跑 typecheck/test 不等于可以发布，Vue 模板表达式可能在生产构建阶段才暴露解析问题。
- 本地 harness 的根命令必须和真实发布链路一致，否则会形成“绿灯错觉”。
- Hook 脚本必须显式 `set -e`，否则中间检查失败后仍可能因为最后一条命令成功而放行提交。
- README、证据文件和实际 workflow 容易漂移；当项目明确不采用 CI 时，应避免保留 CI badge 或 CI 承诺。

## 可沉淀模式

- 对前端发布风险，最小验证集应包含 `pnpm lint`、`pnpm typecheck`、`pnpm test`、`pnpm build`。
- 对全栈仓库，本地根脚本应提供一条 `pnpm verify` 聚合关键检查，CI 可选但本地门禁不可虚设。
- 对 Husky hooks，优先调用项目本地依赖，例如 `pnpm exec lint-staged`，避免 `npx` 带来的版本和网络漂移。
