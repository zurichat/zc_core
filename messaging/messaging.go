package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	socketio "github.com/googollee/go-socket.io"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

func remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}

var Server = socketio.NewServer(nil)

func Messaging() {
	// go Server.Serve()
	// defer Server.Close()
	/////////////////////////////////////////////////started Socket.io server//////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////main db///////////////////////////////////////////////////////////////
	fmt.Println("Starting db connection")
	RoomCollection, er := utils.GetMongoDbCollection("zc_core", "rooms")
	if er != nil {
		fmt.Println(er)
	}
	// MessageCollection
	_, err := utils.GetMongoDbCollection("zc_core", "messages")
	if err != nil {
		fmt.Println(err)

	}
	fmt.Println("Connected to MongoDB!")
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

	////////////////////////////////////////////On connection///////////////////////////////////////////////////////////////////////

	Server.OnConnect("/socket.io/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		// Server.JoinRoom("/socket.io/", "slack", s)

		// s.Send("slack", "message", args ...interface{})
		response := GetMessageSuccess("Connection Successful", ' ')
		s.Emit("connection", response)
		return nil
	})
	////////////////////////////////////////////On connection///////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////Enter a conversation //////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "enter_converstion", func(s socketio.Conn, msg string) {
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
		}
		var conversation Room
		// organisationId := fmt.Sprintf("%v", q["organisationId"])
		converstionType := fmt.Sprintf("%v", q["converstionType"])
		converstionid := fmt.Sprintf("%v", q["converstionid"])
		// requesterId := fmt.Sprintf("%v", q["requesterId"])
		if converstionType == "inbox" {
			err := RoomCollection.FindOne(context.TODO(), bson.M{"_id": converstionid}).Decode(&conversation)
			if err != nil {
				fmt.Println(err)
				response := GetCustomMessageError(fmt.Sprintf("%v", err), 400)
				s.Emit("open_todolist", response)
			}
		} else {

		}

	})

	////////////////////////////////////////////Enter a conversation //////////////////////////////////////////////////////////////

}
