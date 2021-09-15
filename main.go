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
	"github.com/rs/cors"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/data"
	"zuri.chat/zccore/external"
	"zuri.chat/zccore/marketplace"
	"zuri.chat/zccore/messaging"
	"zuri.chat/zccore/organizations"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/realtime"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

func Router(Server *socketio.Server) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	// Setup and init
	r.HandleFunc("/", VersionHandler)
	r.HandleFunc("/v1/welcome", auth.IsAuthenticated(Index)).Methods("GET")
	r.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")

	// Authentication
	r.HandleFunc("/auth/login", auth.LoginIn).Methods("POST")
	r.HandleFunc("/auth/test", auth.AuthTest).Methods("POST")
	r.HandleFunc("/auth/logout", auth.LogOutUser).Methods("POST", "GET")
	r.HandleFunc("/auth/verify-token", auth.IsAuthenticated(auth.VerifyTokenHandler)).Methods("GET", "POST")

	// Organization
	r.HandleFunc("/organizations", auth.IsAuthenticated(organizations.Create)).Methods("POST")
	r.HandleFunc("/organizations", auth.IsAuthenticated(organizations.GetOrganizations)).Methods("GET")
	r.HandleFunc("/organizations/{id}", organizations.GetOrganization).Methods("GET")
	r.HandleFunc("/organizations/{id}", auth.IsAuthenticated(organizations.DeleteOrganization)).Methods("DELETE")
	r.HandleFunc("/organizations/url/{url}", organizations.GetOrganizationByURL).Methods("GET")

	r.HandleFunc("/organizations/{id}/plugins", organizations.AddOrganizationPlugin).Methods("POST")
	r.HandleFunc("/organizations/{id}/plugins", organizations.GetOrganizationPlugins).Methods("GET")
	r.HandleFunc("/organizations/{id}/plugins/{plugin_id}", organizations.GetOrganizationPlugin).Methods("GET")

	r.HandleFunc("/organizations/{id}/url", auth.IsAuthenticated(organizations.UpdateUrl)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/name", auth.IsAuthenticated(organizations.UpdateName)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/logo", auth.IsAuthenticated(organizations.UpdateLogo)).Methods("PATCH")

	r.HandleFunc("/organizations/{id}/members", auth.IsAuthenticated(organizations.CreateMember)).Methods("POST")
	r.HandleFunc("/organizations/{id}/members", auth.IsAuthenticated(organizations.GetMembers)).Methods("GET")
	r.HandleFunc("/organizations/{id}/members/{mem_id}", auth.IsAuthenticated(organizations.GetMember)).Methods("GET")
	r.HandleFunc("/organizations/{id}/members/{mem_id}", auth.IsAuthenticated(organizations.DeleteMember)).Methods("DELETE")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/status", auth.IsAuthenticated(organizations.UpdateMemberStatus)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/photo", auth.IsAuthenticated(organizations.UpdateProfilePicture)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/profile", auth.IsAuthenticated(organizations.UpdateProfile)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/presence", auth.IsAuthenticated(organizations.TogglePresence)).Methods("POST")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings", auth.IsAuthenticated(organizations.UpdateMemberSettings)).Methods("PATCH")

	// Data
	r.HandleFunc("/data/write", data.WriteData)
	r.HandleFunc("/data/read", data.NewRead).Methods("POST")
	r.HandleFunc("/data/read/{plugin_id}/{coll_name}/{org_id}", data.ReadData).Methods("GET")
	r.HandleFunc("/data/delete", data.DeleteData).Methods("POST")
	r.HandleFunc("/data/collections/{plugin_id}", data.ListCollections).Methods("GET")
	r.HandleFunc("/data/collections/{plugin_id}/{org_id}", data.ListCollections).Methods("GET")

	// Plugins
	r.HandleFunc("/plugins/register", plugin.Register).Methods("POST")

	// Marketplace
	r.HandleFunc("/marketplace/plugins", marketplace.GetAllPlugins).Methods("GET")
	r.HandleFunc("/marketplace/plugins/{id}", marketplace.GetPlugin).Methods("GET")
	r.HandleFunc("/marketplace/plugins/{id}", marketplace.RemovePlugin).Methods("DELETE")

	// Users
	r.HandleFunc("/users", user.Create).Methods("POST")
	r.HandleFunc("/users/{user_id}", auth.IsAuthenticated(user.UpdateUser)).Methods("PATCH")
	r.HandleFunc("/users/{user_id}", auth.IsAuthenticated(user.GetUser)).Methods("GET")
	r.HandleFunc("/users/{user_id}", auth.IsAuthenticated(user.DeleteUser)).Methods("DELETE")
	r.HandleFunc("/users/search/{query}", auth.IsAuthenticated(user.SearchOtherUsers)).Methods("GET")
	r.HandleFunc("/users", auth.IsAuthenticated(user.GetUsers)).Methods("GET")

	// Realtime communications
	r.HandleFunc("/realtime/test", realtime.Test).Methods("GET")
	r.HandleFunc("/realtime/auth", realtime.Auth).Methods("POST")
	r.Handle("/socket.io/", Server)

	// Email subscription
	r.HandleFunc("/external/subscribe", external.EmailSubscription).Methods("POST")

	//ping endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.GetSuccess("Server is live", nil, w)
	}).Methods("GET", "POST")

	//api documentation
	r.PathPrefix("/").Handler(http.StripPrefix("/docs", http.FileServer(http.Dir("./docs/"))))

	// Home
	http.Handle("/", r)

	return r
}

func main() {
	////////////////////////////////////Socket  events////////////////////////////////////////////////
	var Server = socketio.NewServer(nil)
	messaging.SocketEvents(Server)
	////////////////////////////////////Socket  events////////////////////////////////////////////////

	// load .env file if it exists
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	fmt.Println("Environment variables successfully loaded. Starting application...")

	if err := utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		fmt.Println("Could not connect to MongoDB")
	}

	// get PORT from environment variables
	port, _ := os.LookupEnv("PORT")
	if port == "" {
		port = "8000"
	}

	r := Router(Server)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Handler:      LoggingMiddleware(c.Handler(r)),
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

// should redirect permanently to the docs page
func Index(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.AuthUser)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("Welcome %s to Zuri Core Developer.", user.Email))
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{w, 200}
		start := time.Now()
		h.ServeHTTP(lrw, r)
		end := time.Now()
		duration := end.Sub(start)
		log.Printf("[%s] | %s | %d | %dms\n", r.Method, r.URL.Path, lrw.statusCode, duration.Milliseconds())
	})
}
