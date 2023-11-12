package tools

import (
	"LineBot_Go/app/config"
	"LineBot_Go/app/dao"
	"LineBot_Go/app/logger"
	"LineBot_Go/app/model"
	"context"

	f "fmt"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitBot() *linebot.Client {
	logger.Info("InitBot()")
	bot, err := linebot.New(
		config.Channel_secret,
		config.Channel_access_token)
	if err != nil {
		logger.Error(f.Sprintf("Error occurs when InitBot() because %s", err))
		panic(err)
	}
	return bot
}

func InitDbConn(tableName string) *dao.DAO {
	logger.Info("InitDbConn()")
	logger.Info(f.Sprintf("tableName: %s", tableName))
	// Set MongoDB authentication options
	cred := options.Credential{
		Username: config.Db_username,
		Password: config.Db_password,
	}

	// Set MongoDB connection options, including authentication
	clientOptions := options.Client().ApplyURI(config.Db_url).
		SetAuth(cred)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logger.Error(f.Sprintf("Error occurs when InitDbConn() because %s", err))
		panic(err)
	}

	// Initialize UserDAO
	Dao := dao.NewDAO(client, config.Db_name, tableName)
	return Dao
}

func ProcessTextMessage(bot *linebot.Client, replyToken, messageText string, userID string) {
	logger.Info("processTextMessage()")
	logger.Info(f.Sprintf("messageText: %s", messageText))
	logger.Info(f.Sprintf("userID: %s", userID))
	// Init database
	userDao := InitDbConn(config.User_table_name)
	// Check user exist or not
	existingUser_userID, getUserIDErr := userDao.GetUserByUserIDDAO(userID)
	if getUserIDErr != nil {
		logger.Error(f.Sprintf("Error occurs when GetUserByUserIDDAO(): %s", getUserIDErr))
		replyText := "系統內部發生錯誤，請稍後再試"
		bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
		return
	}
	// Set chatgpt max question length
	max_ch_len := config.Question_max_len_ch
	max_en_len := config.Question_max_len_ch * config.Chinese_len
	// If user is not exists
	if existingUser_userID.User_id == "" {
		logger.Info("This user has not registered")
		// trim blank
		messageText_trim := strings.ReplaceAll(messageText, " ", "")
		// If no comma in the input, ask user to register again
		if !strings.Contains(messageText_trim, ",") {
			logger.Info("No comma found in the input.")
			replyText := "歡迎使用M800！您尚未註冊，請先註冊才能繼續使用，輸入你的Line ID和年齡，用逗號分隔，例如：line_id,age"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}

		// Split message
		inputs := strings.Split(messageText_trim, ",")
		if len(inputs) != 2 {
			logger.Info("Invalid format. Two values separated by a comma are expected.")
			// Format incorrect
			replyText := "請使用逗號分隔Line ID和年齡，例如：line_id,age"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}

		lineID, ageStr := inputs[0], inputs[1]
		// Check line_id is blank or not
		if lineID == "" {
			logger.Info("LineID is blank")
			// lineID is blank
			replyText := "LineID不能為空白請重新輸入"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}
		// Check line_id is used or not
		existingUser_lineID, getLineIDErr := userDao.GetUserByLineIDDAO(lineID)
		if getUserIDErr != nil {
			logger.Error(f.Sprintf("Error occurs when GetUserByLineIDDAO(): %s", getLineIDErr))
			replyText := "系統內部發生錯誤，請稍後再試"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}
		// line_id has been used ask to try another one
		if existingUser_lineID.Line_id != "" {
			logger.Info(f.Sprintf("This Line_id has been used."))
			replyText := "這個LineID已經有人使用過了，請更換一個註冊"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}
		// Change age to int and check age
		age, err := strconv.Atoi(ageStr)
		if err != nil || age < 0 || age > 150 {
			logger.Info("Invalid age format or out of range.")
			// Age format is incorrect, enter age from 0 to 150 integer
			replyText := "輸入的年齡格式不正確，重新輸入0~150的數字"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}

		// Save User Data
		if SaveUserData(userID, lineID, ageStr) {
			logger.Info("Registration successful")
			successMsg := "註冊成功！Line ID和年齡已儲存，"
			hintMsg := f.Sprintf("接下來可使用單句的ChatBot對話系統，不可以超過%s個中文字或%s個英文字", strconv.Itoa(max_ch_len), strconv.Itoa(max_en_len))
			replyText := successMsg + hintMsg
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
		} else {
			logger.Info("User storage failed. Please start over.")
			replyText := "儲存失敗，請重新開始輸入"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
		}

	} else {
		// Get line_id
		lineID := existingUser_userID.Line_id
		// Get message length
		message_length := len(messageText)
		if message_length > max_en_len {
			logger.Info("Message is to long")
			logger.Info(f.Sprintf("message_length: %s", strconv.Itoa(message_length)))
			replyText := f.Sprintf("您輸入的訊息太長，不能超過%s個中文字或%s個英文字", strconv.Itoa(max_ch_len), strconv.Itoa(max_en_len))
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
			return
		}
		// Has already registered, use LLM to response
		if SaveMessageData(userID, lineID, messageText) {
			logger.Info("Message saves successfully")
			replyText := TalkToChatGPT(messageText)
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()

		} else {
			logger.Info("Message storage failed. Please start over.")
			replyText := "系統儲存發生錯誤，稍後再試"
			bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyText)).Do()
		}

	}

}

