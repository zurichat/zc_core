package data

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginId, collName, orgId := vars["plugin_id"], vars["coll_name"], vars["org_id"]

	if !pluginHasCollection(pluginId, orgId, collName) {
		utils.GetError(errors.New("collection not found"), http.StatusNotFound, w)
		return
	}

	prefixedCollName := getPrefixedCollectionName(pluginId, orgId, collName)
	filter := parseURLQuery(r) // queries will have to be sanitized
	filter["deleted"] = bson.M{"$ne": true}
	docs, err := utils.GetMongoDbDocs(prefixedCollName, filter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if _, exists := filter["_id"]; exists && len(docs) == 1 {
		utils.GetSuccess("success", docs[0], w)
		return
	}
	utils.GetSuccess("success", docs, w)
}

// used to add a prefix to a collection name so we can generate unique collection names for different organizations for a plugin.
func getPrefixedCollectionName(pluginID, orgID, collName string) string {
	return fmt.Sprintf("%s:%s:%s", pluginID, orgID, collName)
}

func parseURLQuery(r *http.Request) map[string]interface{} {
	m := M{}
	for k, v := range r.URL.Query() {
		if k == "id" || k == "_id" {
			m["_id"], _ = primitive.ObjectIDFromHex(v[0])
			continue
		}
		m[k] = v[0]
	}
	return m
}

// Testing
type readDataRequest struct {
	PluginID       string                 `json:"plugin_id"`
	CollectionName string                 `json:"collection_name"`
	OrganizationID string                 `json:"organization_id"`
	ObjectID       string                 `json:"object_id,omitempty"`
	Filter         map[string]interface{} `json:"filter"`
	*ReadOptions   `json:"options,omitempty"`
}

type ReadOptions struct {
	Limit *int64                 `json:"limit,omitempty"`
	Skip  *int64                 `json:"skip,omitempty"`
	Sort  map[string]interface{} `json:"sort,omitempty"`
}

func NewRead(w http.ResponseWriter, r *http.Request) {
	reqData := new(readDataRequest)

	// Parse Body of request into readDataRequest struct
	if err := utils.ParseJsonFromRequest(r, reqData); err != nil {
		utils.GetError(fmt.Errorf("error processing request: %v", err), http.StatusUnprocessableEntity, w)
		return
	}

	// Checks if such collection was created by the plugin
	if !pluginHasCollection(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName) {
		utils.GetError(errors.New("collection not found"), http.StatusNotFound, w)
		return
	}

	prefixedCollName := getPrefixedCollectionName(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName)

	if reqData.ObjectID != "" {
		id, err := primitive.ObjectIDFromHex(reqData.ObjectID)
		if err != nil {
			utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
			return
		}
		doc, err := utils.GetMongoDbDoc(prefixedCollName, bson.M{"_id": id, "deleted": M{"$ne": true}})
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("success", doc, w)
	} else {
		filter := reqData.Filter
		if filter == nil {
			filter = M{}
		}
		var opts *options.FindOptions

		if r := reqData.ReadOptions; r != nil {
			opts = SetOptions(*r)
		}

		filter["deleted"] = bson.M{"$ne": true}
		docs, err := utils.GetMongoDbDocs(prefixedCollName, filter, opts)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("success", docs, w)
	}
}

func SetOptions(r ReadOptions) *options.FindOptions {
	findOptions := options.Find()
	if r.Limit != nil {
		findOptions.SetLimit(*r.Limit)
	}
	if r.Skip != nil {
		findOptions.SetSkip(*r.Skip)
	}
	if len(r.Sort) > 0 {
		findOptions.SetSort(r.Sort)
	}
	return findOptions
}
