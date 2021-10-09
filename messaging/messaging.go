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

const ERRORCODE =  400

func Connect(s socketio.Conn) {
	s.SetContext("")
	fmt.Println("connected:", s.ID())
	iid, _ := strconv.Atoi(s.ID())

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

	RoomID = insertResult.InsertedID.(primitive.ObjectID).Hex()
	}

	response := GetMessageSuccess("Connection Successful", "No data")
	s.Emit("connection", response)
}

func EnterDefaultConversation(server *socketio.Server, s socketio.Conn, msg string) {
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

func BroadCastToDefaultConversation(server *socketio.Server, s socketio.Conn, msg string) {
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
		RoomID:      cid,
		MessageType: "message",
		CreatedAt:   fmt.Sprintf("%v", time.Now().Unix()),
	}

	Dbname := utils.Env("DB_NAME")
	MessageCollection, _ := utils.GetMongoDbCollection(Dbname, "messages")
	_, err := MessageCollection.InsertOne(context.TODO(), NewMessage)

	if err != nil {
		fmt.Println(err)
		res := GetCustomMessageError(fmt.Sprintf("%v", err), ERRORCODE)
		s.Emit("default_conversation", res)
	} else {
		response := GetMessageSuccess("Broadcast Successful", NewMessage)
		server.BroadcastToRoom("/socket.io/", conversationID, "default_conversation", response)
	}
}

func CreateRoom(server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	
	if e != nil {
		fmt.Println(e)
	}
	
	var private = "true"
	
	var members []primitive.ObjectID
	
	roomType := fmt.Sprintf("%v", q["roomType"])

	if	roomType != "inbox" {
		if roomType != "group" {
			if roomType != "channel" {
				response := GetCustomMessageError("Invalid  room type: try inbox/group/channel", ERRORCODE)
				s.Emit("create_room", response)
			}
		}
	} else {
		var receiverID primitive.ObjectID
	
		userid, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%v", q["userId"]))
		members = append(members, userid)
		if roomType == "inbox" {
			recid, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%v", q["receiverID"]))
			receiverID = recid
			members = append(members, receiverID)
		} else if roomType == "group" {
			for _, mem := range q["members"].([]string) {
				memID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%v", mem))
				members = append(members, memID)
			}
		}
		if roomType == "channel" {
			private = "false"
		}

		newRoom := Room{
			OwnerID:   userid,
			RoomName:  fmt.Sprintf("%v", q["roomName"]),
			RoomType:  roomType,
			Members:   members,
			CreatedAt: fmt.Sprintf("%v", time.Now().Unix()),
			Archived:  "false",
			Private:   private,
		}

		Dbname := utils.Env("DB_NAME")
		RoomCollection, er := utils.GetMongoDbCollection(Dbname, "rooms")
	
		if er != nil {
			response := GetCustomMessageError(fmt.Sprintf("%v", er), ERRORCODE)
			s.Emit("create_room", response)
		}
		insertResult, err := RoomCollection.InsertOne(context.TODO(), newRoom)
	
		if err != nil {
			response := GetCustomMessageError(fmt.Sprintf("%v", err), ERRORCODE)
			s.Emit("create_room", response)
		} else {
			roomid := insertResult.InsertedID.(primitive.ObjectID).Hex()
			GetMessageSuccess("Created Room successfully", roomid)
		}
	}
}

func EnterRoom(server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	
	if e != nil {
		fmt.Println(e)
	}

	roomID := fmt.Sprintf("%v", q["roomId"])
	rid, _ := primitive.ObjectIDFromHex(roomID)
	
	fmt.Println("Entering ROOM: " + roomID)
	s.Join(roomID)
	
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

// Leave room functionality.
func LeaveRoom(server *socketio.Server, s socketio.Conn, msg string) {
	q := make(map[string]interface{})
	e := json.Unmarshal([]byte(msg), &q)
	
	if e != nil {
		fmt.Println(e)
	}

	// leave room
	roomID := fmt.Sprintf("%v", q["roomId"])
	s.Leave(roomID)

	// get database, delete room data
	var filtered []bson.M
	
	Dbname := utils.Env("DB_NAME")

	MessageCollection, _ := utils.GetMongoDbCollection(Dbname, "messages")
	_, err := MessageCollection.DeleteMany(context.TODO(), bson.M{"roomid": RoomID})
	
	if err != nil {
		fmt.Println(err)
	} else {
		response := GetMessageSuccess("Left Room", filtered)
		s.Emit("left_conversation", response)
	}
}

func SocketEvents(server *socketio.Server) *socketio.Server {
	// Connection Related Events 
	server.OnConnect("/socket.io/", func(s socketio.Conn) error {
		Connect(s)
		return nil
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	// Connection Related Events 

	// Default Room Events 

	server.OnEvent("/socket.io/", "enter_default_conversation", func(s socketio.Conn, msg string) {
		EnterDefaultConversation(server, s, msg)
	})
	server.OnEvent("/socket.io/", "default_conversation", func(s socketio.Conn, msg string) {
		BroadCastToDefaultConversation(server, s, msg)
	})
	// Default Room Events 

	// Main Events

	// create room
	server.OnEvent("/socket.io/", "create_room", func(s socketio.Conn, msg string) {
		CreateRoom(server, s, msg)
	})
	// create room
	// Enter room
	server.OnEvent("/socket.io/", "enter_room", func(s socketio.Conn, msg string) {
		EnterRoom(server, s, msg)
	})
	// Enter room

	// Main Events
	return server
}
