package marketplace

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/models"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func GetAllApprovedPlugins(w http.ResponseWriter, r *http.Request) {

	docs, err := utils.GetMongoDbDocs(models.PluginCollectionName, M{"approved": true})
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	// we will not expose all plugin info at marketplace, only name, description and iconurl
	mm := make([]M, len(docs))
	for i, doc := range docs {
		m := M{
			"id":          doc["_id"],
			"name":        doc["name"],
			"description": doc["description"],
			"icon_url":    doc["icon_url"],
		}
		mm[i] = m
	}
	utils.GetSuccess("success", mm, w)
}

func GetOneApprovedPlugin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objId, _ := primitive.ObjectIDFromHex(id)
	doc, err := utils.GetMongoDbDoc(models.PluginCollectionName, M{"_id": objId, "approved": true})
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	// we will not expose all plugin info at marketplace, only name, description and iconurl
	m := M{
		"id":          doc["_id"],
		"name":        doc["name"],
		"description": doc["description"],
		"icon_url":    doc["icon_url"],
		"install_url": doc["install_url"],
	}

	utils.GetSuccess("success", m, w)
}
