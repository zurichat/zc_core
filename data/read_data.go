package data

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

// The idea is to use certain characters to signify queries
// we treat the `$` character as a query identifier then the characters that follow before a double underscore
// or colon is a query modifier e.g gte means greater/equal to. The next character is the field name and that the one after the = sign is the value.
// e.g ?$gte:first_name="meh" or ($gte__first_name="meh")
// We will split the field
type MongoQuery struct {
	Lt        string
	Gt        string
	Gte       string
	Lte       string
	In        string
	Nin       string
	Eq        string
	Ne        string
	And       string
	Or        string
	Not       string
	Nor       string
	All       string
	ElemMatch string
}

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
	ObjectID       *string                `json:"object_id,omitempty"`
	Filter         map[string]interface{} `json:"filter"`
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

	// If Object ID exists then filter by only that
	if reqData.ObjectID != nil {
		id, err := primitive.ObjectIDFromHex(*reqData.ObjectID)
		if err != nil {
			utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
			return
		}
		doc, err := utils.GetMongoDbDoc(prefixedCollName, bson.M{"_id": id})
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("success", doc, w)
		// Else use the filter object to filter
	} else {
		docs, err := utils.GetMongoDbDocs(prefixedCollName, reqData.Filter)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("success", docs, w)
	}
}
