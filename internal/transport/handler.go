package http

import (
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	gqlHandler "github.com/graphql-go/handler"

	"zuri.chat/zccore/agora"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/blog"
	"zuri.chat/zccore/contact"
	"zuri.chat/zccore/data"
	"zuri.chat/zccore/external"
	"zuri.chat/zccore/marketplace"
	"zuri.chat/zccore/organizations"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/realtime"
	"zuri.chat/zccore/report"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

type Handler struct {
	Router   *mux.Router
	SocketIO *socketio.Server
}

func NewHandler(server *socketio.Server) *Handler {
	return &Handler{
		SocketIO: server,
	}
}

func (h *Handler) SetupRoutes() {
	h.Router = mux.NewRouter().StrictSlash(true)

	// Load handlers(this to reduce dependency circle issue, might reverse if not working)
	configs := utils.NewConfigurations()
	mailService := service.NewZcMailService(configs)

	orgs := organizations.NewOrganizationHandler(configs, mailService)
	exts := external.NewExternalHandler(configs, mailService)
	reps := report.NewReportHandler(configs, mailService)
	au := auth.NewAuthHandler(configs, mailService)
	us := user.NewUserHandler(configs, mailService)
	gql := utils.NewGraphQlHandler(configs)

	// Agora
	ah := agora.NewAgoraHandler(configs)

	client := utils.GetDefaultMongoClient()
	ps := plugin.NewMongoService(client)
	ph := plugin.NewHandler(ps)

	// Setup and init
	h.Router.HandleFunc("/", VersionHandler)
	h.Router.HandleFunc("/loadapp/{appid}", LoadApp).Methods("GET")

	// Blog
	h.Router.HandleFunc("/posts", blog.GetPosts).Methods("GET")
	h.Router.HandleFunc("/posts", blog.CreatePost).Methods("POST")
	h.Router.HandleFunc("/posts/{post_id}", blog.UpdatePost).Methods("PUT")
	h.Router.HandleFunc("/posts/{post_id}", blog.DeletePost).Methods("DELETE")
	h.Router.HandleFunc("/posts/{post_id}", blog.GetPost).Methods("GET")
	h.Router.HandleFunc("/posts/{post_id}/like/{user_id}", blog.LikeBlog).Methods("PATCH")
	h.Router.HandleFunc("/posts/{post_id}/comments", blog.GetBlogComments).Methods("GET")
	h.Router.HandleFunc("/posts/{post_id}/comments", blog.CommentBlog).Methods("POST")
	h.Router.HandleFunc("/posts/search", blog.SearchBlog).Methods("GET")
	h.Router.HandleFunc("/posts/mail", blog.MailingList).Methods("POST")

	// Authentication
	h.Router.HandleFunc("/auth/login", au.LoginIn).Methods(http.MethodPost)
	h.Router.HandleFunc("/auth/logout", au.LogOutUser).Methods(http.MethodPost)
	h.Router.HandleFunc("/auth/logout/other-sessions", au.LogOutOtherSessions).Methods(http.MethodPost)
	h.Router.HandleFunc("/auth/verify-token", au.IsAuthenticated(au.VerifyTokenHandler)).Methods(http.MethodGet, http.MethodPost)
	h.Router.HandleFunc("/auth/confirm-password", au.IsAuthenticated(au.ConfirmUserPassword)).Methods(http.MethodPost)
	h.Router.HandleFunc("/auth/social-login/{provider}/{access_token}", au.SocialAuth).Methods(http.MethodGet)

	h.Router.HandleFunc("/account/verify-account", au.VerifyAccount).Methods(http.MethodPost)
	h.Router.HandleFunc("/account/request-password-reset-code", au.RequestResetPasswordCode).Methods(http.MethodPost)
	h.Router.HandleFunc("/account/verify-reset-password", au.VerifyPasswordResetCode).Methods(http.MethodPost)
	h.Router.HandleFunc("/account/update-password/{verification_code:[0-9]+}", au.UpdatePassword).Methods(http.MethodPost)

	// Organization
	h.Router.HandleFunc("/organizations", au.IsAuthenticated(orgs.Create)).Methods("POST")                                              // works
	h.Router.HandleFunc("/organizations", au.IsAuthenticated(orgs.GetOrganizations)).Methods("GET")                                     // works
	h.Router.HandleFunc("/organizations/{id}", au.IsAuthenticated(orgs.GetOrganization)).Methods("GET")                                 // works
	h.Router.HandleFunc("/organizations/{id}", au.IsAuthenticated(au.IsAuthorized(orgs.DeleteOrganization, "admin"))).Methods("DELETE") // worksxxx
	h.Router.HandleFunc("/organizations/url/{url}", orgs.GetOrganizationByURL).Methods("GET")                                           // works

	h.Router.HandleFunc("/organizations/{id}/url", au.IsAuthenticated(orgs.UpdateURL)).Methods("PATCH")   // works
	h.Router.HandleFunc("/organizations/{id}/name", au.IsAuthenticated(orgs.UpdateName)).Methods("PATCH") // works
	h.Router.HandleFunc("/organizations/{id}/logo", au.IsAuthenticated(orgs.UpdateLogo)).Methods("PATCH") // works

	h.Router.HandleFunc("/organizations/{id}/settings", au.IsAuthenticated(orgs.UpdateOrganizationSettings)).Methods("PATCH")     // works
	h.Router.HandleFunc("/organizations/{id}/permission", au.IsAuthenticated(orgs.UpdateOrganizationPermission)).Methods("PATCH") //works
	h.Router.HandleFunc("/organizations/{id}/auth", au.IsAuthenticated(orgs.UpdateOrganizationAuthentication)).Methods("PATCH")   // works
	h.Router.HandleFunc("/organizations/{id}/change-owner", au.IsAuthenticated(au.IsAuthorized(orgs.TransferOwnership, "owner"))).Methods("PATCH")

	h.Router.HandleFunc("/organizations/{id}/prefixes", au.IsAuthenticated(orgs.UpdateOrganizationPrefixes)).Methods("PATCH")       // fixed
	h.Router.HandleFunc("/organizations/{id}/slackbotresponses", au.IsAuthenticated(orgs.UpdateSlackBotResponses)).Methods("PATCH") // works
	h.Router.HandleFunc("/organizations/{id}/customemoji", au.IsAuthenticated(orgs.AddSlackCustomEmoji)).Methods("PATCH")           // works

	// Organization: Guest Invites
	h.Router.HandleFunc("/organizations/{id}/send-invite", au.IsAuthenticated(au.IsAuthorized(orgs.SendInvite, "admin"))).Methods("POST")  //works
	h.Router.HandleFunc("/organizations/{id}/invite-stats", au.IsAuthenticated(au.IsAuthorized(orgs.InviteStats, "admin"))).Methods("GET") // none
	h.Router.HandleFunc("/organizations/invites/{uuid}", orgs.CheckGuestStatus).Methods(http.MethodGet)                                    // none
	h.Router.HandleFunc("/organizations/guests/{uuid}", orgs.GuestToOrganization).Methods(http.MethodPost)                                 // test

	h.Router.HandleFunc("/organizations/{id}/plugins", au.IsAuthenticated(orgs.AddOrganizationPlugin)).Methods("POST")                  //works
	h.Router.HandleFunc("/organizations/{id}/plugins", au.IsAuthenticated(orgs.GetOrganizationPlugins)).Methods("GET")                  //works
	h.Router.HandleFunc("/organizations/{id}/plugins/{plugin_id}", au.IsAuthenticated(orgs.GetOrganizationPlugin)).Methods("GET")       //works
	h.Router.HandleFunc("/organizations/{id}/plugins/{plugin_id}", au.IsAuthenticated(orgs.RemoveOrganizationPlugin)).Methods("DELETE") //ask

	h.Router.HandleFunc("/organizations/{id}/members", au.IsAuthenticated(au.IsAuthorized(orgs.CreateMember, "admin"))).Methods("POST") // done
	h.Router.HandleFunc("/organizations/{id}/members", orgs.GetMembers).Methods("GET")                                                  // should work
	h.Router.HandleFunc("/organizations/{id}/members/multiple", au.IsAuthenticated(orgs.GetmultipleMembers)).Methods("GET")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}", au.IsAuthenticated(orgs.GetMember)).Methods("GET")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}", au.IsAuthenticated(au.IsAuthorized(orgs.DeactivateMember, "admin"))).Methods("DELETE")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/reactivate", au.IsAuthenticated(au.IsAuthorized(orgs.ReactivateMember, "admin"))).Methods("POST")

	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/status", au.IsAuthenticated(orgs.UpdateMemberStatus)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/status/remove-history/{history_index}", au.IsAuthenticated(orgs.RemoveStatusHistory)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/photo/{action}", au.IsAuthenticated(orgs.UpdateProfilePicture)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/profile", au.IsAuthenticated(orgs.UpdateProfile)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/uploadfile", au.IsAuthenticated(orgs.UploadFile)).Methods("POST")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/presence", au.IsAuthenticated(orgs.TogglePresence)).Methods("POST")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings", au.IsAuthenticated(orgs.UpdateMemberSettings)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/role", au.IsAuthenticated(au.IsAuthorized(orgs.UpdateMemberRole, "admin"))).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings/notification", au.IsAuthenticated(orgs.UpdateNotification)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings/theme", au.IsAuthenticated(orgs.UpdateUserTheme)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings/message-media", au.IsAuthenticated(orgs.UpdateMemberMessageAndMediaSettings)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings/accessibility", au.IsAuthenticated(orgs.UpdateMemberAccessibilitySettings)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings/languages-and-region", au.IsAuthenticated(orgs.UpdateLanguagesAndRegions)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/settings/advanced", au.IsAuthenticated(orgs.UpdateMemberAdvancedSettings)).Methods("PATCH")

	h.Router.HandleFunc("/organizations/{id}/reports", au.IsAuthenticated(reps.AddReport)).Methods("POST")
	h.Router.HandleFunc("/organizations/{id}/reports", au.IsAuthenticated(reps.GetReports)).Methods("GET")
	h.Router.HandleFunc("/organizations/{id}/reports/{report_id}", au.IsAuthenticated(reps.GetReport)).Methods("GET")

	h.Router.HandleFunc("/organizations/{id}/billing/settings", au.IsAuthenticated(orgs.UpdateBillingSettings)).Methods("PATCH")
	h.Router.HandleFunc("/organizations/{id}/billing/contact", au.IsAuthenticated(orgs.UpdateBillingContact)).Methods("PATCH")

	//organization: payment
	h.Router.HandleFunc("/organizations/{id}/add-token", au.IsAuthenticated(orgs.AddToken)).Methods("POST") //works
	h.Router.HandleFunc("/organizations/{id}/token-transactions", au.IsAuthenticated(orgs.GetTokenTransaction)).Methods("GET")
	h.Router.HandleFunc("/organizations/{id}/upgrade-to-pro", au.IsAuthenticated(orgs.UpgradeToPro)).Methods("POST")                   //works
	h.Router.HandleFunc("/organizations/{id}/charge-tokens", au.IsAuthenticated(orgs.ChargeTokens)).Methods("POST")                    //work
	h.Router.HandleFunc("/organizations/{id}/checkout-session", au.IsAuthenticated(orgs.CreateCheckoutSession)).Methods("POST")        //work
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/cards", au.IsAuthenticated(orgs.AddCard)).Methods("POST")                //works
	h.Router.HandleFunc("/organizations/{id}/members/{mem_id}/cards/{card_id}", au.IsAuthenticated(orgs.DeleteCard)).Methods("DELETE") //work

	// Data
	h.Router.HandleFunc("/data/write", data.WriteData)
	h.Router.HandleFunc("/data/read", data.NewRead).Methods("POST")
	h.Router.HandleFunc("/data/read/{plugin_id}/{coll_name}/{org_id}", data.ReadData).Methods("GET")
	h.Router.HandleFunc("/data/delete", data.DeleteData).Methods("POST")
	h.Router.HandleFunc("/data/collections/info/{plugin_id}/{coll_name}/{org_id}", data.CollectionDetail).Methods("GET")

	// Plugins
	h.Router.HandleFunc("/plugins/register", ph.Register).Methods("POST")
	h.Router.HandleFunc("/plugins/{id}", ph.Update).Methods("PATCH")
	h.Router.HandleFunc("/plugins/{id}", ph.Delete).Methods("DELETE")
	h.Router.HandleFunc("/plugins/{id}/sync", plugin.SyncUpdate).Methods("PATCH")

	// Marketplace
	h.Router.HandleFunc("/marketplace/plugins", marketplace.GetAllPlugins).Methods("GET")
	h.Router.HandleFunc("/marketplace/plugins/popular", marketplace.GetPopularPlugins).Methods("GET")
	h.Router.HandleFunc("/marketplace/plugins/recommended", marketplace.GetRecomendedPlugins).Methods("GET")
	h.Router.HandleFunc("/marketplace/plugins/search", marketplace.Search).Methods("GET")
	h.Router.HandleFunc("/marketplace/plugins/{id}", marketplace.GetPlugin).Methods("GET")
	h.Router.HandleFunc("/marketplace/plugins/urls/url", marketplace.GetPluginByURL).Methods("GET")
	h.Router.HandleFunc("/marketplace/plugins/{id}", marketplace.RemovePlugin).Methods("DELETE")

	// Users
	h.Router.HandleFunc("/users", us.Create).Methods("POST")
	h.Router.HandleFunc("/users/{user_id}", au.IsAuthenticated(au.IsAuthorized(us.UpdateUser, "zuri_admin"))).Methods("PATCH")
	h.Router.HandleFunc("/users/{user_id}", au.IsAuthenticated(au.IsAuthorized(us.GetUser, "zuri_admin"))).Methods("GET")
	h.Router.HandleFunc("/users/{user_id}", au.IsAuthenticated(au.IsAuthorized(us.DeleteUser, "zuri_admin"))).Methods("DELETE")
	h.Router.HandleFunc("/users", au.IsAuthenticated(au.IsAuthorized(us.GetUsers, "zuri_admin"))).Methods("GET")
	h.Router.HandleFunc("/users/{email}/organizations", au.IsAuthenticated(us.GetUserOrganizations)).Methods("GET")

	h.Router.HandleFunc("/guests/invite", us.CreateUserFromUUID).Methods("POST")

	// Contact Us
	h.Router.HandleFunc("/contact", au.OptionalAuthentication(contact.MailUs)).Methods("POST")

	// Realtime communications
	h.Router.HandleFunc("/realtime/test", realtime.Test).Methods("GET")
	h.Router.HandleFunc("/realtime/auth", realtime.Auth).Methods("POST")
	h.Router.HandleFunc("/realtime/refresh", realtime.Refresh).Methods("POST")
	h.Router.HandleFunc("/realtime/publish-event", realtime.PublishEvent).Methods("POST")
	h.Router.Handle("/socket.io/", h.SocketIO)

	// Email subscription
	h.Router.HandleFunc("/external/subscribe", exts.EmailSubscription).Methods("POST")
	h.Router.HandleFunc("/external/unsubscribe/{email}", exts.UnsubscribeEmail).Methods("GET")
	h.Router.HandleFunc("/external/download-client", exts.DownloadClient).Methods("GET")
	h.Router.HandleFunc("/external/send-mail", exts.SendMail).Queries("custom_mail", "{custom_mail:[0-9]+}").Methods("POST")

	// file upload
	h.Router.HandleFunc("/upload/file/{plugin_id}", au.IsAuthenticated(service.UploadOneFile)).Methods("POST")
	h.Router.HandleFunc("/upload/files/{plugin_id}", au.IsAuthenticated(service.UploadMultipleFiles)).Methods("POST")
	h.Router.HandleFunc("/upload/mesc/{apk_sec}/{exe_sec}", au.IsAuthenticated(service.MescFiles)).Methods("POST")
	h.Router.HandleFunc("/delete/file/{plugin_id}", au.IsAuthenticated(service.DeleteFile)).Methods("DELETE")
	h.Router.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))

	// Agora token generator
	h.Router.HandleFunc("/rtc/{channelName}/{role}/{tokentype}", ah.GetRtcToken).Methods("GET")

	// graphql
	schema, _ := graphql.NewSchema(gql.LoadGraphQlSchema())
	gh := gqlHandler.New(&gqlHandler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
	h.Router.Handle("/graphql", gh)

	// Ping endpoint
	h.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.GetSuccess("Server is live", nil, w)
	})

	// Home
	http.Handle("/", h.Router)

	// Docs
	h.Router.PathPrefix("/").Handler(http.StripPrefix("/docs", http.RedirectHandler("https://docs.zuri.chat/", http.StatusMovedPermanently)))
}
