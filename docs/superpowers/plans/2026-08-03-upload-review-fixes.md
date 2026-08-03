# Upload Review Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 验证并修复上传幂等改造审查中可复现的迁移、并发、兼容性、轮询和死代码问题。

**Architecture:** 后端迁移只产生必要回填并容忍并发建索引的 MySQL 1061；初始化失败保留可重试任务；旧客户端缺少分片哈希时由 HTTP 边界计算。前端用内部 AbortController 统一用户取消与十分钟总截止时间，并为 merge 请求关闭全局 30 秒传输超时。

**Tech Stack:** Go 1.25、GORM、go-sqlmock、Gin、Vue 3、TypeScript、Axios、Vitest。

## Global Constraints

- 保留用户现有 staged 改动，不重置、不提交、不暂存本轮修复。
- `docs/brain/` 与 `docs/skills/` 只读。
- 每个行为修复先写回归测试并确认按预期失败，再写最小实现。
- CR-007 在标准 JavaScript run-to-completion 语义下不能形成可触发竞态；仍以二次 `signal.aborted` 检查和取消回归测试加固同一路径。

---

### Task 1: 迁移回填与并发建索引

**Files:**
- Modify: `apps/cloud-drive-backend/internal/database/migrate.go`
- Test: `apps/cloud-drive-backend/internal/database/migrate_test.go`

**Interfaces:**
- Consumes: `[]model.UploadTask`、`*mysql.MySQLError`。
- Produces: `buildUploadIdentityBackfills` 只返回值发生变化的行；`normalizeUploadIdentityIndexError(error) error` 把 MySQL 1061 转成 `nil`。

- [x] **Step 1: 写回填跳过稳定行和 1061 容忍测试**

```go
func TestBuildUploadIdentityBackfills_SkipsTasksWhoseIdentityIsAlreadyBackfilled(t *testing.T) {
    task := model.UploadTask{ID: 1, UserID: 7, FileName: "report.pdf", FileHash: "abcdef", FileSize: 5, ChunkSize: 5, TotalChunks: 1, FileType: "text/plain"}
    key, requestHash, _, err := utils.BuildUploadIdentity(&task)
    require.NoError(t, err)
    task.IdempotencyKey, task.RequestHash = key, requestHash
    backfills, err := buildUploadIdentityBackfills([]model.UploadTask{task})
    require.NoError(t, err)
    assert.Empty(t, backfills)
}

func TestNormalizeUploadIdentityIndexError_IgnoresDuplicateIndex(t *testing.T) {
    err := fmt.Errorf("wrapped: %w", &mysql.MySQLError{Number: 1061, Message: "Duplicate key name"})
    assert.NoError(t, normalizeUploadIdentityIndexError(err))
}
```

- [x] **Step 2: 运行数据库包测试并确认失败原因**

Run: `go test ./internal/database -count=1`

Expected: 稳定行仍返回一个 backfill，且错误归一化函数尚不存在。

- [x] **Step 3: 只返回变化行并在建索引处归一化 1061**

```go
if task.IdempotencyKey == idempotencyKey && task.RequestHash == requestHash {
    continue
}

func normalizeUploadIdentityIndexError(err error) error {
    var mysqlErr *mysql.MySQLError
    if errors.As(err, &mysqlErr) && mysqlErr.Number == 1061 {
        return nil
    }
    return err
}
```

- [x] **Step 4: 运行数据库包测试确认通过**

Run: `go test ./internal/database -count=1`

### Task 2: 初始化目录失败时保留任务

**Files:**
- Modify: `apps/cloud-drive-backend/internal/service/file.go`
- Test: `apps/cloud-drive-backend/internal/service/file_test.go`

**Interfaces:**
- Consumes: `fileService.InitUploadFile` 与可阻塞 `MkdirAll` 的测试目录。
- Produces: 目录创建失败仍返回原错误，但不删除已公开给并发请求的上传任务。

- [x] **Step 1: 写回归测试，通过 GORM delete callback 观察误删副作用**

```go
deleted := false
require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("test:observe-delete", func(*gorm.DB) {
    deleted = true
}))
_, err = svc.InitUploadFile(req)
require.Error(t, err)
assert.False(t, deleted)
```

- [x] **Step 2: 运行服务测试并确认当前实现触发 delete callback**

Run: `go test ./internal/service -run TestInitUploadFile_RetainsTaskWhenChunkDirectoryCreationFails -count=1`

Expected: FAIL，`deleted` 实际为 `true`。

- [x] **Step 3: 删除 `MkdirAll` 失败分支中的无条件任务删除**

```go
if err := os.MkdirAll(chunkDir, 0755); err != nil {
    return nil, err
}
```

- [x] **Step 4: 运行服务包测试确认通过**

Run: `go test ./internal/service -count=1`

### Task 3: 分片哈希向旧客户端兼容

**Files:**
- Modify: `apps/cloud-drive-backend/internal/dto/file.go`
- Modify: `apps/cloud-drive-backend/internal/handler/file.go`
- Test: `apps/cloud-drive-backend/internal/handler/file_test.go`
- Regenerate: `apps/cloud-drive-backend/docs/docs.go`
- Regenerate: `apps/cloud-drive-backend/docs/swagger.json`
- Regenerate: `apps/cloud-drive-backend/docs/swagger.yaml`

**Interfaces:**
- Consumes: 可选 multipart 字段 `chunk_hash` 与已读入的 `fullData`。
- Produces: 缺少哈希时以 `hex.EncodeToString(sha256.Sum256(fullData)[:])` 等价结果填入 `UploadChunkReq.ChunkHash`；显式哈希仍由服务校验。

