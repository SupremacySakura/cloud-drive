package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"zhixu-backend/internal/response"
	"zhixu-backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIConfigHandler struct{ Service *service.AIConfigService }

func (h *AIConfigHandler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAIConfigInvalid):
		response.FailWithStatus(c, 400, response.CodeInvalidParam, err.Error())
	case errors.Is(err, service.ErrAIConfigConflict):
		response.FailWithStatus(c, 409, response.CodeConflict, err.Error())
	case errors.Is(err, service.ErrAIConfigNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		response.FailWithStatus(c, 404, response.CodeNotFound, "模型配置不存在")
	case errors.Is(err, service.ErrModelKeyUnavailable):
		response.FailWithStatus(c, 503, response.CodeServerError, "服务端尚未配置模型密钥加密，请联系部署管理员")
	default:
		response.FailWithStatus(c, 500, response.CodeServerError, "模型配置服务暂时不可用")
	}
}

func bindAIConfig(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	data, err := io.ReadAll(c.Request.Body)
	data = bytes.TrimSpace(data)
	if err != nil || len(data) == 0 || data[0] != '{' {
		response.FailWithStatus(c, 400, response.CodeInvalidParam, "请提供配置 JSON 对象")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		response.FailWithStatus(c, 400, response.CodeInvalidParam, "配置 JSON 无效或包含未知字段")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		response.FailWithStatus(c, 400, response.CodeInvalidParam, "请提供单个配置 JSON 对象")
		return false
	}
	return true
}

// Get godoc
// @Summary 获取当前用户模型配置（不返回密钥）
// @Tags ai-settings
// @Success 200 {object} response.Response{data=service.AIConfigView}
// @Security ApiKeyAuth
// @Router /settings/ai [get]
func (h *AIConfigHandler) Get(c *gin.Context) {
	user, ok := knowledgeUser(c)
	if !ok {
		return
	}
	view, err := h.Service.Get(c.Request.Context(), user)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, view)
}

// Catalog godoc
// @Summary 获取供应商预设目录（不是实时模型清单）
// @Tags ai-settings
// @Success 200 {object} response.Response
// @Security ApiKeyAuth
// @Router /settings/ai/catalog [get]
func (h *AIConfigHandler) Catalog(c *gin.Context) {
	if _, ok := knowledgeUser(c); !ok {
		return
	}
	response.Success(c, service.AIProviderCatalog())
}

// Create godoc
// @Summary 创建模型供应商及模型清单
// @Tags ai-settings
// @Accept json
// @Param data body service.ProviderInput true "供应商与模型配置"
// @Success 201 {object} response.Response{data=service.AIConfigView}
// @Security ApiKeyAuth
// @Router /settings/ai/providers [post]
func (h *AIConfigHandler) Create(c *gin.Context) { h.saveProvider(c, false) }

// Update godoc
// @Summary 替换供应商配置与完整模型清单，保留同名模型 ID
// @Tags ai-settings
// @Accept json
// @Param id path int true "供应商数据库 ID"
// @Param data body service.ProviderInput true "供应商与模型配置"
// @Success 200 {object} response.Response{data=service.AIConfigView}
// @Security ApiKeyAuth
// @Router /settings/ai/providers/{id} [put]
func (h *AIConfigHandler) Update(c *gin.Context) { h.saveProvider(c, true) }

func (h *AIConfigHandler) saveProvider(c *gin.Context, update bool) {
	user, ok := knowledgeUser(c)
	if !ok {
		return
	}
	var id uint
	if update {
		id, ok = knowledgeID(c, c.Param("id"))
		if !ok {
			return
		}
	}
	var input service.ProviderInput
	if !bindAIConfig(c, &input) {
		return
	}
	view, err := h.Service.SaveProvider(c.Request.Context(), user, id, input)
	if err != nil {
		h.fail(c, err)
		return
	}
	status := 201
	if update {
		status = 200
	}
	response.SuccessWithStatus(c, status, view)
}

type DeleteAIProviderRequest struct {
	ExpectedVersion uint64 `json:"expected_version"`
}

// Delete godoc
// @Summary 删除未被默认模型使用的供应商和模型
// @Tags ai-settings
// @Accept json
// @Param id path int true "供应商数据库 ID"
// @Param data body DeleteAIProviderRequest true "当前配置版本"
// @Success 200 {object} response.Response{data=service.AIConfigView}
// @Security ApiKeyAuth
// @Router /settings/ai/providers/{id} [delete]
func (h *AIConfigHandler) Delete(c *gin.Context) {
	user, ok := knowledgeUser(c)
	if !ok {
		return
	}
	id, ok := knowledgeID(c, c.Param("id"))
	if !ok {
		return
	}
	var input DeleteAIProviderRequest
	if !bindAIConfig(c, &input) {
		return
	}
	view, err := h.Service.DeleteProvider(c.Request.Context(), user, id, input.ExpectedVersion)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, view)
}

// Defaults godoc
// @Summary 保存默认对话与 Embedding 模型，切换向量配置需要确认
// @Tags ai-settings
// @Accept json
// @Param data body service.DefaultsInput true "默认模型和配置版本"
// @Success 200 {object} response.Response{data=service.AIConfigView}
// @Security ApiKeyAuth
// @Router /settings/ai/defaults [put]
func (h *AIConfigHandler) Defaults(c *gin.Context) {
	user, ok := knowledgeUser(c)
	if !ok {
		return
	}
	var input service.DefaultsInput
	if !bindAIConfig(c, &input) {
		return
	}
	view, err := h.Service.SaveDefaults(c.Request.Context(), user, input)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, view)
}
