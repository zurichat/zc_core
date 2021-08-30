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
	"zuri.chat/zccore/utils"
)

func Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", VersionHandler)
	r.HandleFunc("/v1/welcome", Index).Methods("GET")
	r.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")
	r.HandleFunc("/data/write", data.WriteData)
	r.HandleFunc("/data/read/{plugin_id}/{coll_name}/{org_id}", data.ReadData).Methods("GET")
	r.HandleFunc("/organisation/create", organizations.Create).Methods("POST")

	http.Handle("/", r)

	return r
}

func main() {
	// load .env file once
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	if err := utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
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