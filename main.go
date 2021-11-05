package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	transportHttp "zuri.chat/zccore/internal/transport"
	"zuri.chat/zccore/logger"

	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/cors"
	"zuri.chat/zccore/messaging"
	"zuri.chat/zccore/utils"
)

type App struct {
	Port string
}

func (app *App) Run() error {
	// Socket  events
	var Server = socketio.NewServer(nil)

	messaging.SocketEvents(Server)

	// Set Stripe api key
	stripe.Key = os.Getenv("STRIPE_KEY")

	if err := utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		return fmt.Errorf("Could not connect to MongoDB: \n%v", err)
	}

	// sentry: enables reporting messages, errors, and panics.
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://82e17f3bba86400a9a38e2437b884d4a@o1013682.ingest.sentry.io/5979019",
	})

	if err != nil {
		return fmt.Errorf("sentry.Init: %s", err)
	}

	// transporter
	handler := transportHttp.NewHandler(Server)
	handler.SetupRoutes()

	c := cors.AllowAll()

	h := transportHttp.RequestDurationMiddleware(handler.Router)

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, c.Handler(h)),
		Addr:         ":" + app.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//nolint:errcheck //CODEI8: ignore error check
	go Server.Serve()

	logger.Info("Socket Served")

	defer Server.Close()

	logger.Info("Zuri Chat API running on port %s", app.Port)

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}


func main() {
	// load .env file if it exists
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("Error loading .env file: %v", err)
	}

	logger.Info("Environment variables successfully loaded. Starting application...")

	// get PORT from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := App{ Port: port }

	log.Fatal(app.Run())
	
	// if err := app.Run(); err != nil {
	// 	fmt.Println("Error occur while starting the Zuri Chat API.")
	// 	log.Fatal(err)
	// }
}
