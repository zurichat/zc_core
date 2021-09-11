package data

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const (
	_PluginCollectionName            = "plugins"
	_PluginCollectionsCollectionName = "plugin_collections"
	_OrganizationCollectionName      = "organizations"
)

// PluginCollections is used internally to keep track collections a plugin created.
type PluginCollections struct {
	ID             primitive.ObjectID `bson:"_id"`
	PluginID       string             `bson:"plugin_id"`
	OrganizationID string             `bson:"organization_id"`
	CollectionName string             `bson:"collection_name"`
	CreatedAt      time.Time          `bson:"created_at"`
}

func pluginHasCollection(pluginID, orgID, collectionName string) bool {
	filter := M{
		"plugin_id":       pluginID,
		"collection_name": collectionName,
		"organization_id": orgID,
	}
	_, err := utils.GetMongoDbDoc(_PluginCollectionsCollectionName, filter)
	if err == nil {
		return true
	}
	return false
}

func createPluginCollectionRecord(pluginID, orgID, collectionName string) error {
	doc := M{
		"plugin_id":       pluginID,
		"organization_id": orgID,
		"collection_name": collectionName,
		"created_at":      time.Now(),
	}

	if _, err := utils.CreateMongoDbDoc(_PluginCollectionsCollectionName, doc); err != nil {
		return err
	}
	return nil
}

func getPluginCollections(pluginId string) ([]bson.M, error) {
	docs, err := utils.GetMongoDbDocs(_PluginCollectionsCollectionName, bson.M{"plugin_id": pluginId})
	if err != nil {
		return nil, fmt.Errorf("Error finding collection records for this plugin: %v", err)
	}
	for _, doc := range docs {
		delete(doc, "_id")
	}
	return docs, nil
}

func getPluginCollectionsForOrganization(pluginId, orgId string) ([]bson.M, error) {
	docs, err := utils.GetMongoDbDocs(_PluginCollectionsCollectionName, bson.M{
		"plugin_id":       pluginId,
		"organization_id": orgId,
	})
	if err != nil {
		return nil, fmt.Errorf("Error finding collection records for this plugin: %v", err)
	}
	for _, doc := range docs {
		delete(doc, "_id")
	}
	return docs, nil
}
