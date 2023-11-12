package main

import (
	"LineBot_Go/app/config"
	cnt "LineBot_Go/app/controller"
	"LineBot_Go/app/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Init gin engine
	r := gin.Default()
	logger.Info("Web Starts Sucessfully!")
	r.POST("/callback", cnt.LineBotWebhookHandler)
	r.Run(":" + config.Web_port)
}
