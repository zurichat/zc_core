package messaging

import (
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	OwnerId    primitive.ObjectID   `json:"ownerid,omitempty" bson:"ownerid,omitempty"`
	RoomName   string               `json:"roomname,omitempty" bson:"roomname,omitempty"`
	RoomType   string               `json:"roomtype,omitempty" bson:"roomtype,omitempty"` //inbox, group, channel
	Members    []primitive.ObjectID `json:"members,omitempty" bson:"members,omitempty"`
	CreatedAt  string               `json:"createat,omitempty" bson:"createat,omitempty"`
	Archived   string               `json:"archived,omitempty" bson:"archived,omitempty"` // true/false
	ArchivedBy primitive.ObjectID   `json:"archivedby,omitempty" bson:"archivedby,omitempty"`
	ArchiveAt  string               `json:"archiveat,omitempty" bson:"archiveat,omitempty"`
	Private    string               `json:"roomprivacy,omitempty" bson:"roomprivacy,omitempty"` // {true, false} inbox, group is private by default
}

type MessageContent struct {
	StringContent string   `json:"stringcontent,omitempty" bson:"stringcontent,omitempty"`
	Files         []string `json:"files,omitempty" bson:"files,omitempty"`
}

type Reaction struct {
	Emoji   string               `json:"emoji,omitempty" bson:"emoji,omitempty"`
	Count   int                  `json:"count,omitempty" bson:"count,omitempty"`
	UserIds []primitive.ObjectID `json:"userids,omitempty" bson:"userids,omitempty"`
}

type Message struct {
	Content     MessageContent       `json:"content,omitempty" bson:"content,omitempty"`
	SenderId    primitive.ObjectID   `json:"senderid,omitempty" bson:"senderid,omitempty"`
	SenderName  string               `json:"sendername,omitempty" bson:"sendername,omitempty"`
	RoomId      primitive.ObjectID   `json:"roomid,omitempty" bson:"roomid,omitempty"`
	CreatedAt   string               `json:"createdat,omitempty" bson:"createdat,omitempty"`
	Read        string               `json:"read,omitempty" bson:"read,omitempty"` // true or false
	ReadAt      string               `json:"readat,omitempty" bson:"readat,omitempty"`
	Edited      string               `json:"edited,omitempty" bson:"edited,omitempty"` // true or false
	EditedAt    string               `json:"editedat,omitempty" bson:"editedat,omitempty"`
	Deleted     string               `json:"deleted,omitempty" bson:"deleted,omitempty"` // true or false
	DeletedAt   string               `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	Status      string               `json:"status,omitempty" bson:"status,omitempty"` // pending, sent ...
	ReceivedBy  []primitive.ObjectID `json:"receivedby,omitempty" bson:"receivedby,omitempty"`
	Reactions   []Reaction           `json:"reactions,omitempty" bson:"reactions,omitempty"`
	MessageType string               `json:"messagetype,omitempty" bson:"messagetype,omitempty"` //message/comment
	MessageId   primitive.ObjectID   `json:"messageid,omitempty" bson:"messageid,omitempty"`
}

// type MessageResponse struct {
// 	Status  bool
// 	Data    interface{}
// 	Message string
// }

// ErrorResponse : This is error model.
type ErrorResponse struct {
	StatusCode   int    `json:"status" bson:"status"`
	ErrorMessage string `json:"message" bson:"message"`
}

// SuccessResponse : This is success model.
type SuccessResponse struct {
	StatusCode int         `json:"status" bson:"status"`
	Message    string      `json:"message" bson:"message"`
	Data       interface{} `json:"data" bson:"data"`
}

// GetError : This is helper function to prepare error model.
func GetMessageError(err error, StatusCode int) interface{} {
	var response = ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   StatusCode,
	}

	return response
}

// GetError : This is helper function to prepare error model.
func GetCustomMessageError(err string, StatusCode int) interface{} {
	var response = ErrorResponse{
		ErrorMessage: err,
		StatusCode:   StatusCode,
	}

	return response
}

// GetSuccess : This is helper function to prepare success model.
func GetMessageSuccess(msg string, data interface{}) interface{} {
	var response = SuccessResponse{
		Message:    msg,
		StatusCode: http.StatusOK,
		Data:       data,
	}
	return response
}