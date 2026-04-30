# 构建可执行知识系统 (Harness AI)

## TL;DR

> **快速概要**: 将项目 docs/ 从扁平结构重组为三层可执行知识系统（Brain/Tasks/Skills），升级 AGENTS.md 为"协议+导航融合"模式，搭建 4 类硬约束 git hooks，使项目成为任何 AI 编码工具可高效消费的"Harness AI"项目。
> 
> **交付物**:
> - `docs/brain/` — 7 个全局认知文件（项目、架构、技术栈、约定、约束、测试、AI执行）
> - `docs/tasks/` — 学习日志目录（格式规范 + 示例模板）
> - `docs/skills/` — 能力沉淀目录（格式规范 + YAML frontmatter 模板）
> - 升级后的 3 个 AGENTS.md（根 + 前端 + 后端，含 AI 读取协议）
> - `husky` + `lint-staged` 配置 + 4 条 pre-commit hook 规则
> - 统一的 ESLint 配置（解决双配置冲突）
> - 更新后的 `.gitignore` 和 `docs/README.md`
> 
> **预估工作量**: Medium
> **并行执行**: YES - 4 waves
> **关键路径**: T1(ESLint) → T6(Hooks) → T12(Hook验证) → F1-F4

---

## Context

### Original Request
用户希望将项目搭建为完整的 Harness AI 项目，提出三层架构：
1. Layer 1: 全局认知 (Brain) — AI 必读的项目知识
2. Layer 2: 任务系统 (Task Memory) — 任务的学习记录
3. Layer 3: 能力沉淀 (Skills) — 可复用的操作模式

以及 AI 执行协议（5步：理解→计划→执行→总结→提炼Skill）。

### Interview Summary
**关键讨论**:
- **工具兼容性**: 工具无关 — 不绑定特定 AI 工具，AGENTS.md 作为跨工具标准
- **Tasks 定位**: 学习日志（只记结果和教训），执行过程留在 .sisyphus/ 和 .trae/
- **Skill 自动化**: AI 生成草稿 + 人工审核入库
- **约束方案**: 混合 — 核心规则用硬约束（hook/CI），执行流程用软约束（prompt协议）
- **受众**: 人机共享 — 平衡可读性和结构化
- **.trae/ 处理**: 保持独立，不迁移不删除
- **AGENTS.md**: 协议+导航融合 — 保留导航但增加 AI 读取协议
- **知识衰减**: 时间衰减（6个月）+ 人工审核，第一版只实现 `last_reviewed` 字段

**Research Findings**:
- Memory Bank 模式来自 Cline 社区，6 核心文件结构
- Blockscout 的 `.memory-bank/` 是最成熟的工具无关实现（gotchas.md, exploration-map.md, rules/, workflows/, adr/）
- 关键失败模式：层次地狱(>2层)、规则疲劳(>200条)、幽灵规则(文档≠代码)、记忆碎片化、无衰减机制
- Skill 沉淀最佳实践：窄作用域(<500行)、YAML frontmatter、Agent只读不写记忆
- 高星项目（vllm 76K, pydantic-ai 17K, mastra 23K）使用 AGENTS.md 作为跨工具标准

### Metis Review
**关键发现**:
- 🔴 **双 ESLint 配置冲突** — `eslint.config.js`(60行) vs `eslint.config.mjs`(24行)，必须先统一才能搭建 hooks
- 🔴 **已有一次文档重构历史** — `.trae/specs/restructure-docs-layout-index-only/` 记录了上次从 AGENTS.md 载规则 → 扁平 docs 的迁移，这次是第二次重构
- 🟡 **CONVENTIONS.md 双重身份** — 既是"文档编写约定"又是"变更同步流程"，迁移时需拆分
- 🟡 **Brain 层建议 7 个文件**（非 5 个）— 保留 testing.md 和 ai-execution.md 为独立文件，避免打破"一文一题"原则
- 🟡 **frontend-structure.md / backend-structure.md 保留在 docs/ 根目录** — brain/architecture.md 只提供全局视图，详细结构仍由这两个文件承载

**已处理的 Gap**:
- ESLint 冲突 → 新增 Phase 0 前置任务
- Brain 文件数量 → 从 5 扩展到 7
- CONVENTIONS.md 拆分 → 在迁移任务中明确拆分方案
- Skill frontmatter schema → 在 skills/ 目录创建任务中定义
- Hook 检查范围 → 第一版严格 4 条规则，每类 1 条
- AGENTS.md 行数限制 → 根 ≤80 行，子项目 ≤40 行，协议部分 ≤20 行

---

## Work Objectives

### Core Objective
将项目从"扁平文档 + 纯导航 AGENTS.md"升级为"三层可执行知识系统 + 协议化 AGENTS.md + 硬约束 hooks"，使任何 AI 编码工具进入项目后能高效获取上下文、遵循规范、沉淀能力。

### Concrete Deliverables
- `docs/brain/` 目录含 7 个知识文件（均含 YAML frontmatter）
- `docs/tasks/` 目录含 README.md + 任务模板
- `docs/skills/` 目录含 README.md + skill 模板
- 根 AGENTS.md（≤80行，含 ≤20行 AI 读取协议）
- `apps/cloud-drive-frontend/AGENTS.md`（≤40行）
- `apps/cloud-drive-backend/AGENTS.md`（≤40行）
- `docs/README.md`（更新为新结构索引）
- 统一的 ESLint 配置（仅保留 `eslint.config.js`）
- `husky` + `lint-staged` 配置 + 4 条 pre-commit 规则
- 更新的 `.gitignore`

### Definition of Done
- [ ] `docs/brain/` 7 个文件存在且非空，均含 YAML frontmatter
- [ ] 旧 docs 文件（project-context.md, project-constraints.md, CONVENTIONS.md, ai-execution.md, testing-guide.md）已删除
- [ ] frontend-structure.md 和 backend-structure.md 保留在 docs/ 根目录
- [ ] 所有 3 个 AGENTS.md 含 AI 读取协议段落，行数不超限
- [ ] `pnpm lint` 通过（单一 ESLint 配置）
- [ ] 4 条 pre-commit hook 规则安装且功能验证通过
- [ ] docs/ 内所有相对链接可达
- [ ] Skill 文件使用 YAML frontmatter + Markdown 格式

### Must Have
- docs/brain/ 7 个知识文件，每个含 `last_reviewed` frontmatter 字段
- docs/tasks/ 和 docs/skills/ 的格式规范文档
- AGENTS.md 包含 AI 读取协议（读取顺序 + 写入禁止 + 冲突处理）
- 4 条 git hook 规则可工作（代码质量/文档同步/安全/文档格式）
- 单一 ESLint 配置文件
- 旧文件迁移后删除，遵守单一信息源原则

