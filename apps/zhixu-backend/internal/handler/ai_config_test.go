package handler

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"

	"zhixu-backend/internal/model"
	"zhixu-backend/internal/repository"
	"zhixu-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type handlerAIStore struct{}

func (handlerAIStore) Read(_ context.Context, user uint) (*repository.AIConfigState, error) {
	return &repository.AIConfigState{Providers: []model.ModelProvider{{ID: 1, UserID: user, Ciphertext: "private-ciphertext"}}}, nil
}
func (handlerAIStore) Mutate(_ context.Context, user uint, fn func(*repository.AIConfigState) error) (*repository.AIConfigState, error) {
	state := &repository.AIConfigState{Settings: model.UserAISettings{UserID: user}}
	if err := fn(state); err != nil {
		return nil, err
	}
	state.Settings.ConfigVersion++
	return state, nil
}

func TestAIConfigHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	credentials, err := service.NewModelCredentials(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	require.NoError(t, err)
	h := &AIConfigHandler{Service: &service.AIConfigService{Store: handlerAIStore{}, Credentials: credentials}}
	for _, tc := range []struct {
		name, method, path, body string
		user                     uint
		status                   int
	}{
		{"unauthorized", "GET", "/settings", "", 0, 401},
		{"safe response", "GET", "/settings", "", 7, 200},
		{"cross user update", "PUT", "/providers/1", `{"provider_id":"gateway","name":"网关","base_url":"https://example.com/v1","protocol":"chat","auth_type":"none","models":[{"model_id":"chat","kind":"chat"}]}`, 8, 404},
		{"unknown user field", "POST", "/providers", `{"user_id":8}`, 7, 400},
		{"trailing json", "POST", "/providers", `{} {}`, 7, 400},
		{"null input", "POST", "/providers", `null`, 7, 400},
		{"invalid id", "PUT", "/providers/nope", `{}`, 7, 400},
		{"create", "POST", "/providers", `{"provider_id":"gateway","name":"网关","base_url":"https://example.com/v1","protocol":"chat","auth_type":"key","api_key":"secret-marker","models":[{"model_id":"chat","kind":"chat"}]}`, 7, 201},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.user != 0 {
					c.Set("user_id", tc.user)
				}
				c.Next()
			})
			r.GET("/settings", h.Get)
			r.POST("/providers", h.Create)
			r.PUT("/providers/:id", h.Update)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			require.NotContains(t, w.Body.String(), "secret-marker")
			require.NotContains(t, w.Body.String(), "private-ciphertext")
		})
	}
}
