# 普通对话

## 首版范围

入口 `/home/chat`，JWT 登录保护。使用本人配置的对话模型，可切换本次模型而不改变默认配置。单个临时会话，不建新表、不持久化消息、不读取知识库，不调用工具或显示私有思维链。

前端纯文本显示消息，支持多轮、流式回复、停止、复制、手动重试。未完成的回复不进入后续上下文，须重试或新建；新建/切换模型清空消息需确认。重试会移除该回复之后的消息，避免上下文分叉。页面卸载取消请求。

## 服务链路

浏览器 POST `/api/chat` → Go JWT `/chat` → Mastra `/plain-chat` → Go 一次性模型网关 → 配置中的供应商。

Go 根据用户与 provider_models.id 解析模型和凭据，0 表示使用默认。密钥解密后仅由 Go 注入供应商 Authorization，浏览器及 Mastra 不持有供应商 API Key。Mastra 使用独立 Agent，maxSteps=1、maxRetries=0，没有工具、workspace、memory 或业务观测存储；原脚手架 Agent 保留但不参与此链路。

网关授权为随机 256-bit 一次性能力，只在活动请求内有效；请求完成、取消或 120 秒超时撤销。模型 ID 固定为配置值，不接受任意地址或工具字段。授权放请求头，不进入 URL 日志。网关只允许 chat/completions 或 responses，Responses 请求 store=false。

公有模型使用 HTTPS，DNS 结果全部验证后直接拨号到已验证 IP，禁用代理及重定向；私网仅开放部署侧精确允许的基础地址，链路本地/云元数据等地址始终禁止。错误正文不透传第三方响应，避免凭据泄漏。

## 启动

本机 MySQL、Go、Vite 保持原有 `pnpm dev` 流程；另一个终端执行：

```bash
pnpm dev:ai
```

默认 AI URL 为 http://127.0.0.1:4111，内部 Go 回调为 http://127.0.0.1:9000。修改端口时在 Go 环境设置 CHAT_AI_URL、CHAT_BACKEND_URL；AI 进程也需同样的 CHAT_BACKEND_URL（该脚本不会读取根 .env.development）。Mastra 固定本机监听，不将开发 Studio 暴露公网。

本地无认证模型的例子：部署环境变量 MODEL_ALLOWED_PRIVATE_ENDPOINTS=http://127.0.0.1:11434/v1，并在用户配置中使用完全相同的地址。该变量逗号分隔；不得配置宽泛通配符。仍需先设置 MODEL_CONFIG_ENCRYPTION_KEY 才能使用加密保存的 API Key。

本轮只提供原生本机服务链路，现有 Docker Compose 不包含 AI 服务，不能直接视为完整对话部署方案。

## 接口

POST `/chat` JSON：model_id 与 messages，message 只有 role 和 content；角色 user/assistant 交替且以 user 结束，最多 79 条，每条 8000 字符、总计 60000 字符、正文 256 KiB。只允许文字上下文。

200 返回 application/x-ndjson，逐行 JSON：

```json
{"type":"meta","model_id":1,"provider":"供应商","model":"模型 ID"}
{"type":"delta","text":"文字片段"}
{"type":"done"}
```

失败返回 error/message 事件，或开始流前返回标准错误响应。前端必须收到 done 才认为完成，EOF 不等于成功。没有自动付费重试，用户手动重试可能产生费用；取消只能尽力终止上游，不能保证供应商停止计费。

## 验证

前端测试、Lint、类型检查和构建通过；Go 回归测试覆盖用户隔离、消息限制、一次性授权、私网精确允许、供应商请求固定模型和工具过滤。Mastra 无工具 Agent 对本地模拟 OpenAI 流读取通过，未调用收费模型。真实账号、真实供应商及 Responses 协议联调待用户配置后验证。AI 生产构建被运行中的开发服务保护阻止，未覆盖其输出目录；类型与新增文件 Lint 通过。

Go 源码无热更新，需重启 pnpm dev。Brain 中旧名称/测试现状未按代码更新，遵循只读协议以代码与本文为准。
