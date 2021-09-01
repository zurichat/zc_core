package plugin

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/models"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

var validate = validator.New()

func Register(w http.ResponseWriter, r *http.Request) {
	reqData := models.Plugin{}
	if err := utils.ParseJsonFromRequest(r, &reqData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(reqData); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	doc, _ := utils.StructToMap(reqData, "bson")
	res, err := utils.CreateMongoDbDoc(models.PluginCollectionName, doc)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	id, _ := res.InsertedID.(primitive.ObjectID)
	w.WriteHeader(http.StatusCreated)
	utils.GetSuccess("success", M{"plugin_id": id.Hex()}, w)
	go approvePlugin(id.Hex())
}

func GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objID, _ := primitive.ObjectIDFromHex(id)
	doc, err := utils.GetMongoDbDoc(models.PluginCollectionName, M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}
	utils.GetSuccess("success", doc, w)
}

// a hack to simulate plugin approval, it basically waits 10 seconds after creation and approves the plugin
func approvePlugin(id string) {
	time.Sleep(10 * time.Second)
	update := M{"approved": true, "approved_at": time.Now()}
	_, err := utils.UpdateOneMongoDbDoc(models.PluginCollectionName, id, update)
	if err != nil {
		log.Println("error approving plugin")
		return
	}
	log.Printf("Plugin %s approved\n", id)
}