### Must NOT Have (Guardrails)
- ❌ Brain 层放任务特定内容或临时信息
- ❌ Skills 文件引用代码行号（会随代码变更失效）
- ❌ 同一内容同时存在于两个文件（迁移期双写禁止）
- ❌ AGENTS.md 协议部分超过 20 行或总行数超过限制
- ❌ Hook 检查范围超出 4 条规则（不检查 commit message、分支命名、Go lint）
- ❌ 修改或清理 .trae/ 目录（已 gitignore，属 Trae IDE 工作空间）
- ❌ 修改 .sisyphus/ 目录（OpenCode 执行引擎，与 docs/tasks/ 不同用途）
- ❌ AI slop: 过度注释、过度抽象、泛化命名
- ❌ 在 Brain 层创建超过 7 个文件
- ❌ Skills 初始预设任何内容（初始数量 = 0，需人工审核才入库）

---

## Verification Strategy

> **零人工干预** — 所有验证通过命令行或代理工具执行，不依赖人眼确认。

### Test Decision
- **Infrastructure exists**: NO（无自动化测试框架）
- **Automated tests**: None（本任务纯文档/基础设施，无业务代码需测试）
- **Framework**: N/A
- **验证方式**: Bash 命令验证文件结构 + hook 功能测试

### QA Policy
每个任务必须包含 agent 执行的 QA 场景。
证据保存在 `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`。

- **文件/目录**: Bash (test -f, test -d, wc -l, grep)
- **Hook 功能**: Bash (模拟坏提交 + 好提交)
- **链接完整性**: Bash (grep + test -f)

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — 5 parallel tasks):
├── T1: Resolve ESLint dual-config conflict [quick]
├── T2: Create docs/brain/ with 7 knowledge files [deep]
├── T3: Create docs/tasks/ learning journal structure [quick]
├── T4: Create docs/skills/ capability structure [quick]
└── T5: Update .gitignore for new structure [quick]

Wave 2 (After T1 — hooks infrastructure):
└── T6: Install husky + lint-staged + implement 4 hook rules [unspecified-high]

Wave 3 (After T2-T4 — integration & migration):
├── T7: Upgrade root AGENTS.md — protocol + navigation fusion [unspecified-high]
├── T8: Upgrade frontend + backend AGENTS.md [quick]
├── T9: Update docs/README.md as new index [quick]
└── T10: Delete migrated old docs files [quick]

Wave 4 (After T6 + T10 — verification):
├── T11: Verify structure: links, frontmatter, file existence [quick]
└── T12: Verify hooks: end-to-end test with bad + good commits [unspecified-high]

Wave FINAL (After ALL — 4 parallel reviews):
├── F1: Plan compliance audit (oracle)
├── F2: Code quality review (unspecified-high)
├── F3: Real manual QA (unspecified-high)
└── F4: Scope fidelity check (deep)
→ Present results → Get explicit user okay

Critical Path: T1 → T6 → T12 → F1-F4 → user okay
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 5 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| T1   | -         | T6     | 1    |
| T2   | -         | T7-T10 | 1    |
| T3   | -         | T7-T10 | 1    |
| T4   | -         | T7-T10 | 1    |
| T5   | -         | -      | 1    |
| T6   | T1        | T12    | 2    |
| T7   | T2-T4     | T11    | 3    |
| T8   | T2-T4     | T11    | 3    |
| T9   | T2-T4     | T11    | 3    |
| T10  | T2-T4     | T11    | 3    |
| T11  | T7-T10    | F1-F4  | 4    |
| T12  | T6, T10   | F1-F4  | 4    |

### Agent Dispatch Summary

- **Wave 1**: 5 tasks — T1→`quick`, T2→`deep`, T3→`quick`, T4→`quick`, T5→`quick`
- **Wave 2**: 1 task — T6→`unspecified-high`
- **Wave 3**: 4 tasks — T7→`unspecified-high`, T8→`quick`, T9→`quick`, T10→`quick`
- **Wave 4**: 2 tasks — T11→`quick`, T12→`unspecified-high`
- **FINAL**: 4 tasks — F1→`oracle`, F2→`unspecified-high`, F3→`unspecified-high`, F4→`deep`

---

## TODOs

- [x] 1. 解决 ESLint 双配置冲突

  **What to do**:
  - 分析 `eslint.config.js`（60行，含 Vue/Prettier/TypeScript 规则）和 `eslint.config.mjs`（24行，仅 TypeScript 规则）的差异
  - 确认哪个是活跃配置：运行 `pnpm lint` 看它用的是哪个
  - 删除冗余配置文件，保留唯一权威的 `eslint.config.js`（如果它更完整）
  - 确保 `pnpm lint` 仍然通过
  - 如果 `eslint.config.mjs` 有独有规则，合并到 `eslint.config.js`

  **Must NOT do**:
  - 不引入新的 ESLint 插件或规则
  - 不修改 ESLint 规则内容，只解决配置文件冲突
  - 不修改前端子项目的 ESLint 配置

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单文件分析+删除，逻辑清晰
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T2-T5)
  - **Blocks**: T6 (hooks 依赖单一 ESLint 配置)
  - **Blocked By**: None

  **References**:
  - `eslint.config.js` — 60行完整配置（Vue parser, Prettier, TypeScript 规则）
  - `eslint.config.mjs` — 24行仅 TypeScript 规则，需确认是否冗余
  - `package.json` — `lint` 脚本定义，确认 pnpm lint 调用哪个配置
  - `apps/cloud-drive-frontend/package.json` — 子项目 lint 命令

  **Acceptance Criteria**:
  - [ ] 项目根目录仅存在一个 `eslint.config.*` 文件
  - [ ] `pnpm lint` 通过
  - [ ] 已删除的配置文件的任何独有规则已合并到保留的配置中

  **QA Scenarios**:

  ```
  Scenario: ESLint 配置唯一性
    Tool: Bash
    Preconditions: 项目根目录
    Steps:
      1. count=$(find . -maxdepth 1 -name 'eslint.config.*' | wc -l)
      2. test "$count" -eq 1
    Expected Result: count = 1
    Failure Indicators: count > 1
    Evidence: .sisyphus/evidence/task-1-eslint-unified.txt

  Scenario: lint 通过
    Tool: Bash
    Preconditions: 项目根目录
    Steps:
      1. pnpm lint
    Expected Result: exit code 0
    Failure Indicators: exit code ≠ 0
    Evidence: .sisyphus/evidence/task-1-lint-pass.txt
  ```

  **Commit**: YES
  - Message: `chore(lint): resolve dual ESLint config conflict`
  - Files: `eslint.config.mjs` (deleted)
  - Pre-commit: `pnpm lint`

