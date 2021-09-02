package data

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

const (
	_PluginCollectionName            = "plugins"
	_PluginCollectionsCollectionName = "plugin_collections"
	_OrganizationCollectionName      = "organizations"
)

type writeDataRequest struct {
	PluginID       string                 `json:"plugin_id"`
	CollectionName string                 `json:"collection_name"`
	OrganizationID string                 `json:"organization_id"`
	BulkWrite      bool                   `json:"bulk_write"`
	ObjectID       string                 `json:"object_id,omitempty"`
	Filter         map[string]interface{} `json:"filter"`
	Payload        interface{}            `json:"payload,omitempty"`
}

func WriteData(w http.ResponseWriter, r *http.Request) {
	reqData := new(writeDataRequest)
	if err := utils.ParseJsonFromRequest(r, reqData); err != nil {
		utils.GetError(fmt.Errorf("error processing request: %v", err), http.StatusUnprocessableEntity, w)
		return
	}

	if _, err := plugin.FindPluginByID(reqData.PluginID); err != nil {
		msg := "plugin with this id does not exist"
		utils.GetError(errors.New(msg), http.StatusNotFound, w)
		return
	}

	if !recordExists(_OrganizationCollectionName, reqData.OrganizationID) {
		msg := "organization with this id does not exist"
		utils.GetError(errors.New(msg), http.StatusNotFound, w)
		return
	}

	// if plugin is writing to this collection the first time, we create a record linking this collection to the plugin.
	if !pluginHasCollection(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName) {
		createPluginCollectionRecord(reqData.PluginID, reqData.OrganizationID, reqData.CollectionName)
	}

	switch r.Method {
	case "POST":
		reqData.handlePost(w, r)
	case "PUT":
		reqData.handlePut(w, r)
	case "DELETE":
		reqData.handleDelete(w, r)
	default:
		fmt.Fprint(w, "Data write endpoint")
	}
}

func (wdr *writeDataRequest) handlePost(w http.ResponseWriter, r *http.Request) {
	var err error
	var writeCount int64
	if wdr.BulkWrite {
		writeCount, err = insertMany(wdr.prefixCollectionName(), wdr.Payload)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	} else {
		if err := insertOne(wdr.prefixCollectionName(), wdr.Payload); err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
		writeCount = 1
	}

	w.WriteHeader(http.StatusCreated)
	utils.GetSuccess("success", M{"insert_count": writeCount}, w)
}

func (wdr *writeDataRequest) handlePut(w http.ResponseWriter, r *http.Request) {
	var err error
	var writeCount int64
	if wdr.BulkWrite {
		writeCount, err = updateMany(wdr.prefixCollectionName(), wdr.Filter, wdr.Payload)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	} else {
		if err := updateOne(wdr.prefixCollectionName(), wdr.ObjectID, wdr.Payload); err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
		writeCount = 1
	}

	utils.GetSuccess("success", M{"modified_count": writeCount}, w)
}

func (wdr *writeDataRequest) handleDelete(w http.ResponseWriter, r *http.Request) {
	var err error
	var deletedCount int64
	if wdr.BulkWrite {
		deletedCount, err = deleteMany(wdr.prefixCollectionName(), wdr.Filter)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	} else {
		if err := deleteOne(wdr.prefixCollectionName(), wdr.ObjectID); err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
		deletedCount = 1
	}

	utils.GetSuccess("success", M{"deleted_count": deletedCount}, w)
}

func (wdr *writeDataRequest) prefixCollectionName() string {
	return getPrefixedCollectionName(wdr.PluginID, wdr.OrganizationID, wdr.CollectionName)
}

func insertMany(collName string, data interface{}) (int64, error) {
	docs, ok := data.([]interface{})
	if !ok {
		return 0, errors.New("type assertion error")
	}
	// call mongodb insert many here
	res, err := utils.CreateManyMongoDbDocs(collName, docs)
	if err != nil {
		return 0, err
	}
	l := len(res.InsertedIDs)
	return int64(l), nil
}

func insertOne(collName string, data interface{}) error {
	doc, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("type assertion error")
	}
	if _, err := utils.CreateMongoDbDoc(collName, doc); err != nil {
		return err
	}
	return nil
}

func updateOne(collName, id string, upd interface{}) error {
	update, ok := upd.(map[string]interface{})
	if !ok {
		return errors.New("type assertion error")
	}
	if _, err := utils.UpdateOneMongoDbDoc(collName, id, update); err != nil {
		return err
	}
	return nil
}

func updateMany(collName string, filter map[string]interface{}, upd interface{}) (int64, error) {
	update, ok := upd.(map[string]interface{})
	if !ok {
		return 0, errors.New("type assertion error")
	}
	// do update many
	res, err := utils.UpdateManyMongoDbDocs(collName, filter, update)
	if err != nil {
		return 0, err
	}

	return res.ModifiedCount, nil
}

func deleteOne(collName, id string) error {
	return nil
}

func deleteMany(collName string, filter map[string]interface{}) (int64, error) {
	return 0, nil
}

func recordExists(collName, id string) bool {
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err := utils.GetMongoDbDoc(collName, M{"_id": objId})
	if err == nil {
		return true
	}
	return false
}

func pluginHasCollection(pluginID, orgID, collectionName string) bool {
	filter := M{
		"plugin_id":       pluginID,
		"collection_name": collectionName,
		"organization_id": orgID,
	}
	_, err := utils.GetMongoDbDoc(_PluginCollectionsCollectionName, filter)
	if err == nil {
		return true
	}
	return false
}

func createPluginCollectionRecord(pluginID, orgID, collectionName string) error {
	doc := M{
		"plugin_id":       pluginID,
		"organization_id": orgID,
		"collection_name": collectionName,
		"created_at":      time.Now(),
	}

	if _, err := utils.CreateMongoDbDoc(_PluginCollectionsCollectionName, doc); err != nil {
		return err
	}
	return nil
}
