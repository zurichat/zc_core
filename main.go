package main

import (
	"errors"
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

	// load .env file if it exists
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	fmt.Println("Environment variables successfully loaded. Starting application...")

	// Set Stripe api key
	stripe.Key = os.Getenv("STRIPE_KEY")

	if err = utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		return errors.New("Could not connect to MongoDB")
	}

	// sentry: enables reporting messages, errors, and panics.
	err = sentry.Init(sentry.ClientOptions{
		Dsn: "https://82e17f3bba86400a9a38e2437b884d4a@o1013682.ingest.sentry.io/5979019",
	})

	if err != nil {
		return fmt.Errorf("sentry.Init: %s", err)
	}

	// transporter
	handler := transportHttp.NewHandler()
	handler.SetupRoutes(Server)	

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

	fmt.Println("Socket Served")

	defer Server.Close()

	fmt.Println("Zuri Chat API running on port ", app.Port)

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}


func main() {
	// get PORT from environment variables
	port, _ := os.LookupEnv("PORT")
	if port == "" {
		port = "8000"
	}

	app := App{ Port: port }
	
	if err := app.Run(); err != nil {
		fmt.Println("Error occur while starting the Zuri Chat API.")
		log.Fatal(err)
	}
}
