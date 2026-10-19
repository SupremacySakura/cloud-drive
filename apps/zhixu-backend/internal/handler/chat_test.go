package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zhixu-backend/internal/model"
	"zhixu-backend/internal/repository"
	"zhixu-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestChatGatewayUsesServerModelAndSingleUseCredential(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var received map[string]json.RawMessage
	var authorization string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"你好\"}}]}\n\n"))
	}))
	defer provider.Close()
	s := &service.ChatService{AllowedPrivateEndpoints: provider.URL}
	token, revoke, err := s.Register(&service.ChatCapability{
		Context: context.Background(), Key: "server-secret",
		Provider: model.ModelProvider{BaseURL: provider.URL, Protocol: "chat"},
		Model:    model.ProviderModel{ModelID: "real-model"},
	})
	require.NoError(t, err)
	defer revoke()
	h := &ChatHandler{Service: s}
	r := gin.New()
	r.POST("/internal/chat/v1/*endpoint", h.Gateway)
	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/internal/chat/v1/chat/completions", strings.NewReader(`{"model":"attacker-model","messages":[{"role":"user","content":"你好"}],"tools":[{}]}`))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	w := request()
	require.Equal(t, 200, w.Code)
	require.Equal(t, "Bearer server-secret", authorization)
	require.Equal(t, `"real-model"`, string(received["model"]))
	require.NotContains(t, received, "tools")
	require.NotContains(t, w.Body.String(), "server-secret")
	require.Equal(t, 404, request().Code)
}

func TestChatSendAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/chat", (&ChatHandler{}).Send)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/chat", strings.NewReader("{}")))
	require.Equal(t, 401, w.Code)
}

type chatHandlerStore struct{ handlerAIStore }

func (chatHandlerStore) Read(_ context.Context, user uint) (*repository.AIConfigState, error) {
	return &repository.AIConfigState{
		Providers: []model.ModelProvider{{ID: 1, UserID: user, Name: "本地模型", BaseURL: "http://127.0.0.1:11434/v1", AuthType: "none", Protocol: "chat"}},
		Models:    []model.ProviderModel{{ID: 2, ProviderID: 1, UserID: user, Kind: "chat", ModelID: "test-model"}},
		Settings:  model.UserAISettings{ChatModelID: 2},
	}, nil
}

func TestChatSendStreamsViaMastraRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var capability string
	ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/plain-chat", r.URL.Path)
		var payload struct {
			Capability string `json:"capability"`
			Gateway    string `json:"gateway"`
			Model      string `json:"model"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		capability = payload.Capability
		require.Len(t, capability, 64)
		require.Equal(t, "test-model", payload.Model)
		require.Equal(t, "http://127.0.0.1:9000/internal/chat/v1", payload.Gateway)
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte("{\"type\":\"delta\",\"text\":\"你好\"}\n{\"type\":\"done\"}\n"))
	}))
	defer ai.Close()
	s := &service.ChatService{Store: chatHandlerStore{}}
	h := &ChatHandler{Service: s, AIURL: ai.URL, InternalBackendURL: "http://127.0.0.1:9000"}
	r := gin.New()
	r.POST("/chat", func(c *gin.Context) { c.Set("user_id", uint(7)); h.Send(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/chat", strings.NewReader(`{"messages":[{"role":"user","content":"你好"}]}`)))
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), `"type":"meta"`)
	require.Contains(t, w.Body.String(), `"type":"done"`)
	require.NotContains(t, w.Body.String(), capability)
	require.Nil(t, s.Capability(capability))
}
