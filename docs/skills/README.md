---
last_reviewed: 2026-04-30
---

# Skills 能力沉淀

本目录用于沉淀可复用的操作模式，积累经过验证的最佳实践。

## 目录职责

- 记录可复用的操作模式（Skill），而非具体实现
- 为常见任务提供标准化的执行指南
- 降低重复探索成本，提高执行效率

## Skill 文件格式规范

每个 Skill 文件需包含 YAML frontmatter 和 Markdown 正文：

```yaml
---
name: skill-name          # kebab-case，唯一标识
trigger: "描述何时使用此 skill"
created: YYYY-MM-DD
last_reviewed: YYYY-MM-DD
applicable_to: [frontend, backend, docs, infra]
---
```

**正文结构：**

```markdown
## 步骤

1. 步骤一
2. 步骤二

## 示例

示例说明

## 边界情况

- 边界情况一
- 边界情况二
```

## 入库流程

1. **AI 生成草稿**：根据任务执行经验，AI 生成新的 Skill 草稿
2. **人工审核确认**：由人工审查草稿的准确性、完整性和适用性
3. **正式写入**：审核通过后写入 `docs/skills/` 目录

## 入库标准

- 至少被 **3 次不同任务**验证过的模式才能正式入库
- 必须是通用的、可复用的操作模式
- 必须有清晰的触发条件和边界情况

## 容量限制

- **文件数上限**：10 个
- **单文件行数上限**：500 行

## 知识衰减策略

- `last_reviewed` 字段标记最后审核时间
- 每 6 个月需重新审核一次，确认仍适用
- 若模式已过时，标记为 deprecated 并保留原文
- 不再适用的 Skill 移至 `archive/` 目录

## 目录结构

```
docs/skills/
├── README.md           # 本文件
├── .template.md        # Skill 文件模板
└── *.md                # 具体 Skill 文件（初始为空）
```