- [x] **Step 1: 写不带 `chunk_hash` 的 handler 回归测试**

```go
mockFileSvc.uploadFileChunkStreamFunc = func(_ uint, chunk *dto.UploadChunkReq, _ io.Reader, _ int64) error {
    assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", chunk.ChunkHash)
    return nil
}
```

- [x] **Step 2: 运行 handler 测试并确认请求当前返回参数错误**

Run: `go test ./internal/handler -run TestUploadFileChunk_ComputesHashForLegacyClient -count=1`

Expected: FAIL，service mock 未被调用且响应不是成功码。

- [x] **Step 3: 取消 DTO 必填约束并在 handler 中回填 SHA-256**

```go
if strings.TrimSpace(req.ChunkHash) == "" {
    sum := sha256.Sum256(fullData)
    req.ChunkHash = hex.EncodeToString(sum[:])
}
```

- [x] **Step 4: 将 Swagger `chunk_hash` 标记为可选并重新生成文档**

Run: `swag init -g cmd/server/main.go`

- [x] **Step 5: 运行 handler 包测试确认通过**

Run: `go test ./internal/handler -count=1`

### Task 4: Merge 总超时、传输超时与取消

**Files:**
- Modify: `apps/cloud-drive-frontend/src/services/apis/file.ts`
- Test: `apps/cloud-drive-frontend/src/services/apis/file.test.ts`

**Interfaces:**
- Consumes: 调用方可选 `AbortSignal`、Axios request config。
- Produces: 十分钟总等待截止时间、明确的“文件合并超时，请稍后重试”错误、`timeout: 0` 的 merge 请求和可靠的取消清理。

- [x] **Step 1: 写 merge 请求关闭 30 秒超时、总截止时间和等待中取消测试**

```ts
expect(mergeCall?.[2]).toMatchObject({ timeout: 0 })
await vi.advanceTimersByTimeAsync(10 * 60 * 1_000)
expect(await Promise.race([outcome, Promise.resolve('pending')])).toBe(
  'rejected:文件合并超时，请稍后重试',
)
```

- [x] **Step 2: 运行 API 测试并确认缺少 timeout override 与总截止时间**

Run: `pnpm exec vitest run src/services/apis/file.test.ts`

Expected: FAIL，merge config 没有 `timeout: 0`，十分钟后 Promise 仍 pending。

- [x] **Step 3: 用内部 AbortController 合并用户取消与十分钟定时器**

```ts
const mergeController = new AbortController()
const timeout = window.setTimeout(() => {
  timedOut = true
  mergeController.abort()
}, MERGE_MAX_WAIT_MS)
```

- [x] **Step 4: 对每个 merge 请求传 `{ signal: mergeController.signal, timeout: 0 }` 并在 finally 清理**

```ts
signal?.removeEventListener('abort', onAbort)
window.clearTimeout(timeout)
```

- [x] **Step 5: 在等待函数注册 abort listener 后再次检查 `signal.aborted`**

```ts
signal?.addEventListener('abort', onAbort, { once: true })
if (signal?.aborted) onAbort()
```

- [x] **Step 6: 运行前端 API 测试确认通过**

Run: `pnpm exec vitest run src/services/apis/file.test.ts`

### Task 5: 删除三个已确认无调用方的 repository 入口

**Files:**
- Modify: `apps/cloud-drive-backend/internal/repository/file_repo.go`
- Modify: `apps/cloud-drive-backend/internal/repository/file_repo_test.go`

**Interfaces:**
- Consumes: 全仓 `rg` 调用点结果。
- Produces: 删除 `GetUploadTaskByHashAndUserID`、`GetUploadTaskByHashAndUserIDAndFolderID`、`CheckFileExistsInFolder` 及只验证这些死入口的测试。

- [x] **Step 1: 再次确认三个符号只在定义与自身测试出现**

Run: `rg -n "GetUploadTaskByHashAndUserID|GetUploadTaskByHashAndUserIDAndFolderID|CheckFileExistsInFolder" apps/cloud-drive-backend`

- [x] **Step 2: 删除方法及孤立测试**

保留仍有职责不同的 `GetUploadTaskByHash`，避免扩大审查范围。

- [x] **Step 3: 运行 repository 包测试确认通过**

Run: `go test ./internal/repository -count=1`

### Task 6: 全量验证与差异复核

**Files:**
- Review: all files above plus existing staged changes.

**Interfaces:**
- Consumes: 根目录本地门禁与 Git staged/unstaged 状态。
- Produces: 可复核的测试证据和不混淆用户 staged 内容的最终报告。

- [x] **Step 1: 格式化本轮 Go 文件**

Run: `gofmt -w internal/database/migrate.go internal/database/migrate_test.go internal/service/file.go internal/service/file_test.go internal/dto/file.go internal/handler/file.go internal/handler/file_test.go internal/repository/file_repo.go internal/repository/file_repo_test.go`

- [x] **Step 2: 运行后端全量测试与竞态测试**

Run: `go test ./... -count=1`

Run: `go test -race ./... -count=1`

- [x] **Step 3: 运行根目录完整门禁**

Run: `pnpm verify`

- [x] **Step 4: 对比 staged 与 unstaged 差异并确认无无关修改**

Run: `git status --short && git diff --check && git diff --stat && git diff --cached --stat`

- [x] **Step 5: 不执行 git add/commit，向用户报告每条 CR 的结论、改动与验证结果**
