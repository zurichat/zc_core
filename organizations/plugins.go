package organizations

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/models"
	"zuri.chat/zccore/utils"
)

func GetPlugins(w http.ResponseWriter, r *http.Request) {
	orgId := mux.Vars(r)["org_id"]
	objId, _ := primitive.ObjectIDFromHex(orgId)

	_, err := utils.GetMongoDbDoc(models.OrganizationCollectionName, bson.M{"_id": objId})
	if err != nil {
		// org not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}
	org := models.Organization{ID: objId}
	org.PopulatePlugins()
	utils.GetSuccess("success", org.Plugins, w)
}