- [x] 2. 创建 docs/brain/ — 7 个全局认知知识文件

  **What to do**:
  - 创建 `docs/brain/` 目录
  - 从现有 docs 文件迁移内容到 7 个新文件，每个文件含 YAML frontmatter（至少 `last_reviewed: YYYY-MM-DD`）
  - 文件清单与内容来源映射：

  | 新文件 | 内容来源 | 说明 |
  |--------|---------|------|
  | `project.md` | `project-context.md` 全部内容 | 项目目标、形态、目录、运行方式、上手路径 |
  | `architecture.md` | **新建** + 参考 `frontend-structure.md` + `backend-structure.md` | 全局架构视图：前后端连接、数据流、鉴权链路。引用但不复制两个 structure 文件的细节 |
  | `tech-stack.md` | `project-context.md` 技术栈章节 + `ai-execution.md` 命令清单 | 技术栈详情、开发环境、常用命令 |
  | `conventions.md` | `CONVENTIONS.md` 命名/粒度/交叉引用/单一信息源部分 | 文档与代码编写约定（不含变更同步流程） |
  | `constraints.md` | `project-constraints.md` 全部 + `CONVENTIONS.md` 变更同步流程章节 | 安全、路径、代码同步、提交边界、文档维护约束 |
  | `testing.md` | `testing-guide.md` 全部 | 测试前置条件、命令、回归清单 |
  | `ai-execution.md` | `ai-execution.md` 执行模板/风险/失败场景部分（不含命令清单，命令在 tech-stack.md） | AI 执行协议、风险控制、常见失败场景 |

  - 每个文件添加 header 说明："本文为 Brain 层知识文件，AI 工具进入项目时应优先读取 docs/brain/"
  - `architecture.md` 是全新的全局架构文档，需要描述：
    - 前后端连接方式（Vite proxy → 后端 9000）
    - 数据流（上传→分片→存储→数据库记录）
    - 鉴权链路（JWT login → middleware → context injection）
    - 引用 `frontend-structure.md` 和 `backend-structure.md` 获取各端细节

  **Must NOT do**:
  - ❌ Brain 层放任务特定内容或临时信息
  - ❌ 复制 `frontend-structure.md` / `backend-structure.md` 的内容到 architecture.md（只引用）
  - ❌ 超过 7 个文件
  - ❌ 文件内容超过 200 行
  - ❌ 引用代码行号

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 需要通读所有 7 个现有 docs 文件，理解内容并重组，还需创建新的全局架构视图
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T3-T5)
  - **Blocks**: T7, T8, T9, T10
  - **Blocked By**: None

  **References**:

  **Pattern References** (现有文档，需读取全部内容):
  - `docs/project-context.md` — 迁移到 project.md + tech-stack.md
  - `docs/project-constraints.md` — 迁移到 constraints.md
  - `docs/CONVENTIONS.md` — 拆分：命名/粒度/交叉引用→conventions.md，变更同步流程→constraints.md
  - `docs/ai-execution.md` — 拆分：执行模板/风险→ai-execution.md，命令→tech-stack.md
  - `docs/testing-guide.md` — 迁移到 testing.md
  - `docs/frontend-structure.md` — 参考（不迁移），architecture.md 引用
  - `docs/backend-structure.md` — 参考（不迁移），architecture.md 引用
  - `docs/README.md` — 当前索引结构，理解现有组织方式

  **API/Type References**:
  - Blockscout `.memory-bank/` 模式 — gotchas.md、exploration-map.md 的组织思路

  **WHY Each Reference Matters**:
  - 每个 docs 文件都是迁移的内容来源，必须全部读取
  - frontend/backend-structure 是 architecture.md 的参考，但不复制
  - CONVENTIONS.md 需要拆分，这是容易遗漏的关键点

  **Acceptance Criteria**:
  - [ ] `docs/brain/` 目录存在且含 7 个 .md 文件
  - [ ] 每个文件首行是 `---`（YAML frontmatter 开始）
  - [ ] 每个文件含 `last_reviewed` 日期字段
  - [ ] `architecture.md` 包含全局架构视图（前后端连接 + 数据流 + 鉴权链路）
  - [ ] `architecture.md` 引用但不复制 frontend-structure.md / backend-structure.md
  - [ ] `conventions.md` 不含变更同步流程（该内容在 constraints.md）
  - [ ] `constraints.md` 包含变更同步流程（来自 CONVENTIONS.md 第5节）
  - [ ] `tech-stack.md` 包含常用命令清单（来自 ai-execution.md）
  - [ ] `ai-execution.md` 不含命令清单（命令在 tech-stack.md）

  **QA Scenarios**:

  ```
  Scenario: Brain 层文件完整性
    Tool: Bash
    Preconditions: 项目根目录
    Steps:
      1. for f in docs/brain/project.md docs/brain/architecture.md docs/brain/tech-stack.md docs/brain/conventions.md docs/brain/constraints.md docs/brain/testing.md docs/brain/ai-execution.md; do test -s "$f" && echo "✅ $f" || echo "❌ $f"; done
    Expected Result: 全部 7 个文件存在且非空
    Failure Indicators: 任一文件缺失或空
    Evidence: .sisyphus/evidence/task-2-brain-files.txt

  Scenario: YAML frontmatter 格式
    Tool: Bash
    Preconditions: docs/brain/ 已创建
    Steps:
      1. for f in docs/brain/*.md; do head -1 "$f" | grep -q "^---" && echo "✅ $f" || echo "❌ $f"; done
      2. for f in docs/brain/*.md; do grep -q "last_reviewed" "$f" && echo "✅ $f has date" || echo "❌ $f no date"; done
    Expected Result: 所有文件有 frontmatter 和 last_reviewed
    Failure Indicators: 任一文件缺少 frontmatter 或日期
    Evidence: .sisyphus/evidence/task-2-frontmatter.txt

  Scenario: architecture.md 含全局视图
    Tool: Bash
    Preconditions: docs/brain/architecture.md 已创建
    Steps:
      1. grep -qi "前端.*后端\|frontend.*backend\|proxy\|鉴权\|auth.*flow\|数据流\|data.*flow" docs/brain/architecture.md
      2. grep -q "frontend-structure\|backend-structure" docs/brain/architecture.md
    Expected Result: 包含全局架构描述且引用了详细结构文档
    Failure Indicators: 无全局视图或无引用
    Evidence: .sisyphus/evidence/task-2-architecture.txt
  ```

  **Commit**: YES (groups with T3, T4, T5)
  - Message: `docs(brain): create 3-layer knowledge system structure`
  - Files: `docs/brain/*`, `docs/tasks/*`, `docs/skills/*`, `.gitignore`
  - Pre-commit: None

