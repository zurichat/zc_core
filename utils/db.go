package utils

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBHandle struct {
	client    *mongo.Client
	sync.Once // to create connection exactly once
}

var defaultMongoHandle = &MongoDBHandle{}

func ConnectToDB(clusterURL string) error {
	return defaultMongoHandle.Connect(clusterURL)
}

func (mh *MongoDBHandle) Connect(clusterURL string) error {
	var Err error

	// do this only once
	mh.Once.Do(func() {
		clientOptions := options.Client().ApplyURI(clusterURL)

		client, err := mongo.NewClient(clientOptions)
		if err != nil {
			Err = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			return
		}

		err = client.Ping(context.Background(), readpref.Primary())
		if err != nil {
			return
		}
		mh.client = client
	})
	return Err
}

func (mh *MongoDBHandle) GetCollection(collectionName string) *mongo.Collection {
	DbName := Env("DB_NAME")
	return mh.client.Database(DbName).Collection(collectionName)
}

func (mh *MongoDBHandle) Client() *mongo.Client {
	return mh.client
}

//"mongodb+srv://zuri:<password>@cluster0.hepte.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"

//GetMongoDbCollection get collection inside your db, this function can be exorted
func GetMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {
	client := defaultMongoHandle.Client()

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}

// get MongoDb documents for a collection
func GetMongoDbDocs(collectionName string, filter map[string]interface{}) ([]bson.M, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

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
func GetMongoDbDoc(collectionName string, filter map[string]interface{}) (bson.M, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	var data bson.M
	if err := collection.FindOne(ctx, MapToBson(filter)).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func CreateMongoDbDoc(collectionName string, data map[string]interface{}) (*mongo.InsertOneResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	res, err := collection.InsertOne(ctx, MapToBson(data))

	if err != nil {
		log.Fatal(err)
	}

	return res, nil
}
