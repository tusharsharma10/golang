package server

import (
	"context"
	"net/http"
	"os"
	"restapi/internal/middlewares"

	"restapi/db"
	"restapi/logger"

	"github.com/gin-gonic/gin"

	"restapi/internal/controller/transaction"
)

func NewRouter(env string) *gin.Engine {
	logger.Init("restapi", os.Getenv("LOG_LEVEL"))

	logger.Debug(context.Background(), "starting server...", logger.Z{
		"mode": os.Getenv("GIN_MODE"),
	})

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	router.Use(gin.Recovery())

	if gin.Mode() != gin.ReleaseMode {
		router.Use(gin.Logger())
	}

	// do not cache anything by default
	// router.Use(middlewares.AttachTransactionIDMiddleware())

	registerRoutes(env, router)

	return router
}

const (
	maxOpenConn = 2
	maxIdleConn = 2
)

func registerRoutes(env string, router *gin.Engine) {
	mysqlDB := db.Conn(env, true, -1, -1)
	masterDBHandle := db.Conn(env, false, maxOpenConn, maxIdleConn, "MASTER")

	transactionController := transaction.NewTransactionController(mysqlDB, masterDBHandle)

	dopamineGroup := router.Group("api/v1")
	{
		dopamineGroup.GET("/healthcheck", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"Test": "Successful"})
		})

		actionRoutes := dopamineGroup.Group("transaction")
		{
			actionRoutes.GET("/all", middlewares.AuthInternalRoutes(), transactionController.Info)

		}

	}
}
