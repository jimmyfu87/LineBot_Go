package controller

import (
	"LineBot_Go/app/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPushMessageHandler(t *testing.T) {
	// Create fake gin.Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create push message but "jimmy" has to be in database
	c.Request = httptest.NewRequest(http.MethodPost, "/push_message", nil)
	c.Request.PostForm = map[string][]string{
		"line_id": {"jimmy"},
		"message": {"Hello, World!"},
	}

	// Call function
	PushMessageHandler(c)

	// Check result
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLineBotWebhookHandler(t *testing.T) {
	// Simulate payload, but userid need to exist
	target_userID := "target_userID"
	lineWebhookJSON := "{\"to\": \"" + target_userID + "\" ,\"messages\":{\"type\": \"text\",\"text\": \"Hello,Ting~這是Line Bot API測試訊息\"}}"
	channelSecret := config.Channel_secret
	signature := CalculateSignature(channelSecret, lineWebhookJSON)

	// Create fake gin.Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set correct request header
	c.Request, _ = http.NewRequest("POST", "/callback", strings.NewReader(lineWebhookJSON))
	c.Request.Header.Set("Authorization", "Bearer "+config.Channel_access_token)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Line-Signature", signature)

	// Call function
	LineBotWebhookHandler(c)

	// Check result
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLineBotWebhookHandlerWithoutSignature(t *testing.T) {
	// Simulate payload, but userid need to exist
	target_userID := "target_userID"
	lineWebhookJSON := "{\"to\": \"" + target_userID + "\" ,\"messages\":{\"type\": \"text\",\"text\": \"Hello,Ting~這是Line Bot API測試訊息\"}}"
	// Create fake gin.Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set request header without signature
	c.Request, _ = http.NewRequest("POST", "/callback", strings.NewReader(lineWebhookJSON))
	c.Request.Header.Set("Authorization", "Bearer "+config.Channel_access_token)
	c.Request.Header.Set("Content-Type", "application/json")

	// Call LineBotWebhookHandler
	LineBotWebhookHandler(c)

	// Check result
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLineBotWebhookHandlerOtherError(t *testing.T) {
	// Simulate payload, but userid need to exist
	target_userID := "target_userID"
	lineWebhookJSON := "{\"to\": \"" + target_userID + "\" ,\"messages\":\"type\": \"text\",\"text\": \"Hello,Ting~這是Line Bot API測試訊息\"}"
	channelSecret := config.Channel_secret
	signature := CalculateSignature(channelSecret, lineWebhookJSON)

	// Create fake gin.Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set request header with wrong webhook json
	c.Request, _ = http.NewRequest("POST", "/callback", strings.NewReader(lineWebhookJSON))
	c.Request.Header.Set("Authorization", "Bearer "+config.Channel_access_token)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Line-Signature", signature)

	// Call LineBotWebhookHandler
	LineBotWebhookHandler(c)

	// Check result
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func CalculateSignature(channelSecret, requestBody string) string {
	key := []byte(channelSecret)
	message := []byte(requestBody)

	// Get Hash
	hash := hmac.New(sha256.New, key)
	hash.Write(message)
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return signature
}
