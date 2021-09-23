package plugin

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"zuri.chat/zccore/utils"
)

type M = map[string]interface{}

var validate = validator.New()

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	p := Plugin{}
	if err := utils.ParseJsonFromRequest(r, &p); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(p); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	if err := CreatePlugin(r.Context(), &p); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("success", M{"plugin_id": p.ID.Hex()}, w)
	go approvePlugin(p.ID.Hex())
}

// a hack to simulate plugin approval, it basically waits 10 seconds after creation and approves the plugin
func approvePlugin(id string) {
	time.Sleep(10 * time.Second)
	update := M{"approved": true, "deleted": false, "approved_at": time.Now().String()}
	_, err := utils.UpdateOneMongoDbDoc(PluginCollectionName, id, update)
	if err != nil {
		log.Println("error approving plugin")
		return
	}
	log.Printf("Plugin %s approved\n", id)
}

func Update(w http.ResponseWriter, r *http.Request) {
	pp := PluginPatch{}
	id := mux.Vars(r)["id"]
	if err := utils.ParseJsonFromRequest(r, &pp); err != nil {
		utils.GetError(errors.WithMessage(err, "error processing request"), 422, w)
		return
	}
	if err := updatePlugin(r.Context(), id, pp); err != nil {
		utils.GetError(errors.WithMessage(err, "cannot update, bad request"), 400, w)
		return
	}

	utils.GetSuccess("updated plugin successfully", nil, w)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, err := utils.UpdateOneMongoDbDoc("plugins", id, M{"deleted": true, "deleted_at": time.Now().String()})
	if err != nil {
		utils.GetError(errors.WithMessage(err, "error deleting plugin"), 400, w)
		return
	}
	w.WriteHeader(204)
	w.Header().Set("content-type", "application/json")
	utils.GetSuccess("plugin deleted", nil, w)
}
