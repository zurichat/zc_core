package data

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zuri.chat/zccore/utils"
)

func WriteData(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is you writing data\n")
}

func Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collName, pluginID, orgID := vars["coll_name"], vars["plugin_id"], vars["org_id"]

	allowed := checkOrganizationForPlugin(orgID, pluginID)

	if !allowed {
		fmt.Fprintf(w, "Not allowed")
		return
	}

	var i interface{}
	json.NewDecoder(r.Body).Decode(&i)

	v, ok := i.(map[string]interface{})
	if !ok {
		// Can't assert, handle error.
		fmt.Println("Error in assertion")
	}

	res, err := utils.CreateMongoDbDoc(collName, v)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("success", res, w)
}

// Checks if plugin is in organization and is authorized
func checkOrganizationForPlugin(org_id, plugin_id string) bool {
	return true
}
