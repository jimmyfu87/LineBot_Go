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
	r.POST("/push_message", cnt.PushMessageHandler)
	r.GET("/get_all_messages/:line_id", cnt.GetAllMessagesByLineIDHandler)
	r.Run(":" + config.Web_port)
}
