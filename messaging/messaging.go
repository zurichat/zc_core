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
			RoomType:  "channel",
			RoomName:  "Default",
			CreatedAt: fmt.Sprintf("%v", time.Now().Unix()),
			Archived:  "false",
			Private:   "false",
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
	// fmt.Println(response)
	s.Emit("connection", response)

}

func EnterDefaultConversation(Server *socketio.Server, s socketio.Conn, msg string) {
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
			s.Emit("enter_default_conversation", response)
		}
	}

}

func BroadCastToDefaultConversation(Server *socketio.Server, s socketio.Conn, msg string) {
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
		s.Emit("default_conversation", res)
	} else {

		// fmt.Println(insertResult.InsertedID)
		response := GetMessageSuccess("Broadcast Sucessful", NewMessage)
		// s.Emit("create_todolist", response)
		Server.BroadcastToRoom("/socket.io/", conversationID, "default_conversation", response)
	}

	// conversation
}

func CreateRoom(Server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	if e != nil {
		fmt.Println(e)
	}
	var private string = "true"
	var members []primitive.ObjectID
	room_type := fmt.Sprintf("%v", q["room_type"])
	if room_type != "inbox" || room_type != "group" || room_type != "channel" {
		response := GetCustomMessageError("Invalid  room type: try inbox/group/channel", 400)
		s.Emit("create_room", response)
	} else {
		var receiverId primitive.ObjectID
		userid, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%v", q["userId"]))
		members = append(members, userid)
		if room_type == "inbox" {
			recid, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%v", q["receiverId"]))
			receiverId = recid
			members = append(members, receiverId)
		} else if room_type == "group" {
			// membersArr := make([]string, len( q["members"]))
			for _, mem := range q["members"].([]string) {
				memId, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%v", mem))
				members = append(members, memId)
			}
		}
		if room_type == "channel" {
			private = "false"
		}

		newRoom := Room{
			OwnerId:   userid,
			RoomName:  fmt.Sprintf("%v", q["room_name"]),
			RoomType:  room_type,
			Members:   members,
			CreatedAt: fmt.Sprintf("%v", time.Now().Unix()),
			Archived:  "false",
			Private:   private,
		}

		Dbname := utils.Env("DB_NAME")
		RoomCollection, er := utils.GetMongoDbCollection(Dbname, "rooms")
		if er != nil {
			// log.Fatal(er)
			response := GetCustomMessageError(fmt.Sprintf("%v", er), 400)
			s.Emit("create_room", response)
		}
		insertResult, err := RoomCollection.InsertOne(context.TODO(), newRoom)
		if err != nil {
			response := GetCustomMessageError(fmt.Sprintf("%v", err), 400)
			s.Emit("create_room", response)
		} else {
			roomid := insertResult.InsertedID.(primitive.ObjectID).Hex()
			GetMessageSuccess("Created Room successfully", roomid)
		}

	}
}

func EnterRoom(Server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	if e != nil {
		fmt.Println(e)
	}

	roomId := fmt.Sprintf("%v", q["roomId"])
	rid, _ := primitive.ObjectIDFromHex(roomId)
	fmt.Println("Entering ROOM: " + roomId)
	s.Join(roomId)
	var filtered []bson.M
	Dbname := utils.Env("DB_NAME")
	MessageCollection, _ := utils.GetMongoDbCollection(Dbname, "messages")
	filterCursor, err := MessageCollection.Find(context.TODO(), bson.M{"roomid": rid})
	if err != nil {
		fmt.Println(err)
	} else {
		if err = filterCursor.All(context.TODO(), &filtered); err != nil {
			fmt.Println(err)
		} else {
			response := GetMessageSuccess("Entered Room", filtered)
			s.Emit("enter_default_conversation", response)
		}
	}

}

func SocketEvents(Server *socketio.Server) *socketio.Server {
	///////////////////////////////////Connection Related Events//////////////////////////////////////////////
	Server.OnConnect("/socket.io/", func(s socketio.Conn) error {
		Connect(s)
		return nil
	})
	Server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	Server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	///////////////////////////////////Connection Related Events//////////////////////////////////////////////

	////////////////////////////////////Default Room Evennts///////////////////////////////////////////

	Server.OnEvent("/socket.io/", "enter_default_conversation", func(s socketio.Conn, msg string) {
		EnterDefaultConversation(Server, s, msg)
	})
	Server.OnEvent("/socket.io/", "default_conversation", func(s socketio.Conn, msg string) {
		BroadCastToDefaultConversation(Server, s, msg)
	})
	//////////////////////////////////Default Room Events///////////////////////////////////////////////

	/////////////////////////////////Main Events////////////////////////////////////////////////////////

	//------------------------------create room---------------------------------------------------//
	Server.OnEvent("/socket.io/", "create_room", func(s socketio.Conn, msg string) {
		CreateRoom(Server, s, msg)
	})
	//------------------------------create room---------------------------------------------------//
	//------------------------------Enter room---------------------------------------------------//
	Server.OnEvent("/socket.io/", "enter_room", func(s socketio.Conn, msg string) {
		EnterRoom(Server, s, msg)
	})
	//------------------------------Enter room---------------------------------------------------//

	/////////////////////////////////Main Events////////////////////////////////////////////////////////
	return Server

}
