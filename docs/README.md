# 文档中心（索引页）

> 本页只做导航与阅读顺序，不承载长篇规则正文。

## 1. 信息架构

`docs` 采用“单层主题 + 统一命名”的最小结构，当前一级主题如下：

- 前端结构：[frontend-structure.md](./frontend-structure.md)
- 后端结构：[backend-structure.md](./backend-structure.md)
- AI 执行方式：[ai-execution.md](./ai-execution.md)
- 项目上下文：[project-context.md](./project-context.md)
- 项目约束：[project-constraints.md](./project-constraints.md)
- 测试提示：[testing-guide.md](./testing-guide.md)
- 文档约定：[CONVENTIONS.md](./CONVENTIONS.md)

## 2. 推荐阅读顺序

1. [project-context.md](./project-context.md)
2. [project-constraints.md](./project-constraints.md)
3. [ai-execution.md](./ai-execution.md)
4. [frontend-structure.md](./frontend-structure.md) / [backend-structure.md](./backend-structure.md)
5. [testing-guide.md](./testing-guide.md)

## 3. 索引页与正文边界

本页（索引）允许包含：

- 主题目录与链接
- 阅读顺序
- 每篇文档一句话职责说明

本页（索引）禁止包含：

- 详细流程步骤
- 完整命令清单
- 可独立成段的规则正文

正文页允许包含：

- 主题定义、背景、流程、示例、命令
- 与该主题直接相关的维护入口

正文页禁止包含：

- 重复粘贴其他主题全文
- 与主题无关的跨域细节

## 4. 主题职责速览

- `frontend-structure.md`：前端目录结构、页面路由、状态与 API 分层。
- `backend-structure.md`：后端分层职责、路由到服务链路、数据模型与存储入口。
- `ai-execution.md`：AI/自动化执行步骤、命令约束、常见执行路径。
- `project-context.md`：项目目标、技术栈、运行方式与关键路径概览。
- `project-constraints.md`：安全、配置、路径、协作方面的硬性约束。
- `testing-guide.md`：测试前置条件、建议顺序、回归检查清单。

## 5. 场景化走读（验证）

- 新成员入门：按“推荐阅读顺序”可在 5 分钟内定位项目目标、启动路径和子项目入口。
- AI 执行任务：可从 `project-constraints.md` + `ai-execution.md` 获得约束与执行模板，再跳转子域结构文档定位代码。
- 测试准备：可从 `testing-guide.md` 获取前置条件、命令与回归顺序，形成闭环验证路径。
