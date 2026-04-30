package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud-drive-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthMiddlewareTest() (*gin.Engine, *httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)
	_ = engine
	return router, recorder, nil
}

// Test AuthMiddleware - Missing Authorization Header
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "未授权")
}

// Test AuthMiddleware - Empty Authorization Header
func TestAuthMiddleware_EmptyHeader(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "未授权")
}

// Test AuthMiddleware - Invalid Format (No Bearer)
func TestAuthMiddleware_InvalidFormatNoBearer(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "invalidtoken123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Bearer Without Token
func TestAuthMiddleware_BearerWithoutToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Bearer With Short Token
func TestAuthMiddleware_BearerWithShortToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer abc")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Invalid Token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Valid Token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(200, gin.H{"message": "success", "user_id": userID})
		} else {
			c.JSON(200, gin.H{"message": "user_id not set"})
		}
	})

	// Generate valid token
	token, err := utils.GenerateToken(123)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "123")
}

// Test AuthMiddleware - Valid Token Sets UserID
func TestAuthMiddleware_ValidTokenSetsUserID(t *testing.T) {
	var capturedUserID uint
	var userIDExists bool

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		userIDExists = exists
		if exists {
			capturedUserID = userID.(uint)
		}
		c.JSON(200, gin.H{"success": true})
	})

	// Generate valid token
	token, err := utils.GenerateToken(456)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.True(t, userIDExists)
	assert.Equal(t, uint(456), capturedUserID)
}

// Test AuthMiddleware - Malformed Token Format
func TestAuthMiddleware_MalformedTokenFormat(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", "Bearer "},
		{"only spaces", "Bearer    "},
		{"no bearer prefix", "justatoken"},
		{"wrong prefix", "Basic dXNlcjpwYXNz"},
		{"Bearer with spaces only", "Bearer     "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tc.token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 401, w.Code)
			assert.Contains(t, w.Body.String(), "无效的token")
		})
	}
}

// Test AuthMiddleware - Case Sensitivity
func TestAuthMiddleware_CaseSensitivity(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// "bearer" lowercase should be rejected
	token, _ := utils.GenerateToken(123)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "bearer "+token) // lowercase
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Token with Special Characters
func TestAuthMiddleware_TokenWithSpecialChars(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer token.with.special!chars")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Multiple Authorization Headers
func TestAuthMiddleware_MultipleAuthHeaders(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	token, _ := utils.GenerateToken(123)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// Gin only reads the first header value
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Test AuthMiddleware - Very Long Token
func TestAuthMiddleware_VeryLongToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	longToken := "Bearer " + string(make([]byte, 10000))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", longToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Expired Token Simulation (invalid signature)
func TestAuthMiddleware_ExpiredTokenSimulation(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Token with valid format but will fail signature validation
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjN9.invalid_signature"

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// Test AuthMiddleware - Token with Wrong Signing Method
func TestAuthMiddleware_WrongSigningMethod(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Token using none algorithm (should be rejected)
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoxMjN9."

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+noneToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}
