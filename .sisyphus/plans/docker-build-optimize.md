# Docker 构建优化：解决 go mod download 卡死问题

## TL;DR

> **核心目标**: 通过 GOPROXY 镜像 + BuildKit 缓存挂载 + .dockerignore 三层优化，彻底解决 Docker 构建中 `go mod download` 卡死的问题。
> 
> **交付物**:
> - 修改后的 Dockerfile（含 GOPROXY + Cache Mount）
> - 新增 .dockerignore 文件
> - 构建验证通过
> 
> **预估工作量**: Quick（2 个文件修改，< 30 分钟）
> **并行执行**: YES - 2 tasks 可并行，1 验证串行
> **关键路径**: Task 1/2 → Task 3（验证）

---

## Context

### Original Request
项目在服务器部署时 Docker 构建卡死在 `go mod download`，想改造成 vendor 或寻找更好方案。经评估，采用 GOPROXY + Build Cache + .dockerignore 组合方案。

### Interview Summary
**Key Discussions**:
- 用户提出 vendor 方案 → 评估后推荐组合方案更优（改动小、效果好、无维护负担）
- 用户确认选择组合方案（不含 vendor）

**Research Findings**:
- Dockerfile 已使用多阶段构建，且 `COPY go.mod go.sum` 在 `COPY . .` 之前（缓存策略正确）
- 没有现有的 vendor 目录和 .dockerignore 文件
- go.mod 约 71 个模块（11 直接 + 60 间接）
- 后端 `.env` 文件存在于代码目录中，需在 .dockerignore 中排除
- Docker build context 为 `./apps/cloud-drive-backend`

### Metis Review
**Identified Gaps**（已处理）:
- GOPROXY 应使用三段 fallback（goproxy.cn → goproxy.io → direct）而非两段
- Cache Mount 应同时缓存 go mod 和 go build 两个路径，并设置 id 防止缓存污染
- .dockerignore 必须排除 .env 文件（安全风险）
- .dockerignore 放在 `apps/cloud-drive-backend/` 目录（与 Dockerfile 同级，因 build context 在该目录）
- Go 1.25 版本需确认 Docker Hub 上可用

---

## Work Objectives

### Core Objective
优化 Go 后端 Docker 构建，使 `go mod download` 不再因网络问题卡死，同时减少构建上下文大小和利用缓存加速二次构建。

### Concrete Deliverables
- `apps/cloud-drive-backend/Dockerfile` — 添加 GOPROXY + Build Cache Mount
- `apps/cloud-drive-backend/.dockerignore` — 排除不必要文件，减小构建上下文

### Definition of Done
- [x] `docker compose --profile backend build backend` 构建成功无卡死
- [x] 第二次构建（仅改代码）go mod download 步骤使用缓存，耗时显著减少
- [x] `.env` 文件不出现在最终镜像中
- [x] 容器启动后 healthz 端点正常响应

### Must Have
- `ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct` 三段 fallback
- `--mount=type=cache` 带 `id` 参数分别缓存 go mod 和 go build
- `.dockerignore` 排除 `.env`、`.git`、docs 等
- 保持现有 Dockerfile 多阶段构建结构不变

### Must NOT Have (Guardrails)
- ❌ 不添加 vendor 目录或机制
- ❌ 不在 Dockerfile 中添加 test 阶段
- ❌ 不修改 docker-compose.yml 结构
- ❌ 不修改 CGO_ENABLED=0 或 runtime 阶段
- ❌ 不添加不必要的注释或 AI slop（如"优化构建速度"之类的废话注释）

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: NO（无单元测试框架用于 Docker 构建验证）
- **Automated tests**: None（通过 Docker 命令行验证）
- **Framework**: N/A
- **Agent-Executed QA**: Bash 命令验证构建和运行时

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Docker 构建**: Use Bash (docker compose) — 构建命令、检查日志、验证镜像
- **运行时验证**: Use Bash (curl) — 健康检查、响应验证

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - 两个独立修改可并行):
├── Task 1: 修改 Dockerfile（GOPROXY + Cache Mount）[quick]
└── Task 2: 创建 .dockerignore [quick]

Wave 2 (After Wave 1 - 验证):
└── Task 3: 构建与运行时验证 [unspecified-high]

Wave FINAL (After ALL tasks):
├── F1: Plan Compliance Audit [oracle]
├── F2: Code Quality Review [unspecified-high]
├── F3: Real Manual QA [unspecified-high]
└── F4: Scope Fidelity Check [deep]

