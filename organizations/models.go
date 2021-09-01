package organizations

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/models"
	"zuri.chat/zccore/utils"
)

const (
	OrganizationCollectionName = "organizations"
)


type Organization struct {
<<<<<<< HEAD
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


=======
	ID        primitive.ObjectID       `bson:"_id"`
	Name      string                   `bson:"name"`
	Email     string                   `bson:"email"`
	CreatorID string                   `bson:"creator_id"`
	Plugins   []map[string]interface{} `bson:"plugins"`
	Admins    []string                 `bson:"admins"`
	Settings  map[string]interface{}   `bson:"settings"`
	ImageURL  string                   `bson:"image_url"`
	CreatedAt time.Time                `bson:"created_at"`
	UpdatedAt time.Time                `bson:"updated_at"`
}

func (o *Organization) PopulatePlugins() {
	f := bson.M{"organization_id": o.ID.Hex()}
	docs, err := utils.GetMongoDbDocs(models.InstalledPluginsCollectionName, f)
	if err != nil {
		return
	}
	for _, doc := range docs {
		p := doc["plugin"].(bson.M)
		o.Plugins = append(o.Plugins, p)
	}
}

type InstalledPlugin struct {
	ID             primitive.ObjectID     `bson:"_id"`
	PluginID       string                 `bson:"plugin_id"`
	Plugin         map[string]interface{} `bson:"plugin"`
	OrganizationID string                 `bson:"organization_id"`
	AddedBy        string                 `bson:"added_by"`
	ApprovedBy     string                 `bson:"approved_by"`
	InstalledAt    time.Time              `bson:"installed_at"`
	UpdatedAt      time.Time              `bson:"updated_at"`
}

type OrganizationAdmin struct {
	ID             primitive.ObjectID `bson:"id"`
	OrganizationID string             `bson:"organization_id"`
	UserID         string             `bson:"user_id"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}
>>>>>>> 1bdef2a18b0a3cbd7c78dd942ac9eb158f42a1d3
