package organizations

import (
	"encoding/json"
	"net/http"
	"zuri.chat/zccore/utils"
)

func GetPluginsOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	plugins_collection := "plugins"

	decoder := json.NewDecoder(r.Body)
    var org map[string]interface{}
    err := decoder.Decode(&org)
    if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
    }


	/* filter := map[string]interface{}{
		"organisation_id" : "1",
	} */
	
	result, err := utils.GetMongoDbDocs(plugins_collection, org)

	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
	}

	utils.GetSuccess("Plugins returned successfully", result, w)
}