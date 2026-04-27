# 项目上下文

## 来源

- 根目录 `AGENTS.md`（项目概览、启动命令、架构与端口关系）。
- 前后端子项目 `AGENTS.md`（各子域入口与关键运行信息）。

## 适用范围

- 面向首次进入仓库的开发者或 AI，快速建立“这是什么项目、怎么跑起来、从哪里读起”的全局认知。

## 更新入口

- 当项目技术栈、目录拓扑、端口约定、联调方式发生变更时，优先更新本文。
- 变更涉及子域细节时，同时同步到对应主题文档：见 `frontend-structure.md`、`backend-structure.md`、`project-constraints.md`。

## 项目目标与形态

- Cloud Drive 是一个前后端分离的文件存储应用，核心能力包括登录鉴权、文件上传与管理、分享相关流程。
- 仓库采用 `pnpm workspace` 单仓组织，当前业务代码集中在 `apps/`，`packages/` 预留为共享包目录（当前为空）。

## 技术栈总览

- 前端：Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Axios。
- 后端：Go + Gin + GORM + MySQL + JWT + Swaggo。
- 协作方式：前端开发服务器通过代理将 `/api/*` 转发到后端服务。

## 目录与职责

- `apps/cloud-drive-frontend`：前端 SPA，包含页面、路由、状态管理与 API 调用层。
- `apps/cloud-drive-backend`：后端 HTTP API，包含路由、业务服务、数据访问与鉴权中间件。
- `data/`：运行时文件存储目录（分片与文件内容）。
- `docs/`：文档中心与主题正文。

## 运行与联调关系

- 后端默认端口 `9000`（可通过 `PORT` 环境变量覆盖）。
- 前端开发端口 `5173`，通过 Vite 代理与后端联调。
- 建议启动顺序：先后端，再前端；随后验证登录态检查与文件相关接口是否可达。

## 新成员最短上手路径

1. 在仓库根目录执行 `pnpm install` 安装依赖。
2. 在 `apps/cloud-drive-backend` 执行 `go run cmd/server/main.go` 启动后端。
3. 在 `apps/cloud-drive-frontend` 执行 `pnpm dev` 启动前端。
4. 阅读顺序建议：`project-constraints.md` → `ai-execution.md` → `frontend-structure.md` / `backend-structure.md` → `testing-guide.md`。
