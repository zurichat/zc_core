package data

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// This are contains fields that can be used to perform Read queries
type DataFilter struct {
	ID *primitive.ObjectID `schema:"id" bson:"_id"`
}

type M map[string]interface{}

func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collName, pluginID, orgID := vars["coll_name"], vars["plugin_id"], vars["org_id"]

	pluginData, err := findPluginDataByCollection(collName)

	if err != nil {
		utils.GetError(fmt.Errorf("error retreiving data: %v", err), http.StatusInternalServerError, w)
		return
	}

	if err := checkPluginDataCollectionOwner(pluginData, pluginID); err != nil {
		utils.GetError(err, http.StatusForbidden, w)
		return
	}

	// to get data for specific orgs, ensure org_id is in the url
	// every data a plugin is trying to write should have an org_id field so only data pertaining to that org is retrieved
	// extra filter parameters should passed by query params.
	filter := parseURLQuery(r)
	filter["org_id"], _ = primitive.ObjectIDFromHex(orgID)
	data, err := utils.GetMongoDbDocs(collName, filter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("success", data, w)
}

func findPluginDataByCollection(collName string) (bson.M, error) {
	filter := M{"collection_name": collName}
	return utils.GetMongoDbDoc(_COLLECTION_NAME, filter)
}

func checkPluginDataCollectionOwner(pluginData bson.M, pluginID string) error {
	ownerPluginID, ok := pluginData["owner_plugin_id"].(primitive.ObjectID)
	if !ok {
		// invalid plugin id
		return fmt.Errorf("invalid plugin ID")
	}

	if ownerPluginID.Hex() != pluginID {
		// not allowed to access collection that is not yours.
		return fmt.Errorf("not allowed to access this data")
	}
	return nil
}

var decoder = schema.NewDecoder()

func parseURLQuery(r *http.Request) map[string]interface{} {
	m := M{}
	df := &DataFilter{}
	if len(r.URL.Query()) == 0 {
		return m
	}
	_ = decoder.Decode(df, r.URL.Query())
	m, _ = utils.StructToMap(df, "bson")
	return m
}
