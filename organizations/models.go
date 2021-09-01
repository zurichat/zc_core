package organizations

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const (
	OrganizationCollectionName = "organizations"
)

type Organization struct {
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
	docs, err := utils.GetMongoDbDocs("installed_plugins", f)
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
