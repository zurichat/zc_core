package data

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginId, collName, orgId := vars["plugin_id"], vars["coll_name"], vars["org_id"]

	if !pluginHasCollection(pluginId, collName) {
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

func pluginHasCollection(pluginID, collectionName string) bool {
	filter := M{
		"plugin_id":       pluginID,
		"collection_name": collectionName,
	}
	_, err := utils.GetMongoDbDoc("plugin_collections", filter)
	if err == nil {
		return true
	}
	return false
}

func createPluginCollectionRecord(pluginID, collectionName string) error {
	doc := M{
		"plugin_id":       pluginID,
		"collection_name": collectionName,
		"created_at":      time.Now(),
	}

	if _, err := utils.CreateMongoDbDoc("plugin_collections", doc); err != nil {
		return err
	}

	return nil
}

func parseURLQuery(r *http.Request) map[string]interface{} {
	m := M{}
	for k, v := range r.URL.Query() {
		m[k] = v[0]
	}
	return m
}
