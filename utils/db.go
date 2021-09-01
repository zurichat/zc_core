package utils

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBHandle struct {
	client *mongo.Client
}

var defaultMongoHandle = &MongoDBHandle{}

var once sync.Once

func ConnectToDB(clusterURL string) error {
	var err error
	once.Do(func() {
		err = defaultMongoHandle.Connect(clusterURL)
	})
	return err
}

func (mh *MongoDBHandle) Connect(clusterURL string) error {
	clientOptions := options.Client().ApplyURI(clusterURL)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return err
	}
	mh.client = client
	return nil
}

func (mh *MongoDBHandle) GetCollection(collectionName string) *mongo.Collection {
	DbName := Env("DB_NAME")
	return mh.client.Database(DbName).Collection(collectionName)
}

// GetCollection return collection for the db in DB_NAME env variable.
func GetCollection(collectionName string) *mongo.Collection {
	return defaultMongoHandle.GetCollection(collectionName)
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
		return nil, err
	}

	return res, nil
}

func CreateManyMongoDbDocs(collectionName string, data []interface{}) (*mongo.InsertManyResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	res, err := collection.InsertMany(ctx, data)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Update single MongoDb document for a collection
func UpdateOneMongoDbDoc(collectionName string, ID string, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	id, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.M{"_id": id}

	//updateOne sets the fields, without using $set the entire document will be overwritten
	update_data := bson.M{"$set": MapToBson(data)}
	res, err := collection.UpdateOne(ctx, filter, update_data)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Update many MongoDb documents for a collection
func UpdateManyMongoDbDocs(collectionName string, filter map[string]interface{}, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	update_data := bson.M{"$set": MapToBson(data)}

	res, err := collection.UpdateMany(ctx, MapToBson(filter), update_data)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Replace a document with new data but preserve its id
func ReplaceMongoDbDoc(collectionName string, filter map[string]interface{}, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	res, err := collection.ReplaceOne(ctx, MapToBson(filter), MapToBson(data))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete single MongoDb document for a collection
func DeleteOneMongoDoc(collectionName string, ID string) (*mongo.DeleteResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	id, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.M{"_id": id}
	res, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete many MongoDb documents for a collection
func DeleteManyMongoDoc(collectionName string, filter map[string]interface{}) (*mongo.DeleteResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	res, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return nil, err
	}

	return res, nil
}
