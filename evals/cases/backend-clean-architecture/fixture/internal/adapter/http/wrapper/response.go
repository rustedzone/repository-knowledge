package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, gin.H{"data": data})
}

func Error(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}
