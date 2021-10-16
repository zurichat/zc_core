package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"

	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/cors"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/blog"
	"zuri.chat/zccore/contact"
	"zuri.chat/zccore/data"
	"zuri.chat/zccore/external"
	"zuri.chat/zccore/marketplace"
	"zuri.chat/zccore/messaging"
	"zuri.chat/zccore/organizations"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/realtime"
	"zuri.chat/zccore/report"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

func Router(server *socketio.Server) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	// Load handlers(Doing this to reduce dependency circle issue, might reverse if not working)
	configs := utils.NewConfigurations()
	mailService := service.NewZcMailService(configs)

	orgs := organizations.NewOrganizationHandler(configs, mailService)
	exts := external.NewExternalHandler(configs, mailService)
	reps := report.NewReportHandler(configs, mailService)
	au := auth.NewAuthHandler(configs, mailService)
	us := user.NewUserHandler(configs, mailService)
	gql := utils.NewGraphQlHandler(configs)

	// Setup and init
	r.HandleFunc("/", VersionHandler)
	r.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")

	// Blog
	r.HandleFunc("/posts", blog.GetPosts).Methods("GET")
	r.HandleFunc("/posts", blog.CreatePost).Methods("POST")
	r.HandleFunc("/posts/{post_id}", blog.UpdatePost).Methods("PUT")
	r.HandleFunc("/posts/{post_id}", blog.DeletePost).Methods("DELETE")
	r.HandleFunc("/posts/{post_id}", blog.GetPost).Methods("GET")
	r.HandleFunc("/posts/{post_id}/like/{user_id}", blog.LikeBlog).Methods("PATCH")
	r.HandleFunc("/posts/{post_id}/comments", blog.GetBlogComments).Methods("GET")
	r.HandleFunc("/posts/{post_id}/comments", blog.CommentBlog).Methods("POST")
	r.HandleFunc("/posts/search", blog.SearchBlog).Methods("GET")
	r.HandleFunc("/posts/mail", blog.MailingList).Methods("POST")

	// Authentication
	r.HandleFunc("/auth/login", au.LoginIn).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout", au.LogOutUser).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout/other-sessions", au.LogOutOtherSessions).Methods(http.MethodPost)
	r.HandleFunc("/auth/verify-token", au.IsAuthenticated(au.VerifyTokenHandler)).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/auth/confirm-password", au.IsAuthenticated(au.ConfirmUserPassword)).Methods(http.MethodPost)
	r.HandleFunc("/auth/social-login/{provider}/{access_token}", au.SocialAuth).Methods(http.MethodGet)

	r.HandleFunc("/account/verify-account", au.VerifyAccount).Methods(http.MethodPost)
	r.HandleFunc("/account/request-password-reset-code", au.RequestResetPasswordCode).Methods(http.MethodPost)
	r.HandleFunc("/account/verify-reset-password", au.VerifyPasswordResetCode).Methods(http.MethodPost)
	r.HandleFunc("/account/update-password/{verification_code:[0-9]+}", au.UpdatePassword).Methods(http.MethodPost)

	// Organization
	r.HandleFunc("/organizations", au.IsAuthenticated(orgs.Create)).Methods("POST")
	r.HandleFunc("/organizations", au.IsAuthenticated(orgs.GetOrganizations)).Methods("GET")
	r.HandleFunc("/organizations/{id}", au.IsAuthenticated(orgs.GetOrganization)).Methods("GET")
	r.HandleFunc("/organizations/{id}", au.IsAuthenticated(au.IsAuthorized(orgs.DeleteOrganization, "admin"))).Methods("DELETE")
	r.HandleFunc("/organizations/url/{url}", orgs.GetOrganizationByURL).Methods("GET")

	r.HandleFunc("/organizations/{id}/url", au.IsAuthenticated(orgs.UpdateURL)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/name", au.IsAuthenticated(orgs.UpdateName)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/logo", au.IsAuthenticated(orgs.UpdateLogo)).Methods("PATCH")

	r.HandleFunc("/organizations/{id}/settings", au.IsAuthenticated(orgs.UpdateOrganizationSettings)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/permission", au.IsAuthenticated(orgs.UpdateOrganizationPermission)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/auth", au.IsAuthenticated(orgs.UpdateOrganizationAuthentication)).Methods("PATCH")

	r.HandleFunc("/organizations/{id}/prefixes", au.IsAuthenticated(orgs.UpdateOrganizationPrefixes)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/slackbotresponses", au.IsAuthenticated(orgs.UpdateSlackBotResponses)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/customemoji", au.IsAuthenticated(orgs.AddSlackCustomEmoji)).Methods("PATCH")

	// Organization: Guest Invites
	r.HandleFunc("/organizations/{id}/send-invite", au.IsAuthenticated(au.IsAuthorized(orgs.SendInvite, "admin"))).Methods("POST")
	r.HandleFunc("/organizations/{id}/invite-stats", au.IsAuthenticated(au.IsAuthorized(orgs.InviteStats, "admin"))).Methods("GET")
	r.HandleFunc("/organizations/invites/{uuid}", orgs.CheckGuestStatus).Methods(http.MethodGet)
	r.HandleFunc("/organizations/guests/{uuid}", orgs.GuestToOrganization).Methods(http.MethodPost)

	r.HandleFunc("/organizations/{id}/plugins", au.IsAuthenticated(orgs.AddOrganizationPlugin)).Methods("POST")
	r.HandleFunc("/organizations/{id}/plugins", au.IsAuthenticated(orgs.GetOrganizationPlugins)).Methods("GET")
	r.HandleFunc("/organizations/{id}/plugins/{plugin_id}", au.IsAuthenticated(orgs.GetOrganizationPlugin)).Methods("GET")
	r.HandleFunc("/organizations/{id}/plugins/{plugin_id}", au.IsAuthenticated(orgs.RemoveOrganizationPlugin)).Methods("DELETE")

	r.HandleFunc("/organizations/{id}/members", au.IsAuthenticated(au.IsAuthorized(orgs.CreateMember, "admin"))).Methods("POST")
	r.HandleFunc("/organizations/{id}/members", orgs.GetMembers).Methods("GET")
	r.HandleFunc("/organizations/{id}/members/multiple", au.IsAuthenticated(orgs.GetmultipleMembers)).Methods("GET")
	r.HandleFunc("/organizations/{id}/members/{mem_id}", au.IsAuthenticated(orgs.GetMember)).Methods("GET")
	r.HandleFunc("/organizations/{id}/members/{mem_id}", au.IsAuthenticated(au.IsAuthorized(orgs.DeactivateMember, "admin"))).Methods("DELETE")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/reactivate", au.IsAuthenticated(au.IsAuthorized(orgs.ReactivateMember, "admin"))).Methods("POST")

	r.HandleFunc("/organizations/{id}/members/{mem_id}/status", au.IsAuthenticated(orgs.UpdateMemberStatus)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/photo/{action}", au.IsAuthenticated(orgs.UpdateProfilePicture)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/profile", au.IsAuthenticated(orgs.UpdateProfile)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/presence", au.IsAuthenticated(orgs.TogglePresence)).Methods("POST")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings", au.IsAuthenticated(orgs.UpdateMemberSettings)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/role", au.IsAuthenticated(au.IsAuthorized(orgs.UpdateMemberRole, "admin"))).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/notification", au.IsAuthenticated(orgs.UpdateNotification)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/theme", au.IsAuthenticated(orgs.UpdateUserTheme)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/mark-as-read", au.IsAuthenticated(orgs.UpdateMarkAsRead)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/message-media", au.IsAuthenticated(orgs.UpdateMemberMessageAndMediaSettings)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/accessibility", au.IsAuthenticated(orgs.UpdateMemberAccessibilitySettings)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/languages-and-region", au.IsAuthenticated(orgs.SetLanguagesAndRegions)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/settings/advanced", au.IsAuthenticated(orgs.UpdateMemberAdvancedSettings)).Methods("PATCH")

	r.HandleFunc("/organizations/{id}/reports", au.IsAuthenticated(reps.AddReport)).Methods("POST")
	r.HandleFunc("/organizations/{id}/reports", au.IsAuthenticated(reps.GetReports)).Methods("GET")
	r.HandleFunc("/organizations/{id}/reports/{report_id}", au.IsAuthenticated(reps.GetReport)).Methods("GET")
	r.HandleFunc("/organizations/{id}/change-owner", au.IsAuthenticated(orgs.TransferOwnership)).Methods("PATCH")
	r.HandleFunc("/organizations/{id}/billing", au.IsAuthenticated(orgs.SaveBillingSettings)).Methods("PATCH")

	//organization: payment
	r.HandleFunc("/organizations/{id}/add-token", au.IsAuthenticated(orgs.AddToken)).Methods("POST")
	r.HandleFunc("/organizations/{id}/token-transactions", au.IsAuthenticated(orgs.GetTokenTransaction)).Methods("GET")
	r.HandleFunc("/organizations/{id}/upgrade-to-pro", au.IsAuthenticated(orgs.UpgradeToPro)).Methods("POST")
	r.HandleFunc("/organizations/{id}/charge-tokens", au.IsAuthenticated(orgs.ChargeTokens)).Methods("POST")
	r.HandleFunc("/organizations/{id}/checkout-session", au.IsAuthenticated(orgs.CreateCheckoutSession)).Methods("POST")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/cards", au.IsAuthenticated(orgs.AddCard)).Methods("POST")
	r.HandleFunc("/organizations/{id}/members/{mem_id}/cards/{card_id}", au.IsAuthenticated(orgs.DeleteCard)).Methods("DELETE")

	// Data
	r.HandleFunc("/data/write", data.WriteData)
	r.HandleFunc("/data/read", data.NewRead).Methods("POST")
	r.HandleFunc("/data/read/{plugin_id}/{coll_name}/{org_id}", data.ReadData).Methods("GET")
	r.HandleFunc("/data/delete", data.DeleteData).Methods("POST")
	r.HandleFunc("/data/collections/info/{plugin_id}/{coll_name}/{org_id}", data.CollectionDetail).Methods("GET")

	// Plugins
	r.HandleFunc("/plugins/register", plugin.Register).Methods("POST")
	r.HandleFunc("/plugins/{id}", plugin.Update).Methods("PATCH")
	r.HandleFunc("/plugins/{id}", plugin.Delete).Methods("DELETE")
	r.HandleFunc("/plugins/{id}/sync", plugin.SyncUpdate).Methods("PATCH")

	// Marketplace
	r.HandleFunc("/marketplace/plugins", marketplace.GetAllPlugins).Methods("GET")
	r.HandleFunc("/marketplace/plugins/popular", marketplace.GetPopularPlugins).Methods("GET")
	r.HandleFunc("/marketplace/plugins/recommended", marketplace.GetRecomendedPlugins).Methods("GET")
	r.HandleFunc("/marketplace/plugins/search", marketplace.Search).Methods("GET")
	r.HandleFunc("/marketplace/plugins/{id}", marketplace.GetPlugin).Methods("GET")
	r.HandleFunc("/marketplace/plugins/{id}", marketplace.RemovePlugin).Methods("DELETE")

	// Users
	r.HandleFunc("/users", us.Create).Methods("POST")
	r.HandleFunc("/users/{user_id}", au.IsAuthenticated(au.IsAuthorized(us.UpdateUser, "zuri_admin"))).Methods("PATCH")
	r.HandleFunc("/users/{user_id}", au.IsAuthenticated(au.IsAuthorized(us.GetUser, "zuri_admin"))).Methods("GET")
	r.HandleFunc("/users/{user_id}", au.IsAuthenticated(au.IsAuthorized(us.DeleteUser, "zuri_admin"))).Methods("DELETE")
	r.HandleFunc("/users", au.IsAuthenticated(au.IsAuthorized(us.GetUsers, "zuri_admin"))).Methods("GET")
	r.HandleFunc("/users/{email}/organizations", au.IsAuthenticated(us.GetUserOrganizations)).Methods("GET")

	r.HandleFunc("/guests/invite", us.CreateUserFromUUID).Methods("POST")

	// Contact Us
	r.HandleFunc("/contact", au.OptionalAuthentication(contact.MailUs, au)).Methods("POST")

	// Realtime communications
	r.HandleFunc("/realtime/test", realtime.Test).Methods("GET")
	r.HandleFunc("/realtime/auth", realtime.Auth).Methods("POST")
	r.HandleFunc("/realtime/refresh", realtime.Refresh).Methods("POST")
	r.HandleFunc("/realtime/publish-event", realtime.PublishEvent).Methods("POST")
	r.Handle("/socket.io/", server)

	// Email subscription
	r.HandleFunc("/external/subscribe", exts.EmailSubscription).Methods("POST")
	r.HandleFunc("/external/download-client", exts.DownloadClient).Methods("GET")
	r.HandleFunc("/external/send-mail", exts.SendMail).Queries("custom_mail", "{custom_mail:[0-9]+}").Methods("POST")

	// file upload
	r.HandleFunc("/upload/file/{plugin_id}", au.IsAuthenticated(service.UploadOneFile)).Methods("POST")
	r.HandleFunc("/upload/files/{plugin_id}", au.IsAuthenticated(service.UploadMultipleFiles)).Methods("POST")
	r.HandleFunc("/upload/mesc/{apk_sec}/{exe_sec}", au.IsAuthenticated(service.MescFiles)).Methods("POST")
	r.HandleFunc("/delete/file/{plugin_id}", au.IsAuthenticated(service.DeleteFile)).Methods("DELETE")
	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))

	// graphql
	schema, _ := graphql.NewSchema(gql.LoadGraphQlSchema())
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
	r.Handle("/graphql", h)

	// Ping endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.GetSuccess("Server is live", nil, w)
	})

	// Home
	http.Handle("/", r)

	// Docs
	r.PathPrefix("/").Handler(http.StripPrefix("/docs", http.RedirectHandler("https://docs.zuri.chat/", http.StatusMovedPermanently)))

	return r
}

