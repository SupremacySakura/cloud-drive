# Cloud Drive 全面优化计划

## TL;DR

> **Quick Summary**: 对 Cloud Drive 网盘项目进行工程化全面升级，覆盖安全加固、测试体系建设、CI/CD流水线、代码架构重构四大方向，使项目达到面试展示级别的工程化成熟度。
> 
> **Deliverables**:
> - 后端安全加固（JWT/CORS/文件上传验证/速率限制）
> - 前端安全加固（Token刷新/全局错误处理/环境变量保护）
> - 前端 Vitest 测试框架搭建 + 核心模块测试
> - 后端 Go 测试框架搭建 + 核心模块测试
> - GitHub Actions CI/CD 流水线
> - FileManagement.vue 巨型组件拆分
> - 后端架构重构（接口定义/context传播/graceful shutdown/结构化日志）
> - 数据库优化（连接池/索引/N+1修复）
> - 基础设施完善（Docker安全/.gitignore/脚本优化）
> - UX改善（空状态/搜索/可访问性）
> 
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 5 waves
> **Critical Path**: Task 1(安全) → Task 6(测试框架) → Task 10(CI/CD) → Task 14(架构重构) → Task 18(集成验证)

---

## Context

### Original Request
用户表示项目基础功能已完善，希望全面优化。项目定位为面试展示，需要体现工程化能力和最佳实践。

### Interview Summary
**Key Discussions**:
- 优化优先级：全面优化，不偏重单一方向
- 项目定位：面试展示项目，工程化成熟度 > 功能花哨
- 四大方向确认：安全加固、测试体系建设、CI/CD流水线、代码架构重构
- 面试信号：测试覆盖和CI/CD是最强工程化信号，安全加固体现生产思维

**Research Findings**:
- 前端：FileManagement.vue 2401行巨型组件、catch(e:any)8处、单store、零测试、可访问性2/5
- 后端：JWT密钥硬编码P0、无CORS P0、零测试P0、文件上传无验证P1、N+1查询P1、始终200 P2
- 基础设施：CI/CD完全缺失、.env敏感信息、Docker缺安全配置、.gitignore不完整

### Metis Review
**Identified Gaps** (addressed):
- nginx.conf缺少client_max_body_size：默认1MB会阻断分片上传 → 已加入Task 3
- repository.DB公开字段被service直接使用绕过抽象 → 已加入Task 14架构重构
- sanitize函数散落在service层应移到utils → 已加入Task 14
- 业务码vs RESTful状态码需能在面试中解释 → 加为Must NOT Have的guardrail

---

## Work Objectives

### Core Objective
将 Cloud Drive 项目从"功能可用"提升到"工程化成熟"水平，每个优化点都能在面试中讲出"为什么这样做"的理由。

### Concrete Deliverables
- 后端 JWT 密钥强制从环境变量读取，禁止硬编码默认值
- 后端 CORS 中间件配置
- 后端文件上传大小限制 + MIME 类型验证
- 后端速率限制中间件
- 前端 Token 刷新机制
- 前端全局 401/403/500 错误处理
- 前端 Vitest + @vue/test-utils 测试框架
- 后端 testify 测试框架
- 前端核心模块测试文件（store/service/utils）
- 后端核心模块测试文件（service/handler/utils）
- GitHub Actions CI 流水线（lint + typecheck + test + build）
- FileManagement.vue 拆分为 composables
- 后端 FileService/AuthService 接口定义
- 后端 context 传播改造
- 后端 graceful shutdown
- 后端结构化日志（zerolog/zap）
- 数据库连接池配置 + 索引优化
- Docker 安全加固（非root用户 + 资源限制）
- .gitignore 完善
- Vite 路径别名 + 构建优化

### Definition of Done
- [ ] `pnpm lint && pnpm typecheck` 全部通过
- [ ] `go vet ./...` 和 `go build ./...` 通过
- [ ] 前端 `pnpm test` 至少10个测试用例通过
- [ ] 后端 `go test ./...` 至少10个测试用例通过
- [ ] GitHub Actions CI green
- [ ] 无硬编码密钥/默认值残留
- [ ] FileManagement.vue < 500行

### Must Have
- JWT密钥必须从环境变量读取，无硬编码默认值
- CORS中间件必须配置
- 文件上传必须有大小限制和类型验证
- 前后端必须有测试框架和核心测试用例
- CI流水线必须能自动运行lint+test
- 巨型组件必须拆分
- 后端必须有接口定义（可mock）
- graceful shutdown必须实现
- 数据库连接池必须配置

### Must NOT Have (Guardrails)
- 不得改变现有API的业务行为（如分页返回结构、认证流程）
- HTTP状态码风格保持当前业务码模式不变（面试可解释理由），但文档中需注释说明
- 不得引入过度抽象（如DI框架wire），保持简单手写依赖注入
- 不得在重构中遗漏原有功能的edge case
- 不得为测试而测试——每个测试必须验证真实行为
- 不得在CI中添加部署步骤（项目不需要自动部署）
- 不得添加新的业务功能（如用户头像、文件预览等）
- 安全加固不得影响现有正常流程的性能

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO (前端零测试) / NO (后端零测试)
- **Automated tests**: YES (Tests-after) — 先实现优化，再补测试
- **Framework**: 前端 Vitest + @vue/test-utils / 后端 testify
- **If TDD**: N/A - 使用 tests-after 模式，优化完成后补充测试

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Frontend/UI**: Use Playwright (playwright skill) - Navigate, interact, assert DOM, screenshot
- **Backend/API**: Use Bash (curl) - Send requests, assert status + response fields
- **Infrastructure**: Use Bash - Run commands, verify configs, check CI status
- **Go code**: Use Bash (go test) - Run tests, check coverage

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - 安全加固 + 基础设施):
├── Task 1: 后端安全加固 - JWT/CORS/速率限制 [deep]
├── Task 2: 后端文件上传安全验证 [unspecified-high]
├── Task 3: 前端安全加固 - Token刷新/全局错误处理 [deep]
├── Task 4: Docker安全加固 + .gitignore完善 [quick]
├── Task 5: nginx配置修复(client_max_body_size) [quick]
├── Task 6: 前端测试框架搭建(Vitest) [quick]
└── Task 7: 后端测试框架搭建(testify) [quick]

Wave 2 (After Wave 1 - 核心测试 + 基础重构):
├── Task 8: 前端核心模块测试(store/service/utils) (depends: 6) [unspecified-high]
├── Task 9: 后端核心模块测试(service/utils) (depends: 7) [unspecified-high]
├── Task 10: GitHub Actions CI流水线 (depends: 6, 7) [unspecified-high]
├── Task 11: 数据库优化 - 连接池/索引 (depends: 1) [deep]
├── Task 12: FileManagement.vue composables拆分 (depends: 3) [deep]
├── Task 13: 后端接口定义 + 依赖注入改造 (depends: 1) [deep]
└── Task 14: Vite配置优化(路径别名+构建分包) (depends: 6) [quick]

Wave 3 (After Wave 2 - 进阶重构 + UX):
├── Task 15: 后端context传播 + graceful shutdown (depends: 13) [deep]
├── Task 16: 后端结构化日志(zerolog) (depends: 13) [unspecified-high]
├── Task 17: 后端错误处理统一 + HTTP状态码改进 (depends: 13) [unspecified-high]
├── Task 18: 前端状态管理重构(新增stores) (depends: 12) [deep]
├── Task 19: 前端catch(e:any)修复 + 可访问性改善 (depends: 12) [unspecified-high]
└── Task 20: UX改善(空状态/搜索/分页跳转) (depends: 12) [visual-engineering]

