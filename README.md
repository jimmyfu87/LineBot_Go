# LineBot_Go

### 系統簡介:

- LINEBOT提供以下功能，包含註冊、ChatGPT、主動回覆使用者、取得使用者列表
    
    - 使用者註冊功能

        - 檢查使用者的輸入是否符合要求的格式

        - 將使用者的user_id、使用者自定義的line_id、年齡、建立時間存到MongoDB

    - ChatGPT單句對話功能
        
        - 註冊後可透過Line間接使用ChatGPT單句對話功能，並記錄使用者的問題到MongoDB
    
        - 可以限制使用者問題的字數．及限制ChatGPT回答使用的token數

        - 每句對話是獨立的，無法延續上一句回答繼續對話
    
    - 主動回覆指定使用者訊息
        
        - 使用者可透過push_message的API向指定line_id的使用者
    
    - 取得所有註冊使用者資訊
        
        - 使用者可透過get_all_users的API取得所有使用者的訊息


### 使用方式:
    go run main.go



## 文件路徑圖、程式簡介：
    main.go
    app/
        dao/
            dao.go
        logger/
              logger.go
        model/
             model.go
        tools/
             tools.go 
        config/  
              config.go
        controller/
                   controller.go
                   controller_test.go
    mongodb_data/
    mongo-init.js
    docker-compose.yml
    go.mod
    go.sum

- main.go
    
    - 執行主程式，會啟動一個API引擎，在指定port部署

- app
    - dao.go
        
        - 處理和資料庫連接的資料層
    
    - logger.go
        
        - 定義log的格式，info和error
    
    - tools.go
        
        - 可能會重複使用到函式
    
    - config.go
        
        - 自定義的客製化變數，如問題長度，以及linebot用到的key等
    
    - controller.go
        
        - controller層，接收API傳的參數，以及API的處理
    
    - controller_test.go
        
        - controller層的單元測試，僅完成push_message, webhook的部分測試，且資料庫相依性待解決

- mongodb_data/
    
    - mongodb用來儲存資料的資料夾

- mongo-init.js
    
    - 執行docker-compose up時初始化DB和兩個table，分別是user, message

- docker-compoase
    
    - 建立mongodb container的yml檔，在母路徑執行docker-compose up即可完成mongodb部署

- go.mod
    
    - 程式會使用的外部套件或依賴套件

- go.sum
    
    - 防止go.mod被篡改的校驗go.mod的文件



    
# 程式細節介紹


## `main.go`
### 用途：
- 啟動Gin引擎部署API


## `dao.go`
### 用途： 
- 處理和資料庫的互動，包含和User和Message兩個table的互動
- 建立一個mongodb的collection的dao，即可使用各種dao方法

## `logger.go`
### 用途： 
- 建立log的格式，共定義兩種一個是info另一個是error


## `model.go`
### 用途： 
- 定義User和Message資料的屬性和格式


## `tools.go`
### 用途：
  - 可能會被重複利用的函數 

#### 1. InitBot()

- 利用兩個key建立一個linebot的連線

#### 2. InitDbConn(tableName string)

- 建立和MongoDB的連線後，建立指定table的dao並回傳供資料庫互動使用

#### 3. ProcessTextMessage(bot *linebot.Client, replyToken, messageText string, userID string)
- 處理從linebot送過來的text message的function
- line_id為使用者自定義的暱稱
- user_id為該line用戶已被line官方指定的id

1. 建立和user table的資料庫連線，並檢查這個user是否已經註冊過

2. 如果沒有註冊過，要求依照指定格式line_id,age輸入資料，輸入有空格會自動刪除掉
    
    - 檢查輸入有沒有","，沒有的話要求重新輸入
    
    - 將","相隔的字串進行拆分，如果不是兩個(line_id,age)的話要求重新輸入
    
    - 檢查line_id是否為空，若為空要求重新輸入
    
    - 檢查line_id是否已經被使用過，若使用過要求重新輸入
    
    - 檢查age是否是整數且介於0~150間，若不是要求重新輸入
    
    - 以上檢查都通過，把使用者資料和時間用SaveUserData存到資料庫，註冊前的訊息都不會儲存
    
    - 提供ChatGPT的使用指南