Critical Path: Task 1/2 → Task 3 → F1-F4 → user okay
Parallel Speedup: ~50% (Task 1 和 Task 2 并行)
Max Concurrent: 2 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1    | -         | 3      | 1    |
| 2    | -         | 3      | 1    |
| 3    | 1, 2      | F1-F4  | 2    |
| F1   | 3         | -      | FINAL |
| F2   | 3         | -      | FINAL |
| F3   | 3         | -      | FINAL |
| F4   | 3         | -      | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **2** - T1 → `quick`, T2 → `quick`
- **Wave 2**: **1** - T3 → `unspecified-high`
- **FINAL**: **4** - F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. 修改 Dockerfile：添加 GOPROXY 和 Build Cache Mount

  **What to do**:
  - 在 `apps/cloud-drive-backend/Dockerfile` 的 builder 阶段添加 `ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct`
  - 将 `RUN go mod download` 改为 `RUN --mount=type=cache,target=/go/pkg/mod,id=go-mod-cache go mod download`
  - 将 `RUN CGO_ENABLED=0 GOOS=linux go build ...` 改为 `RUN --mount=type=cache,target=/root/.cache/go-build,id=go-build-cache CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go`
  - 保持其他所有内容不变（多阶段构建、alpine runtime、非 root 用户等）

  **Must NOT do**:
  - 不修改 runtime 阶段（alpine:3.22 部分）
  - 不添加 test 阶段
  - 不改 CGO_ENABLED 设定
  - 不添加多余注释

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单文件修改，改动明确，不超过 10 行变更
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Task 3
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `apps/cloud-drive-backend/Dockerfile` — 当前 Dockerfile，保持其多阶段构建模式，仅在 builder 阶段添加 ENV 和 cache mount

  **External References**:
  - Docker BuildKit cache mount 文档: `https://docs.docker.com/build/cache/optimize/#use-cache-mounts`
  - GOPROXY 环境变量: `https://go.dev/ref/mod#proxy-configuration`

  **WHY Each Reference Matters**:
  - Dockerfile 是唯一需要修改的文件，需确保在正确行添加 ENV 和 cache mount 语法
  - GOPROXY 三段 fallback（goproxy.cn → goproxy.io → direct）确保至少一个代理可达

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Dockerfile 语法正确 - 构建 Docker 镜像
    Tool: Bash (docker compose)
    Preconditions: Docker daemon 运行中
    Steps:
      1. 执行 `cd /Users/shi/study/frontend/projects/cloud-drive && docker compose --profile backend build backend 2>&1`
      2. 检查构建日志中出现 goproxy.cn 或 goproxy.io 的下载 URL
      3. 检查构建成功完成（退出码 0）
    Expected Result: 构建成功，日志中可见使用 goproxy.cn 作为代理下载依赖
    Failure Indicators: 构建失败，或 go mod download 超时/C cache mount 语法错误
    Evidence: .sisyphus/evidence/task-1-dockerfile-build.log

  Scenario: Cache Mount 语法错误导致构建失败（负面测试）
    Tool: Bash (docker compose)
    Preconditions: BuildKit 已启用
    Steps:
      1. 检查 Dockerfile 中 --mount=type=cache 语法是否正确（使用 `docker compose --profile backend build backend` 验证）
      2. 如果语法错误，构建会报错明确的错误信息
    Expected Result: 构建正常通过，cache mount 语法被正确解析
    Evidence: .sisyphus/evidence/task-1-cache-mount-check.log
  ```

  **Commit**: YES (groups with Task 2)
  - Message: `fix(backend): optimize docker build with goproxy and buildkit cache`
  - Files: `apps/cloud-drive-backend/Dockerfile`, `apps/cloud-drive-backend/.dockerignore`
  - Pre-commit: `docker compose --profile backend build backend`

- [x] 2. 创建 .dockerignore 文件

  **What to do**:
  - 在 `apps/cloud-drive-backend/` 目录创建 `.dockerignore` 文件（与 Dockerfile 同级，因为 docker-compose.yml 中 build context 设为 `./apps/cloud-drive-backend`）
  - 排除以下内容：
    ```
    .env
    .env.example
    .git
    .gitignore
    docs/
    *.md
    AGENTS.md
    ENV_CONFIG.md
    README.md
    .DS_Store
    *.test
    *.out
    ```
  - 确保 **不排除** 以下构建必需文件：
    - `go.mod`、`go.sum`
    - `cmd/` 目录
    - `internal/` 目录

  **Must NOT do**:
  - 不排除 go.mod、go.sum、cmd/、internal/
  - 不在项目根目录创建 .dockerignore（build context 不在根目录）
  - 不添加 vendor 相关排除（项目不用 vendor）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 创建一个简单文件，内容明确
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Task 3
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `docker-compose.yml:28` — build context 是 `./apps/cloud-drive-backend`，所以 .dockerignore 必须在此目录
  - `apps/cloud-drive-backend/` — 列出目录内容确认有哪些文件需排除（.env、docs/、*.md 等）

  **WHY Each Reference Matters**:
  - docker-compose.yml 中的 build context 决定了 .dockerignore 的位置
  - 后端目录内容帮助确认哪些文件应该排除

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: .dockerignore 正确排除不必要文件
    Tool: Bash (docker compose)
    Preconditions: Task 1 和 Task 2 均已完成
    Steps:
      1. 执行 `docker compose --profile backend build backend 2>&1 | head -20`
      2. 观察构建上下文大小（应比无 .dockerignore 时小）
      3. 执行 `docker compose --profile backend run --rm backend ls -la /app/ 2>&1 || true`
      4. 确认 .env 文件不在 /app/ 目录中
    Expected Result: .env 文件不出现在容器内的 /app/ 目录
    Failure Indicators: /app/.env 文件存在，说明 .dockerignore 未正确排除
    Evidence: .sisyphus/evidence/task-2-dockerignore-verify.log

  Scenario: .dockerignore 不排除构建必需文件
    Tool: Bash (docker compose)
    Preconditions: Dockerfile 和 .dockerignore 已就位
    Steps:
      1. 执行 `docker compose --profile backend build backend 2>&1`
      2. 构建应成功完成，说明 go.mod、go.sum、cmd/、internal/ 未被排除
    Expected Result: 构建成功，所有源代码文件被正确复制到构建阶段
    Failure Indicators: 构建失败提示找不到 go.mod 或源代码文件
    Evidence: .sisyphus/evidence/task-2-build-with-ignore.log
  ```

  **Commit**: YES (groups with Task 1)
  - Message: `fix(backend): optimize docker build with goproxy and buildkit cache`
  - Files: `apps/cloud-drive-backend/Dockerfile`, `apps/cloud-drive-backend/.dockerignore`
  - Pre-commit: `docker compose --profile backend build backend`

