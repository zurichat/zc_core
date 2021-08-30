package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Plugin struct {
	ID           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	Description  string             `bson:"description"`
	InstallCount int64              `bson:"install_count"`
	Approved     string             `bson:"approved"`
	ApprovedAt   time.Time          `bson:"approved_at"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

// PluginCollection is used internally to keep track collections a plugin created.
type PluginCollection struct {
	ID             primitive.ObjectID `bson:"_id"`
	PluginID       string             `bson:"plugin_id"`
	CollectionName string             `bson:"collection_name"`
	CreatedAt      time.Time          `bson:"created_at"`
}

func NewPluginModel(name, description string) *Plugin {
	return &Plugin{
		Name:        name,
		Description: description,
	}
}
