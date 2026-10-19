import { Agent } from "@mastra/core/agent";
import { registerApiRoute } from "@mastra/core/server";
import { z } from "zod";

const inputSchema = z
  .object({
    model: z.string().min(1).max(200),
    protocol: z.enum(["chat", "responses"]),
    gateway: z.string().url(),
    capability: z.string().regex(/^[a-f0-9]{64}$/),
    messages: z
      .array(
        z.object({
          role: z.enum(["user", "assistant"]),
          content: z.string().min(1).max(8000),
        }),
      )
      .min(1)
      .max(79),
  })
  .strict();

export const plainChatRoute = registerApiRoute("/plain-chat", {
  method: "POST",
  handler: async (c) => {
    const parsed = inputSchema.safeParse(await c.req.json().catch(() => null));
    if (!parsed.success) return c.json({ error: "Invalid chat input" }, 400);
    const input = parsed.data;
    const gateway = new URL(input.gateway);
    // The AI service only calls the local Go gateway. Never accept an arbitrary provider URL.
    const backend = new URL(
      process.env.CHAT_BACKEND_URL || "http://127.0.0.1:9000",
    );
    if (
      gateway.origin !== backend.origin ||
      gateway.pathname !== "/internal/chat/v1" ||
      gateway.username ||
      gateway.password ||
      gateway.search ||
      gateway.hash
    ) {
      return c.json({ error: "Invalid gateway" }, 403);
    }
    const controller = new AbortController();
    const abort = () => controller.abort();
    c.req.raw.signal.addEventListener("abort", abort, { once: true });
    const timeout = setTimeout(abort, 115_000);
    const encoder = new TextEncoder();
    const stream = new ReadableStream<Uint8Array>({
      async start(output) {
        const send = (event: object) =>
          output.enqueue(encoder.encode(JSON.stringify(event) + "\n"));
        try {
          // Standalone agent: no registered tools, workspace, memory, or observability storage.
          const agent = new Agent({
            id: "plain-chat",
            name: "普通对话",
            instructions:
              "你是一个友好的文字对话助手。只根据本次会话回答，不声称访问文件、知识库或调用工具。",
            model: {
              providerId: "openai",
              modelId: input.model,
              url: input.gateway,
              apiKey: input.capability,
              api: input.protocol,
            },
          });
          const messages = input.messages.map((m) =>
            m.role === "user"
              ? { role: "user" as const, content: m.content }
              : { role: "assistant" as const, content: m.content },
          );
          const result = await agent.stream(messages, {
            maxSteps: 1,
            abortSignal: controller.signal,
            modelSettings: { maxRetries: 0 },
          });
          let failed = false;
          const reader = result.fullStream.getReader();
          try {
            for (;;) {
              const { value, done } = await reader.read();
              if (done) break;
              if (value.type === "text-delta")
                send({ type: "delta", text: value.payload.text });
              if (value.type === "error") {
                failed = true;
                break;
              }
            }
          } finally {
            await reader.cancel().catch(() => {});
          }
          if (failed || controller.signal.aborted) {
            send({
              type: "error",
              message: controller.signal.aborted
                ? "模型响应超时或请求已停止"
                : "模型调用失败，请检查密钥、地址、模型 ID 和额度；可手动重试",
            });
          } else send({ type: "done" });
        } catch {
          if (!controller.signal.aborted)
            send({
              type: "error",
              message: "模型调用失败，请检查服务与模型配置后手动重试",
            });
        } finally {
          clearTimeout(timeout);
          c.req.raw.signal.removeEventListener("abort", abort);
          try {
            output.close();
          } catch {
            /* Client already cancelled. */
          }
        }
      },
      cancel() {
        abort();
        clearTimeout(timeout);
      },
    });
    return new Response(stream, {
      headers: {
        "Content-Type": "application/x-ndjson",
        "Cache-Control": "no-store",
      },
    });
  },
});
