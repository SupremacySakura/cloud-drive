---
last_reviewed: 2026-04-30
---

本文为 Brain 层知识文件，AI 工具进入项目时应优先读取 docs/brain/

# 技术栈总览（摘自项目上下文）

- 前端：Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Axios。
- 后端：Go + Gin + GORM + MySQL + JWT + Swaggo。
- 协作方式：前端开发服务器通过代理将 `/api/*` 转发到后端服务。

## AI 执行命令清单（摘自 ai-execution.md）

- 依赖安装：`pnpm install`（仓库根目录）。
- 前端开发：`pnpm dev`（`apps/cloud-drive-frontend`）。
- 前端构建：`pnpm build`（`apps/cloud-drive-frontend`）。
- 根目录校验：`pnpm lint`、`pnpm typecheck`。
- 后端运行：`go run cmd/server/main.go`（`apps/cloud-drive-backend`）。
- 后端测试：`go test ./...`（`apps/cloud-drive-backend`）。

(End of file - total 44 lines)
