# DESIGN.md — 浅色 Apple 风格设计规范

> 本文是前端 UI 改造的**唯一设计权威来源**。所有视觉与动效决策以本文为准；
> 本文未覆盖的场景，回到 Apple 设计原则推导，而不是凭感觉新增样式。
>
> - 范围：`apps/cloud-drive-frontend`
> - 目标：浅色优先（light-first）的类 Apple 设计，界面"像物理世界一样即时、连续、可打断"
> - 基线技术栈：Vue 3 + Tailwind CSS v4 + `@iconify/vue`（见 `docs/brain/tech-stack.md`）

---

## 1. 设计原则（改造的判断依据）

改造中每个决策都要能对应到下面的一条原则（源自 Apple《Designing Fluid Interfaces》，WWDC 2018）：

1. **响应（Response）** — 消灭延迟。按钮在 pointer-down 瞬间给出高亮，不等 click；反馈在交互过程中持续存在，不在结束时才出现。
2. **直接操纵（Direct Manipulation）** — 拖什么动什么，1:1 跟随手指/指针，且尊重按下点的偏移。
3. **可打断性（Interruptibility）** — 最重要的原则。任何动画在任意时刻都能被抓取、反转；**永远从当前呈现值出发**，不是从目标值出发。
4. **弹簧优先（Springs）** — 手势驱动的东西用弹簧而非固定时长动画。弹簧天然携带速度、可重定向。
5. **速度交接（Velocity Handoff）** — 手势结束时，动画以手指的瞬时速度继续，拖和动画之间无接缝。
6. **动量投影（Momentum Projection）** — 松手后按速度投影落点，吸附到投影最近的边界，而不是从松手点找最近边界。
7. **空间一致性** — 从哪进就从哪出；弹层从触发它的元素生长（`transform-origin` 锚定触发源）。
8. **橡皮筋（Rubber-banding）** — 边界处渐进抵抗，不硬停。
9. **材质与深度** — 用半透明毛玻璃材质做浮动层，内容在底下滚动；材质厚度表达层级。
10. **动效服务于人** — 默认无回弹（damping 1.0）；只有用户"扔"出去的东西才允许回弹（damping ≈ 0.8）。

---

## 2. 设计令牌（Design Tokens）

所有数值进 Tailwind v4 的 `@theme`（权威实现：`apps/cloud-drive-frontend/src/style.css`），
页面/组件代码**禁止**散落硬编码色值与魔法数字。旧 `tailwind.config.js` 已废弃（文件内保留迁移对照，不再被引用）。

> **实现备注（Tailwind v4 树摇）**：v4 只把"被用到的" `@theme` 变量编译进产物。模板里用 `text-display`、`shadow-card`、`bg-material` 等工具类时会自动生成并发射对应变量；但在 Vue SFC `<style>` 里直接写 `var(--color-material)` 这类自定义 CSS 时，若模板中没有用到对应工具类，变量会被树摇掉、样式静默失效。约定：**令牌一律通过工具类消费**；确需自定义 CSS 引用变量时，给该元素同时挂上对应工具类（如 `bg-material`）确保变量被发射。存量兼容令牌 `--color-background-light/dark`、`--font-display` 仍可用，新代码禁用。

### 2.1 颜色 — 浅色语义令牌

| 令牌 | 值 | 用途 |
| --- | --- | --- |
| `--color-canvas` | `#f5f5f7` | 应用主背景（Apple 浅灰画布） |
| `--color-surface` | `#ffffff` | 卡片、弹层、面板 |
| `--color-surface-secondary` | `#f2f2f7` | 分组列表底、hover 底、输入框底 |
| `--color-surface-tertiary` | `#eaeaef` | 禁用/占位填充、分段控件底 |
| `--color-material` | `rgba(255,255,255,0.72)` | 浮动导航/工具栏材质（配 blur 20px saturate 180%） |
| `--color-material-thick` | `rgba(255,255,255,0.85)` | 滚动内容下方的小型浮动控件 |
| `--color-hairline` | `rgba(0,0,0,0.08)` | 1px 分隔线、卡片描边 |
| `--color-label` | `rgba(60,60,67,1)` | 主文本 |
| `--color-label-secondary` | `rgba(60,60,67,0.60)` | 次要文本（元数据、说明） |
| `--color-label-tertiary` | `rgba(60,60,67,0.33)` | 占位、禁用文本 |
| `--color-scrim` | `rgba(0,0,0,0.40)` | 模态遮罩 |

品牌与功能色（保留现有主色，语义色对齐 iOS 系统色）：

