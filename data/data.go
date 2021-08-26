package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PluginData (Data) is how we keep track of plugins and their related data
type PluginData struct {
	ID              primitive.ObjectID `json:"_id"`
	OwnerPluginID   primitive.ObjectID `json:"owner_plugin_id"`
	PluginAuthToken string             `json:"plugin_auth_token"`
	CollectionName  string             `json:"collection_name"`

	// an unordered sequence/list/array of mongodb collection data
	CollectionData []interface{} `json:"collection_data"`
	CreatedAt      time.Time     `json:"created_at"`
}
