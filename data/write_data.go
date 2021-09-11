package data

import (
	"errors"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
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

	w.Header().Set("content-type", "application/json")

	switch r.Method {
	case "POST":
		reqData.handlePost(w, r)
	case "PUT":
		reqData.handlePut(w, r)
	default:
		fmt.Fprint(w, `{"data_write": "Data write request"}`)
	}
}

func (wdr *writeDataRequest) handlePost(w http.ResponseWriter, r *http.Request) {
	var writeCount int64
	data := M{}
	if wdr.BulkWrite {
		res, err := insertMany(wdr.prefixCollectionName(), wdr.Payload)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
		writeCount = int64(len(res.InsertedIDs))
		data["object_ids"] = res.InsertedIDs
	} else {
		res, err := insertOne(wdr.prefixCollectionName(), wdr.Payload)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
		writeCount = 1
		data["object_id"] = res.InsertedID
	}
	data["insert_count"] = writeCount
	w.WriteHeader(http.StatusCreated)
	utils.GetSuccess("success", data, w)
}

func (wdr *writeDataRequest) handlePut(w http.ResponseWriter, r *http.Request) {
	var err error
	res := &mongo.UpdateResult{}
	if wdr.BulkWrite {
		res, err = updateMany(wdr.prefixCollectionName(), wdr.Filter, wdr.Payload)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	} else {
		res, err = updateOne(wdr.prefixCollectionName(), wdr.ObjectID, wdr.Payload)
		if err != nil {
			utils.GetError(fmt.Errorf("an error occured: %v", err), http.StatusInternalServerError, w)
			return
		}
	}
	data := M{
		"matched_documents":  res.MatchedCount,
		"modified_documents": res.ModifiedCount,
	}
	utils.GetSuccess("success", data, w)
}

func (wdr *writeDataRequest) prefixCollectionName() string {
	return getPrefixedCollectionName(wdr.PluginID, wdr.OrganizationID, wdr.CollectionName)
}

func insertMany(collName string, data interface{}) (*mongo.InsertManyResult, error) {
	docs, ok := data.([]interface{})
	if !ok {
		return nil, errors.New("invalid object type, payload must be an array of objects")
	}
	return utils.CreateManyMongoDbDocs(collName, docs)
}

func insertOne(collName string, data interface{}) (*mongo.InsertOneResult, error) {
	doc, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid object type, payload must be a valid JSON object")
	}
	return utils.CreateMongoDbDoc(collName, doc)
}

func updateOne(collName, id string, upd interface{}) (*mongo.UpdateResult, error) {
	update, ok := upd.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid object type")
	}
	return utils.UpdateOneMongoDbDoc(collName, id, update)
}

func updateMany(collName string, filter map[string]interface{}, upd interface{}) (*mongo.UpdateResult, error) {
	update, ok := upd.(map[string]interface{})
	if !ok {
		return nil, errors.New("type assertion error")
	}
	return utils.UpdateManyMongoDbDocs(collName, filter, update)
}

func recordExists(collName, id string) bool {
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err := utils.GetMongoDbDoc(collName, M{"_id": objId})
	if err == nil {
		return true
	}
	return false
}

func MustObjectIDFromHex(hex string) primitive.ObjectID {
	objID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return objID
}
