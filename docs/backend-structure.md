# 后端结构说明

## 来源

- `apps/cloud-drive-backend/AGENTS.md`（入口、分层、鉴权、存储与命令）。
- 根目录 `AGENTS.md`（整体架构、后端端口与关键风险提示）。

## 适用范围

- 适用于 `apps/cloud-drive-backend` 的目录结构、分层链路、鉴权流程与数据存储入口说明。

## 更新入口

- 当后端目录分层、路由注册方式、中间件策略、模型结构或存储路径变更时更新本文。
- 接口形态变化需同步检查前端调用层说明，见 `frontend-structure.md`。

## 启动入口与目录骨架

- 启动入口：`cmd/server/main.go`，负责初始化数据库、注册路由并启动 HTTP 服务。
- 核心目录：
- `internal/router/`：路由注册与模块入口。
- `internal/handler/`：HTTP 层参数解析与响应封装。
- `internal/service/`：业务逻辑编排。
- `internal/repository/`：数据访问层。
- `internal/model/`：GORM 模型定义。
- `internal/middleware/`：鉴权、跨域等中间件。

## 分层调用链路

- 典型链路为 `router -> handler -> service -> repository -> database`。
- `handler` 不承载复杂业务规则，主要负责输入输出与错误映射。
- `service` 聚焦业务流程与领域规则，组合多个仓储操作。
- `repository` 聚焦数据库读写与查询组织，避免泄漏 HTTP 语义。

## 认证与文件模块主流程

- 认证模块：登录/注册接口由 `auth` 相关 handler 暴露，JWT 校验由中间件完成。
- 受保护接口从 `Authorization: Bearer <token>` 读取 token，通过中间件解析后注入上下文。
- 文件模块：由路由入口接收上传或管理请求，经 service 层处理分片/元数据逻辑，最终落盘到存储路径并写入数据库记录。

## 数据库与文件存储关键路径

- 数据库连接初始化位于 `internal/database/db.go`，当前配置需逐步迁移到环境变量。
- 文件存储路径在路由初始化相关逻辑中存在硬编码，当前默认落到仓库 `data/`。
- 生产化建议：将 DSN、JWT 密钥、存储路径统一改为环境变量注入，避免机器绑定。