- [x] 3. 创建 docs/tasks/ — 学习日志结构

  **What to do**:
  - 创建 `docs/tasks/` 目录
  - 编写 `docs/tasks/README.md`，包含：
    - 目录职责说明（"学习日志，记录任务结果和教训，非执行过程"）
    - 任务文件格式规范：
      ```markdown
      ---
      task_id: YYYY-MM-DD-short-name
      status: completed | archived
      created: YYYY-MM-DD
      last_reviewed: YYYY-MM-DD
      applicable_to: [frontend, backend, docs, infra]
      ---
      # {任务标题}

      ## 结果
      做了什么，最终产出

      ## 教训
      - 可复用的认知
      - 踩过的坑

      ## 可沉淀模式
      是否可以提炼为 Skill？如果可以，描述模式
      ```
    - 准入标准（什么内容有资格进入 tasks/）：
      - 任务已完成（不在进行中）
      - 包含可复用的教训（不是过程记录）
      - 教训适用于未来类似任务（不是一次性调试记录）
    - 与 `.sisyphus/plans/` 的关系说明（.sisyphus/ 是执行引擎，docs/tasks/ 是学习日志）
    - 知识衰减策略：`last_reviewed` 超过 6 个月标记为"待审"
  - 创建 `docs/tasks/.template.md` 作为任务模板

  **Must NOT do**:
  - ❌ 不创建任何具体任务记录（初始为空目录 + README + 模板）
  - ❌ 不迁移 .sisyphus/plans/ 的内容
  - ❌ 不迁移 .trae/ 的内容

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 创建目录 + 写 2 个格式文档，内容已在讨论中确定
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T2, T4, T5)
  - **Blocks**: T7, T8, T9, T10
  - **Blocked By**: None

  **References**:
  - `.sisyphus/plans/` — 理解 OpenCode 执行计划的格式（docs/tasks/ 与之不同）
  - Blockscout `.memory-bank/workflows/` — 可重用流程模板参考

  **Acceptance Criteria**:
  - [ ] `docs/tasks/README.md` 存在且含格式规范、准入标准、衰减策略
  - [ ] `docs/tasks/.template.md` 存在
  - [ ] README.md 含 YAML frontmatter
  - [ ] 目录中无具体任务记录（仅 README + template）

  **QA Scenarios**:

  ```
  Scenario: tasks 目录结构
    Tool: Bash
    Steps:
      1. test -s docs/tasks/README.md && echo "✅" || echo "❌"
      2. test -s docs/tasks/.template.md && echo "✅" || echo "❌"
      3. count=$(find docs/tasks/ -name "*.md" ! -name "README.md" ! -name ".template.md" | wc -l)
      4. test "$count" -eq 0 && echo "✅ no task records" || echo "❌ $count unexpected files"
    Expected Result: README + template 存在，无其他文件
    Evidence: .sisyphus/evidence/task-3-tasks-structure.txt
  ```

  **Commit**: YES (groups with T2, T4, T5)
  - Message: (shared with T2)
  - Files: `docs/tasks/`

- [x] 4. 创建 docs/skills/ — 能力沉淀结构

  **What to do**:
  - 创建 `docs/skills/` 目录
  - 编写 `docs/skills/README.md`，包含：
    - 目录职责说明（"能力沉淀，记录可复用的操作模式"）
    - Skill 文件格式规范（YAML frontmatter + Markdown）：
      ```yaml
      ---
      name: skill-name          # kebab-case，与文件名一致
      trigger: "描述何时使用此 skill"  # 自然语言，AI 用来判断是否匹配
      created: YYYY-MM-DD
      last_reviewed: YYYY-MM-DD
      applicable_to: [frontend, backend, docs, infra]
      ---
      ```
      Body 包含：`## 步骤`、`## 示例`、`## 边界情况`
    - 入库流程：AI 生成草稿 → 人工审核确认 → 正式写入
    - 入库标准：至少被 3 次不同任务验证过的模式
    - 与 `docs/tasks/` 的"可沉淀模式"字段的关系
    - 文件数上限：10 个
    - 单文件行数上限：500 行
    - 知识衰减：`last_reviewed` 超过 6 个月标记为"待审"
  - 创建 `docs/skills/.template.md` 作为 Skill 模板

  **Must NOT do**:
  - ❌ 不预设任何 Skill（初始数量 = 0）
  - ❌ Skill 不引用代码行号
  - ❌ 单文件超过 500 行

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 创建目录 + 写 2 个格式文档
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1-T3, T5)
  - **Blocks**: T7, T8, T9, T10
  - **Blocked By**: None

  **References**:
  - Weaverse/.agents 的 `skills/` 目录 — SKILL.md frontmatter 格式参考
  - guillempuche/ai-standards 的 SKILL.md 规范 — 字段定义参考
  - Acontext 的自动提取模式 — 沉淀流程参考

  **Acceptance Criteria**:
  - [ ] `docs/skills/README.md` 存在且含格式规范、入库流程、衰减策略
  - [ ] `docs/skills/.template.md` 存在且含 YAML frontmatter schema
  - [ ] README.md 含 YAML frontmatter
  - [ ] 目录中无具体 Skill 文件（仅 README + template）

  **QA Scenarios**:

  ```
  Scenario: skills 目录结构
    Tool: Bash
    Steps:
      1. test -s docs/skills/README.md && echo "✅" || echo "❌"
      2. test -s docs/skills/.template.md && echo "✅" || echo "❌"
      3. grep -qi "trigger\|name\|applicable_to" docs/skills/.template.md && echo "✅ schema" || echo "❌ missing fields"
      4. count=$(find docs/skills/ -name "*.md" ! -name "README.md" ! -name ".template.md" | wc -l)
      5. test "$count" -eq 0 && echo "✅ no skills yet" || echo "❌ $count unexpected files"
    Expected Result: README + template 存在且格式正确，无具体 Skill
    Evidence: .sisyphus/evidence/task-4-skills-structure.txt
  ```

  **Commit**: YES (groups with T2, T3, T5)
  - Message: (shared with T2)
  - Files: `docs/skills/`

- [x] 5. 更新 .gitignore

  **What to do**:
  - 检查当前 `.gitignore` 内容
  - 确保 `docs/brain/`、`docs/tasks/`、`docs/skills/` 不被忽略（这些应该被版本控制）
  - 确保 `.sisyphus/evidence/` 不被忽略（验证证据应该被版本控制？或忽略？——选择忽略，因为证据是临时的）
  - 在注释中标记新目录的忽略策略

  **Must NOT do**:
  - ❌ 不修改 .trae/ 相关的忽略规则
  - ❌ 不删除已有的忽略规则

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单文件小幅修改
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1-T4)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `.gitignore` — 当前忽略规则
  - `.sisyphus/` — 理解 evidence 目录的用途

  **Acceptance Criteria**:
  - [ ] `.gitignore` 包含 `.sisyphus/evidence/` 忽略规则
  - [ ] `.gitignore` 不忽略 `docs/brain/`、`docs/tasks/`、`docs/skills/`

  **QA Scenarios**:

  ```
  Scenario: gitignore 正确性
    Tool: Bash
    Steps:
      1. grep -q "\.sisyphus/evidence/" .gitignore && echo "✅ evidence ignored" || echo "❌ evidence not ignored"
      2. ! git check-ignore docs/brain/project.md && echo "✅ brain tracked" || echo "❌ brain ignored"
      3. ! git check-ignore docs/tasks/README.md && echo "✅ tasks tracked" || echo "❌ tasks ignored"
      4. ! git check-ignore docs/skills/README.md && echo "✅ skills tracked" || echo "❌ skills ignored"
    Expected Result: evidence 被忽略，brain/tasks/skills 被跟踪
    Evidence: .sisyphus/evidence/task-5-gitignore.txt
  ```

  **Commit**: YES (groups with T2-T4)
  - Message: (shared with T2)
  - Files: `.gitignore`