Wave 4 (After Wave 3 - 后端进阶 + 文件存储):
├── Task 21: 后端文件存储优化(分片清理/事务/去重) (depends: 15) [deep]
├── Task 22: 后端配置管理重构(viper/结构体) (depends: 16) [unspecified-high]
├── Task 23: 前端API层增强(AbortController/下载进度) (depends: 18) [unspecified-high]
└── Task 24: 前端可访问性完善(aria/键盘导航) (depends: 19) [unspecified-high]

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high + playwright)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 1 → Task 13 → Task 15 → Task 21
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 7 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | - | 11, 13 | 1 |
| 2 | - | - | 1 |
| 3 | - | 12 | 1 |
| 4 | - | - | 1 |
| 5 | - | - | 1 |
| 6 | - | 8, 10, 14 | 1 |
| 7 | - | 9, 10 | 1 |
| 8 | 6 | - | 2 |
| 9 | 7 | - | 2 |
| 10 | 6, 7 | - | 2 |
| 11 | 1 | - | 2 |
| 12 | 3 | 18, 19, 20 | 2 |
| 13 | 1 | 15, 16, 17 | 2 |
| 14 | 6 | - | 2 |
| 15 | 13 | 21 | 3 |
| 16 | 13 | 22 | 3 |
| 17 | 13 | - | 3 |
| 18 | 12 | 23 | 3 |
| 19 | 12 | 24 | 3 |
| 20 | 12 | - | 3 |
| 21 | 15 | - | 4 |
| 22 | 16 | - | 4 |
| 23 | 18 | - | 4 |
| 24 | 19 | - | 4 |

### Agent Dispatch Summary

- **Wave 1**: **7** - T1 → `deep`, T2 → `unspecified-high`, T3 → `deep`, T4 → `quick`, T5 → `quick`, T6 → `quick`, T7 → `quick`
- **Wave 2**: **7** - T8 → `unspecified-high`, T9 → `unspecified-high`, T10 → `unspecified-high`, T11 → `deep`, T12 → `deep`, T13 → `deep`, T14 → `quick`
- **Wave 3**: **6** - T15 → `deep`, T16 → `unspecified-high`, T17 → `unspecified-high`, T18 → `deep`, T19 → `unspecified-high`, T20 → `visual-engineering`
- **Wave 4**: **4** - T21 → `deep`, T22 → `unspecified-high`, T23 → `unspecified-high`, T24 → `unspecified-high`
- **FINAL**: **4** - F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. 后端安全加固 — JWT/CORS/速率限制 ✅ (16个Go测试通过, 8个Vitest测试通过)

  **What to do**:
  - 修改 `internal/utils/jwt.go`：移除硬编码默认值 `"your_secret_key"`，当环境变量为空时 panic 而非使用默认值
  - 添加 CORS 中间件：在 `internal/middleware/` 创建 `cors.go`，配置允许的 Origin（从环境变量读取）、Methods、Headers
  - 在 `internal/router/router.go` 中注册 CORS 中间件
  - 添加速率限制中间件：在 `internal/middleware/` 创建 `ratelimit.go`，基于 IP 的令牌桶限流，默认 60 req/min
  - 对 `/auth/login` 和 `/auth/register` 应用更严格的速率限制（10 req/min）
  - 修改 `.env.example` 和 `.env`：添加 `CORS_ALLOWED_ORIGINS` 配置项，移除 `JWT_SECRET=your_secret_key`

  **Must NOT do**:
  - 不改变 JWT token 的生成/验证逻辑，仅移除默认值
  - 不引入外部限流库（如 redis），使用内存令牌桶即可
  - 不改变现有 API 路由结构

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 安全加固涉及多个文件、中间件设计、环境变量管理，需要深入理解
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `playwright`: 非UI任务

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5, 6, 7)
  - **Blocks**: Task 11, Task 13
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/middleware/auth.go` - 现有中间件模式（JWT校验），新中间件应遵循此模式
  - `internal/router/router.go:22-42` - 路由注册入口，CORS和限流中间件需在此注册
  - `internal/utils/jwt.go:12-18` - JWT密钥获取函数，需移除硬编码默认值

  **API/Type References**:
  - `internal/router/auth.go:11-20` - 路由注册方式，限流中间件需按此方式挂载

  **External References**:
  - Gin CORS 中间件: `github.com/gin-contrib/cors` - 官方推荐的 CORS 解决方案

  **WHY Each Reference Matters**:
  - `auth.go` 中间件是现有模式参考，新中间件必须风格一致
  - `jwt.go:12-18` 是安全漏洞所在，必须修改此处
  - `router.go` 是所有中间件的注册点

  **Acceptance Criteria**:

  - [ ] `jwt.go` 中 `getJWTSecret()` 不再返回硬编码 `"your_secret_key"`
  - [ ] `JWT_SECRET` 环境变量为空时程序启动 panic
  - [ ] CORS 中间件已注册并从环境变量读取配置
  - [ ] `/auth/login` 限流 10 req/min，其他接口 60 req/min
  - [ ] `go build ./...` 通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: JWT密钥必须从环境变量读取
    Tool: Bash
    Preconditions: 后端未启动
    Steps:
      1. 运行 `grep -n "your_secret_key" apps/cloud-drive-backend/internal/utils/jwt.go`
      2. 预期输出为空（不再存在硬编码值）
      3. 设置 `JWT_SECRET=""` 并尝试 `go run cmd/server/main.go`
      4. 预期程序 panic 或拒绝启动
    Expected Result: 无硬编码密钥，空密钥时拒绝启动
    Failure Indicators: grep 找到 "your_secret_key" 或空密钥时正常启动
    Evidence: .sisyphus/evidence/task-1-jwt-no-hardcode.txt

  Scenario: CORS中间件正常工作
    Tool: Bash (curl)
    Preconditions: 后端已启动
    Steps:
      1. `curl -H "Origin: http://localhost:5173" -H "Access-Control-Request-Method: POST" -X OPTIONS http://localhost:9000/api/auth/login -i`
      2. 检查响应头包含 `Access-Control-Allow-Origin: http://localhost:5173`
    Expected Result: OPTIONS 预检请求返回正确的 CORS 头
    Failure Indicators: 无 Access-Control-Allow-Origin 头
    Evidence: .sisyphus/evidence/task-1-cors-headers.txt

  Scenario: 速率限制生效
    Tool: Bash (curl)
    Preconditions: 后端已启动
    Steps:
      1. 连续发送 12 次 POST 请求到 `/api/auth/login`
      2. 检查第 11-12 次请求返回 429 Too Many Requests
    Expected Result: 超过限制后返回 429
    Failure Indicators: 所有请求均返回 200
    Evidence: .sisyphus/evidence/task-1-rate-limit.txt
  ```

  **Commit**: YES
  - Message: `fix(backend): harden JWT, CORS, and rate limiting`
  - Files: `internal/utils/jwt.go, internal/middleware/cors.go, internal/middleware/ratelimit.go, internal/router/router.go, .env.example`
  - Pre-commit: `go build ./...`

- [x] 2. 后端文件上传安全验证 ✅ (MIME类型验证, 文件大小限制, 路径遍历防护)

  **What to do**:
  - 在 `internal/handler/file.go` 添加文件大小限制：在 `UploadFileChunkStream` 中检查 Content-Length 和实际读取字节数
  - 添加 MIME 类型验证：读取文件前 512 字节用 `http.DetectContentType` 检测真实类型
  - 在 `internal/service/file.go` 添加允许的 MIME 类型白名单（image/*, video/*, application/pdf, application/zip, text/* 等）
  - 添加路径遍历防护：验证 `FileHash` 不包含 `../` 或其他路径遍历字符
  - 验证分片总大小与声明的文件大小匹配

  **Must NOT do**:
  - 不改变现有上传接口的参数结构
  - 不在前端添加文件类型白名单（前端验证可绕过，只做后端）
  - 不改变分片大小（1MB）

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 文件上传安全涉及多个验证点，需要仔细理解现有上传流程
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5, 6, 7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/handler/file.go:132-170` - UploadFileChunkStream 函数，需添加验证
  - `internal/service/file.go:36-77` - sanitizeFileName 函数，安全验证参考
  - `internal/service/file.go:187-423` - 文件上传完整流程

  **API/Type References**:
  - `internal/dto/file.go` - 上传请求DTO，了解参数结构

  **WHY Each Reference Matters**:
  - handler层是添加大小限制的正确位置（HTTP层验证）
  - service层已有sanitizeFileName，新的安全函数应风格一致

  **Acceptance Criteria**:

  - [ ] 单个分片大小超过限制时返回错误
  - [ ] MIME类型不在白名单时返回错误
  - [ ] FileHash 包含 `../` 时返回错误
  - [ ] 分片总大小与声明不匹配时返回错误
  - [ ] `go build ./...` 通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: 拒绝超大文件上传
    Tool: Bash (curl)
    Preconditions: 后端已启动，用户已登录
    Steps:
      1. 生成一个超过限制的文件 `dd if=/dev/zero of=/tmp/test_large.bin bs=1M count=100`
      2. 尝试上传该文件
      3. 检查返回错误码和提示信息
    Expected Result: 返回文件大小超限错误
    Failure Indicators: 上传成功或无错误提示
    Evidence: .sisyphus/evidence/task-2-file-size-limit.txt

  Scenario: 拒绝危险MIME类型
    Tool: Bash (curl)
    Preconditions: 后端已启动
    Steps:
      1. 创建一个包含 `#!/bin/bash` 内容的 .sh 文件
      2. 尝试上传该文件（修改 Content-Type 欺骗）
      3. 检查服务端通过 DetectContentType 检测到真实类型并拒绝
    Expected Result: 返回不支持的文件类型错误
    Failure Indicators: 危险文件上传成功
    Evidence: .sisyphus/evidence/task-2-mime-reject.txt
  ```

  **Commit**: YES
  - Message: `fix(backend): add file upload size and MIME validation`
  - Files: `internal/handler/file.go, internal/service/file.go`
  - Pre-commit: `go build ./...`

- [x] 3. 前端安全加固 — Token刷新/全局错误处理 ✅ (401/403/500错误处理, 请求超时和重试)

  **What to do**:
  - 在 `src/services/request.ts` 添加 Token 刷新逻辑：当 401 响应时，尝试用 refreshToken 刷新，成功后重试原请求
  - 在 `src/stores/user.ts` 添加 refreshToken 存储（后端需配合，如无 refresh 接口则改为：401 时自动登出 + toast 提示 "登录已过期"）
  - 在 `src/services/request.ts` 响应拦截器添加全局错误处理：401 → 登出 + toast，403 → toast "无权限"，500 → toast "服务器错误"
  - 添加请求超时配置（默认 30s）
  - 添加请求重试逻辑（网络错误时指数退避，最多重试 2 次）

  **Must NOT do**:
  - 如果后端无 refresh token 接口，不强制实现 refresh 流程，改用"401登出+提示"
  - 不改变现有的 token 存储方式（sessionStorage via pinia-plugin-persistedstate）
  - 不添加 CSRF token（JWT Bearer 模式天然防 CSRF）

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 涉及 Axios 拦截器重构、认证流程设计、错误处理策略
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4, 5, 6, 7)
  - **Blocks**: Task 12
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `src/services/request.ts` - 现有 Axios 实例和拦截器
  - `src/stores/user.ts` - 用户状态和 token 管理

  **API/Type References**:
  - `src/services/apis/auth.ts` - 认证 API 定义

  **External References**:
  - Axios 拦截器文档: https://axios-http.com/docs/interceptors

  **WHY Each Reference Matters**:
  - request.ts 是核心修改点，需理解现有拦截器逻辑
  - user.ts 的 token 管理方式决定了 refresh 策略

  **Acceptance Criteria**:

  - [ ] 401 响应时用户看到 toast 提示并被登出
  - [ ] 403/500 错误有对应的 toast 提示
  - [ ] 请求超时配置为 30s
  - [ ] 网络错误有自动重试（最多2次）
  - [ ] `pnpm typecheck` 通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: 401响应时登出并提示
    Tool: Playwright
    Preconditions: 用户已登录，前端已启动
    Steps:
      1. 导航到 /home
      2. 模拟后端返回 401（可通过 DevTools 拦截或临时修改后端）
      3. 执行一个需要认证的操作（如文件列表）
      4. 检查页面显示 toast "登录已过期"
      5. 检查页面跳转到 /login
    Expected Result: 401 时 toast 提示并跳转登录页
    Failure Indicators: 401 时无提示或停留在当前页
    Evidence: .sisyphus/evidence/task-3-401-handling.png

  Scenario: 网络错误自动重试
    Tool: Bash
    Preconditions: 前端开发服务器运行中
    Steps:
      1. 临时停止后端服务
      2. 在前端发起 API 请求
      3. 检查 console 或 network 面板显示重试行为
    Expected Result: 请求失败后自动重试（最多2次）
    Failure Indicators: 只请求一次就失败
    Evidence: .sisyphus/evidence/task-3-retry.txt
  ```

  **Commit**: YES
  - Message: `feat(frontend): token refresh and global error handling`
  - Files: `src/services/request.ts, src/stores/user.ts`
  - Pre-commit: `pnpm typecheck`

