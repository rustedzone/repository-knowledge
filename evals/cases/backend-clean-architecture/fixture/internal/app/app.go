package app

import (
	"database/sql"

	"example.invalid/access-service/internal/adapter/http/controller"
	"example.invalid/access-service/internal/adapter/http/middleware"
	"example.invalid/access-service/internal/adapter/keto"
	"example.invalid/access-service/internal/adapter/postgres"
	rediscache "example.invalid/access-service/internal/adapter/redis"
	"example.invalid/access-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

func New() (*gin.Engine, error) {
	database, err := sql.Open("postgres", "fixture-dsn")
	if err != nil {
		return nil, err
	}
	approvalRepository := postgres.NewPostgresApprovalRepository(database)
	groupRepository := postgres.NewPostgresGroupRepository(database)
	approvalCache := rediscache.NewApprovalCache()
	ketoClient := keto.NewKetoTupleClient()
	approvalUsecase := usecase.NewApprovalUsecase(approvalRepository, approvalCache, ketoClient)
	_ = usecase.NewGroupUsecase(groupRepository, ketoClient)
	approvalController := controller.NewApprovalController(approvalUsecase)

	router := gin.New()
	router.Use(middleware.OathkeeperIdentity())
	router.POST("/approvals/:approvalID/approve", approvalController.Approve)
	return router, nil
}
