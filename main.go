package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"zuri.chat/zccore/data"
	"zuri.chat/zccore/messaging"
	"zuri.chat/zccore/organizations"
)

func Router(Server *socketio.Server) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", VersionHandler)
	// r.Handle("/", http.FileServer(http.Dir("./views/chat/")))
	r.HandleFunc("/v1/welcome", Index).Methods("GET")
	r.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")
	r.HandleFunc("/data/write", data.WriteData)
	r.HandleFunc("/data/read", data.ReadData)
	r.HandleFunc("/organisation/create", organizations.Create).Methods("POST")
	r.Handle("/socket.io/", Server)

	http.Handle("/", r)

	return r
}

// function to check if a file exists, usefull in checking for .env
func file_exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func main() {
	////////////////////////////////////Socket  events////////////////////////////////////////////////
	var Server = socketio.NewServer(nil)
	Server.OnConnect("/socket.io/", func(s socketio.Conn) error {
		messaging.Connect(s)
		return nil
	})
	Server.OnEvent("/socket.io/", "enter_converstion", func(s socketio.Conn, msg string) {
		messaging.EnterConversation(Server, s, msg)
	})
	Server.OnEvent("/socket.io/", "conversation", func(s socketio.Conn, msg string) {
		messaging.BroadCastToConversation(Server, s, msg)
	})
	Server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	Server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	////////////////////////////////////Socket  events////////////////////////////////////////////////

	// load .env file if it exists
	if file_exists(".env") {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// get PORT from environment variables
	port, ok := os.LookupEnv("PORT")

	// if there is no PORT in environment variables default to port 8000
	if !ok {
		port = "8000"
	}

	r := Router(Server)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go Server.Serve()
	fmt.Println("Socket Served")
	defer Server.Close()

	fmt.Println("Zuri Chat API running on port ", port)
	log.Fatal(srv.ListenAndServe())
}

func LoadApp(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appId := params["appid"]

	fmt.Printf("URL called with Param: %s", appId)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<div><b>Hello</b> World <button style='color: green;'>Click me!</button></div>: App = %s\n", appId)
}
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Zuri Chat API - Version 0.0001\n")

}
func Index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	// http.HandleFunc("/v1/welcome", Index)
	fmt.Fprintf(w, "Welcome to Zuri Core Index")
}
