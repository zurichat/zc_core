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
	"zuri.chat/zccore/marketplace"
	"zuri.chat/zccore/messaging"
	"zuri.chat/zccore/organizations"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

func Router(Server *socketio.Server) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/", VersionHandler)
	// r.Handle("/", http.FileServer(http.Dir("./views/chat/")))
	r.HandleFunc("/v1/welcome", Index).Methods("GET")
	r.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")
	r.HandleFunc("/organisation/create", organizations.Create).Methods("POST")
	r.HandleFunc("/organisation/{org_id}", organizations.Create).Methods("POST")
	r.Handle("/socket.io/", Server)
	r.HandleFunc("/data/write", data.WriteData).Methods("POST", "PUT", "DELETE")
	r.HandleFunc("/data/read/{plugin_id}/{coll_name}/{org_id}", data.ReadData).Methods("GET")
	r.HandleFunc("/plugin/register", plugin.Register).Methods("POST")
	r.HandleFunc("/plugin/{id}", plugin.GetByID).Methods("GET")
	r.HandleFunc("/marketplace/plugins", marketplace.GetAllApprovedPlugins).Methods("GET")
	r.HandleFunc("/marketplace/plugins/{id}", marketplace.GetOneApprovedPlugin).Methods("GET")
	r.HandleFunc("/marketplace/install", marketplace.InstallPluginToOrg).Methods("POST")

	http.Handle("/", r)

	return r
}

func main() {
	////////////////////////////////////Socket  events////////////////////////////////////////////////
	var Server = socketio.NewServer(nil)
	Server.OnConnect("/socket.io/", func(s socketio.Conn) error {
		messaging.Connect(s)
		return nil
	})
	Server.OnEvent("/socket.io/", "enter_conversation", func(s socketio.Conn, msg string) {
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

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println("Environment variables successfully loaded. Starting application...")

	if err := utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		log.Fatal(err)
	}

	// fetch variables from environment
	DATABASE_NAME, _ := os.LookupEnv("DB_NAME")
	ORGANIZATION_COLLECTION, _ := os.LookupEnv("ORGANIZATION_COLLECTION")

	orgCollection, err := utils.GetMongoDbCollection(DATABASE_NAME, ORGANIZATION_COLLECTION)

	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	orgRepo := organizations.NewOrgRepository(orgCollection)
	OrgService := organizations.NewOrgService(orgRepo)

	// get PORT from environment variables
	port, _ := os.LookupEnv("PORT")
	if port == "" {
		port = "8000"
	}

	r := Router(Server)

	// organization route handler
	organizations.NewOrgController(r, OrgService)

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