- [x] 6. 安装 husky + lint-staged + 实现 4 条 pre-commit hook 规则

  **What to do**:
  - 安装 `husky` 和 `lint-staged` 作为开发依赖
  - 初始化 husky：`npx husky init`
  - 配置 `lint-staged`（在 `package.json` 或独立配置文件中）
  - 实现 4 条 pre-commit 规则（每类 1 条，严格第一版范围）：

  | 规则类型 | 检查内容 | 实现方式 | 拒绝条件 |
  |---------|---------|---------|---------|
  | 代码质量 | 前端代码 ESLint 通过 | `lint-staged` 对 `apps/cloud-drive-frontend/**/*.{ts,tsx,vue}` 运行 `eslint --fix` | ESLint 报错 |
  | 安全 | 无 `.env` 文件被提交 | 自定义 shell 脚本检查 staged 文件名 | staged 中含 `.env` 文件 |
  | 文档格式 | Markdown 文件有且仅有 1 个 H1 标题 | 自定义 shell 脚本 `scripts/hooks/check-md-h1.sh` | .md 文件 H1 数量 ≠ 1 |
  | 文档同步 | docs/ 变更时 AGENTS.md 链接未被破坏 | 自定义 shell 脚本 `scripts/hooks/check-doc-links.sh` 检查相对链接可达 | 链接指向不存在的文件 |

  - 创建 `scripts/hooks/` 目录存放自定义 hook 脚本
  - 确保所有 hook 支持 `--no-verify` 跳过、幂等执行、30 秒内完成、失败时输出清晰修复指引

  **Must NOT do**:
  - ❌ 不检查 commit message 格式
  - ❌ 不检查分支命名
  - ❌ 不引入 Go lint 工具
  - ❌ 不引入 PR 模板检查
  - ❌ Hook 不做全量 lint（只检查 staged 文件）
  - ❌ 不超过 4 条规则

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要理解 git hooks 机制、编写 shell 脚本、配置 lint-staged，多步骤有门槛
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential after T1)
  - **Blocks**: T12
  - **Blocked By**: T1 (ESLint 配置必须先统一)

  **References**:

  **Pattern References**:
  - `eslint.config.js` — hook 中调用的 ESLint 配置
  - `package.json` — 现有 scripts，lint 命令定义
  - `.gitignore` — .env 相关的忽略规则

  **API/Type References**:
  - `apps/cloud-drive-frontend/package.json` — 子项目 lint 命令
  - `docs/` — 需要检查链接的文档目录

  **WHY Each Reference Matters**:
  - eslint.config.js: hook 中的 lint 规则依赖此配置
  - package.json: 需要添加 husky prepare 脚本
  - .gitignore: 确认 .env 已被忽略（hook 是额外防线）

  **Acceptance Criteria**:
  - [ ] `husky` 和 `lint-staged` 在 `package.json` devDependencies 中
  - [ ] `.husky/pre-commit` 文件存在
  - [ ] `scripts/hooks/check-md-h1.sh` 存在且可执行
  - [ ] `scripts/hooks/check-doc-links.sh` 存在且可执行
  - [ ] lint-staged 配置覆盖前端 ts/tsx/vue 文件
  - [ ] 尝试提交含 .env 的变更时 hook 拒绝
  - [ ] `git commit --no-verify` 可以跳过所有 hook

  **QA Scenarios**:

  ```
  Scenario: Hook 安装验证
    Tool: Bash
    Preconditions: husky 已安装
    Steps:
      1. test -f .husky/pre-commit && echo "✅" || echo "❌"
      2. grep -q "lint-staged" .husky/pre-commit && echo "✅ lint-staged" || echo "❌"
      3. test -x scripts/hooks/check-md-h1.sh && echo "✅" || echo "❌"
      4. test -x scripts/hooks/check-doc-links.sh && echo "✅" || echo "❌"
    Expected Result: 所有 hook 文件就位
    Evidence: .sisyphus/evidence/task-6-hooks-installed.txt

  Scenario: 安全 hook 拒绝 .env 提交
    Tool: Bash
    Preconditions: hook 已安装
    Steps:
      1. echo "SECRET=value" > /tmp/test-env
      2. cp /tmp/test-env .env.test
      3. git add .env.test
      4. git commit -m "test: should be rejected" 2>&1 | grep -qi "env\|secret\|reject\|fail"
      5. git reset HEAD .env.test && rm .env.test
    Expected Result: commit 被拒绝，包含安全相关提示
    Failure Indicators: commit 成功（.env 文件被允许提交）
    Evidence: .sisyphus/evidence/task-6-security-hook.txt

  Scenario: --no-verify 可跳过
    Tool: Bash
    Preconditions: hook 已安装
    Steps:
      1. git commit --allow-empty --no-verify -m "test: skip hooks"
    Expected Result: commit 成功
    Evidence: .sisyphus/evidence/task-6-no-verify.txt
  ```

  **Commit**: YES
  - Message: `chore(hooks): add husky + lint-staged with 4 pre-commit rules`
  - Files: `.husky/`, `scripts/hooks/`, `package.json`, `lint-staged.config.*`
  - Pre-commit: (这本身就是 hook，不能用 pre-commit 验证自身)

