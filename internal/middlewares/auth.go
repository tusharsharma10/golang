package middlewares

import (
	"restapi/logger"

	"github.com/gin-gonic/gin"
)

func AuthInternalRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {

		// apiKey := c.Request.Header.Get("x-api-key")

		// if apiKey != os.Getenv("INTERNAL_API_KEY") {
		// 	logger.Error(c, "access denied", logger.Z{
		// 		"apiKey": apiKey,
		// 	})
		// 	c.AbortWithStatus(http.StatusUnauthorized)

		// 	return
		// }

		logger.Info(c, "Hi from middleware", nil)

		c.Next()
	}
}
