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
    Apps            [] string   `json:""`
	// Settings
	// DateUpdated    time.Time
}