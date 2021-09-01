package organizations

import (
	"time"
)

type Organization struct {
	_id            	string		`json:"id" bson:"_id"`
	UserId          string		`json:"user_id" bson:"user_id"`
	OwnerEmail  	string		`json:"owner_email" bson:"owner_email"`
	Url  			string		`json:"url" bson:"url"`
	Name			string		`json:"name" bson:"name"`
	LogoUrl  		string		`json:"logo_url" bson:"logo_url"`
	DateCreated     time.Time	`json:"date_created" bson:"date_created"`
	DateUpdated     time.Time	`json:"date_updated" bson:"date_updated"`
	// Apps
	// Settings
	// DateUpdated    time.Time
}

// Channel Model
type Channel struct {
	_id            	string		`json:"id" bson:"_id"`
	Name			string		`json:"name" bson:"name"`
	Description     string		`json:"description" bson:"description"`
	isPrivate       bool        `json:"is_private" bson:"is_private"`
	Bookmark 	    []string	`json:"bookmark" bson:"bookmark"`
	User  			[]string	`json:"user" bson:"user"`
	DateCreated     time.Time	`json:"date_created" bson:"date_created"`
	DateUpdated     time.Time	`json:"date_updated" bson:"date_updated"`
}


