package utils

import (
	"time"

	"github.com/spf13/viper"
)

// centralize config file using viper
type Configurations struct {
	ClusterUrl          string
	DbName              string
	SecretKey           string
	SessionKey          string
	SessionDbCollection string
	SessionMaxAge       int
	UserDbCollection    string
	SendGridApiKey      string

	ESPType       string
	SmtpUsername  string
	SmtpPassword  string
	SendgridEmail string

	MailGunKey         string
	MailGunDomain      string
	MailGunSenderEmail string

	EmailSubscriptionTemplate  string
	ConfirmEmailTemplate       string
	PasswordResetTemplate      string
	DownloadClientTemplate     string
	WorkspaceInviteTemplate    string
	TokenBillingNoticeTemplate string
	WorkSpaceInviteTemplate	   string
	WorkSpaceWelcomeTemplate   string

	CentrifugoKey      string
	CentrifugoEndpoint string

	GoogleOAuthUrl   string
	GoogleOAuthV3Url string
	FacebookOAuthUrl string

	HmacSampleSecret string
}

func NewConfigurations() *Configurations {
	// Load environmental variables
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.ReadInConfig()

	viper.AutomaticEnv()

	mgURL := viper.GetString("CLUSTER_URL")
	viper.SetDefault("DB_NAME", "zurichat")
	viper.SetDefault("SECRET_KEY", "5d5c7f94e29ba12a21f682be310d3af4")
	viper.SetDefault("SESSION_KEY", "f6822af94e29ba112be310d3af45d5c7")
	viper.SetDefault("HMAC_SECRET", "u7b8be9bd9b9ebd9b9dbdbee")
	viper.SetDefault("SESSION_MAX_AGE", time.Now().Unix()+(31536000*200))
	viper.SetDefault("USER_COLLECTION", "users")
	viper.SetDefault("SESSION_COLLECTION", "session_store")
	viper.SetDefault("CONFIRM_EMAIL_TEMPLATE", "./templates/confirm_email.html")
	viper.SetDefault("PASSWORD_RESET_TEMPLATE", "./templates/password_reset.html")
	viper.SetDefault("EMAIL_SUBSCRIPTION_TEMPLATE", "./templates/email_subscription.html")
	viper.SetDefault("DOWNLOAD_CLIENT_TEMPLATE", "./templates/download_clients.html")
	viper.SetDefault("WORKSPACE_INVITE_TEMPLATE", "./templates/workspace_invite.html")
	viper.SetDefault("TOKEN_BILLING_NOTICE_TEMPLATE", "./templates/token_billing_notice.html")
	viper.SetDefault("WORKSPACE_INVITE_TEMPLATE", "./templates/workspace_inivte.html")
	viper.SetDefault("WORKSPACE_WELCOME_TEMPLATE", "./templates/workspace_welcome.html")
	viper.SetDefault("GOOGLE_OAUTH_V3", "https://www.googleapis.com/oauth2/v3/userinfo?access_token=:access_token")

	configs := &Configurations{
		ClusterUrl:          mgURL,
		DbName:              viper.GetString("DB_NAME"),
		SecretKey:           viper.GetString("SECRET_KEY"),
		SessionKey:          viper.GetString("SESSION_KEY"),
		SessionDbCollection: viper.GetString("SESSION_COLLECTION"),
		SessionMaxAge:       viper.GetInt("SESSION_MAX_AGE"),
		UserDbCollection:    viper.GetString("USER_COLLECTION"),
		SendGridApiKey:      viper.GetString("SENDGRID_API_KEY"),
		ESPType:             viper.GetString("ESP_TYPE"),

		ConfirmEmailTemplate:       viper.GetString("CONFIRM_EMAIL_TEMPLATE"),
		PasswordResetTemplate:      viper.GetString("PASSWORD_RESET_TEMPLATE"),
		DownloadClientTemplate:     viper.GetString("DOWNLOAD_CLIENT_TEMPLATE"),
		WorkspaceInviteTemplate:    viper.GetString("WORKSPACE_INVITE_TEMPLATE"),
		EmailSubscriptionTemplate:  viper.GetString("EMAIL_SUBSCRIPTION_TEMPLATE"),
		TokenBillingNoticeTemplate: viper.GetString("TOKEN_BILLING_NOTICE_TEMPLATE"),
		WorkSpaceInviteTemplate: 	viper.GetString("WORKSPACE_INVITE_TEMPLATE"),
		WorkSpaceWelcomeTemplate: 	viper.GetString("WORKSPACE_WELCOME_TEMPLATE"),


		SmtpUsername:  viper.GetString("SMTP_USERNAME"),
		SmtpPassword:  viper.GetString("SMTP_PASSWORD"),
		SendgridEmail: viper.GetString("SENDGRID_EMAIL"),

		MailGunKey:         viper.GetString("MAILGUN_KEY"),
		MailGunDomain:      viper.GetString("MAILGUN_DOMAIN"),
		MailGunSenderEmail: viper.GetString("MAILGUN_EMAIL"),

		CentrifugoKey:      viper.GetString("CENTRIFUGO_KEY"),
		CentrifugoEndpoint: viper.GetString("CENTRIFUGO_ENDPOINT"),

		GoogleOAuthUrl:   viper.GetString("GOOGLE_OAUTH"),
		GoogleOAuthV3Url: viper.GetString("GOOGLE_OAUTH_V3"),
		FacebookOAuthUrl: viper.GetString("FACEBOOK_OAUTH"),

		HmacSampleSecret: viper.GetString("HMAC_SECRET"),
	}

	return configs
}