| 令牌 | 值 | 用途 |
| --- | --- | --- |
| `--color-primary` | `#10b674` | 品牌绿：主按钮、选中态、进度、链接 |
| `--color-primary-hover` | `#0da267` | 主按钮 hover |
| `--color-primary-pressed` | `#0b8f5b` | 主按钮按下 |
| `--color-primary-tint` | `rgba(16,182,116,0.12)` | 图标底、选中底、tag 底 |
| `--color-danger` | `#ff3b30` | 删除、危险操作 |
| `--color-warning` | `#ff9500` | 警告（进度、图标） |
| `--color-warning-tint` | `rgba(255,149,0,0.12)` | 警告底（tag、图标底） |
| `--color-warning-strong` | `#c93400` | 警告文字（ tint 底上的深橙，保证对比度） |
| `--color-info` | `#007aff` | 提示性链接、信息态 |

规则：
- 大面积背景只用 `canvas/surface` 两级；颜色永远放在**实心层**上，不放在半透明前景上。
- 浅色材质上不用浅灰文字 —— 用高对比 label 色 + 略重字重保证"活力感（vibrancy）"可读。
- 禁止浅色材质叠浅色材质（legibility 会崩）。

### 2.2 字体

```css
--font-sans: -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Inter',
  'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', system-ui, sans-serif;
```

字级（字距随字号变化 —— 大字号收紧、小字号微放，**禁止全站一个 letter-spacing**）：

| 令牌 | 字号/行高 | 字重 | 字距 | 用途 |
| --- | --- | --- | --- | --- |
| `display` | 34 / 40 | 700 | -0.02em | 登录页大标题、仪表盘欢迎语 |
| `title-1` | 28 / 34 | 700 | -0.02em | 页面大标题 |
| `title-2` | 22 / 28 | 700 | -0.01em | 页面标题（文件管理等） |
| `title-3` | 20 / 25 | 600 | -0.01em | 卡片标题、弹窗标题 |
| `headline` | 17 / 24 | 600 | 0 | 列表主文案、按钮文字 |
| `body` | 15 / 22 | 400 | 0 | 正文、描述 |
| `subhead` | 14 / 20 | 400 | +0.01em | 次要说明、表单项 |
| `caption` | 12 / 16 | 400 | +0.02em | 元数据（大小、时间）、角标 |

规则：层级靠**字重 + 字号 + 行高组合**建立，不靠颜色花招；间距用 rem/em 随字缩放。

### 2.3 圆角、间距、阴影

圆角（Apple 大圆角语言）：

| 令牌 | 值 | 用途 |
| --- | --- | --- |
| `radius-sm` | `8px` | 小按钮、tag、输入框 |
| `radius-md` | `10px` | 卡片、列表行（Apple 标准 10pt） |
| `radius-lg` | `14px` | 弹窗内卡片、预览容器 |
| `radius-xl` | `20px` | 底部抽屉、大型弹层 |
| `radius-full` | `9999px` | 主按钮、胶囊 tag、Toast |

间距：4px 基准网格；页面水平内边距 24px（移动端 16px）；卡片内边距 16–20px；区块间距 24 / 32px。

阴影（柔和、双层、低透明度）：

| 令牌 | 值 | 用途 |
| --- | --- | --- |
| `shadow-card` | `0 1px 2px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06)` | 卡片 |
| `shadow-popover` | `0 2px 8px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.12)` | 菜单、弹窗 |
| `shadow-floating` | `0 4px 24px rgba(0,0,0,0.10)` | 悬浮面板（上传面板） |

规则：阴影随表面增大而加深 —— 弹层比小卡片"更厚"。

### 2.4 动效令牌

| 令牌 | 值 | 用途 |
| --- | --- | --- |
| `duration-micro` | `100–150ms` | hover、按下反馈 |
| `duration-standard` | `200–250ms` | 颜色/透明度状态切换 |
| `ease-out` | `cubic-bezier(0.22, 0.61, 0.36, 1)` | 进入 |
| `ease-in-out` | `cubic-bezier(0.65, 0, 0.35, 1)` |  reversible 切换 |
| 镜像规则 | 反向用反转的贝塞尔控制点 | 进/出沿同一路径 |

弹簧参数（设计师参数：damping / response，**response 不是 duration**）：

| 场景 | damping | response |
| --- | --- | --- |
| 默认 UI（无回弹） | `1.0` | `0.3–0.4` |
| 拖放/ reposition | `1.0` | `0.4` |
| 抽屉/底部弹层 | `0.8` | `0.3` |
| 有动量的甩动（flick） | `0.8` | `0.3–0.4` |

技术映射：
- 纯状态切换（hover/选中/展开）→ CSS transition，用 2.4 的 duration/easing。
- **手势驱动（拖拽、抽屉滑动、滑动删除）→ 必须用弹簧库**。Vue 生态选 `motion-v`（Motion 官方 Vue 版），手势结束时传入松手速度 `velocity`；不要用 CSS transition/keyframes 做手势动画（无法中途抓取反转）。
- 手势落点用动量投影：`projected = current + (v/1000) * d / (1 - d)`，`d = 0.998`，吸附到投影最近的边界。