func main() {
	// Socket  events
	var Server = socketio.NewServer(nil)

	messaging.SocketEvents(Server)

	// load .env file if it exists
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	fmt.Println("Environment variables successfully loaded. Starting application...")

	// Set Stripe api key
	stripe.Key = os.Getenv("STRIPE_KEY")

	if err = utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		fmt.Println("Could not connect to MongoDB")
	}

	// sentry: enables reporting messages, errors, and panics.
	err = sentry.Init(sentry.ClientOptions{
		Dsn: "https://82e17f3bba86400a9a38e2437b884d4a@o1013682.ingest.sentry.io/5979019",
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// get PORT from environment variables
	port, _ := os.LookupEnv("PORT")
	if port == "" {
		port = "8000"
	}

	r := Router(Server)

	c := cors.AllowAll()

	h := RequestDurationMiddleware(r)

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, c.Handler(h)),
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//nolint:errcheck //CODEI8: ignore error check
	go Server.Serve()

	fmt.Println("Socket Served")

	defer Server.Close()

	fmt.Println("Zuri Chat API running on port ", port)
	//nolint:gocritic //CODEI8: please provide soln -> lint throw exitAfterDefer: log.Fatal will exit, and `defer Server.Close()` will not run
	log.Fatal(srv.ListenAndServe())
}

func LoadApp(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appID := params["appid"]

	fmt.Printf("URL called with Param: %s", appID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<div><b>Hello</b> World <button style='color: green;'>Click me!</button></div>: App = %s\n", appID)
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Zuri Chat API - Version 0.0255\n")
}

func RequestDurationMiddleware(h http.Handler) http.Handler {
	const durationLimit = 10

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)

		postToSlack := func() {
			m := make(map[string]interface{})
			m["timeTaken"] = duration.Seconds()

			if duration.Seconds() < durationLimit {
				return
			}

			scheme := "http"

			if r.TLS != nil {
				scheme += "s"
			}

			m["endpoint"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)
			m["timeTaken"] = duration.Seconds()

			b, _ := json.Marshal(m)
			resp, err := http.Post("https://companyfiles.zuri.chat/api/v1/slack/message", "application/json", strings.NewReader(string(b)))

			if err != nil {
				return
			}

			if resp.StatusCode != 200 {
				fmt.Printf("got error %d", resp.StatusCode)
			}

			defer resp.Body.Close()
		}

		if strings.Contains(r.Host, "api.zuri.chat") {
			go postToSlack()
		}
	})
}
