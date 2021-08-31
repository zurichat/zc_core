package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

var RoomID interface{}

func remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}

func Connect(s socketio.Conn) {
	s.SetContext("")
	fmt.Println("connected:", s.ID())
	iid, _ := strconv.Atoi(s.ID())
	// Server.JoinRoom("/socket.io/", "slack", s)
	if iid == 1 {
		Dbname := utils.Env("DB_NAME")
		RoomCollection, er := utils.GetMongoDbCollection(Dbname, "rooms")
		if er != nil {
			log.Fatal(er)
		}
		Newroom := Room{
			RoomType:    "channel",
			RoomName:    "Default",
			CreatedAt:   fmt.Sprintf("%v", time.Now().Unix()),
			Archived:    "false",
			RoomPrivacy: "public",
		}
		insertResult, err := RoomCollection.InsertOne(context.TODO(), Newroom)
		if err != nil {
			log.Fatal(err)
		}
		// RoomID = fmt.Sprintf("%v", insertResult.InsertedID)
		RoomID = insertResult.InsertedID.(primitive.ObjectID).Hex()
	}

	// s.Send("slack", "message", args ...interface{})
	response := GetMessageSuccess("Connection Successful", "No data")
	fmt.Println(response)
	s.Emit("connection", response)

}

func EnterConversation(Server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	if e != nil {
		fmt.Println(e)
	}

	// userName := fmt.Sprintf("%v", q["name"])
	conversationID := fmt.Sprintf("%v", RoomID)
	cid, _ := primitive.ObjectIDFromHex(conversationID)
	fmt.Println("Entering ROOM: " + conversationID)
	s.Join(conversationID)
	var filtered []bson.M
	Dbname := utils.Env("DB_NAME")
	MessageCollection, _ := utils.GetMongoDbCollection(Dbname, "messages")
	filterCursor, err := MessageCollection.Find(context.TODO(), bson.M{"roomid": cid})
	if err != nil {
		fmt.Println(err)
	} else {
		if err = filterCursor.All(context.TODO(), &filtered); err != nil {
			fmt.Println(err)
		} else {
			response := GetMessageSuccess("Entered Conversation", filtered)
			s.Emit("enter_converstion", response)
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// var conversation Room
	// // organisationId := fmt.Sprintf("%v", q["organisationId"])
	// converstionType := fmt.Sprintf("%v", q["converstionType"])
	// converstionid := fmt.Sprintf("%v", q["converstionid"])
	// // requesterId := fmt.Sprintf("%v", q["requesterId"])
	// if converstionType == "inbox" {
	// 	err := RoomCollection.FindOne(context.TODO(), bson.M{"_id": converstionid}).Decode(&conversation)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		response := GetCustomMessageError(fmt.Sprintf("%v", err), 400)
	// 		s.Emit("open_todolist", response)
	// 	}
	// } else {

	// }
	/////////////////////////////////////////////////////////////////////////////////////////////////////
}

func BroadCastToConversation(Server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	if e != nil {
		fmt.Println(e)
	}

	NewMessageContent := MessageContent{
		StringContent: fmt.Sprintf("%v", q["content"]),
	}

	conversationID := fmt.Sprintf("%v", RoomID)
	cid, _ := primitive.ObjectIDFromHex(conversationID)
	NewMessage := Message{
		Content:     NewMessageContent,
		SenderName:  fmt.Sprintf("%v", q["name"]),
		RoomId:      cid,
		MessageType: "message",
		CreatedAt:   fmt.Sprintf("%v", time.Now().Unix()),
	}

	// fmt.Println(q)
	Dbname := utils.Env("DB_NAME")
	MessageCollection, _ := utils.GetMongoDbCollection(Dbname, "messages")
	_, err := MessageCollection.InsertOne(context.TODO(), NewMessage)
	if err != nil {
		// log.Fatal(err)
		fmt.Println(err)
		res := GetCustomMessageError(fmt.Sprintf("%v", err), 400)
		s.Emit("conversation", res)
	} else {

		// fmt.Println(insertResult.InsertedID)
		response := GetMessageSuccess("Broadcast Sucessful", NewMessage)
		// s.Emit("create_todolist", response)
		Server.BroadcastToRoom("/socket.io/", conversationID, "conversation", response)
	}

	// conversation
}

func Messaging() {
	// var Server = socketio.NewServer(nil)
	// go Server.Serve()
	// defer Server.Close()
	/////////////////////////////////////////////////started Socket.io server//////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////main db///////////////////////////////////////////////////////////////
	// fmt.Println("Starting db connection")
	// RoomCollection, er := utils.GetMongoDbCollection("zc_core", "rooms")
	// if er != nil {
	// 	fmt.Println(er)
	// }
	// // MessageCollection
	// _, err := utils.GetMongoDbCollection("zc_core", "messages")
	// if err != nil {
	// 	fmt.Println(err)

	// }
	// fmt.Println("Connected to MongoDB!")
	// owner, _ := primitive.ObjectIDFromHex("iuwifuouwfobwubuw")
	// Newroom := Room{
	// 	OwnerId:     owner,
	// 	RoomType:    "channel",
	// 	CreatedAt:   fmt.Sprintf("%v", time.Now().Unix()),
	// 	Archived:    "false",
	// 	RoomPrivacy: "public",
	// }
	// Result, err := RoomCollection.InsertOne(context.TODO(), Newroom)
	// if err != nil {
	// 	// log.Fatal(err)
	// 	fmt.Println(err)
	// }
	// response := GetMessageSuccess("Query Successful", ' ')
	// fmt.Println(response)
	// fmt.Println(Result)
	///////////////////////////////////////////////////////////main db///////////////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////Enter a conversation //////////////////////////////////////////////////////////////

	// Server.OnEvent("/socket.io/", "enter_converstion", func(s socketio.Conn, msg string) {
	// 	q := make(map[string]interface{})
	// 	e := json.Unmarshal([]byte(msg), &q)
	// 	if e != nil {
	// 		fmt.Println(e)
	// 	}
	// 	var conversation Room
	// 	// organisationId := fmt.Sprintf("%v", q["organisationId"])
	// 	converstionType := fmt.Sprintf("%v", q["converstionType"])
	// 	converstionid := fmt.Sprintf("%v", q["converstionid"])
	// 	// requesterId := fmt.Sprintf("%v", q["requesterId"])
	// 	if converstionType == "inbox" {
	// 		err := RoomCollection.FindOne(context.TODO(), bson.M{"_id": converstionid}).Decode(&conversation)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			response := GetCustomMessageError(fmt.Sprintf("%v", err), 400)
	// 			s.Emit("open_todolist", response)
	// 		}
	// 	} else {

	// 	}

	// })

	////////////////////////////////////////////Enter a conversation //////////////////////////////////////////////////////////////

}
