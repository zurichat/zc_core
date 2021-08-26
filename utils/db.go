package utils

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//"mongodb+srv://zuri:<password>@cluster0.hepte.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"

// If you want to export your function. You must to start upper case function name. Otherwise you won't see your function when you import that on other class.
//getMongoDbConnection get connection of mongodb
func getMongoDbConnection() (*mongo.Client, context.Context, error) {
	cluster_url := Env("CLUSTER_URL")

	clientOptions := options.Client().ApplyURI(cluster_url)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, ctx, nil
}

//GetMongoDbCollection get collection inside your db, this function can be exorted
func GetMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {
	client, _, err := getMongoDbConnection()

	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}

// get MongoDb documents for a collection
func GetMongoDbDocs(CollectionName string, filter map[string]interface{}) ([]bson.M, error) {
	DbName := Env("DB_NAME")
	client, ctx, err := getMongoDbConnection()
	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	var data []bson.M
	filterCursor, err := collection.Find(ctx, MapToBson(filter))
	if err != nil {
		return nil, err
	}
	if err = filterCursor.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// get single MongoDb document for a collection
func GetMongoDbDoc(CollectionName string, filter map[string]interface{}) (bson.M, error) {
	DbName := Env("DB_NAME")
	client, ctx, err := getMongoDbConnection()
	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	var data bson.M
	if err = collection.FindOne(ctx, MapToBson(filter)).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func CreateMongoDbDoc(CollectionName string, data map[string]interface{}) (*mongo.InsertOneResult, error) {
	DbName := Env("DB_NAME")
	client, ctx, err := getMongoDbConnection()
	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)
	res, err := collection.InsertOne(ctx, MapToBson(data))

	if err != nil {
		log.Fatal(err)
	}

	return res, nil
}

// update single MongoDb document for a collection
func UpdateOneMongoDbDoc(CollectionName string, ID string, data map[string]interface{}) (*mongo.UpdateResult, error) {
	DbName := Env("DB_NAME")
	client, ctx, err := getMongoDbConnection()
	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	id, _ := primitive.ObjectIDFromHex(ID)
	filter:= bson.M{"_id": id}
	res, err := collection.UpdateOne(ctx, filter, MapToBson(data))

	if err != nil {
		log.Fatal(err)
	}

	return res, nil
}