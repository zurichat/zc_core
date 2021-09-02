package data

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

// The idea is to use certain characters to signify queries
// we treat the `$` character as a query identifier then the characters that follow before a double underscore
// or colon is a query modifier e.g gte means greater/equal to. The next character is the field name and that the one after the = sign is the value.
// e.g ?$gte:first_name="meh" or ($gte__first_name="meh")
// We will split the field
type MongoQuery struct {
	LT  string
	GT  string
	GTE string
	LTE string
}

func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginId, collName, orgId := vars["plugin_id"], vars["coll_name"], vars["org_id"]

	if !pluginHasCollection(pluginId, orgId, collName) {
		utils.GetError(errors.New("record not found"), http.StatusNotFound, w)
		return
	}
	// proceed to perform read operation taking queries passed from request
	// queries are created via query parameters in the url
	prefixedCollName := getPrefixedCollectionName(pluginId, orgId, collName)
	filter := parseURLQuery(r)
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
		m[k] = v[0]
	}
	return m
}
