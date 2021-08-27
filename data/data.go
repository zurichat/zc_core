package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const _COLLECTION_NAME = "plugin_data"

// PluginData (Data) is how we keep track of plugins and manage access to their required data
type PluginData struct {
	ID              primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	OwnerPluginID   primitive.ObjectID `bson:"owner_plugin_id" json:"owner_plugin_id"`
	PluginAuthToken string             `bson:"plugin_auth_token" json:"plugin_auth_token"`
	CollectionName  string             `bson:"collection_name" json:"collection_name"`
	CollectionData  []interface{}      `bson:"collection_data" json:"collection_data"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

func createPluginData(pd *PluginData, ownerPluginHexID string) error {
	// find owner plugin
	filter := make(map[string]interface{})
	objID, _ := primitive.ObjectIDFromHex(ownerPluginHexID)
	filter["_id"] = objID
	p, err := utils.GetMongoDbDoc("plugins", filter)
	if err != nil {
		return err
	}
	// attach to plugin data
	pd.OwnerPluginID = p["_id"].(primitive.ObjectID) // type assertion since the result is interface{}

	m, _ := utils.StructToMap(pd, "bson")
	m["created_at"] = time.Now()

	res, err := utils.CreateMongoDbDoc(_COLLECTION_NAME, m)
	if err != nil {
		return err
	}

	pd.ID = res.InsertedID.(primitive.ObjectID)
	pd.CreatedAt = m["created_at"].(time.Time)
	return nil
}
