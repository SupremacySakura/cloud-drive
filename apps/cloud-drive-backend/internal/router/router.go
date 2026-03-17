package router

import "github.com/gin-gonic/gin"

func SetUpRouter() *gin.Engine {
	r := gin.Default()
	RegisterUserRouter(r)
	return r
}