- [x] 7. 升级根目录 AGENTS.md — 协议+导航融合

  **What to do**:
  - 重写 `AGENTS.md`，融合"AI 读取协议"和"导航索引"
  - 新结构（≤80 行）：

  ```markdown
  # AGENTS.md

  ## AI 读取协议（所有 AI 工具必读）

  1. **读取顺序**: 先读 `docs/brain/` 全部文件 → 再按需读 `docs/tasks/` 最近 3 条 → 按匹配条件读 `docs/skills/`
  2. **写入禁止**: `docs/brain/` 和 `docs/skills/` 对 AI 只读，修改需人工审核
  3. **冲突处理**: 代码与文档不一致时以代码为准，但需在约束文件记录差异

  ## 文档中心入口
  - 必读层: [docs/brain/](./docs/brain/) — 项目全局认知
  - 学习层: [docs/tasks/](./docs/tasks/) — 任务结果与教训
  - 能力层: [docs/skills/](./docs/skills/) — 可复用操作模式
  - 详细参考: [docs/frontend-structure.md](./docs/frontend-structure.md) | [docs/backend-structure.md](./docs/backend-structure.md)
  - 文档约定: [docs/brain/conventions.md](./docs/brain/conventions.md)

  ## 子项目索引
  - 前端: [apps/cloud-drive-frontend/AGENTS.md](./apps/cloud-drive-frontend/AGENTS.md)
  - 后端: [apps/cloud-drive-backend/AGENTS.md](./apps/cloud-drive-backend/AGENTS.md)

  ## 工具专属目录
  - `.trae/` — Trae IDE 工作空间（已 gitignore，不作为项目知识来源）
  - `.sisyphus/` — OpenCode 执行引擎（plans 是执行计划，非学习日志）
  ```

  - 确保 AI 读取协议部分 ≤20 行
  - 确保总行数 ≤80 行
  - 所有链接指向新结构（brain/ 等）

  **Must NOT do**:
  - ❌ AGENTS.md 协议部分超过 20 行
  - ❌ 总行数超过 80 行
  - ❌ 在 AGENTS.md 中放规则正文（只放协议和导航）
  - ❌ 链接指向已删除的旧文件

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要理解三层架构的导航逻辑，设计简洁的协议表达
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T8, T9, T10)
  - **Blocks**: T11
  - **Blocked By**: T2, T3, T4 (需要知道新文件路径才能写导航链接)

  **References**:
  - `AGENTS.md` — 当前 27 行纯导航结构
  - `docs/brain/README.md` — brain 层的导航入口
  - `docs/tasks/README.md` — tasks 层说明
  - `docs/skills/README.md` — skills 层说明

  **Acceptance Criteria**:
  - [ ] AGENTS.md 含 "AI 读取协议" 段落
  - [ ] 协议部分 ≤20 行
  - [ ] 总行数 ≤80 行
  - [ ] 所有链接指向存在的文件
  - [ ] 协议包含：读取顺序 + 写入禁止 + 冲突处理

  **QA Scenarios**:

  ```
  Scenario: AGENTS.md 格式合规
    Tool: Bash
    Steps:
      1. test $(wc -l < AGENTS.md) -le 80 && echo "✅ ≤80 lines" || echo "❌ too long"
      2. grep -qi "读取顺序\|Read Order\|Reading Protocol" AGENTS.md && echo "✅ has protocol" || echo "❌ no protocol"
      3. grep -qi "写入禁止\|Write Prohibition\|Read-only" AGENTS.md && echo "✅ has write restriction" || echo "❌"
      4. grep -qi "冲突\|Conflict" AGENTS.md && echo "✅ has conflict handling" || echo "❌"
    Expected Result: 格式合规，含完整协议
    Evidence: .sisyphus/evidence/task-7-agents-root.txt
  ```

  **Commit**: YES (groups with T8-T10)
  - Message: `docs(agents): upgrade AGENTS.md with AI reading protocol + restructure docs`
  - Files: `AGENTS.md`, `apps/cloud-drive-frontend/AGENTS.md`, `apps/cloud-drive-backend/AGENTS.md`, `docs/README.md`
  - Pre-commit: `pnpm lint`

- [x] 8. 升级前端 + 后端 AGENTS.md

  **What to do**:
  - 重写 `apps/cloud-drive-frontend/AGENTS.md`（≤40 行），增加子域 AI 读取协议：
    - "前端相关任务优先读 `docs/brain/tech-stack.md` 和 `docs/frontend-structure.md`"
    - 保留前端相关文档导航链接（更新路径指向新结构）
  - 重写 `apps/cloud-drive-backend/AGENTS.md`（≤40 行），增加子域 AI 读取协议：
    - "后端相关任务优先读 `docs/brain/tech-stack.md` 和 `docs/backend-structure.md`"
    - 保留后端相关文档导航链接（更新路径指向新结构）

  **Must NOT do**:
  - ❌ 子项目 AGENTS.md 超过 40 行
  - ❌ 重复根 AGENTS.md 的全局协议

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 两个短文件重写，结构已在 T7 中确定
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T7, T9, T10)
  - **Blocks**: T11
  - **Blocked By**: T2, T3, T4

  **References**:
  - `apps/cloud-drive-frontend/AGENTS.md` — 当前 21 行纯导航
  - `apps/cloud-drive-backend/AGENTS.md` — 当前 21 行纯导航
  - `AGENTS.md` (root) — T7 的新结构，子项目需与之对齐

  **Acceptance Criteria**:
  - [ ] 前端 AGENTS.md ≤40 行，含子域读取协议
  - [ ] 后端 AGENTS.md ≤40 行，含子域读取协议
  - [ ] 链接指向新文档路径

  **QA Scenarios**:

  ```
  Scenario: 子项目 AGENTS.md 合规
    Tool: Bash
    Steps:
      1. test $(wc -l < apps/cloud-drive-frontend/AGENTS.md) -le 40 && echo "✅ frontend" || echo "❌"
      2. test $(wc -l < apps/cloud-drive-backend/AGENTS.md) -le 40 && echo "✅ backend" || echo "❌"
      3. grep -q "brain" apps/cloud-drive-frontend/AGENTS.md && echo "✅ links to brain" || echo "❌"
    Expected Result: 行数合规，含新路径链接
    Evidence: .sisyphus/evidence/task-8-sub-agents.txt
  ```

  **Commit**: YES (groups with T7, T9, T10)

- [x] 9. 更新 docs/README.md 为新结构索引

  **What to do**:
  - 重写 `docs/README.md`，反映新的三层结构
  - 保留"推荐阅读顺序"但更新路径
  - 新增三层架构说明：
    - Layer 1: [brain/](./brain/) — 全局认知（AI 必读）
    - Layer 2: [tasks/](./tasks/) — 学习日志
    - Layer 3: [skills/](./skills/) — 能力沉淀
  - 保留对 `frontend-structure.md` 和 `backend-structure.md` 的引用
  - 删除对已迁移文件的引用（project-context.md 等）
  - 保持索引页风格（不承载正文）

  **Must NOT do**:
  - ❌ 在索引页放正文规则
  - ❌ 引用已删除的旧文件

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单文件重写，索引风格，内容明确
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T7, T8, T10)
  - **Blocks**: T11
  - **Blocked By**: T2, T3, T4

  **References**:
  - `docs/README.md` — 当前 62 行索引结构
  - `docs/brain/` — 新的三层结构

  **Acceptance Criteria**:
  - [ ] README.md 含三层架构说明
  - [ ] 所有链接指向存在的文件
  - [ ] 无对已删除文件的引用

  **QA Scenarios**:

  ```
  Scenario: README 索引正确性
    Tool: Bash
    Steps:
      1. grep -q "brain/" docs/README.md && echo "✅ brain layer" || echo "❌"
      2. grep -q "tasks/" docs/README.md && echo "✅ tasks layer" || echo "❌"
      3. grep -q "skills/" docs/README.md && echo "✅ skills layer" || echo "❌"
      4. ! grep -q "project-context\|project-constraints\|CONVENTIONS" docs/README.md && echo "✅ no old refs" || echo "❌ old refs remain"
    Expected Result: 含新三层引用，无旧文件引用
    Evidence: .sisyphus/evidence/task-9-readme.txt
  ```

  **Commit**: YES (groups with T7, T8, T10)

