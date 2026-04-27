# 前端结构说明

## 来源

- `apps/cloud-drive-frontend/AGENTS.md`（入口、路由、状态、接口约定）。
- 根目录 `AGENTS.md`（前后端联调关系、通用命令约定）。

## 适用范围

- 适用于 `apps/cloud-drive-frontend` 目录结构、页面路由、状态管理与请求层协作。

## 更新入口

- 当页面结构、路由配置、鉴权流程、请求拦截器或前端构建命令变更时更新本文。
- 接口契约变化需同时检查 `backend-structure.md` 与 `project-context.md` 的对应描述。

## 目录结构与职责

- `src/main.ts`：应用启动入口，注册 Pinia、路由及持久化插件。
- `src/router/`：路由定义与导航守卫。
- `src/pages/`：页面级组件（登录、首页、文件管理、上传、分享等）。
- `src/stores/`：Pinia 状态管理，当前以用户登录态为核心。
- `src/services/apis/`：按业务域拆分 API 封装。
- `src/services/request.ts`：Axios 实例与请求/响应拦截器。

## 路由与页面映射

- 公共路由：`/login`、`/register`、`/require-login`。
- 受保护主路由：`/home` 及其多数子路由。
- 特例：`/home/pickup-codes` 为免登录访问路由。
- 鉴权守卫在进入受保护页面前调用登录态检查接口，失败后重定向到登录提示页。

## 状态与请求层分工

- Pinia（`src/stores/user.ts`）负责保存用户状态与 token，并持久化到本地存储。
- 请求层（`src/services/request.ts`）负责统一注入 `Authorization` 请求头与处理通用响应逻辑。
- 业务 API 文件只关心参数与返回结构，不重复处理 token 拼接逻辑。

## UI 与样式约定

- 样式体系基于 Tailwind，主色值为 `#10b674`。
- 暗色模式采用 class 方案（根节点 `dark` 类控制）。
- 页面优先使用组合式 API 与 TypeScript 严格模式，减少隐式类型风险。
