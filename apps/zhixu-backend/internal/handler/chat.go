package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"zhixu-backend/internal/response"
	"zhixu-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	Service            *service.ChatService
	AIURL              string
	InternalBackendURL string
}

func chatEvent(c *gin.Context, event any) {
	data, _ := json.Marshal(event)
	_, _ = c.Writer.Write(append(data, '\n'))
	c.Writer.Flush()
}

// Send godoc
// @Summary 临时无工具多轮对话，返回 NDJSON 流
// @Tags chat
// @Accept json
// @Produce application/x-ndjson
// @Param data body service.ChatInput true "用户模型 ID 与会话"
// @Success 200 {string} string "meta / delta / done / error 事件"
// @Security ApiKeyAuth
// @Router /chat [post]
func (h *ChatHandler) Send(c *gin.Context) {
	user, ok := knowledgeUser(c)
	if !ok {
		return
	}
	var input service.ChatInput
	if !bindAIConfig(c, &input) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()
	cap, err := h.Service.Resolve(ctx, user, input)
	if err != nil {
		(&AIConfigHandler{}).fail(c, err)
		return
	}
	token, revoke, err := h.Service.Register(cap)
	if err != nil {
		response.FailWithStatus(c, 500, response.CodeServerError, "对话服务不可用")
		return
	}
	defer revoke()
	payload, _ := json.Marshal(gin.H{
		"messages": input.Messages, "model": cap.Model.ModelID, "protocol": cap.Provider.Protocol,
		"gateway":    strings.TrimRight(h.InternalBackendURL, "/") + "/internal/chat/v1",
		"capability": token,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(h.AIURL, "/")+"/plain-chat", bytes.NewReader(payload))
	if err != nil {
		response.FailWithStatus(c, 503, response.CodeServerError, "AI 服务地址配置无效")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// No automatic retries: a retry could duplicate a paid model invocation.
	client := &http.Client{Transport: &http.Transport{Proxy: nil, ResponseHeaderTimeout: 15 * time.Second}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	upstream, err := client.Do(req)
	if err != nil {
		response.FailWithStatus(c, 503, response.CodeServerError, "AI 服务未启动或连接超时，请检查 Mastra 服务")
		return
	}
	defer upstream.Body.Close()
	if upstream.StatusCode != 200 || !strings.Contains(upstream.Header.Get("Content-Type"), "application/x-ndjson") {
		response.FailWithStatus(c, 503, response.CodeServerError, "AI 服务暂时不可用")
		return
	}
	c.Header("Content-Type", "application/x-ndjson; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Accel-Buffering", "no")
	c.Status(200)
	chatEvent(c, gin.H{"type": "meta", "model_id": cap.Model.ID, "model": cap.Model.ModelID, "provider": cap.Provider.Name})
	buffer := make([]byte, 4096)
	for {
		n, readErr := upstream.Body.Read(buffer)
		if n > 0 {
			if _, err := c.Writer.Write(buffer[:n]); err != nil {
				return
			}
			c.Writer.Flush()
		}
		if readErr != nil {
			if readErr != io.EOF && ctx.Err() == nil {
				chatEvent(c, gin.H{"type": "error", "message": "回复连接中断，请手动重试"})
			}
			return
		}
	}
}

// Gateway uses a short-lived, unguessable capability bound to the active authenticated request.
// It never exposes the provider key to Mastra, nor accepts arbitrary target URLs or models.
func (h *ChatHandler) Gateway(c *gin.Context) {
	cap := h.Service.ClaimCapability(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	if cap == nil {
		c.Status(404)
		return
	}
	path := strings.TrimPrefix(c.Param("endpoint"), "/")
	expected := "chat/completions"
	if cap.Provider.Protocol == "responses" {
		expected = "responses"
	}
	if path != expected {
		c.Status(404)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512*1024)
	var body map[string]json.RawMessage
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.Status(400)
		return
	}
	// Only language generation fields are forwarded; no tools or remote resources.
	clean := map[string]json.RawMessage{}
	for _, key := range []string{"messages", "input", "stream", "temperature", "top_p", "max_tokens", "max_output_tokens", "stream_options"} {
		if value, ok := body[key]; ok {
			clean[key] = value
		}
	}
	clean["model"], _ = json.Marshal(cap.Model.ModelID)
	clean["stream"] = json.RawMessage("true")
	if expected == "responses" {
		clean["store"] = json.RawMessage("false")
	}
	payload, _ := json.Marshal(clean)
	ctx, cancel := context.WithCancel(cap.Context)
	defer cancel()
	client, cleanup, err := h.Service.ModelClient(ctx, cap.Provider.BaseURL)
	if err != nil {
		c.JSON(403, gin.H{"error": gin.H{"message": "模型地址未获允许或无法解析"}})
		return
	}
	defer cleanup()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cap.Provider.BaseURL, "/")+"/"+expected, bytes.NewReader(payload))
	if err != nil {
		c.Status(400)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if cap.Key != "" {
		req.Header.Set("Authorization", "Bearer "+cap.Key)
	}
	upstream, err := client.Do(req)
	if err != nil {
		c.JSON(504, gin.H{"error": gin.H{"message": "模型连接失败或超时"}})
		return
	}
	defer upstream.Body.Close()
	if upstream.StatusCode != 200 {
		// Never forward provider error bodies: they can contain credentials or request content.
		c.JSON(upstream.StatusCode, gin.H{"error": gin.H{"message": "模型供应商拒绝请求，请检查认证、地址、模型与额度"}})
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Accel-Buffering", "no")
	c.Status(200)
	buffer := make([]byte, 4096)
	for {
		n, err := upstream.Body.Read(buffer)
		if n > 0 {
			if _, e := c.Writer.Write(buffer[:n]); e != nil {
				return
			}
			c.Writer.Flush()
		}
		if err != nil {
			return
		}
	}
}
