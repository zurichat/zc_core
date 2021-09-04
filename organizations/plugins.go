package organizations

import (
	"net/http"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"zuri.chat/zccore/utils"
)

func GetOrganizationPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := "organizations"

	orgId := mux.Vars(r)["org_id"]
	objId, _ := primitive.ObjectIDFromHex(orgId)

	doc, err := utils.GetMongoDbDoc(collection, bson.M{"_id": objId})
	if err != nil {
		// org not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	orgName := GetOrgPluginCollectionName(doc["name"].(string))

	docs, err := utils.GetMongoDbDocs(orgName, nil)
  
	if err != nil {
		// org plugins not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("Plugins returned successfully", docs, w)
}
