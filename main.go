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
	"zuri.chat/zccore/app"
	"zuri.chat/zccore/messaging"
	"zuri.chat/zccore/organizations"
)

// Added a PathPrefix to route all endpoints via "/v1"
func Router() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
  	wsr := r.PathPrefix("/v1").Subrouter()	
    r.HandleFunc("/", VersionHandler).Methods("GET")
	wsr.HandleFunc("/app", app.AppHandler)
	wsr.HandleFunc("/messaging", messaging.Message).Methods("GET")
	wsr.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")
	wsr.HandleFunc("/data/write", data.WriteData)
	wsr.HandleFunc("/data/read", data.ReadData)
	wsr.HandleFunc("/organisation/create", organizations.Create).Methods("POST")

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
