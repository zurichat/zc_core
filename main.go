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
	r.HandleFunc("/data/read", data.ReadData)

	http.Handle("/", r)

	return r
}

func sanityCheck() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	envProps := []string{
		"ORGANIZATION_COLLECTION",
		"DB_NAME",
		"CLUSTER_URL",
		"PORT",
	}
	
	for _, k := range envProps {
		_, ok := os.LookupEnv(k)		
		if !ok {
			log.Fatal(fmt.Sprintf("Environment variable %s not defined. Terminating application...", k))
		}
	}
	fmt.Println("Environment variables successfully loaded. Starting application...")
}

func main() {
	// check that all required variables are loaded
	sanityCheck()

	// fecth variables from environment
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

	r := Router()
	// organization route handler
	organizations.NewOrgController(r, OrgService)

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
	http.HandleFunc("/v1/welcome", Index)
	fmt.Fprintf(w, "Welcome to Zuri Core")
}