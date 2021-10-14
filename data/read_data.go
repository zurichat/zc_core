package data

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zuri.chat/zccore/utils"
)

// ReadData handles the data retrieval operation for plugins using GET requests.
func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID, collName, orgID := vars["plugin_id"], vars["coll_name"], vars["org_id"]

	actualCollName := mongoCollectionName(pluginID, collName)

	filter := parseURLQuery(r)
	filter["deleted"] = bson.M{"$ne": true}
	filter["organization_id"] = orgID
	docs, err := utils.GetMongoDBDocs(actualCollName, filter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if _, exists := filter["_id"]; exists && len(docs) == 1 {
		delete(docs[0], "organization_id")
		utils.GetSuccess("success", docs[0], w)
		return
	}

	for _, doc := range docs {
		delete(doc, "organization_id")
	}

	utils.GetSuccess("the use of this endpoint is being deprecated, switch to the POST method.", docs, w)
}

func getPrefixedCollectionName(pluginID, orgID, collName string) string {
	return fmt.Sprintf("%s:%s:%s", pluginID, orgID, collName)
}

func parseURLQuery(r *http.Request) map[string]interface{} {
	m := utils.M{}

	for k, v := range r.URL.Query() {
		if k == "id" || k == "_id" {
			m["_id"], _ = primitive.ObjectIDFromHex(v[0])
			continue
		}

		m[k] = v[0]
	}

	return m
}

type readDataRequest struct {
	PluginID       string                 `json:"plugin_id"`
	CollectionName string                 `json:"collection_name"`
	OrganizationID string                 `json:"organization_id"`
	ObjectID       string                 `json:"object_id,omitempty"`
	ObjectIDs      []string               `json:"object_ids,omitempty"`
	Filter         map[string]interface{} `json:"filter,omitempty"`
	RawQuery       map[string]interface{}            `json:"raw_query,omitempty"`
	ReadOptions    *readOptions           `json:"options,omitempty"`
}

type readOptions struct {
	Limit      *int64                 `json:"limit,omitempty"`
	Skip       *int64                 `json:"skip,omitempty"`
	Sort       map[string]interface{} `json:"sort,omitempty"`
	Projection map[string]interface{} `json:"projection,omitempty"`
}

func (r *readDataRequest) containsID() bool {
	return r.ObjectID != "" || idInFilter(bson.M(r.Filter))
}

func (r *readDataRequest) getIDString() string {
	if r.ObjectID != "" {
		return r.ObjectID
	}

	if id, exists := r.Filter["_id"]; exists {
		return id.(string)
	}

	if id, exists := r.Filter["id"]; exists {
		return id.(string)
	}
	return ""
}

// NewRead handles data retrieval process using POST requests, providing flexibility for the query.
func NewRead(w http.ResponseWriter, r *http.Request) {
	reqData := new(readDataRequest)

	if err := utils.ParseJSONFromRequest(r, reqData); err != nil {
		utils.GetError(fmt.Errorf("error processing request: %v", err), http.StatusUnprocessableEntity, w)
		return
	}

	filter := bson.M(reqData.Filter)

	if filter == nil {
		filter = bson.M{}
	}

	filter["deleted"] = bson.M{"$ne": true}
	filter["organization_id"] = reqData.OrganizationID

	actualCollName := mongoCollectionName(reqData.PluginID, reqData.CollectionName)

	if reqData.containsID() {
		id := reqData.getIDString()
		var opts *options.FindOneOptions
		filter["_id"] = mustObjectIDFromHex(id)

		if r := reqData.ReadOptions; r != nil {
			opts = setFindOneOptions(*r)
		}

		doc, err := findOne(actualCollName, filter, opts)

		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		delete(doc, "organization_id")
		utils.GetSuccess("success", doc, w)
		return
	}

	var opts *options.FindOptions

	if r := reqData.ReadOptions; r != nil {
		opts = setOptions(*r)
	}

	if reqData.ObjectIDs != nil {
		filter["_id"] = bson.M{"$in": hexToObjectIDs(reqData.ObjectIDs)}
	}

	if reqData.RawQuery != nil {
		filter = reqData.RawQuery
		filter["organization_id"] = reqData.OrganizationID
	} 
	
	docs, err := findMany(actualCollName, filter, opts)
	

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	for _, doc := range docs {
		delete(doc, "organization_id")
	}

	utils.GetSuccess("success", docs, w)
}

func findOne(collName string, filter bson.M, opts ...*options.FindOneOptions) (bson.M, error) {
	return utils.GetMongoDBDoc(collName, filter, opts...)
}

func findMany(collName string, filter bson.M, opts ...*options.FindOptions) ([]bson.M, error) {
	return utils.GetMongoDBDocs(collName, filter, opts...)
}

func idInFilter(filter bson.M) bool {
	_, exists := filter["_id"]
	_, exists2 := filter["id"]

	return exists || exists2
}

func hexToObjectIDs(ids []string) []primitive.ObjectID {
	objIDs := make([]primitive.ObjectID, len(ids))

	for i, id := range ids {
		objIDs[i] = mustObjectIDFromHex(id)
	}

	return objIDs
}

func setOptions(r readOptions) *options.FindOptions {
	findOptions := options.Find()

	if r.Limit != nil {
		findOptions.SetLimit(*r.Limit)
	}

	if r.Skip != nil {
		findOptions.SetSkip(*r.Skip)
	}

	if len(r.Sort) > 0 {
		findOptions.SetSort(r.Sort)
	}

	if r.Projection != nil {
		findOptions.SetProjection(r.Projection)
	}

	return findOptions
}

func setFindOneOptions(r readOptions) *options.FindOneOptions {
	findOneOpts := options.FindOne()

	if r.Skip != nil {
		findOneOpts.SetSkip(*r.Skip)
	}

	if len(r.Sort) > 0 {
		findOneOpts.SetSort(r.Sort)
	}

	if r.Projection != nil {
		findOneOpts.SetProjection(r.Projection)
	}

	return findOneOpts
}