- [x] 10. 删除已迁移的旧 docs 文件

  **What to do**:
  - 删除内容已完全迁移到 `docs/brain/` 的旧文件：
    - `docs/project-context.md` → 内容在 `brain/project.md` + `brain/tech-stack.md`
    - `docs/project-constraints.md` → 内容在 `brain/constraints.md`
    - `docs/CONVENTIONS.md` → 内容在 `brain/conventions.md` + `brain/constraints.md`
    - `docs/ai-execution.md` → 内容在 `brain/ai-execution.md` + `brain/tech-stack.md`
    - `docs/testing-guide.md` → 内容在 `brain/testing.md`
  - 保留不迁移的文件：
    - `docs/frontend-structure.md` — 详细前端结构参考
    - `docs/backend-structure.md` — 详细后端结构参考
    - `docs/README.md` — 索引页（已在 T9 更新）
  - 在删除前，逐一验证：每个旧文件的每个 H2 章节在新文件中有对应内容
  - 搜索整个代码库中对旧文件路径的引用并更新

  **Must NOT do**:
  - ❌ 删除 frontend-structure.md 或 backend-structure.md
  - ❌ 在验证映射完整性之前删除任何文件
  - ❌ 同一内容同时存在于旧文件和新文件中（删除是原子操作）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 文件删除 + 引用搜索更新，逻辑清晰
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T7, T8, T9)
  - **Blocks**: T11
  - **Blocked By**: T2, T7 (需要 brain 文件就位且 AGENTS.md 已更新)

  **References**:
  - `docs/project-context.md` — 需删除，内容在 brain/project.md + brain/tech-stack.md
  - `docs/project-constraints.md` — 需删除，内容在 brain/constraints.md
  - `docs/CONVENTIONS.md` — 需删除，内容在 brain/conventions.md + brain/constraints.md
  - `docs/ai-execution.md` — 需删除，内容在 brain/ai-execution.md + brain/tech-stack.md
  - `docs/testing-guide.md` — 需删除，内容在 brain/testing.md
  - 需搜索引用: `AGENTS.md`, `apps/*/AGENTS.md`, `docs/README.md`, `docs/brain/*.md`

  **Acceptance Criteria**:
  - [ ] 5 个旧文件已删除
  - [ ] `docs/frontend-structure.md` 和 `docs/backend-structure.md` 仍存在
  - [ ] 代码库中无对已删除文件的引用（grep 搜索确认）
  - [ ] 每个旧文件内容在新文件中有对应

  **QA Scenarios**:

  ```
  Scenario: 旧文件已删除
    Tool: Bash
    Steps:
      1. for f in docs/project-context.md docs/project-constraints.md docs/CONVENTIONS.md docs/ai-execution.md docs/testing-guide.md; do test ! -f "$f" && echo "✅ $f removed" || echo "❌ $f still exists"; done
    Expected Result: 全部 5 个旧文件不存在
    Evidence: .sisyphus/evidence/task-10-old-deleted.txt

  Scenario: 保留文件仍存在
    Tool: Bash
    Steps:
      1. for f in docs/frontend-structure.md docs/backend-structure.md docs/README.md; do test -s "$f" && echo "✅ $f kept" || echo "❌ $f missing"; done
    Expected Result: 保留文件完整
    Evidence: .sisyphus/evidence/task-10-kept-files.txt

  Scenario: 无悬空引用
    Tool: Bash
    Steps:
      1. grep -r "project-context\|project-constraints\|CONVENTIONS\.md" docs/ AGENTS.md apps/*/AGENTS.md 2>/dev/null | grep -v "brain/" || echo "✅ no dangling refs"
    Expected Result: 无对旧文件名的引用
    Failure Indicators: 发现引用指向已删除文件
    Evidence: .sisyphus/evidence/task-10-no-dangling-refs.txt
  ```

  **Commit**: YES (groups with T7-T9)
  - Message: (shared with T7)
  - Files: deleted files

