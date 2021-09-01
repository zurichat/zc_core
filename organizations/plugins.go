package organizations

import (
	"fmt"
	"net/http"

<<<<<<< HEAD
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

	orgName := InstalledPluginsCollectionName(doc["name"].(string))

	docs, err := utils.GetMongoDbDocs(orgName, nil)
	fmt.Println(orgName)
	if err != nil {
		// org plugins not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("Plugins returned successfully", docs, w)
=======
	// "github.com/gorilla/mux"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

func GetPlugins(w http.ResponseWriter, r *http.Request) {
	// orgId := mux.Vars(r)["org_id"]
	// objId, _ := primitive.ObjectIDFromHex(orgId)

	// _, err := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objId})
	// if err != nil {
	// 	// org not found.
	// 	utils.GetError(err, http.StatusNotFound, w)
	// 	return
	// }
	// org := Organization{ID: orgId}
	// org.PopulatePlugins()
	utils.GetSuccess("success", nil, w)
>>>>>>> 936244317d814074672f9a335dbef6094421bc0c
}
