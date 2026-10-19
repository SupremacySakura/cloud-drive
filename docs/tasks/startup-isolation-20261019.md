---
task_id: startup-isolation-20261019
status: completed
created: 2026-10-19
applicable_to: "本地开发、Docker、配置隔离、数据库升级"
---

# 启动隔离与审查修复

## 实现与约束

- 本地 `pnpm dev` 使用 `.env.development`，默认端口为 Vite 5173、后端 9000、本机 MySQL 3306；所有服务均为本机服务，无 Docker 依赖。文件使用 `data/dev/files`，数据库数据由本机 MySQL 管理。
- 全栈 Docker 使用 `.env`，新配置默认端口为前端 8080、后端宿主机 9001、MySQL 3307。现有环境变量覆盖示例默认值。
- 不自动覆盖已有生产配置，不清空数据库，不删除文件卷。开发与 Docker 数据库名均为 `zhixu`，通过不同 MySQL 实例和端口隔离；历史库仅允许通过显式配置临时兼容。
- 后端表迁移完成后幂等创建查询索引，而非依赖只在 MySQL 首次启动执行的初始化 SQL。
- 配置与使用步骤的权威入口为根 README。

## 文档差异记录（待人工审核）

- `docs/brain/project.md` 和 `tech-stack.md` 仍使用旧目录名称；实际目录为 `apps/zhixu-*`。
- `docs/brain/testing.md` 的测试现状已过时，目前已有前后端自动化测试。
- 上述 Brain 文件及约束文件对 AI 只读，本次未直接修改。新的配置隔离、端口与升级约束记录于本任务及 README，需人工审核同步。

## 验证

- `pnpm lint`、`pnpm typecheck`、`pnpm build`、`pnpm test`（128 项）及 `pnpm test:backend` 全部通过。
- 全栈 Compose 配置解析通过；未显式指定库名时统一指向 `zhixu`。
- 早期验证使用开发 MySQL 容器，已按需求纠正为纯本机模式并删除开发 Compose 文件；该次验证不代表本机 MySQL 联调完成。
- 模拟本地后端退出，Vite 同步停止，9000 与 5173 端口全部释放；Docker 的 `8080/api/healthz` 仍正常。
- 本机 Docker `.env` 只把前端端口改为 8080，保留原后端 9001、MySQL 3307 和全部其他配置；未清空旧数据或重建后端。
- 未执行生产数据库更名或真实旧库迁移。Docker 后端仍是原镜像，新增迁移逻辑会在下次构建部署时生效。
- 纯本机模式：开发脚本语法和无 Docker 调用检查通过；当前终端没有可用的本机 mysql 客户端，未验证本机数据库完整联调。README 已补充本机安装、启动、建库和账号配置步骤。

## 项目定位补充

- 毕设背景文档位于 `/Users/shi/study/university/毕业设计`：产品设计文档、架构设计方案-v2、需求规格说明书。
- 当前项目名称为知序 `zhixu`，主体是个人知识管理与演化；原网盘仅为历史技术底座，不应作为当前命名或默认配置。
- 保留 Vue、Go、MySQL 底座及现有功能，规划以独立 Mastra AI 服务和 Qdrant 支撑检索问答、知识关联及演化能力。规划功能不应被描述为已实现。
