package organizations

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrganizationCollectionName = "organizations"
)

type Organization struct {
	_id       string					`json:"id" bson:"_id"`
	Name      string                    `json:"name" bson:"name"`
	Email     string                   	`json:"email" bson:"email"`
	CreatorID string                   	`json:"creator_id" bson:"creator_id"`
	Plugins   []map[string]interface{} 	`json:"plugins" bson:"plugins"`
	Admins    []string                 	`json:"admins" bson:"admins"`
	Settings  map[string]interface{}   	`json:"settings" bson:"settings"`
	ImageURL  string                   	`json:"image_url" bson:"image_url"`
	CreatedAt time.Time                	`json:"created_at" bson:"created_at"`
	UpdatedAt time.Time                	`json:"updated_at" bson:"updated_at"`
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
