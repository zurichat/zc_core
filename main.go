package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"zuri.chat/zccore/data"
	"zuri.chat/zccore/organizations"
)

func Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", VersionHandler)
	r.HandleFunc("/v1/welcome", Index).Methods("GET")
	r.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")
	r.HandleFunc("/data/write", data.WriteData)
	r.HandleFunc("/data/read", data.ReadData)
	r.HandleFunc("/organisation/create", organizations.Create).Methods("POST")
	r.HandleFunc("/organisation/plugins", organizations.GetPluginsOrganizations).Methods("GET")

	http.Handle("/", r)

	return r
}

// function to check if a file exists, usefull in checking for .env
func file_exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func main() {
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

	r := Router()

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

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