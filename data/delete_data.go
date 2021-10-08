package data

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

type deleteDataRequest struct {
	PluginID       string                 `json:"plugin_id"`
	CollectionName string                 `json:"collection_name"`
	OrganizationID string                 `json:"organization_id"`
	BulkDelete     bool                   `json:"bulk_delete"`
	ObjectID       string                 `json:"object_id,omitempty"`
	Filter         map[string]interface{} `json:"filter"`
}

// DeleteData handles request for plugins to delete their data.
func DeleteData(w http.ResponseWriter, r *http.Request) {
	reqData := new(deleteDataRequest)

	if err := utils.ParseJsonFromRequest(r, reqData); err != nil {
		utils.GetError(fmt.Errorf("error processing request: %v", err), http.StatusUnprocessableEntity, w)
		return
	}

	if _, err := plugin.FindPluginByID(r.Context(), reqData.PluginID); err != nil {
		utils.GetError(fmt.Errorf("error retrieving plugin with id %v", reqData.PluginID), http.StatusNotFound, w)
		return
	}

	if !pluginHasCollection(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName) {
		utils.GetError(errors.New("collection does not exist"), http.StatusNotFound, w)
		return
	}

	reqData.handleDelete(w, r)
}

func (ddr *deleteDataRequest) prefixCollectionName() string {
	return getPrefixedCollectionName(ddr.PluginID, ddr.OrganizationID, ddr.CollectionName)
}

func (ddr *deleteDataRequest) handleDelete(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")

	filter := make(map[string]interface{})

	if ddr.BulkDelete {
		filter = ddr.Filter
	} else {
		filter["_id"] = mustObjectIDFromHex(ddr.ObjectID)
	}

	deletedCount, err := deleteMany(ddr.prefixCollectionName(), filter)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occurred: %v", err), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("success", utils.M{"deleted_count": deletedCount}, w)
}

func deleteMany(collName string, filter map[string]interface{}) (int64, error) {
	update := make(map[string]interface{})
	update["deleted"] = true
	update["deleted_at"] = time.Now()
	filter["deleted"] = bson.M{"$ne": true}

	res, err := updateMany(collName, filter, update)

	if err != nil {
		return 0, err
	}

	return res.ModifiedCount, nil
}
