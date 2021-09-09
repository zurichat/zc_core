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
	utils.GetSuccess("success", docs, w)
}

// used to add a prefix to a collection name so we can generate unique collection names for different organizations for a plugin.
func getPrefixedCollectionName(pluginID, orgID, collName string) string {
	return fmt.Sprintf("%s:%s:%s", pluginID, orgID, collName)
}

func parseURLQuery(r *http.Request) map[string]interface{} {
	m := M{}
	for k, v := range r.URL.Query() {
		if k == "_id" {
			m[k], _ = primitive.ObjectIDFromHex(v[0])
			continue
		}
		m[k] = v[0]
	}
	return m
}
