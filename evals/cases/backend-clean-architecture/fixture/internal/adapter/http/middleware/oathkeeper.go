package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const IdentityKey = "oathkeeper_identity"

func OathkeeperIdentity() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		identity := ctx.GetHeader("X-User")
		if identity == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Oathkeeper identity"})
			return
		}
		ctx.Set(IdentityKey, identity)
		ctx.Next()
	}
}
