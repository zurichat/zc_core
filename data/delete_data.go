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

func DeleteData(w http.ResponseWriter, r *http.Request) {
	reqData := new(deleteDataRequest)

	if err := utils.ParseJsonFromRequest(r, reqData); err != nil {
		utils.GetError(fmt.Errorf("error processing request: %v", err), http.StatusUnprocessableEntity, w)
		return
	}

	if _, err := plugin.FindPluginByID(r.Context(), reqData.PluginID); err != nil {
		msg := "plugin with this id does not exist"
		utils.GetError(errors.New(msg), http.StatusNotFound, w)
		return
	}

	//if !recordExists(_OrganizationCollectionName, reqData.OrganizationID) {
	//msg := "organization with this id does not exist"
	//utils.GetError(errors.New(msg), http.StatusNotFound, w)
	//return
	//}

	// if plugin is writing to this collection the first time, we create a record linking this collection to the plugin.
	if !pluginHasCollection(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName) {
		createPluginCollectionRecord(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName)
	}

	reqData.handleDelete(w, r)
}

func (ddr *deleteDataRequest) prefixCollectionName() string {
	return getPrefixedCollectionName(ddr.PluginID, ddr.OrganizationID, ddr.CollectionName)
}

func (ddr *deleteDataRequest) handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var err error
	var deletedCount int64
	if ddr.BulkDelete {
		deletedCount, err = deleteMany(ddr.prefixCollectionName(), ddr.Filter)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	} else {
		if deletedCount, err = deleteOne(ddr.prefixCollectionName(), ddr.ObjectID); err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	}

	utils.GetSuccess("success", M{"deleted_count": deletedCount}, w)
}

func deleteOne(collName, id string) (int64, error) {
	update := make(map[string]interface{})
	update["deleted"] = true
	update["deleted_at"] = time.Now()
	res, err := updateMany(collName, bson.M{"_id": MustObjectIDFromHex(id), "deleted": bson.M{"$ne": true}}, update)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
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
