package plugin

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

var validate = validator.New()

func Register(w http.ResponseWriter, r *http.Request) {
	p := Plugin{}
	if err := utils.ParseJsonFromRequest(r, &p); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(p); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	if err := CreatePlugin(&p); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	utils.GetSuccess("success", M{"plugin_id": p.ID.Hex()}, w)
	go approvePlugin(p.ID.Hex())
}

func GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p, err := FindPluginByID(id)
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}
	utils.GetSuccess("success", p, w)
}

// a hack to simulate plugin approval, it basically waits 10 seconds after creation and approves the plugin
func approvePlugin(id string) {
	time.Sleep(10 * time.Second)
	update := M{"approved": true, "approved_at": time.Now().String()}
	_, err := utils.UpdateOneMongoDbDoc(PluginCollectionName, id, update)
	if err != nil {
		log.Println("error approving plugin")
		return
	}
	log.Printf("Plugin %s approved\n", id)
}
