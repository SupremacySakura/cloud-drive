---
task_id: markdown-knowledge-20261019
status: completed
created: 2026-10-19
applicable_to: "Markdown 入库、持久化任务、租约与用户隔离"
---

# Markdown 知识闭环

## 实现

- 独立 `knowledge_items` 和 `knowledge_import_tasks`，启动时 AutoMigrate。
- 唯一来源键及创建事务保证同一用户同一文件只有一个知识条目和任务。
- `/knowledge` 提交/列表、详情、按文件状态查询和失败重试接口，全部使用 JWT。
- 后台队列使用 MySQL 8 行锁、过期租约及 token fencing，支持进程重启恢复。
- UTF-8 Markdown 限量读取、原正文保存、换行和 BOM 规范化；原文件保留，已完成知识不级联删除。
- 前端加入弹窗、任务面板、文件状态、知识列表及安全正文阅读已接入；Mastra 不在本轮修改范围。
- 默认删除文件保留知识，新增可选同步删除；知识列表支持独立删除并保留原文件。任务与条目物理删除，事务及 worker fencing 防止半删除和结果复活。
- 知识状态筛选与文件工具栏共享自定义 `FilterMenu`，替换原生 select。
- API 契约见 [Markdown 知识后端](../markdown-knowledge-backend.md)，界面行为见 [Markdown 知识前端](../markdown-knowledge-frontend.md)。

## 验证范围

- 后端全量测试、静态检查与构建通过；自动化测试包含 sqlmock 事务与临时文件读取。
- 前端全量 142 项测试、Lint、类型检查及生产构建通过；新增删除选项、删除请求、菜单键盘交互回归。
- 全量竞态检测数分钟未输出结果，已停止，不能视为通过。
- 没有连接或修改本机开发、Docker 或生产数据库。真实 MySQL 迁移及 HTTP 联调仍需验证。
- 原设计稿已暂存，本次保留原样，不暂存、不提交本轮代码。
- 本轮删除功能未做真实账号联调；Go 开发进程需手动重启才加载新代码，不能把旧进程响应视为新代码结果。

## 文档差异

- Brain 中旧项目名称及测试现状仍过时，见 startup-isolation-20261019 记录。Brain 和 Skills 只读，本轮未修改。
- 原设计稿对空文件描述“更新文件后重试”；当前文件模块没有覆盖内容接口，因此空文件通常需重新上传有效文件。后端重试仍如实失败，不伪造成功。
