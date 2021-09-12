package utils

import "github.com/spf13/viper"

// centralize config file using viper
type Configuration struct {
	ClusterUrl				string
	DbName					string
	SecretKey				string
	SessionKey				string
	SessionDbCollection		string
	SessionMaxAge			int
	UserDbCollection		string
	SendGridApiKey			string
}

func NewConfiguration() *Configuration {
	// Load environmental variables
	viper.AutomaticEnv()
	
	mgURL := viper.GetString("CLUSTER_URL")
	viper.SetDefault("DB_NAME", "zurichat")
	viper.SetDefault("SECRET_KEY", "5d5c7f94e29ba12a21f682be310d3af4")
	viper.SetDefault("SESSION_KEY", "f6822af94e29ba112be310d3af45d5c7")
	viper.SetDefault("SESSION_MAX_AGE", 60 * 60 * 12)
	viper.SetDefault("USER_COLLECTION", "users")
	viper.SetDefault("SESSION_COLLECTION", "session_store")

	configs := &Configuration{
		ClusterUrl: mgURL,
		DbName: viper.GetString("DB_NAME"),
		SecretKey: viper.GetString("SECRET_KEY"),
		SessionKey: viper.GetString("SESSION_KEY"),
		SessionDbCollection: viper.GetString("SESSION_COLLECTION"),
		SessionMaxAge: viper.GetInt("SESSION_MAX_AGE"),
		UserDbCollection: viper.GetString("USER_COLLECTION"),
		SendGridApiKey: viper.GetString("SENDGRID_API_KEY"),
	}
	
	return configs
}