package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User_id    string             `bson:"user_id" json:"user_id"`
	Line_id    string             `bson:"line_id" json:"line_id"`
	Age        string             `bson:"age" json:"age"`
	CreateTime string             `bson:"createTime" json:"createTime" `
}

type Message struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User_id  string             `bson:"user_id" json:"user_id"`
	Line_id  string             `bson:"line_id" json:"line_id"`
	Question string             `bson:"question" json:"question"`
	SendTime string             `bson:"sendTime" json:"sendTime" `
}
