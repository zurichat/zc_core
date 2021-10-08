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

// ReadData handles the data retrieval operation for plugins using GET requests.
func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID, collName, orgID := vars["plugin_id"], vars["coll_name"], vars["org_id"]

	if !pluginHasCollection(pluginID, orgID, collName) {
		utils.GetError(errors.New("collection not found"), http.StatusNotFound, w)
		return
	}

	prefixedCollName := getPrefixedCollectionName(pluginID, orgID, collName)
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
	m := utils.M{}
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
	ObjectIDs      []string               `json:"object_ids,omitempty"`
	Filter         map[string]interface{} `json:"filter,omitempty"`
	ReadOptions    *readOptions           `json:"options,omitempty"`
}

//
type readOptions struct {
	Limit *int64                 `json:"limit,omitempty"`
	Skip  *int64                 `json:"skip,omitempty"`
	Sort  map[string]interface{} `json:"sort,omitempty"`
}

// NewRead handles data retrieval process using POST requests, providing flexibility for the query
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
		doc, err := utils.GetMongoDbDoc(prefixedCollName, bson.M{"_id": id, "deleted": utils.M{"$ne": true}})
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("success", doc, w)
		return
	}
	filter := reqData.Filter
	if filter == nil {
		filter = utils.M{}
	}
	if reqData.ObjectIDs != nil {
		filter["_id"] = bson.M{"$in": hexToObjectIDs(reqData.ObjectIDs)}
	}
	var opts *options.FindOptions

	if r := reqData.ReadOptions; r != nil {
		opts = setOptions(*r)
	}

	filter["deleted"] = bson.M{"$ne": true}
	docs, err := utils.GetMongoDbDocs(prefixedCollName, filter, opts)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("success", docs, w)
}

func setOptions(r readOptions) *options.FindOptions {
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

func hexToObjectIDs(ids []string) []primitive.ObjectID {
	objIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objIDs[i] = mustObjectIDFromHex(id)
	}
	return objIDs
}