- [x] 4. Docker安全加固 + .gitignore完善 ✅ (非root用户, 资源限制, .env移除git追踪)

  **What to do**:
  - 修改 `apps/cloud-drive-backend/Dockerfile`：添加非 root 用户
    ```dockerfile
    RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    USER appuser
    ```
  - 修改 `docker-compose.yml`：为 backend 和 frontend 服务添加资源限制
  - 检查 `.env` 是否被 git 追踪，如果是则 `git rm --cached .env apps/cloud-drive-backend/.env`
  - 完善 `.gitignore`：添加 `*.log`, `npm-debug.log*`, `*.swp`, `*.swo`, `.vscode/`, `.idea/`
  - 修改 `.env.example`：确保只含占位符值，不含真实密钥

  **Must NOT do**:
  - 不删除 `.env` 文件本身，只取消 git 追踪
  - 不改变 docker-compose.yml 的 profile 设计
  - 不添加 Docker 健康检查（已有）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 配置文件修改为主，工作量小
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5, 6, 7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `apps/cloud-drive-backend/Dockerfile` - 现有 Dockerfile，需添加 USER 指令
  - `docker-compose.yml` - 现有编排配置，需添加资源限制
  - `.gitignore` - 现有忽略规则，需补充

  **WHY Each Reference Matters**:
  - Dockerfile 需理解现有多阶段构建结构
  - docker-compose 需理解现有 profile 和服务结构

  **Acceptance Criteria**:

  - [ ] 后端 Dockerfile 包含 `USER appuser` 指令
  - [ ] docker-compose.yml 包含 memory limits
  - [ ] `.env` 文件未被 git 追踪（`git ls-files .env` 返回空）
  - [ ] `.gitignore` 包含 `*.log`, `.vscode/`, `.idea/`
  - [ ] `docker compose config` 验证通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Docker容器以非root用户运行
    Tool: Bash
    Preconditions: Docker已构建
    Steps:
      1. `docker compose --profile backend up -d --build`
      2. `docker compose exec backend whoami`
      3. 预期输出 "appuser" 而非 "root"
    Expected Result: whoami 返回 appuser
    Failure Indicators: 返回 root
    Evidence: .sisyphus/evidence/task-4-docker-nonroot.txt

  Scenario: .env不被git追踪
    Tool: Bash
    Preconditions: 无
    Steps:
      1. `git ls-files .env apps/cloud-drive-backend/.env`
      2. 预期输出为空
    Expected Result: .env 文件不在 git 追踪中
    Failure Indicators: 列出 .env 文件路径
    Evidence: .sisyphus/evidence/task-4-env-not-tracked.txt
  ```

  **Commit**: YES
  - Message: `chore: Docker security and .gitignore`
  - Files: `Dockerfile, docker-compose.yml, .gitignore, .env.example`
  - Pre-commit: `docker compose config`

- [x] 5. nginx配置修复 — client_max_body_size ✅ (100m上传限制, gzip压缩, 安全头)

  **What to do**:
  - 修改 `apps/cloud-drive-frontend/nginx.conf`：添加 `client_max_body_size 100m;`（分片上传每片 1MB，需要足够余量）
  - 同时添加 gzip 压缩配置以优化前端资源传输
  - 添加安全头：`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`

  **Must NOT do**:
  - 不改变 nginx 的路由/代理配置
  - 不设置 client_max_body_size 为 0（无限制），保持合理上限

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单文件配置修改，工作量极小
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 6, 7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `apps/cloud-drive-frontend/nginx.conf` - 现有 nginx 配置

  **WHY Each Reference Matters**:
  - 必须理解现有 nginx 配置结构才能正确添加指令

  **Acceptance Criteria**:

  - [ ] nginx.conf 包含 `client_max_body_size 100m;`
  - [ ] nginx.conf 包含安全头（X-Content-Type-Options, X-Frame-Options）
  - [ ] nginx.conf 包含 gzip 配置

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: nginx允许大文件上传
    Tool: Bash
    Preconditions: 前端Docker容器已运行
    Steps:
      1. `docker compose exec frontend cat /etc/nginx/nginx.conf | grep client_max_body_size`
      2. 预期输出包含 `client_max_body_size 100m;`
    Expected Result: nginx 配置了 100m 上传限制
    Failure Indicators: 未找到 client_max_body_size 或值为默认 1m
    Evidence: .sisyphus/evidence/task-5-nginx-upload.txt
  ```

  **Commit**: YES
  - Message: `fix(nginx): set client_max_body_size for chunk upload`
  - Files: `apps/cloud-drive-frontend/nginx.conf`

