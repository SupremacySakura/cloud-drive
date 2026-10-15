---
task_id: operation-log-20261014
status: completed
created: 2026-10-14
last_reviewed: 2026-10-15
applicable_to: "操作审计日志、活动流 feed、Go+GORM+sqlmock 测试、Vue3 composable 分页加载"
---

# 首页操作日志（替代假"最近活动"）

## 结果

- 新增 `OperationLogModel` 表：动作/目标类型/目标 ID/名称快照/目录快照/JSON detail/创建时间，复合索引 `(user_id, created_at)`；12 个 action 覆盖上传（含秒传）、新建文件夹、重命名、移动、删除、下载（含打包）、预览、分享链接创建/删除、取件码创建/删除。
- **快照优先**：目标名称与所在目录在写入时快照，目标后续改名/移动/删除不影响历史可读性——这是与旧实现（联查 file_models 按 updated_at 排序）的本质区别。
- 写入策略：service 层业务成功后调用 `recordOperation`，写日志失败仅 zerolog 报错不影响主操作，不挂在业务事务内（避免日志失败回滚业务）；每用户保留最近 500 条（先查 cutoff ID 再 DELETE，规避 MySQL 子查询不支持 LIMIT）。
- 新增 `GET /file/logs?page=&page_size=` 分页接口（vo/dto/swagger 全链路）；overview 的 `recent_activities`（假数据源头）全栈移除。
- 前端：`useOperationLogs` composable（AbortController 竞态防护 + 加载更多 + reset）+ `describeOperationLog` 纯函数（12 action → icon/颜色/标题/副标题/高亮，detail 容错解析）+ DashboardPage「操作日志」feed（48px 行沿用 iCloud 视觉，首屏 20 条）。
- 测试：后端 repo 7 + service 17 + handler 3，前端 32（utils 20 + apis 4 + composable 8），共 59 个新用例。

## 验证

- `pnpm verify` 全链路通过：lint、typecheck、前端 16 文件/126 用例、前端生产构建（vue-tsc + vite）、后端构建、后端 6 测试包全过。
- Swagger 三文件由 `swag init` 重新生成，diff 纯增量（+507 行新接口，-77 行删除的旧定义），证明存量文档原本同步。

## 教训

- **假数据要整体替换而不是修补**：旧"最近活动"是 `file_models` 按 `updated_at` 倒序——删除不可见、文件夹/分享/取件码无记录、改名后历史失真。补字段救不了语义缺失，独立日志表才是正解。
- **sqlmock + GORM 四个高频坑**（本项目反复踩）：① `First()` 会追加主键排序子句，查单条用 `Take()` 避免正则失配；② `Offset(0)` 被 GORM 优化掉，分页查询绑定参数个数随 page 变化，page=1 时只有 (user_id, limit) 两个；③ GORM 单条 Create/Update/Delete 默认包隐式事务，测试要补 `ExpectBegin/ExpectCommit`，注入失败场景则是 `ExpectRollback`；④ `Select("id")` 生成带反引号的 ``SELECT `id` ``，正则必须含反引号。
- **跨实例状态不互斥**：composable 的 AbortController 是实例私有的，竞态防护只在同实例连续调用时生效；测试要建模真实用法（同一实例 load 两次），不能开两个实例指望互相 abort。vitest `toHaveBeenLastCalledWith` 对尾参严格，实际传 2 参的调用断言不能补 `undefined`。
- **方法存在 ≠ 功能生效**：CR 复核抓出 `PruneOperationLogs` 实现了方法+测试却没有任何生产调用点（"每用户保留 500 条"形同虚设）。设计文档承诺的策略，必须在写入路径接线并以端到端断言验证，写完方法后回头 grep 一遍调用点。
- **竞态防护要覆盖所有提交路径**：只给 `load()` 配 AbortController 而漏掉 `loadMore()`，翻页请求在 reset/重新加载后仍会把旧数据回灌新列表。共享同一个在途控制器（load/loadMore/reset 任何新操作都 abort 前一个）比逐个补检查更不容易漏。
- **改动 `.vue` 模板后 typecheck 盖不住模板表达式**，必须跑生产 build 才算验证（AGENTS.md 既有约定，本次再次确认）。

## 可沉淀模式

- **日志/审计表设计四件套**：名称快照 + 目录快照 + JSON detail + 每用户保留上限（先查 cutoff 再批量 DELETE）。快照保可读性，JSON 保扩展性，上限保表不膨胀。
- **写入点收敛到单一 helper**（`recordOperation`）：nil 仓库时静默跳过——存量测试用结构体字面量构造 service、不经构造函数，新依赖对它们零影响。这是给已有很多测试的 service 加依赖的低风险姿势。
- **feed 类 UI 的状态分层**：composable 管数据/分页/竞态（可测），页面只做 map 到展示卡片（`operationLogCards` computed），展示语义进纯函数（`describeOperationLog`）——三层各自独立可测。
- **接口删除前全仓 grep 消费方**（含 swagger 生成物与前端类型），前端页面先行切换、后端字段后移除，顺序不能反。