---

## 3. 材质与层级

浮动功能层（顶部导航、工具栏、上传面板）做成**半透明材质层**，内容从底下滚动穿过：

```css
.material {
  background: var(--color-material);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
}
```

- 材质**入场要"实体化"**：blur 半径 + scale 一起动（如 `scale 0.96→1` + blur `0→20px`），不是单纯淡入。
- 滚动交界不用硬线：浮动层与内容重叠处用渐变遮罩淡出小条 blur 边。
- 模态任务：表面 + `scrim` 遮罩，背景内容压暗；非阻塞并行面板（上传面板）：只偏移 + 半透明，**不加遮罩**。
- 层级（z 序）：基础内容 0 → 粘性工具栏 20 → 悬浮面板 30 → 下拉菜单 40 → 弹窗 50 → Toast 60。

---

## 4. 交互反馈规范

- **按下即反馈**（pointer-down，不是 click）：

```css
.btn-primary { transition: transform 100ms ease-out, background-color 150ms; }
.btn-primary:active { transform: scale(0.97); background: var(--color-primary-pressed); }
```

- 点击热区 ≥ 40px（尽量 44px），目标周围留 ~10px 容差；允许"拖出取消、拖回恢复"。
- 列表行按下用底色变化（`surface-secondary`），整行不做 scale。
- 手势判定：先按 ~10px 迟滞阈值确认方向，再 1:1 跟踪；拖动全程连续反馈，不许只在松手时动画。
- 边界拖拽用橡皮筋函数，不硬停：

```js
const rubberband = (o, d, c = 0.55) => (o * d * c) / (d + c * Math.abs(o));
```

---

## 5. 组件改造规范（对照现有组件）

### 5.1 全局框架

- 顶部导航/侧边栏：`material` 半透明，下方内容滚动穿越；选中项用 `primary-tint` 底 + 品牌绿字。
- 页面最大内容宽 1280px 水平居中，面包屑/标题左对齐。

### 5.2 页面

| 页面 | 改造要点 |
| --- | --- |
| `LoginPage` / `RegisterPage` | 居中卡片（radius-xl、shadow-card），画布浅灰；display 级欢迎标题 + secondary 副文案；主按钮全圆角品牌绿，按下瞬时缩放；错误 inline 校验（不要提交时才报）；成功反馈即时。 |
| `HomePage` / `DashboardPage` | 存储用量用品牌绿环形进度；卡片 hover 微浮起（shadow-card → 加深）；数字用 title 级 + 紧行高。 |
| `FileManagement` | 见 5.3。 |
| `UploadFile` | 大拖放区：虚线描边 + tertiary 底，拖拽悬停时 `primary-tint` 描边品牌绿并轻微放大；文件拖入用橡皮筋反馈。 |
| `FilePickup` / `PickupCodes` | 取件码输入框大字号、等宽数字、分组样式；成功状态用绿色对勾动效（motion 画圈），失败 shake。 |

### 5.3 文件管理域（`components/file-management/`）

- `FileManagerHeader`：title-2 页标题 + secondary 副标题，操作按钮居右成组。
- `FileToolbar`：粘性 `material` 工具栏（内容从底下滚过）；图标按钮 36px 热区；视图切换（列表/网格）用**分段控件**（surface-tertiary 底 + 选中白 pill），切换时两视图交叉淡入 + 内容微位移。
- `FileListView`：iCloud Drive 风格行 —— 行高 48px，leading 文件图标放 `radius-sm` + `primary-tint`（图片/文档类型可用不同浅色底），主文案 headline + 元数据 subhead/secondary，分隔线从图标后内缩（hairline）；hover `surface-secondary`；行尾更多按钮唤起操作菜单。
- `FileGridView`：白卡 `radius-md` + shadow-card；图标大、名称两行截断；选中态品牌绿描边环；hover 轻微上浮加深阴影。
- `FileActionMenu` / 右键菜单：从触发元素生长（`transform-origin` 锚定触发点，`scale 0.95→1` + 淡入，弹簧 damping 1.0 / response 0.3）；destructive 操作置底、danger 色。
- `FileNameDialog` / `MoveFileDialog` / `ConfirmDialog`：桌面端中心弹窗（scale 0.96→1 + 材质模糊背景压暗）；移动端转底部抽屉（可拖拽下滑关闭，速度交接 + 动量投影决定归位还是关闭）；Confirm 只有真正不可逆操作才用（删除），文案直接具体（"删除 3 个文件？"而非"确定吗？"）。
- `FilePreviewDialog`：图片预览支持双指/滚轮缩放拖拽（弹簧回归）；底部操作条用 `material-thick`。
- `UploadTaskPanel`：右下角悬浮面板（shadow-floating、radius-lg、material），可折叠（弹簧）；进度条品牌绿；**可拖拽换位**，边界橡皮筋。
- `ToastNotice`：顶部居中胶囊（material-thick、radius-full），成功/失败图标 + 文案，弹簧滑入；同一时间最多一条。
- `FilePagination`：极简分页；当前页 `primary-tint` 圆 pill。
- `FileEmptyState`：居中插画图标（Iconify）+ title-3 + secondary 说明 + 主按钮 CTA。