3. 如果註冊過，檢查問題是否超過上限，沒超過上限就會儲存後再串ChatGPT回答問題
    
    - 檢查問題的字數是否超過上限，超過上限要求重新輸入
    
    - 儲存問題、傳送問題的時間、user_id到message的table
    
    - 呼叫ChatGPT並回答問題給使用者

 
#### 4. SaveUserData(userID string, lineID string, age string)  

- 將使用者資訊連同註冊時間存放到資料庫

#### 5. SaveMessageData(userID string, question string)  

- 將對話訊息連同傳送時間存放到資料庫

#### 6. TalkToChatGPT(question string)  

- 串接ChatGPT的API並根據問題提供回答

#### ７. GetTimeNow()

- 取得當下的時間

## `config.go`
### 用途： 
- 儲存多個可客製化的參數和linebot的key，讓其餘程式可以重複利用

    - (1) Web_port: server執行的port
    - (2) Channel_secret: LineBot的channel_secret
    - (3) Channel_access_token: LineBot的channel_access_token
    - (4) Db_name: MongoDB的名稱，若更改在mongo-init.js也需要更改
    - (5) Db_username: MongoDB的使用者名稱，若更改在mongo-init.js也需要更改
    - (6) Db_password: MongoDB的使用者密碼，若更改在mongo-init.js也需要更改
    - (7) Db_url: MongoDB的URL
    - (8) User_table_name: MongoDB存放user的table，若更改在mongo-init.js也需要更改
    - (9) Message_table_name: MongoDB存放message的table，若更改在mongo-init.js也需要更改
    - (10) ChatGPT_key: OpenAI的引用ChatGPT使用的API_KEY
    - (11) Model_name: OpenAI的引用ChatGPT使用的Model
    - (12) Question_max_len_ch: 使用ChatGPT中文字最長可以幾個字
    - (13) Answer_max_tokens: ChatGPT回答使用的tokens上限
    - (14) Chinese_len: 假設每個中文字會使用的字符，有可能3~4所以設定為4


## `controller.go`
### 用途： 
- controller層，接收API傳的參數，以及資料的處理
- 建立一個mongodb的collection的dao，即可使用各種dao方法

#### 1. GetAllMessagesByLineIDHandler(c *gin.Context) / GET

- 取得所有user的message資料的API，傳line_id為path parameter，取得後的格式如下

      [
        {
            "id": "6550fac36d581c83230d1cf6",
            "user_id": "U1c9e034b960f72f6252f13fb16751928",
            "line_id": "jimmy",
            "question": "幫我介紹蘋果這家公司",
            "sendTime": "2023-11-13 00:18:11"
        },
        {
            "id": "6550fad46d581c83230d1cf9",
            "user_id": "U1c9e034b960f72f6252f13fb16751928",
            "line_id": "jimmy",
            "question": "幫我介紹臉書這家公司",
            "sendTime": "2023-11-13 00:18:28"
        }
      ] 
#### 2. PushMessageHandler(c *gin.Context) / POST

- 對指定line_id的使用者發送訊息，使用form data格式呼叫
    
    - line_id: 發送對象的自定義ID
    
    - message: 想發送的訊息
 
 #### 3. LineBotWebhookHandler(c *gin.Context) 
 
 - 接受linebot送來的webhook請求並處理的API
 - 檢查訊息格式、是否為文字訊息、line簽名

## `controller_test.go`
### 用途： 
- controller層的單元測試，僅完成push_message, webhook的部分測試，且資料庫相依性待解決

## `mongodb_data/`
### 用途： 
- 儲存mongodb資料的資料夾

## `mongo-init.js`
### 用途： 
- 執行docker-compose up時初始化DB和兩個table，分別是user, message

## `docker-compose.yml`
### 用途：
- 建立mongodb container的yml檔，在母路徑執行docker-compose up即可完成mongodb部署

## `go.mod` 
### 用途：  
- 程式會使用的外部套件或依賴套件

## `go.sum`
### 用途： 
- 防止go.mod被篡改的校驗go.mod的文件