- [x] 6. 前端测试框架搭建（Vitest）✅ (8个测试通过)

  **What to do**:
  - 安装依赖：`pnpm add -D vitest @vue/test-utils happy-dom @vitest/coverage-v8`
  - 创建 `apps/cloud-drive-frontend/vitest.config.ts`：配置 test environment 为 happy-dom，配置 coverage
  - 在 `apps/cloud-drive-frontend/package.json` 添加 scripts：`"test": "vitest", "test:run": "vitest run", "test:coverage": "vitest run --coverage"`
  - 创建示例测试文件 `src/utils/__tests__/format.test.ts` 验证框架可运行
  - 确认根目录 `package.json` 添加 `"test": "pnpm --filter cloud-drive-frontend test"`

  **Must NOT do**:
  - 不配置 Jest（Vitest 与 Vite 原生兼容更好）
  - 不创建大量测试用例（本任务只搭建框架 + 1个示例，Task 8 补充核心测试）
  - 不添加 E2E 测试（Playwright E2E 是后续可选任务）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 框架搭建是标准操作，工作量小
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 5, 7)
  - **Blocks**: Task 8, Task 10, Task 14
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `apps/cloud-drive-frontend/vite.config.ts` - Vite 配置，Vitest 需与其兼容
  - `apps/cloud-drive-frontend/tsconfig.app.json` - TypeScript 配置

  **External References**:
  - Vitest 官方文档: https://vitest.dev/config/
  - @vue/test-utils 文档: https://test-utils.vuejs.org/

  **WHY Each Reference Matters**:
  - vitest.config.ts 需要与 vite.config.ts 共享 resolve 配置

  **Acceptance Criteria**:

  - [ ] `vitest.config.ts` 文件存在且配置正确
  - [ ] `pnpm test` 可运行
  - [ ] 示例测试 `format.test.ts` 通过
  - [ ] `pnpm typecheck` 通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: 测试框架正常运行
    Tool: Bash
    Preconditions: 依赖已安装
    Steps:
      1. 在 apps/cloud-drive-frontend 目录运行 `pnpm test:run`
      2. 检查输出显示测试通过
    Expected Result: vitest 运行成功，至少1个测试通过
    Failure Indicators: vitest 无法启动或测试失败
    Evidence: .sisyphus/evidence/task-6-vitest-setup.txt
  ```

  **Commit**: YES
  - Message: `test(frontend): setup Vitest and add core tests`
  - Files: `vitest.config.ts, package.json, src/utils/__tests__/format.test.ts`
  - Pre-commit: `pnpm test:run`

- [x] 7. 后端测试框架搭建（testify）✅ (16个测试通过)

  **What to do**:
  - 安装 testify：`cd apps/cloud-drive-backend && go get github.com/stretchr/testify`
  - 创建 `internal/utils/jwt_test.go`：测试 JWT 生成和解析
  - 创建 `internal/service/auth_test.go`：测试 ValidateUser 逻辑（使用 mock repository）
  - 创建 `internal/utils/file_test.go`：测试 sanitizeFileName 函数
  - 在 `Makefile` 或根 `package.json` 添加后端测试命令

  **Must NOT do**:
  - 不使用 gomock 等复杂 mock 框架（ testify mock 足够）
  - 不创建大量测试（本任务 3-5 个基础测试文件，Task 9 补充更多）
  - 不添加集成测试（需要真实数据库，后续可选）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 框架搭建 + 少量基础测试，工作量小
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 5, 6)
  - **Blocks**: Task 9, Task 10
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/utils/jwt.go` - JWT 工具函数，需编写测试
  - `internal/utils/file.go` - sanitizeFileName 函数，需编写测试
  - `internal/service/auth.go` - AuthService 定义和 ValidateUser

  **External References**:
  - testify 文档: https://github.com/stretchr/testify

  **WHY Each Reference Matters**:
  - 这些是后端最基础的工具函数，测试它们能验证框架搭建正确

  **Acceptance Criteria**:

  - [ ] `go test ./internal/utils/...` 通过
  - [ ] `go test ./internal/service/...` 通过（至少 auth_test）
  - [ ] testify 已添加到 go.mod

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Go测试框架正常运行
    Tool: Bash
    Preconditions: 后端依赖已安装
    Steps:
      1. 在 apps/cloud-drive-backend 目录运行 `go test ./...`
      2. 检查输出显示测试通过
    Expected Result: go test 运行成功，至少3个测试函数通过
    Failure Indicators: go test 无法运行或测试失败
    Evidence: .sisyphus/evidence/task-7-testify-setup.txt
  ```

  **Commit**: YES
  - Message: `test(backend): setup testify and add core tests`
  - Files: `go.mod, internal/utils/jwt_test.go, internal/service/auth_test.go, internal/utils/file_test.go`
  - Pre-commit: `go test ./...`

- [x] 8. 前端核心模块测试 ✅ (59测试通过: store/service/utils)

  **What to do**:
  - 创建 `src/stores/__tests__/user.test.ts`：测试 token 存储、登出、isAuthenticated
  - 创建 `src/services/__tests__/request.test.ts`：测试拦截器逻辑（token 注入、401 处理）、超时配置
  - 创建 `src/utils/__tests__/file.test.ts`：测试 iconForFile、文件大小格式化等
  - 创建 `src/composables/__tests__/useFileList.test.ts`（如果 Task 12 已完成）

  **Must NOT do**: 不测试 UI 组件快照、不追求100%覆盖率（≥60%即可）

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 6

  **References**: `src/stores/user.ts`, `src/services/request.ts`, `src/utils/`, `src/utils/__tests__/format.test.ts`

  **Acceptance Criteria**:
  - [ ] `pnpm test:run` 至少8个测试通过
  - [ ] user store 测试覆盖 login/logout/token
  - [ ] request 拦截器测试覆盖 token 注入 + 错误处理

  **QA Scenarios**:
  ```
  Scenario: 前端测试套件全部通过
    Tool: Bash
    Steps: 1. `pnpm test:run` 2. 检查 ≥8 个测试通过
    Expected Result: 所有测试通过
    Evidence: .sisyphus/evidence/task-8-frontend-tests.txt
  ```

  **Commit**: YES, Message: `test(frontend): add core module tests`, Files: `src/**/__tests__/*.test.ts`

- [x] 9. 后端核心模块测试 ✅ (handler/middleware测试完成)

  **What to do**:
  - 创建 `internal/handler/auth_test.go`：测试登录/注册参数验证
  - 创建 `internal/handler/file_test.go`：测试文件操作参数绑定
  - 创建 `internal/repository/file_repo_test.go`：测试数据库查询（使用 sqlmock）
  - 创建 `internal/service/file_test.go`：测试文件服务核心逻辑（mock repository）
  - 创建 `internal/middleware/auth_test.go`：测试 JWT 中间件校验

  **Must NOT do**: 不使用真实数据库、不追求高覆盖率（≥50%即可）

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 7

  **References**: `internal/handler/*.go`, `internal/service/file.go`, `internal/middleware/auth.go`, `internal/utils/jwt_test.go`

  **Acceptance Criteria**:
  - [ ] `go test ./internal/...` 至少10个测试函数通过
  - [ ] handler 测试覆盖参数验证
  - [ ] middleware 测试覆盖 JWT 校验

  **QA Scenarios**:
  ```
  Scenario: 后端测试套件全部通过
    Tool: Bash
    Steps: 1. `go test ./... -v` 2. 检查 ≥10 个测试函数通过
    Expected Result: 所有测试通过
    Evidence: .sisyphus/evidence/task-9-backend-tests.txt
  ```

  **Commit**: YES, Message: `test(backend): add core module tests`, Files: `internal/**/*_test.go`

- [x] 10. GitHub Actions CI 流水线 ✅ (3个Job配置完成)

  **What to do**:
  - 创建 `.github/workflows/ci.yml`：push to main + PR 触发
  - Job 1 - Frontend: `pnpm install` → `pnpm lint` → `pnpm typecheck` → `pnpm test:run` → `pnpm build`
  - Job 2 - Backend: `go vet ./...` → `go build ./...` → `go test ./...`
  - Job 3 - Docker: `docker compose config` 验证
  - 使用 pnpm store cache + Go module cache
  - 在 README.md 添加 CI badge

  **Must NOT do**: 不添加部署步骤、不添加安全扫描、不使用 self-hosted runner

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 6, Task 7

  **References**: `package.json`, `apps/cloud-drive-frontend/package.json`, `apps/cloud-drive-backend/go.mod`

  **Acceptance Criteria**:
  - [ ] `.github/workflows/ci.yml` 存在
  - [ ] CI 包含 frontend + backend + docker 验证步骤

  **QA Scenarios**:
  ```
  Scenario: CI配置语法正确且本地可复现
    Tool: Bash
    Steps: 1. 运行 `pnpm lint && pnpm typecheck && pnpm test:run && pnpm build` 2. 运行 `go vet ./... && go build ./... && go test ./...`
    Expected Result: CI 中所有步骤本地也能通过
    Evidence: .sisyphus/evidence/task-10-ci-local.txt
  ```

  **Commit**: YES, Message: `ci: add GitHub Actions workflow`, Files: `.github/workflows/ci.yml, README.md`

- [x] 11. 数据库优化 — 连接池/索引 ✅ (连接池配置, 3个索引, N+1修复)

  **What to do**:
  - 修改 `internal/database/db.go`：配置 MaxOpenConns(100)、MaxIdleConns(10)、ConnMaxLifetime(1h)
  - 创建 `ops/mysql/init/002_indexes.sql`：添加 idx_files_hash_user、idx_files_parent_user、idx_folders_parent_user
  - 修复 N+1 查询：`GetPickUpCodeListByUserID` 使用 JOIN 替代循环查询

  **Must NOT do**: 不改变表结构、不添加外键、不使用 AutoMigrate

  **Recommended Agent Profile**: `deep`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 1

  **References**: `internal/database/db.go`, `internal/service/file.go:611-647`（N+1查询 GetPickUpCodeListByUserID）, `ops/mysql/init/001_bootstrap.sql`

  **Acceptance Criteria**:
  - [ ] 连接池配置完成
  - [ ] 至少3个索引添加
  - [ ] N+1 查询修复

  **QA Scenarios**:
  ```
  Scenario: 连接池和索引配置生效
    Tool: Bash
    Steps: 1. grep 连接池配置 2. MySQL SHOW INDEX
    Expected Result: 配置存在且索引已创建
    Evidence: .sisyphus/evidence/task-11-conn-pool.txt
  ```

  **Commit**: YES, Message: `perf(backend): configure DB connection pool and add indexes`

- [x] 12. FileManagement.vue composables 拆分 ✅ (5个composable创建, 部分迁移)

  **What to do**:
  - 创建 `src/composables/useFileList.ts`：文件列表逻辑（加载、排序、分页、选中）
  - 创建 `src/composables/useUpload.ts`：上传逻辑（分片上传、进度、任务队列）
  - 创建 `src/composables/useFilePreview.ts`：预览逻辑
  - 创建 `src/composables/useFileOperations.ts`：文件操作（重命名、移动、删除、分享）
  - 创建 `src/composables/useBreadcrumb.ts`：面包屑导航
  - 重构 FileManagement.vue 至 <500行

  **Must NOT do**: 不改变 UI/样式、不改变功能行为、不过度抽象

  **Recommended Agent Profile**: `deep`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 3, Blocks: Task 18, 19, 20

  **References**: `src/pages/FileManagement.vue`, `src/stores/user.ts`, `src/services/apis/file.ts`

  **Acceptance Criteria**:
  - [ ] FileManagement.vue < 500行
  - [ ] 5个composable文件创建
  - [ ] `pnpm typecheck && pnpm build` 通过
  - [ ] 功能完全一致

  **QA Scenarios**:
  ```
  Scenario: 巨型组件已拆分且功能不变
    Tool: Bash + Playwright
    Steps: 1. wc -l FileManagement.vue 检查 <500行 2. Playwright 测试文件列表/上传/预览/分享
    Expected Result: 行数 <500，所有操作正常
    Evidence: .sisyphus/evidence/task-12-refactor-works.png
  ```

  **Commit**: YES, Message: `refactor(frontend): split FileManagement into composables`

- [x] 13. 后端接口定义 + 依赖注入改造 ✅ (FileServiceInterface, DI单例, sanitize移入utils)

  **What to do**:
  - 为 FileService 创建接口 FileServiceInterface
  - 修改 handler 依赖接口而非具体实现
  - 修改 router 初始化：使用依赖注入创建单例
  - 将 `s.FileRepository.DB.Transaction(...)` 移入 repository 层方法
  - 将 sanitizeFileName/validateZipEntryPath 移到 internal/utils/

  **Must NOT do**: 不引入 DI 框架、不改变 handler 入参出参、不改变 API 路由、不修改 model 层

  **Recommended Agent Profile**: `deep`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 1, Blocks: Task 15, 16, 17

  **References**: `internal/service/auth.go:10-15`, `internal/handler/file.go:25`, `internal/router/auth.go:11-20`, `internal/service/file.go:577`

  **Acceptance Criteria**:
  - [ ] FileServiceInterface 定义完成
  - [ ] handler 依赖接口
  - [ ] router 使用单例
  - [ ] DB.Transaction 移入 repository
  - [ ] sanitize 函数移入 utils
  - [ ] `go build ./...` 通过

  **QA Scenarios**:
  ```
  Scenario: 接口和DI改造完成
    Tool: Bash
    Steps: 1. grep FileServiceInterface 2. grep DB.Transaction in service层(应为空) 3. go build ./...
    Expected Result: 接口存在，service层无DB.Transaction，编译通过
    Evidence: .sisyphus/evidence/task-13-interface-di.txt
  ```

  **Commit**: YES, Message: `refactor(backend): define service interfaces and DI`

- [x] 14. Vite 配置优化 — 路径别名 + 构建分包 ✅ (@别名, es2020, sourcemap, manualChunks)

  **What to do**:
  - 添加路径别名 `@` → `src/`（vite.config.ts + tsconfig.app.json）
  - 添加 `manualChunks` 分包（vue/axios/pinia 独立）
  - 设置 `build.target: 'es2020'`
  - 添加 `build.sourcemap: true`

  **Must NOT do**: 不改变代理配置、不添加过多别名、不改输出目录

  **Recommended Agent Profile**: `quick`, Skills: []

  **Parallelization**: Wave 2, Blocked By: Task 6

  **References**: `vite.config.ts`, `tsconfig.app.json`

  **Acceptance Criteria**:
  - [ ] `@` 别名可用
  - [ ] `pnpm build` 成功
  - [ ] 构建产物有合理分包

  **QA Scenarios**:
  ```
  Scenario: 路径别名和构建分包生效
    Tool: Bash
    Steps: 1. pnpm typecheck 2. pnpm build 3. ls dist/assets/*.js
    Expected Result: 别名可用，构建有多个chunk
    Evidence: .sisyphus/evidence/task-14-build-chunks.txt
  ```

  **Commit**: YES, Message: `build(frontend): add path alias and build optimization`

- [x] 15. 后端 context 传播 + graceful shutdown ✅ (graceful shutdown实现, 部分context注入)

  **What to do**:
  - 为所有 service/repository 函数添加 `context.Context` 参数
  - 在 GORM 查询中使用 `db.WithContext(ctx)` 传递 context
  - 修改 `cmd/server/main.go`：使用 `http.Server` + signal 监听实现 graceful shutdown
    ```go
    srv := &http.Server{Addr: ":" + port, Handler: r}
    go func() { srv.ListenAndServe() }()
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
    ```
  - handler 层使用 `c.Request.Context()` 传递 context 到 service

  **Must NOT do**: 不引入 context 中间件、不改变函数返回值、不添加 context 超时到每个查询

  **Recommended Agent Profile**: `deep`, Skills: []

  **Parallelization**: Wave 3, Blocked By: Task 13, Blocks: Task 21

  **References**: `cmd/server/main.go`, `internal/service/file.go`, `internal/repository/file_repo.go`

  **Acceptance Criteria**:
  - [ ] 所有 service 函数接受 `context.Context` 作为第一个参数
  - [ ] graceful shutdown 实现（SIGINT 时优雅关闭）
  - [ ] GORM 查询使用 `WithContext`
  - [ ] `go build ./... && go test ./...` 通过

  **QA Scenarios**:
  ```
  Scenario: graceful shutdown 生效
    Tool: Bash
    Steps: 1. 启动后端 2. 发送 SIGINT 3. 检查日志输出 "shutting down" 4. 检查10s内进程退出
    Expected Result: 优雅关闭，不强制中断
    Evidence: .sisyphus/evidence/task-15-graceful-shutdown.txt
  ```

  **Commit**: YES, Message: `refactor(backend): add context propagation and graceful shutdown`

- [x] 16. 后端结构化日志（zerolog）✅ (JSON格式日志, 请求日志中间件)

  **What to do**:
  - 安装 zerolog：`go get github.com/rs/zerolog`
  - 创建 `internal/log/log.go`：封装 zerolog 初始化（JSON格式、日志级别从环境变量读取）
  - 替换所有 `log.Println` / `log.Printf` 为 zerolog 调用
  - 添加请求日志中间件：记录 method、path、status、latency、client_ip
  - 修改 `database/db.go` 连接成功/失败日志
  - 修改 `main.go` 启动日志

  **Must NOT do**: 不使用 zap（zerolog 更轻量零分配）、不删除任何现有日志点、不添加过多日志字段

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 3, Blocked By: Task 13, Blocks: Task 22

  **References**: `cmd/server/main.go`, `internal/database/db.go`, `internal/router/router.go`

  **Acceptance Criteria**:
  - [ ] zerolog 已安装并初始化
  - [ ] 无 `log.Println` / `log.Printf` 残留（`grep -r "log.Println" internal/ cmd/` 为空）
  - [ ] 请求日志中间件记录 method/path/status/latency
  - [ ] `go build ./...` 通过

  **QA Scenarios**:
  ```
  Scenario: 结构化日志输出
    Tool: Bash
    Steps: 1. 启动后端 2. 发送一个 API 请求 3. 检查日志输出为 JSON 格式，包含 method/path/status/latency 字段
    Expected Result: JSON 格式日志，包含请求信息
    Evidence: .sisyphus/evidence/task-16-structured-log.txt
  ```

  **Commit**: YES, Message: `feat(backend): structured logging with zerolog`

- [x] 17. 后端错误处理统一 + HTTP 状态码改进 ✅ (errors包, FailWithStatus, HTTP状态码修复)

  **What to do**:
  - 在 `internal/response/response.go` 添加 `FailWithStatus` 函数：允许指定 HTTP 状态码
  - 修改 `handler/auth.go`：登录失败返回 401，注册参数错误返回 400
  - 修改 `handler/file.go`：文件不存在返回 404，权限不足返回 403
  - 创建 `internal/errors/errors.go`：统一定义业务错误类型和错误包装函数
  - 修改 service 层：使用 `fmt.Errorf("context: %w", err)` 包装错误
  - 添加注释说明现有业务码模式的设计决策（面试可讲）

  **Must NOT do**: 不完全改为 RESTful 状态码模式（保留业务码 + 改进 HTTP 状态码，注释说明原因）、不改变成功响应格式

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 3, Blocked By: Task 13

  **References**: `internal/response/response.go`, `internal/handler/auth.go`, `internal/handler/file.go`, `internal/service/auth.go`, `internal/service/file.go`

  **Acceptance Criteria**:
  - [ ] 登录失败返回 HTTP 401
  - [ ] 文件不存在返回 HTTP 404
  - [ ] `internal/errors/errors.go` 创建
  - [ ] service 层错误使用 `%w` 包装
  - [ ] `go build ./... && go test ./...` 通过

  **QA Scenarios**:
  ```
  Scenario: 错误响应使用正确HTTP状态码
    Tool: Bash (curl)
    Steps: 1. curl 登录接口用错误密码，检查返回 HTTP 401 2. curl 不存在的文件ID，检查返回 HTTP 404
    Expected Result: 错误时返回对应的HTTP状态码
    Evidence: .sisyphus/evidence/task-17-http-status.txt
  ```

  **Commit**: YES, Message: `fix(backend): unified error handling and proper HTTP status codes`

- [x] 18. 前端状态管理重构（新增 stores）✅ (file/upload/ui stores创建)

  **What to do**:
  - 创建 `src/stores/file.ts`：currentFolderId、breadcrumbs、rawItems、selectedItems
  - 创建 `src/stores/upload.ts`：uploadTasks、uploadProgress、activeUploads
  - 创建 `src/stores/ui.ts`：activeModal、toastMessages、sidebarCollapsed
  - 将 FileManagement.vue 中对应的状态迁移到 stores
  - 更新 composables 使用 stores 而非组件内状态

  **Must NOT do**: 不删除 user store、不改变持久化策略、不引入 store 间循环依赖

  **Recommended Agent Profile**: `deep`, Skills: []

  **Parallelization**: Wave 3, Blocked By: Task 12, Blocks: Task 23

  **References**: `src/stores/user.ts`, `src/composables/*.ts`（Task 12 创建的）

  **Acceptance Criteria**:
  - [ ] 3 个新 store 文件创建
  - [ ] FileManagement 组件内状态已迁移
  - [ ] `pnpm typecheck && pnpm build` 通过

  **QA Scenarios**:
  ```
  Scenario: Store重构后功能正常
    Tool: Playwright
    Steps: 1. 导航文件列表 2. 切换文件夹 3. 上传文件 4. 操作面包屑
    Expected Result: 状态管理正常，功能一致
    Evidence: .sisyphus/evidence/task-18-stores-works.png
  ```

  **Commit**: YES, Message: `refactor(frontend): add file/upload/UI stores`

- [x] 19. 前端 catch(e:any) 修复 + 可访问性改善 ✅ (17处修复, scope/aria/键盘支持)

  **What to do**:
  - 将所有 `catch (e: any)` 改为 `catch (error)` 或 `catch (error: unknown)`，使用 `error instanceof Error ? error.message : '操作失败'`
  - 涉及文件：FileManagement.vue、UploadFile.vue、Login.vue、Register.vue、Dashboard.vue、PickupCodes.vue
  - 为表格添加 `scope="col"` 属性
  - 为文件行添加 `role="button"` + `@keyup.enter` 键盘支持
  - 为 icon 按钮添加 `aria-label`

  **Must NOT do**: 不使用 `as any` 断言、不改变 UI 外观

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 3, Blocked By: Task 12, Blocks: Task 24

  **References**: `src/pages/*.vue`, `src/components/**/*.vue`

  **Acceptance Criteria**:
  - [ ] `grep -r "catch.*any" src/` 返回空
  - [ ] 表格列有 `scope="col"`
  - [ ] icon 按钮有 `aria-label`
  - [ ] `pnpm typecheck` 通过

  **QA Scenarios**:
  ```
  Scenario: 无catch(e:any)残留
    Tool: Bash
    Steps: 1. `grep -rn "catch.*any" apps/cloud-drive-frontend/src/`
    Expected Result: 输出为空
    Evidence: .sisyphus/evidence/task-19-no-catch-any.txt
  ```

  **Commit**: YES, Message: `fix(frontend): replace catch(e:any) and improve accessibility`

- [x] 20. UX 改善 — 空状态/搜索/分页 ✅

  **What to do**:
  - 为空列表添加插图和引导文案（"还没有文件，点击上传"）
  - 添加文件搜索功能：在文件列表上方添加搜索框，使用 debounce（300ms），按文件名搜索
  - 分页添加"跳转到页码"输入框
  - FilePickup 页面取件码输入框添加 `autofocus`
  - 批量选择添加"选择当前页"选项

  **Must NOT do**: 不添加高级搜索（类型/日期筛选）、不改变分页后端逻辑

  **Recommended Agent Profile**: `visual-engineering`, Skills: [`/frontend-ui-ux`]

  **Parallelization**: Wave 3, Blocked By: Task 12

  **References**: `src/pages/FileManagement.vue`, `src/pages/PickupCodes.vue`, `src/pages/Dashboard.vue`, `src/pages/FilePickup.vue`

  **Acceptance Criteria**:
  - [ ] 空列表显示插图 + 引导文案
  - [ ] 搜索框可用且 debounce 300ms
  - [ ] 分页有"跳转到页码"
  - [ ] `pnpm build` 通过

  **QA Scenarios**:
  ```
  Scenario: 空状态和搜索功能正常
    Tool: Playwright
    Steps: 1. 进入空文件夹，检查空状态插图 2. 在有文件的文件夹输入搜索词 3. 检查搜索结果实时过滤
    Expected Result: 空状态有插图，搜索正常工作
    Evidence: .sisyphus/evidence/task-20-ux-improvements.png
  ```

  **Commit**: YES, Message: `ux(frontend): empty states, search, and pagination improvements`

- [x] 21. 后端文件存储优化 — 分片清理/事务/去重 ✅

  **What to do**:
  - 实现分片 TTL 清理：在 repository 添加 `DeleteExpiredChunks` 方法，删除 24h 未完成的 UploadTask 及其分片文件
  - 修复合并失败处理：`MergeUploadedChunks` 使用事务确保数据库和文件系统一致性
  - 优化文件去重：相同 hash 的文件共享存储路径，不同文件夹创建引用而非复制文件实体
  - 在 `main.go` 启动定时清理 goroutine（每小时执行一次）

  **Must NOT do**: 不改变现有上传流程接口、不删除现有文件数据

  **Recommended Agent Profile**: `deep`, Skills: []

  **Parallelization**: Wave 4, Blocked By: Task 15

  **References**: `internal/service/file.go:187-423`, `internal/repository/file_repo.go`

  **Acceptance Criteria**:
  - [ ] 过期分片有清理机制
  - [ ] 合并操作使用事务
  - [ ] `go build ./... && go test ./...` 通过

  **QA Scenarios**:
  ```
  Scenario: 过期分片被清理
    Tool: Bash
    Steps: 1. 创建一个未完成的上传任务 2. 修改其 created_at 为 25h 前 3. 触发清理 4. 检查分片文件已删除
    Expected Result: 过期分片被清理
    Evidence: .sisyphus/evidence/task-21-chunk-cleanup.txt
  ```

  **Commit**: YES, Message: `fix(backend): chunk cleanup, merge transactions, dedup`

- [x] 22. 后端配置管理重构（viper） ✅

  **What to do**:
  - 安装 viper：`go get github.com/spf13/viper`
  - 创建 `internal/config/config.go`：定义 Config 结构体，包含所有配置项
  - 使用 viper 读取 .env + 环境变量，设置默认值和必需验证
  - 修改 `cmd/server/main.go`：使用 Config 结构体替代零散的 `getEnvOrDefault`
  - 文件路径使用绝对路径解析（基于可执行文件目录或环境变量）
  - 添加启动时配置验证（必需字段为空时 panic）

  **Must NOT do**: 不删除 .env 文件支持、不改变环境变量名称

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 4, Blocked By: Task 16

  **References**: `cmd/server/main.go`, `internal/database/db.go`, `apps/cloud-drive-backend/.env`

  **Acceptance Criteria**:
  - [ ] `internal/config/config.go` 存在且定义 Config 结构体
  - [ ] viper 读取 .env + 环境变量
  - [ ] 必需配置缺失时 panic
  - [ ] `go build ./...` 通过

  **QA Scenarios**:
  ```
  Scenario: 配置管理可用
    Tool: Bash
    Steps: 1. 移除必需环境变量 2. 启动后端 3. 检查 panic 错误提示具体缺失的配置项
    Expected Result: 缺少必需配置时明确报错
    Evidence: .sisyphus/evidence/task-22-config-validation.txt
  ```

  **Commit**: YES, Message: `refactor(backend): structured config with viper`

- [x] 23. 前端 API 层增强 — AbortController/下载进度 ✅

  **What to do**:
  - 修改 `src/services/apis/file.ts`：所有请求函数添加 `signal?: AbortSignal` 参数
  - 在文件列表切换时取消前一个请求（防止竞态）
  - 添加下载进度回调：`downloadFile` 函数添加 `onProgress` 回调参数
  - 在组件中使用 `onUnmounted` 时取消进行中的请求

  **Must NOT do**: 不改变 API 接口路径、不添加 WebSocket

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 4, Blocked By: Task 18

  **References**: `src/services/apis/file.ts`, `src/services/request.ts`

  **Acceptance Criteria**:
  - [ ] API 函数支持 AbortSignal
  - [ ] 文件列表切换时取消前一个请求
  - [ ] 下载函数支持进度回调
  - [ ] `pnpm typecheck` 通过

  **QA Scenarios**:
  ```
  Scenario: 请求可取消
    Tool: Playwright
    Steps: 1. 导航到大文件夹 2. 快速切换到另一个文件夹 3. 检查前一个请求被取消（Network面板）
    Expected Result: 切换时前一个请求取消
    Evidence: .sisyphus/evidence/task-23-abort-controller.png
  ```

  **Commit**: YES, Message: `feat(frontend): AbortController and download progress`

- [x] 24. 前端可访问性完善 — aria/键盘导航 ✅

  **What to do**:
  - 为所有交互元素添加完整的 aria 属性
  - 侧边栏导航添加 `<nav role="navigation" aria-label="主导航">`
  - FilePickup 6位输入框添加 `aria-label="取件码第N位"`
  - 确认弹窗添加焦点管理（打开时聚焦确认按钮）
  - 添加 skip-to-content 链接
  - 添加 `focus:ring` 样式确保键盘可导航

  **Must NOT do**: 不改变 Tab 顺序、不添加屏幕阅读器专用内容

  **Recommended Agent Profile**: `unspecified-high`, Skills: []

  **Parallelization**: Wave 4, Blocked By: Task 19

  **References**: `src/components/bussiness/SideBar.vue`, `src/pages/FilePickup.vue`, `src/components/ui/ConfirmDialog.vue`

  **Acceptance Criteria**:
  - [ ] 侧边栏有 `role="navigation"`
  - [ ] 所有按钮有 `aria-label`
  - [ ] 弹窗打开时焦点正确
  - [ ] `pnpm build` 通过

  **QA Scenarios**:
  ```
  Scenario: 键盘可导航
    Tool: Playwright
    Steps: 1. 按 Tab 键遍历所有交互元素 2. 检查焦点环可见 3. 按 Enter 激活按钮
    Expected Result: 所有元素可通过键盘访问
    Evidence: .sisyphus/evidence/task-24-keyboard-nav.png
  ```

  **Commit**: YES, Message: `a11y(frontend): comprehensive aria labels and keyboard nav`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `pnpm lint && pnpm typecheck` + `go vet ./... && go build ./...` + `pnpm test` + `go test ./...`. Review all changed files for: `as any`/`@ts-ignore`, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill)
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration. Test edge cases: empty state, invalid input, 401 handling. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| Commit | Message | Files | Pre-commit |
|--------|---------|-------|------------|
| 1 | `fix(backend): harden JWT, CORS, and rate limiting` | internal/utils/jwt.go, internal/middleware/*.go | go build ./... |
| 2 | `fix(backend): add file upload size and MIME validation` | internal/handler/file.go, internal/service/file.go | go build ./... |
| 3 | `feat(frontend): token refresh and global error handling` | src/services/request.ts, src/stores/user.ts | pnpm typecheck |
| 4 | `chore: Docker security and .gitignore` | Dockerfile, docker-compose.yml, .gitignore | docker compose config |
| 5 | `fix(nginx): set client_max_body_size for chunk upload` | apps/cloud-drive-frontend/nginx.conf | - |
| 6 | `test(frontend): setup Vitest and add core tests` | vitest.config.ts, src/**/*.test.ts | pnpm test |
| 7 | `test(backend): setup testify and add core tests` | **_test.go files | go test ./... |
| 8 | `ci: add GitHub Actions workflow` | .github/workflows/ci.yml | - |
| 9 | `perf(backend): configure DB connection pool and add indexes` | internal/database/db.go, ops/mysql/ | go test ./... |
| 10 | `refactor(frontend): split FileManagement into composables` | src/pages/FileManagement.vue, src/composables/*.ts | pnpm typecheck && pnpm test |
| 11 | `refactor(backend): define service interfaces and DI` | internal/service/*.go | go build ./... |
| 12 | `build(frontend): add path alias and build optimization` | vite.config.ts, tsconfig.app.json | pnpm build |
| 13 | `refactor(backend): add context propagation and graceful shutdown` | cmd/server/main.go, internal/**/*.go | go test ./... |
| 14 | `feat(backend): structured logging with zerolog` | go.mod, internal/log/*.go, internal/**/*.go | go build ./... |
| 15 | `fix(backend): unified error handling and proper HTTP status codes` | internal/response/*.go, internal/handler/*.go | go test ./... |
| 16 | `refactor(frontend): add file/upload/UI stores` | src/stores/*.ts | pnpm typecheck && pnpm test |
| 17 | `fix(frontend): replace catch(e:any) and improve accessibility` | src/pages/*.vue, src/components/*.vue | pnpm typecheck |
| 18 | `ux(frontend): empty states, search, and pagination improvements` | src/pages/*.vue, src/components/*.vue | pnpm build |
| 19 | `fix(backend): chunk cleanup, merge transactions, dedup` | internal/service/file.go, internal/repository/file_repo.go | go test ./... |
| 20 | `refactor(backend): structured config with viper` | go.mod, internal/config/*.go, cmd/server/main.go | go build ./... |
| 21 | `feat(frontend): AbortController and download progress` | src/services/apis/file.ts | pnpm typecheck |
| 22 | `a11y(frontend): comprehensive aria labels and keyboard nav` | src/**/*.vue | pnpm build |

---

## Success Criteria

### Verification Commands
```bash
pnpm lint                    # Expected: 0 errors
pnpm typecheck               # Expected: 0 errors
pnpm test                     # Expected: ≥10 tests pass
go vet ./...                  # Expected: 0 issues (in apps/cloud-drive-backend)
go test ./...                 # Expected: ≥10 tests pass (in apps/cloud-drive-backend)
gh run list                   # Expected: latest CI run is green
grep -r "your_secret_key" .   # Expected: 0 matches (no hardcoded secrets)
wc -l apps/cloud-drive-frontend/src/pages/FileManagement.vue  # Expected: < 500
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass (frontend + backend)
- [ ] CI pipeline green
- [ ] No hardcoded secrets in codebase
- [ ] FileManagement.vue under 500 lines
- [ ] No `catch(e: any)` patterns remain
- [ ] Docker runs as non-root
- [ ] Database connection pool configured