图标规则：`@iconify/vue` 全站**统一一个图标族**（推荐圆润线性族），工具栏 20px、行内 16–20px，线宽观感一致；文件类型图标底用浅色 tint。

---

## 6. 可访问性（内置，不是补丁）

组件必须同时响应三个信号：

```css
@media (prefers-reduced-motion: reduce) {
  /* 滑动/弹簧 → 短交叉淡入；去掉回弹 */
  .sheet { transform: none !important; transition: opacity 200ms ease; }
}
@media (prefers-reduced-transparency: reduce) {
  .material { background: var(--color-surface); backdrop-filter: none; }
}
@media (prefers-contrast: more) {
  .card { border: 1px solid var(--color-hairline); background: var(--color-surface); }
}
```

另：对比度达 WCAG AA；不循环慢速大幅动效；暗色模式保留 class 方案能力，但**本次改造以浅色为基准**，暗色令牌后续单独定义，不在浅色改造中夹带。

---

## 7. 实施路线图

> **状态（2026-09）：已全部完成。** 实际执行与路线的偏差记录：
> - 动效层引入 `motion-v`（Motion / AnimatePresence），全局降级由 `App.vue` 的 `<MotionConfig reduced-motion="user">` + `style.css` 三段 prefers-* 媒体查询兜底；
> - 上传面板实现了折叠 + 拖把手换位（拖把手用 `useDragControls`，不干扰列表滚动）；导航抽屉为左滑拖拽关闭（速度方向优先判定）；
> - 已知折衷：部分图标按钮热区为 36px（工具栏标准），略低于 §4 的 40px 目标；`text-label-tertiary`（对比度 ~3:1）仅用于占位/元数据等非关键文本。

1. **令牌层**：`style.css` 引入 `@theme` 定义第 2 节全部令牌；梳理旧 `tailwind.config.js`，新代码一律用令牌（旧值标记 deprecated）。
2. **基础组件**：`ui/ConfirmDialog`、`ToastNotice`、按钮/输入原子样式。
3. **页面改造顺序**：`Login` → `Register` → `Home/Dashboard` → `FileManagement`（含全部子组件）→ `UploadFile` → `Pickup` 系。
4. **动效层**：引入 `motion-v`，把手势场景（抽屉关闭、上传面板拖拽、拖放区）换成弹簧 + 速度交接。
5. **验收**：`pnpm typecheck`、`pnpm lint`、`pnpm build`（在 `apps/cloud-drive-frontend`）全绿 + 走查清单（§8）。

## 8. 走查清单（每个页面合并前过一遍）

- [ ] 所有颜色/圆角/间距/字号来自令牌，无硬编码
- [ ] 按钮在 pointer-down 瞬间有反馈
- [ ] 弹层从触发源生长，进出同路径
- [ ] 手势驱动的动效用弹簧且可中途打断；其余动效可解释 duration/easing 来源
- [ ] 浮动层为半透明材质，无硬分隔线，无"浅材质叠浅材质"
- [ ] 三种 prefers-* 媒体查询行为正确
- [ ] 点击热区 ≥ 40px，文本对比度 AA
- [ ] typecheck / lint / build 通过

---

## 9. 速查表

| 需求 | 做法 | 值 |
| --- | --- | --- |
| 默认动效 | 无回弹弹簧 | damping 1.0 / response 0.3–0.4 |
| 甩动落点 | 动量投影 | `current + (v/1000)·d/(1−d)`，d=0.998 |
| 可打断 | 从当前呈现值出发 | 读实时 transform，不用目标值 |
| 拖动手势 | Pointer Events + capture | 尊重按下点偏移 |
| 边界 | 橡皮筋 | 渐进抵抗 |
| 浮动层 | 半透明材质 | `rgba(255,255,255,0.72)` + blur 20px saturate 180% |
| 卡片 | 白底 + 柔和双层阴影 | radius 10px |
| 主按钮 | 品牌绿 pill | `#10b674`，按下 0.97 缩放 |
| 大标题字距 | 收紧 | -0.02em；正文 0；小字 +0.01~0.02em |

> 原则出处：Apple《Designing Fluid Interfaces》（WWDC 2018）、《The Details of UI Typography》（WWDC 2020）。
> 冲突处理：本文与代码不一致时以本文为准推进改造，改造完成后的新基线回填到 `docs/frontend-structure.md` 的 UI 约定一节。
