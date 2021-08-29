package todolist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// "github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// type Todo struct {
// 	id       string
// 	content  string
// 	time     string
// 	complete bool
// }

type Response struct {
	Status        bool
	TodolistData  Todolist
	TodoData      Todo
	TodoListArray []bson.M
	Message       string
	TodolistId    interface{}
}

type Todo struct {
	Id       string `json:"_id,omitempty" bson:"_id,omitempty"`
	Content  string `json:"content,omitempty" bson:"content,omitempty"`
	Time     string `json:"time,omitempty" bson:"time,omitempty"`
	Complete string `json:"complete,omitempty" bson:"complete,omitempty"`
}

type Todolist struct {
	UserId string `json:"userId,omitempty" bson:"userId,omitempty"`
	Title  string `json:"title,omitempty" bson:"title,omitempty"`
	Todo   []Todo `json:"todo,omitempty" bson:"todo,omitempty"`
	Time   string `json:"time,omitempty" bson:"time,omitempty"`
}

var (
	Newtodo     Todolist
	currentlist Todolist
	Atodolist   Todolist
	TodoID      string
)

func remove(slice []Todo, s int) []Todo {
	return append(slice[:s], slice[s+1:]...)
}

var Server = socketio.NewServer(nil)

func TodoOps() {

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////mongodb connection///////////////////////////////////////////////////////////////////////
	// fmt.Println("Starting db connection")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017")) //Please change the URL
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("zc_core").Collection("todolist")
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	} else {
		// fmt.Println("Sucess")
	}

	// fmt.Println("Connected to MongoDB!")

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////mongodb connection///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////On connection///////////////////////////////////////////////////////////////////////

	Server.OnConnect("/socket.io/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		// Server.JoinRoom("/socket.io/", "slack", s)

		// s.Send("slack", "message", args ...interface{})
		response := Response{
			Status:  true,
			Message: "Connection Successful",
		}
		s.Emit("connection", response)
		return nil
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////On connection///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Create Todolist///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "create_todolist", func(s socketio.Conn, msg string) {
		fmt.Println(msg)
		// var Atodolist Todolist
		q := make(map[string]interface{})
		// q, e := url.ParseQuery(msg)
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {

			fmt.Println(e)
		}
		var emptyTodo []Todo
		Newtodo = Todolist{
			UserId: fmt.Sprintf("%v", q["userid"]),
			Title:  fmt.Sprintf("%v", q["title"]),
			Time:   fmt.Sprintf("%v", time.Now().Unix()),
			Todo:   emptyTodo,
		}
		fmt.Println(q)
		insertResult, err := collection.InsertOne(context.TODO(), Newtodo)
		if err != nil {
			// log.Fatal(err)
			fmt.Println(err)
			response := Response{
				Status:  false,
				Message: fmt.Sprintf("%v", err),
			}

			s.Emit("create_todolist", response)
		} else {

			response := Response{
				Status:     true,
				Message:    "Todolist Created",
				TodolistId: insertResult.InsertedID,
			}
			s.Emit("create_todolist", response)
		}

		// fmt.Println("Created a todolist with id: ", insertResult.InsertedID)
		// var userid string = fmt.Sprintf("%v", Newtodo.Id)
		// fmt.Println(Newtodo)
		// b, err := json.Marshal(m)
		// s.Emit("create_todolist", "created todo with id: "+fmt.Sprintf("%v", insertResult.InsertedID))

	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Create Todolist///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Delete Todolist///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "delete_todolist", func(s socketio.Conn, msg string) {
		fmt.Println(msg)
		// var Atodolist Todolist
		q := make(map[string]interface{})
		// q, e := url.ParseQuery(msg)
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			panic(e)
		}

		// user_id := fmt.Sprintf("%v", q["userid"])
		todolistid := fmt.Sprintf("%v", q["todoListId"])
		todoid, _ := primitive.ObjectIDFromHex(todolistid)
		deleteResult, err := collection.DeleteOne(context.TODO(), bson.M{"_id": todoid})
		if err != nil {
			fmt.Println(err)
			response := Response{
				Status:     false,
				Message:    fmt.Sprintf("%v", err),
				TodolistId: todolistid,
			}
			s.Emit("delete_todolist", response)
		} else {

			if deleteResult.DeletedCount > 0 {
				response := Response{
					Status:     true,
					Message:    "Todolist Deleted",
					TodolistId: todolistid,
				}
				s.Emit("delete_todolist", response)
			} else {
				response := Response{
					Status:     true,
					Message:    "Todolist Not Deleted",
					TodolistId: todolistid,
				}
				s.Emit("delete_todolist", response)
			}

		}

		// fmt.Println((deleteResult))

		// fmt.Println("Created a todolist with id: ", insertResult.InsertedID)
		// // var userid string = fmt.Sprintf("%v", Newtodo.Id)
		// fmt.Println(Newtodo)
		// b, err := json.Marshal(m)
		// s.Emit("create_todolist", "created todo with id: "+fmt.Sprintf("%v", insertResult.InsertedID))
		// s.Emit("delete_todolist", deleteResult)
	})
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Delete Todolist///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Return all todolists for user///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "all_todolists", func(s socketio.Conn, msg string) {
		fmt.Println("All todos")
		var filtered []bson.M
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
		}
		user_id := fmt.Sprintf("%v", q["userid"])

		filterCursor, err := collection.Find(context.TODO(), bson.M{"userId": user_id})
		if err != nil {
			fmt.Println(err)
			response := Response{
				Status:  false,
				Message: fmt.Sprintf("%v", err),
			}
			s.Emit("all_todolists", response)
		} else {
			if err = filterCursor.All(context.TODO(), &filtered); err != nil {
				fmt.Println(err)
				response := Response{
					Status:  false,
					Message: fmt.Sprintf("%v", err),
				}
				s.Emit("all_todolists", response)
			} else {
				response := Response{
					Status:        true,
					Message:       "All Todolists",
					TodoListArray: filtered,
				}
				s.Emit("all_todolists", response)
			}
		}

		// fmt.Println(filtered)
		// s.Emit("all_todolists", filtered)

	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Return all todolists for user///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Open Todolist///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "open_todolist", func(s socketio.Conn, msg string) {
		var currentlist Todolist
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
		}
		// user_id := fmt.Sprintf("%v", q["userid"])
		TodoID = fmt.Sprintf("%v", q["todoListId"])
		todoid, _ := primitive.ObjectIDFromHex(TodoID)

		err := collection.FindOne(context.TODO(), bson.M{"_id": todoid}).Decode(&currentlist)
		if err != nil {
			fmt.Println(err)
			response := Response{
				Status:     false,
				Message:    fmt.Sprintf("%v", err),
				TodolistId: TodoID,
			}
			s.Emit("open_todolist", response)
		} else {
			s.Join(TodoID)
			response := Response{
				Status:       true,
				Message:      "Opened Todolist",
				TodolistData: currentlist,
				TodolistId:   TodoID,
			}
			Server.BroadcastToRoom("/socket.io/", TodoID, "open_todolist", response)
		}

		// fmt.Printf("Found a single document: %+v\n", currentlist)

	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Open Todolist///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////leave Todolist///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "leave_todolist", func(s socketio.Conn, msg string) {
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
		}
		// user_id := fmt.Sprintf("%v", q["userid"])
		TodoID = fmt.Sprintf("%v", q["todoListId"])
		s.Leave(TodoID)

		response := Response{
			Status:     true,
			Message:    "Left Todolist",
			TodolistId: TodoID,
		}
		s.Emit("leave_todolist", response)

	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////leave Todolist///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Add to todolist///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "add_to_todolist", func(s socketio.Conn, msg string) {
		var currentlist Todolist
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
			fmt.Println("after1")
		}
		// user_id := fmt.Sprintf("%v", q["userid"])
		todolistid := fmt.Sprintf("%v", q["todoListId"])
		todocontent := fmt.Sprintf("%v", q["content"])
		ctime := time.Now().Unix()

		todoid, _ := primitive.ObjectIDFromHex(todolistid)
		s.Join(todolistid)
		err := collection.FindOne(context.TODO(), bson.M{"_id": todoid}).Decode(&currentlist)
		if err != nil {
			fmt.Println(err)
			fmt.Println("after2")
			response := Response{
				Status:     false,
				Message:    fmt.Sprintf("%v", err),
				TodolistId: todolistid,
			}
			Server.BroadcastToRoom("/socket.io/", todolistid, "add_to_todolist", response)
		} else {
			rectodo := Todo{
				Id:       uuid.New().String(),
				Content:  todocontent,
				Time:     fmt.Sprintf("%v", ctime),
				Complete: "false",
			}
			currentlist.Todo = append(currentlist.Todo, rectodo)
			_, er := collection.UpdateOne(context.TODO(),
				bson.M{"_id": todoid},
				bson.D{
					{"$set", bson.D{{"Todo", currentlist.Todo}}},
				},
			)
			if er != nil {
				fmt.Println(err)
				fmt.Println("after3")
				response := Response{
					Status:     false,
					Message:    fmt.Sprintf("%v", er),
					TodolistId: todolistid,
				}
				Server.BroadcastToRoom("/socket.io/", todolistid, "add_to_todolist", response)
			} else {
				fmt.Println(err)
				response := Response{
					Status:     true,
					Message:    "Added to todolist",
					TodolistId: todolistid,
					TodoData:   rectodo,
				}
				Server.BroadcastToRoom("/socket.io/", todolistid, "add_to_todolist", response)

			}
		}

		// fmt.Println(result)
		// s.Emit("add_to_todolist", "Opened todo with id: "+fmt.Sprintf("%v", result))
		// Server.BroadcastToRoom("/socket.io/", "todolistid", "add_to_todolist", result)
	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////Add to todolist///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////edit todo///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "edit_todo", func(s socketio.Conn, msg string) {
		var currentlist Todolist
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
		}
		// user_id := fmt.Sprintf("%v", q["userid"])
		todolistid := fmt.Sprintf("%v", q["todoListId"])
		todocontent := fmt.Sprintf("%v", q["content"])

		// completestatus, err := strconv.ParseBool(fmt.Sprintf("%v", q["completeStatus"]))
		completestatus := fmt.Sprintf("%v", q["completeStatus"])
		if err != nil {
			fmt.Println(err)
		}
		contentid := fmt.Sprintf("%v", q["contentId"])
		todoid, _ := primitive.ObjectIDFromHex(todolistid)

		err1 := collection.FindOne(context.TODO(), bson.M{"_id": todoid}).Decode(&currentlist)
		if err != nil {
			fmt.Println(err1)
			response := Response{
				Status:     false,
				Message:    fmt.Sprintf("%v", err1),
				TodolistId: todolistid,
			}
			Server.BroadcastToRoom("/socket.io/", todolistid, "edit_todo", response)
		} else {

			for index, content := range currentlist.Todo {
				if content.Id == contentid {
					pp := &currentlist.Todo[index]
					// p := &s[1]
					pp.Content = todocontent
					pp.Complete = completestatus

					// currentlist.Todo[index].Content = todocontent
					break
				}
			}

			_, err := collection.UpdateOne(context.TODO(),
				bson.M{"_id": todoid},
				bson.D{
					{"$set", bson.D{{"Todo", currentlist.Todo}}},
				},
			)

			if err != nil {
				fmt.Println(err1)
				response := Response{
					Status:     false,
					Message:    fmt.Sprintf("%v", err),
					TodolistId: todolistid,
				}
				Server.BroadcastToRoom("/socket.io/", todolistid, "edit_todo", response)
			} else {
				response := Response{
					Status:     true,
					Message:    "Added to todolist",
					TodolistId: todolistid,
				}
				Server.BroadcastToRoom("/socket.io/", todolistid, "edit_todo", response)
			}
			// var pp Todo

			// fmt.Println(result)
			// s.Emit("edit_todo", fmt.Sprintf("%v", result))
			// Server.BroadcastToRoom("/socket.io/", "todolistid", "edit_todo", result)
		}

	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////edit todo///////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////delete todo///////////////////////////////////////////////////////////////////////

	Server.OnEvent("/socket.io/", "delete_todo", func(s socketio.Conn, msg string) {
		var currentlist Todolist
		q := make(map[string]interface{})
		e := json.Unmarshal([]byte(msg), &q)
		if e != nil {
			fmt.Println(e)
		}
		// user_id := fmt.Sprintf("%v", q["userid"])
		todolistid := fmt.Sprintf("%v", q["todoListId"])
		contentid := fmt.Sprintf("%v", q["contentId"])
		todoid, _ := primitive.ObjectIDFromHex(todolistid)

		err := collection.FindOne(context.TODO(), bson.M{"_id": todoid}).Decode(&currentlist)
		if err != nil {
			fmt.Println(err)
			response := Response{
				Status:     false,
				Message:    fmt.Sprintf("%v", err),
				TodolistId: todolistid,
			}
			Server.BroadcastToRoom("/socket.io/", todolistid, "delete_todo", response)
		} else {
			for index, content := range currentlist.Todo {
				if content.Id == contentid {
					fmt.Println(currentlist.Todo)
					fmt.Println(index)
					sp := &currentlist
					// p := &s[1]
					// pp.Content = todocontent
					new_list := remove(currentlist.Todo, index)
					sp.Todo = new_list

					// currentlist.Todo[index].Content = todocontent
					break
				}
			}

			fmt.Println("Removed")
			fmt.Println(currentlist.Todo)

			_, err := collection.UpdateOne(context.TODO(),
				bson.M{"_id": todoid},
				bson.D{
					{"$set", bson.D{{"Todo", currentlist.Todo}}},
				},
			)
			if err != nil {
				fmt.Println(err)
				response := Response{
					Status:     false,
					Message:    fmt.Sprintf("%v", err),
					TodolistId: todolistid,
				}
				Server.BroadcastToRoom("/socket.io/", todolistid, "delete_todo", response)
			} else {
				response := Response{
					Status:     true,
					Message:    "Deleted From todolist",
					TodolistId: todolistid,
				}
				Server.BroadcastToRoom("/socket.io/", todolistid, "delete_todo", response)
			}
			// fmt.Println(result)
			// s.Emit("delete_todo", fmt.Sprintf("%v", result))

		}
		// var pp Todo

	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////delete todo///////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////test area///////////////////////////////////////////////////////////////////////

	// Server.OnEvent("/", "bye", func(s socketio.Conn) string {
	// 	last := s.Context().(string)
	// 	s.Emit("bye", last)
	// 	s.Close()
	// 	return last
	// })
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////test area///////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////error and disconnect///////////////////////////////////////////////////////////////////////

	Server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	Server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////error and disconnect///////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////routes///////////////////////////////////////////////////////////////////////
	// http.Handle("/socket.io/", Server)
	// http.Handle("/", http.FileServer(http.Dir("./views/chat/")))
	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080" // Default port if not specified
	// }
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////routes///////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////start Server and socket///////////////////////////////////////////////////////////////////////
	go Server.Serve()
	defer Server.Close()
	// log.Println("Serving at localhost:" + port + "...")
	// log.Fatal(http.ListenAndServe(":"+port, nil))

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////start Server and socket///////////////////////////////////////////////////////////////////////
}
