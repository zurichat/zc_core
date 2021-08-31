package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

func Plugins(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusOK)
	var pluguns []Plugins

	response.Header().Set("Content-Type", "application/json")

	DbName, CollectionName := utils.Env("DB_NAME"), "plugins"
	collection, err := utils.GetMongoDbCollection(DbName, CollectionName)
	if err != nil {
		utils.GetError(err, 500, response)
	}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		utils.GetError(err, 400, response)
		return
	}

	defer cursor.Close(ctx)

	plugins = make([]Plugins, 0)
	for cursor.Next(ctx) {
		var plugin Plugins
		cursor.Decode(&plugin)
		plugins = append(plugins, *plugin)
	}

	json.NewEncoder(response).Encode(plugins)
}
