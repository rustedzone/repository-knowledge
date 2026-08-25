package controller

import (
	"context"
	"net/http"

	"example.invalid/access-service/internal/adapter/http/middleware"
	"example.invalid/access-service/internal/adapter/http/wrapper"
	"github.com/gin-gonic/gin"
)

type ApprovalService interface {
	Approve(ctx context.Context, approvalID, actor string) error
}

type ApprovalController struct {
	usecase ApprovalService
}

func NewApprovalController(usecase ApprovalService) *ApprovalController {
	return &ApprovalController{usecase: usecase}
}

func (controller *ApprovalController) Approve(ctx *gin.Context) {
	actor := ctx.GetString(middleware.IdentityKey)
	if err := controller.usecase.Approve(ctx, ctx.Param("approvalID"), actor); err != nil {
		wrapper.Error(ctx, http.StatusUnprocessableEntity, err)
		return
	}
	wrapper.OK(ctx, gin.H{"status": "PROVISIONED"})
}
