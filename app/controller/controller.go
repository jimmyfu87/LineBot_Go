package controller

import (
	"LineBot_Go/app/logger"
	t "LineBot_Go/app/tools"
	f "fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

func LineBotWebhookHandler(c *gin.Context) {
	logger.Info("LineBotWebhookHandler()")
	// Parse linebot webhook request
	bot := t.InitBot()
	events, err := bot.ParseRequest(c.Request)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			logger.Error("Invalid Siganature")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Siganature"})
		} else {
			logger.Error(f.Sprintf("Cannot parse webhook request because %s", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot parse webhook request"})
		}
		return
	}

	// Handle event
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			message, isTextMessage := event.Message.(*linebot.TextMessage)
			if isTextMessage {
				logger.Info("Process messages")
				// Get user_id
				user_id := events[0].Source.UserID
				t.ProcessTextMessage(bot, event.ReplyToken, message.Text, user_id)
			} else {
				logger.Error("This is not a text message")
				replyText := "這不是文字訊息，請重新輸入"
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyText)).Do()
			}
		} else {
			logger.Error("This is not a EventTypeMessage")
			replyText := "訊息格式錯誤，請重新輸入"
			bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyText)).Do()
		}
	}
}