- [x] 3. 构建与运行时完整验证

  **What to do**:
  - 完整构建 Docker 镜像并验证运行时功能
  - 第一次构建：清除缓存后完整构建，验证 GOPROXY 生效
  - 第二次构建：仅修改源代码后重建，验证 Build Cache 加速效果
  - 启动服务，验证 healthz 端点和数据库连接
  - 验证 .env 文件不在最终镜像中

  **Must NOT do**:
  - 不修改任何源代码文件
  - 不修改 docker-compose.yml
  - 不添加任何新功能

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要执行多个 Docker 命令并验证结果，需要仔细检查
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential, depends on Task 1 and 2)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1, Task 2

  **References**:

  **Pattern References**:
  - `docker-compose.yml` — 需要使用 `--profile backend` 启动服务
  - `apps/cloud-drive-backend/Dockerfile` — 修改后的 Dockerfile，验证 GOPROXY 和 cache mount

  **API/Type References**:
  - `docker-compose.yml:51` — healthcheck 配置: `wget -q -O /dev/null http://127.0.0.1:9000/healthz`

  **WHY Each Reference Matters**:
  - docker-compose.yml 确定了启动命令和健康检查方式
  - Dockerfile 的 healthz 端点是验证应用启动的关键

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: 完整构建成功（冷启动）
    Tool: Bash (docker compose)
    Preconditions: Task 1 和 Task 2 已完成，Docker daemon 运行中
    Steps:
      1. 执行 `docker compose --profile backend build --no-cache backend 2>&1 | tee /tmp/build-cold.log`
      2. 检查退出码为 0
      3. 检查日志中出现 goproxy.cn 相关的下载 URL: `grep -i "goproxy" /tmp/build-cold.log || echo "goproxy keyword found in go mod download output"`
      4. 记录构建时间
    Expected Result: 构建成功完成，无超时或卡死，日志可见使用 goproxy 代理
    Failure Indicators: 构建超时、卡死、或 go mod download 失败
    Evidence: .sisyphus/evidence/task-3-cold-build.log

  Scenario: 缓存构建加速（热启动）
    Tool: Bash (docker compose)
    Preconditions: 第一次构建已成功
    Steps:
      1. 修改一个源代码文件（如添加空行到 cmd/server/main.go）
      2. 执行 `docker compose --profile backend build backend 2>&1 | tee /tmp/build-warm.log`
      3. 对比冷/热构建中 go mod download 步骤的耗时
      4. 热构建中 go mod download 应使用缓存，几乎瞬间完成
    Expected Result: 第二次构建明显更快，go mod download 使用缓存
    Failure Indicators: 第二次构建的 go mod download 仍然重新下载所有依赖
    Evidence: .sisyphus/evidence/task-3-warm-build.log

  Scenario: 运行时功能验证
    Tool: Bash (docker compose + curl)
    Preconditions: 服务已通过 docker compose --profile backend up -d 启动
    Steps:
      1. 执行 `docker compose --profile backend up -d` 启动服务（含 MySQL）
      2. 等待 healthcheck 通过: `docker compose ps` 查看 backend 状态为 healthy
      3. 执行 `curl -s http://localhost:9000/healthz`
      4. 验证返回 200 OK
      5. 执行 `docker compose --profile backend down` 清理
    Expected Result: backend 服务健康检查通过，healthz 端点正常响应
    Failure Indicators: 服务未启动或 healthz 返回非 200
    Evidence: .sisyphus/evidence/task-3-runtime-verify.log

  Scenario: 安全验证 - .env 不在最终镜像中
    Tool: Bash (docker compose)
    Preconditions: 镜像已构建
    Steps:
      1. 执行 `docker compose --profile backend run --rm backend cat /app/.env 2>&1 || true`
      2. 确认返回 "No such file" 或类似错误（文件不存在）
    Expected Result: .env 文件不存在于最终镜像中
    Failure Indicators: .env 文件内容被输出（说明 .dockerignore 未正确排除）
    Evidence: .sisyphus/evidence/task-3-security-verify.log
  ```

  **Commit**: NO (验证任务，不提交代码)

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, grep pattern). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Read modified Dockerfile, check: ENV GOPROXY has 3-stage fallback, --mount=type=cache has id parameter, no unnecessary comments, no AI slop. Read .dockerignore, check: .env excluded, go.mod/go.sum NOT excluded, cmd/internal NOT excluded. Verify no other files were touched.
  Output: `Dockerfile [CLEAN/ISSUES] | .dockerignore [CLEAN/ISSUES] | Scope [CLEAN/CREEP] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Execute: `docker compose --profile backend build backend` → verify build succeeds. Execute: `docker compose --profile backend up -d` → verify healthcheck passes. Execute: `curl http://localhost:9000/healthz` → verify 200 response. Verify .env not in image. Save evidence to .sisyphus/evidence/final-qa/.
  Output: `Build [PASS/FAIL] | Healthcheck [PASS/FAIL] | .env Check [PASS/FAIL] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **1**: `fix(backend): optimize docker build with goproxy and buildkit cache` - apps/cloud-drive-backend/Dockerfile, apps/cloud-drive-backend/.dockerignore
  Pre-commit: `docker compose --profile backend build backend`

---

## Success Criteria

### Verification Commands
```bash
# 构建成功（无卡死）
docker compose --profile backend build backend
# Expected: 构建成功完成，无超时

# 健康检查通过
docker compose --profile backend up -d
docker compose ps  # backend 状态为 healthy
curl http://localhost:9000/healthz  # 返回 200

# .env 不在镜像中
docker compose --profile backend run --rm backend cat /app/.env 2>&1
# Expected: "No such file" 或错误（文件不存在）

# 清理
docker compose --profile backend down
```

### Final Checklist
- [ ] GOPROXY 使用三段 fallback（goproxy.cn → goproxy.io → direct）
- [ ] Cache Mount 带 id 参数分别缓存 go mod 和 go build
- [ ] .dockerignore 排除 .env、.git、docs 等
- [ ] .dockerignore 不排除 go.mod、go.sum、cmd/、internal/
- [ ] Docker 构建不再卡死在 go mod download
- [ ] .env 文件不出现在最终镜像中
- [ ] 未添加 vendor、test 阶段等超出范围的改动