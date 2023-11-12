package dao

import (
	"LineBot_Go/app/logger"
	"LineBot_Go/app/model"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DAO struct {
	collection *mongo.Collection
}

func NewDAO(client *mongo.Client, databaseName, collectionName string) *DAO {
	logger.Info("NewUserDAO()")
	logger.Info(fmt.Sprintf("databaseName: %s", databaseName))
	logger.Info(fmt.Sprintf("collectionName: %s", collectionName))
	database := client.Database(databaseName)
	collection := database.Collection(collectionName)
	return &DAO{collection}
}

func (dao *DAO) InsertUserDAO(user model.User) error {
	logger.Info("InsertUserDAO()")
	logger.Info(fmt.Sprintf("user: %s", user))
	_, err := dao.collection.InsertOne(context.TODO(), user)
	return err
}

func (dao *DAO) GetUserByLineIDDAO(lineID string) (model.User, error) {
	logger.Info("GetUserByLineIDDAO()")
	logger.Info(fmt.Sprintf("lineID: %s", lineID))
	filter := bson.M{"line_id": lineID}
	user := model.User{}
	err := dao.collection.FindOne(context.Background(), filter).Decode(&user)
	return user, err
}

func (dao *DAO) GetUserByUserIDDAO(userID string) (model.User, error) {
	logger.Info("GetUserByUserIDDAO()")
	logger.Info(fmt.Sprintf("userID: %s", userID))
	filter := bson.M{"user_id": userID}
	user := model.User{}
	err := dao.collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		logger.Info(fmt.Sprintf("Cannot find user, because %s", err))
		return model.User{}, nil
	}

	return user, err
}

func (dao *DAO) GetAllMessagesByLineIDDAO(lineID string) ([]model.Message, error) {
	logger.Info("GetAllMessageByLineIDDAO()")
	logger.Info(fmt.Sprintf("lineID: %s", lineID))
	var messages []model.Message
	filter := bson.M{"line_id": lineID}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		logger.Info(fmt.Sprintf("Cannot find messages, because %s", err))
		return nil, err
	}
	// Defer closing the cursor
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &messages); err != nil {
		logger.Info(fmt.Sprintf("Assign error, because %s", err))
		return nil, err
	}

	return messages, nil
}

func (dao *DAO) InsertMessageDAO(message model.Message) error {
	logger.Info("InsertMessageDAO()")
	logger.Info(fmt.Sprintf("message: %s", message))
	_, err := dao.collection.InsertOne(context.TODO(), message)
	return err
}
