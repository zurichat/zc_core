package data

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

type Collection struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	PluginID string             `json:"plugin_id" bson:"plugin_id"`
}

// CollectionDetail returns details about a collection.
func CollectionDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pluginID, orgID, collName := vars["plugin_id"], vars["org_id"], vars["coll_name"]

	if orgID == "__none__" {
		orgID = ""
	}

	actualCollName := mongoCollectionName(pluginID, collName)

	coll := utils.GetCollection(actualCollName)
	count, err := coll.CountDocuments(r.Context(), bson.M{"organization_id": orgID})

	if err != nil {
		utils.GetError(fmt.Errorf("unable to get collection details: %v", err), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("success", utils.M{
		"count": count,
	}, w)
}

func SaveCollection(name, pluginID string) error {
	coll := utils.GetCollection("collections_record")
	_, err := coll.InsertOne(context.TODO(), &Collection{Name: name, PluginID: pluginID})

	return err
}

func FindPluginCollections(ctx context.Context, pluginID string) ([]*Collection, error) {
	coll := utils.GetCollection("collections_record")
	cursor, err := coll.Find(ctx, bson.M{"plugin_id": pluginID})

	if err != nil {
		return nil, err
	}

	//nolint:gomnd // just no i'm not removing magic constant.
	results := make([]*Collection, 0, 10)

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
