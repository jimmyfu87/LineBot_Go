package controller

import (
	"LineBot_Go/app/config"
	"LineBot_Go/app/logger"
	t "LineBot_Go/app/tools"
	f "fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

func GetAllMessagesByLineIDHandler(c *gin.Context) {
	logger.Info("GetAllMessagesByLineIDHandler()")
	// Get params
	lineID := c.Param("line_id")
	logger.Info(f.Sprintf("lineID: %s", lineID))
	messageDao := t.InitDbConn(config.Message_table_name)
	// Get all messages
	messages, err := messageDao.GetAllMessagesByLineIDDAO(lineID)
	if err != nil {
		logger.Error(f.Sprintf("Error occurs when GetAllMessagesByLineIDHandler() because %s", err))
		logger.Error(f.Sprintf("lineID: %s", lineID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get user's messages"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func PushMessageHandler(c *gin.Context) {
	logger.Info("PushMessageHandler()")
	// Parse line_id and message from the request
	lineID := c.PostForm("line_id")
	message := c.PostForm("message")
	logger.Info(f.Sprintf("lineID: %s", lineID))
	logger.Info(f.Sprintf("message: %s", message))
	// Check if the line_id parameter is missing
	if lineID == "" {
		logger.Error("Missing line_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing line_id parameter"})
		return
	}

	// Initialize Line Bot
	bot := t.InitBot()

	// Initialize database connection
	userDao := t.InitDbConn(config.User_table_name)

	// Get user by lineID
	user, getUserIDErr := userDao.GetUserByLineIDDAO(lineID)
	if getUserIDErr != nil {
		logger.Error(f.Sprintf("Unable to retrieve user information: %s", getUserIDErr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user information"})
		return
	}

	// Create the message to be sent
	sendingMessage := linebot.NewTextMessage(message)

	// Push the message to the specified line_id
	userID := user.User_id
	_, pushErr := bot.PushMessage(userID, sendingMessage).Do()
	if pushErr != nil {
		logger.Error(f.Sprintf("Unable to send message: %s", pushErr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to send message"})
		return
	}

	// Return a success message
	logger.Info(f.Sprintf("Message successfully sent to user with line_id: %s", lineID))
	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully"})
}
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