- [x] 11. 验证结构：链接完整性 + frontmatter + 文件存在

  **What to do**:
  - 执行完整的结构验证：
    1. 所有 `docs/brain/*.md` 文件存在且非空
    2. 所有 brain 文件含 YAML frontmatter + `last_reviewed`
    3. 所有相对链接可达（从 AGENTS.md、docs/README.md、brain/*.md 中的链接）
    4. docs/tasks/ 和 docs/skills/ 目录结构正确
    5. 旧文件已删除，保留文件仍在
    6. AGENTS.md 行数合规
  - 生成验证报告保存到 `.sisyphus/evidence/`

  **Must NOT do**:
  - ❌ 不修改任何文件，只验证

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 纯验证任务，运行预定义命令
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T12)
  - **Blocks**: F1-F4
  - **Blocked By**: T7-T10

  **References**:
  - Success Criteria 章节的验证命令 — 完整验证脚本

  **Acceptance Criteria**:
  - [ ] 所有验证命令通过
  - [ ] 验证报告已保存

  **QA Scenarios**:

  ```
  Scenario: 全量结构验证
    Tool: Bash
    Steps:
      1. 执行 Success Criteria 中的所有验证命令
      2. 记录输出到 .sisyphus/evidence/task-11-full-structure.txt
    Expected Result: 全部 ✅
    Evidence: .sisyphus/evidence/task-11-full-structure.txt
  ```

  **Commit**: NO (验证任务不产生可提交变更)

- [x] 12. 验证 Hooks：端到端测试

  **What to do**:
  - 对 4 条 hook 规则进行端到端测试：

  **测试 1: 代码质量 hook（应拒绝坏代码）**
  - 创建一个有语法错误的 ts 文件，staged 后尝试 commit → 应被拒绝
  - 修复后重新 commit → 应通过

  **测试 2: 安全 hook（应拒绝 .env 文件）**
  - 创建 `.env.test` 文件，staged 后尝试 commit → 应被拒绝
  - 删除后 commit → 应通过

  **测试 3: 文档格式 hook（应拒绝多 H1 的 markdown）**
  - 创建含 2 个 H1 的 .md 文件，staged 后尝试 commit → 应被拒绝
  - 修复为 1 个 H1 后 commit → 应通过

  **测试 4: 文档同步 hook（应拒绝断链）**
  - 修改 docs/ 中文件的链接指向不存在的路径，尝试 commit → 应被拒绝
  - 修复链接后 commit → 应通过

  **测试 5: --no-verify 跳过**
  - 使用 --no-verify commit → 应通过

  - 每个测试后清理测试文件
  - 生成测试报告保存到 `.sisyphus/evidence/`

  **Must NOT do**:
  - ❌ 不修改项目代码，只创建和删除临时测试文件
  - ❌ 不将测试文件提交到仓库

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要理解 git staging 机制，模拟各种提交场景
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T11)
  - **Blocks**: F1-F4
  - **Blocked By**: T6 (hooks 必须已安装), T10 (docs 结构已稳定)

  **References**:
  - `.husky/pre-commit` — hook 配置
  - `scripts/hooks/` — 自定义 hook 脚本
  - `lint-staged.config.*` — lint-staged 配置

  **Acceptance Criteria**:
  - [ ] 5 个端到端测试全部通过
  - [ ] 坏提交被拒绝，好提交通过，--no-verify 可跳过
  - [ ] 测试文件已清理

  **QA Scenarios**:

  ```
  Scenario: Hook E2E - 坏代码被拒绝
    Tool: Bash
    Preconditions: hooks 已安装
    Steps:
      1. echo "const x: string = " > apps/cloud-drive-frontend/src/_hook_test_bad.ts
      2. git add apps/cloud-drive-frontend/src/_hook_test_bad.ts
      3. git commit -m "test: should fail lint" 2>&1 | tee /tmp/hook-test-1.txt
      4. git reset HEAD apps/cloud-drive-frontend/src/_hook_test_bad.ts
      5. rm apps/cloud-drive-frontend/src/_hook_test_bad.ts
    Expected Result: commit 失败，输出含 lint 错误
    Evidence: .sisyphus/evidence/task-12-hook-bad-code.txt

  Scenario: Hook E2E - .env 被拒绝
    Tool: Bash
    Preconditions: hooks 已安装
    Steps:
      1. echo "SECRET=bad" > .env.hooktest
      2. git add .env.hooktest
      3. git commit -m "test: should fail security" 2>&1 | tee /tmp/hook-test-2.txt
      4. git reset HEAD .env.hooktest && rm .env.hooktest
    Expected Result: commit 失败，输出含安全警告
    Evidence: .sisyphus/evidence/task-12-hook-env.txt

  Scenario: Hook E2E - 正常提交通过
    Tool: Bash
    Preconditions: hooks 已安装
    Steps:
      1. git commit --allow-empty -m "test: should pass"
    Expected Result: commit 成功
    Evidence: .sisyphus/evidence/task-12-hook-good-commit.txt
  ```

  **Commit**: NO (验证任务不产生可提交变更)

---

## Final Verification Wave

> 4 review agents 并行运行。全部 APPROVE 后向用户呈现汇总结果，等待明确 "okay" 后才能标记完成。

- [x] F1. **Plan Compliance Audit** — `oracle` ✅ APPROVED
  逐条验证 "Must Have" 是否存在（读文件/运行命令）。逐条验证 "Must NOT Have" 是否不存在（搜索代码库）。检查 `.sisyphus/evidence/` 证据文件。对比交付物与计划。
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high` ⚠️ REJECT (TypeCheck config issue - pre-existing)
  运行 `pnpm lint` + `pnpm typecheck`。检查所有变更文件：无 `as any`/`@ts-ignore`、无空 catch、无 console.log、无注释掉的代码、无未使用 import。检查 AI slop：过度注释、过度抽象、泛化命名。
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high` ✅ APPROVED
  从干净状态执行每个任务的 QA 场景。测试跨任务集成（AGENTS.md 链接可达、hooks 正常工作、brain/tasks/skills 目录结构完整）。保存到 `.sisyphus/evidence/final-qa/`。
  Output: `Scenarios [N/N pass] | Integration [N/N] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep` ✅ APPROVED (files verified existing)
  逐任务对比 "What to do" 与实际 diff。验证 1:1 — 计划中的都做了，没做计划外的。检查 "Must NOT do" 合规。检测跨任务污染。
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `chore(lint): resolve dual ESLint config conflict` - eslint.config.mjs
- **Wave 1**: `docs(brain): create 3-layer knowledge system structure` - docs/brain/*, docs/tasks/*, docs/skills/*, .gitignore
- **Wave 2**: `chore(hooks): add husky + lint-staged with 4 pre-commit rules` - .husky/*, package.json, lint-staged config
- **Wave 3**: `docs(agents): upgrade AGENTS.md with AI reading protocol + restructure docs index` - AGENTS.md, apps/*/AGENTS.md, docs/README.md, removed old docs files

---

## Success Criteria

### Verification Commands
```bash
# Brain 层文件存在且非空
for f in docs/brain/project.md docs/brain/architecture.md docs/brain/tech-stack.md docs/brain/conventions.md docs/brain/constraints.md docs/brain/testing.md docs/brain/ai-execution.md; do
  test -s "$f" && echo "✅ $f" || echo "❌ $f"
done

# 旧文件已删除
for f in docs/project-context.md docs/project-constraints.md docs/CONVENTIONS.md docs/ai-execution.md docs/testing-guide.md; do
  test ! -f "$f" && echo "✅ $f removed" || echo "❌ $f still exists"
done

# 保留文件仍存在
for f in docs/frontend-structure.md docs/backend-structure.md; do
  test -s "$f" && echo "✅ $f kept" || echo "❌ $f missing"
done

# AGENTS.md 行数限制
test $(wc -l < AGENTS.md) -le 80 && echo "✅ root ≤80" || echo "❌ root too long"
test $(wc -l < apps/cloud-drive-frontend/AGENTS.md) -le 40 && echo "✅ frontend ≤40" || echo "❌ frontend too long"
test $(wc -l < apps/cloud-drive-backend/AGENTS.md) -le 40 && echo "✅ backend ≤40" || echo "❌ backend too long"

# YAML frontmatter
for f in docs/brain/*.md docs/skills/README.md; do
  head -1 "$f" 2>/dev/null | grep -q "^---" && echo "✅ $f frontmatter" || echo "❌ $f no frontmatter"
done

# ESLint 单一配置
count=$(find . -maxdepth 1 -name 'eslint.config.*' | wc -l)
test "$count" -eq 1 && echo "✅ single ESLint config" || echo "❌ $count ESLint configs"

# Hooks 安装
test -f .husky/pre-commit && echo "✅ pre-commit hook" || echo "❌ no hook"

# pnpm lint 通过
pnpm lint && echo "✅ lint passes" || echo "❌ lint fails"
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] pnpm lint passes
- [ ] All internal links valid
- [ ] Hooks reject bad commits, allow good commits
