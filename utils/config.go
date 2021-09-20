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

	ConfirmEmailTemplate  string
	PasswordResetTemplate string
	TeamInvitationTemplate string
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
	viper.SetDefault("SESSION_MAX_AGE", time.Now().Unix()+(31536000*200))
	viper.SetDefault("USER_COLLECTION", "users")
	viper.SetDefault("SESSION_COLLECTION", "session_store")
	viper.SetDefault("CONFIRM_EMAIL_TEMPLATE", "./templates/confirm_email.html")
	viper.SetDefault("PASSWORD_RESET_TEMPLATE", "./templates/password_reset.html")
	viper.SetDefault("TEAM_INVITATION_TEMPLATE", "./templates/team_invite.html")

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

		ConfirmEmailTemplate:  viper.GetString("CONFIRM_EMAIL_TEMPLATE"),
		PasswordResetTemplate: viper.GetString("PASSWORD_RESET_TEMPLATE"),
		TeamInvitationTemplate: viper.GetString("TEAM_INVITATION_TEMPLATE"),

		SmtpUsername:  viper.GetString("SMTP_USERNAME"),
		SmtpPassword:  viper.GetString("SMTP_PASSWORD"),
		SendgridEmail: viper.GetString("SENDGRID_EMAIL"),

		MailGunKey:         viper.GetString("MAILGUN_KEY"),
		MailGunDomain:      viper.GetString("MAILGUN_DOMAIN"),
		MailGunSenderEmail: viper.GetString("MAILGUN_EMAIL"),
	}

	return configs
}