func SaveUserData(userID string, lineID string, age string) bool {
	logger.Info("SaveUserData()")
	logger.Info(f.Sprintf("userID: %s", userID))
	logger.Info(f.Sprintf("lineID: %s", lineID))
	logger.Info(f.Sprintf("age: %s", age))
	// Init database
	userDao := InitDbConn(config.User_table_name)
	// Get time
	createTime := GetTimeNow()
	// Init user
	user := model.User{User_id: userID, Line_id: lineID, Age: age, CreateTime: createTime}
	// Insert User to db
	logger.Info("Inserting user")
	logger.Info(f.Sprintf("user: %s", user))
	err := userDao.InsertUserDAO(user)
	if err != nil {
		logger.Error(f.Sprintf("Error occurs when InsertUser(): %s", err))
		return false
	}
	// Data is saved successfully
	return true
}

func SaveMessageData(userID string, lineID string, question string) bool {
	logger.Info("SaveMessageData()")
	logger.Info(f.Sprintf("userID: %s", userID))
	logger.Info(f.Sprintf("lineID: %s", lineID))
	logger.Info(f.Sprintf("question: %s", question))
	// Init database
	messageDao := InitDbConn(config.Message_table_name)
	// Init user
	sendTime := GetTimeNow()
	message := model.Message{User_id: userID, Line_id: lineID, Question: question, SendTime: sendTime}
	// Insert User to db
	logger.Info("Inserting message")
	logger.Info(f.Sprintf("message: %s", message))
	err := messageDao.InsertMessageDAO(message)
	if err != nil {
		logger.Error(f.Sprintf("Error occurs when InsertMessageDAO(): %s", err))
		return false
	}
	// Data is saved successfully
	return true
}

func TalkToChatGPT(question string) string {
	logger.Info("TalkToChatGPT()")
	logger.Info(f.Sprintf("question: %s", question))
	client := openai.NewClient(config.ChatGPT_key)
	params := openai.ChatCompletionRequest{
		Model:     config.Model_name,
		MaxTokens: config.Answer_max_tokens,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: question},
		},
	}
	// Send request to chatgpt
	response, err := client.CreateChatCompletion(context.Background(), params)
	if err != nil {
		errorMsg := f.Sprintf("Error sending request: %s", err)
		logger.Error(errorMsg)
		return errorMsg
	}
	// Get response
	content := response.Choices[0].Message.Content
	return content
}

func GetTimeNow() string {
	logger.Info("GetTimeNow()")
	location, _ := time.LoadLocation("Asia/Taipei")
	currentTime := time.Now().In(location)
	// Format time
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	return formattedTime
}
