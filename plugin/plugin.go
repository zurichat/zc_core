package plugin

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)


func SyncUpdate(w http.ResponseWriter, r *http.Request) {
	pp := SyncUpdateRequest{}

	ppID, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])

	if err != nil {
		utils.GetError(errors.WithMessage(err, "incorrect id"), http.StatusUnprocessableEntity, w)
		return
	}

	//nolint:govet //dod-san: ignore error shadowing.
	if err := utils.ParseJSONFromRequest(r, &pp); err != nil {
		utils.GetError(errors.WithMessage(err, "error processing request"), http.StatusUnprocessableEntity, w)
		return
	}

	pluginDetails, _ := utils.GetMongoDBDoc(PluginCollectionName, bson.M{"_id": ppID})

	if pluginDetails == nil {
		utils.GetError(errors.WithMessage(fmt.Errorf("plugin not found"), "error processing request"), http.StatusUnprocessableEntity, w)
		return
	}

	var splugin Plugin

	if err = mapstructure.Decode(pluginDetails, &splugin); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	sort.SliceStable(splugin.Queue, func(i, j int) bool {
		return splugin.Queue[i].ID < splugin.Queue[j].ID
	})

	for i := 0; i < len(splugin.Queue); i++ {
		onestruct := splugin.Queue[i]
		if onestruct.ID <= pp.ID {
			splugin.Queue = append(splugin.Queue[:i], splugin.Queue[i+1:]...)
			i-- // Important: decrease index
		}
	}

	updateFields := make(map[string]interface{})

	updateFields["queue"] = splugin.Queue
	_, ee := utils.UpdateOneMongoDBDoc(PluginCollectionName, mux.Vars(r)["id"], updateFields)

	if ee != nil {
		utils.GetError(ee, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("synchronization updated successful", nil, w)
}

