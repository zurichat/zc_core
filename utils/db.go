package utils

import (
	"context"
	"fmt"
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

type errChecker struct {
	err error
}

func (e *errChecker) Check(err error) {
   if e.err != nil {
   	return
   }
   
   e.err = err
}

func ConnectToDB(clusterURL string) error {
    var ec errChecker

	once.Do(func() {
		ec.Check(defaultMongoHandle.Connect(clusterURL))
		ec.Check(CreateUniqueIndex("users", "email", 1))
		ec.Check(CreateUniqueIndex("plugins", "template_url", 1))
	})

	return ec.err
}

func (mh *MongoDBHandle) Connect(clusterURL string) error {
	clientOptions := options.Client().ApplyURI(clusterURL)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return err
	}

	timeOutFactor := 3
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOutFactor)*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return err
	}

	mh.client = client

	return nil
}

func (mh *MongoDBHandle) GetCollection(collectionName string) *mongo.Collection {
	DBName := Env("DB_NAME")
	return mh.client.Database(DBName).Collection(collectionName)
}

// GetCollection return collection for the db in DB_NAME env variable.
func GetCollection(collectionName string) *mongo.Collection {
	return defaultMongoHandle.GetCollection(collectionName)
}

func (mh *MongoDBHandle) Client() *mongo.Client {
	return mh.client
}

// "mongodb+srv://zuri:<password>@cluster0.hepte.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"

// GetMongoDbCollection get collection inside your db, this function can be exorted.
func GetMongoDBCollection(dbname, collectionName string) (*mongo.Collection, error) {
	client := defaultMongoHandle.Client()

	collection := client.Database(dbname).Collection(collectionName)

	return collection, nil
}

// get MongoDb documents for a collection.
func GetMongoDBDocs(collectionName string, filter map[string]interface{}, opts ...*options.FindOptions) ([]bson.M, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	var data []bson.M

	filterCursor, err := collection.Find(ctx, MapToBson(filter), opts...)
	if err != nil {
		return nil, err
	}

	if err := filterCursor.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// get single MongoDb document for a collection.
func GetMongoDBDoc(collectionName string, filter map[string]interface{}, opts ...*options.FindOneOptions) (bson.M, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	var data bson.M
	if err := collection.FindOne(ctx, MapToBson(filter), opts...).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func CreateMongoDBDoc(collectionName string, data map[string]interface{}) (*mongo.InsertOneResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	res, err := collection.InsertOne(ctx, MapToBson(data))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func CreateManyMongoDBDocs(collectionName string, data []interface{}) (*mongo.InsertManyResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	res, err := collection.InsertMany(ctx, data)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Update single MongoDb document for a collection.
func UpdateOneMongoDBDoc(collectionName, id string, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	// updateOne sets the fields, without using $set the entire document will be overwritten
	updateData := bson.M{"$set": MapToBson(data)}
	res, err := collection.UpdateOne(ctx, filter, updateData)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// This methods allows update of any kind e.g array increment, object embedding etc by passing the raw update data.
func GenericUpdateOneMongoDBDoc(collectionName string, id interface{}, updateData map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	filter := bson.M{"_id": id}

	res, err := collection.UpdateOne(ctx, filter, updateData)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Update many MongoDb documents for a collection.
func UpdateManyMongoDBDocs(collectionName string, filter, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	updateData := bson.M{"$set": MapToBson(data)}

	res, err := collection.UpdateMany(ctx, MapToBson(filter), updateData)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Replace a document with new data but preserve its id.
func ReplaceMongoDBDoc(collectionName string, filter, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	res, err := collection.ReplaceOne(ctx, MapToBson(filter), MapToBson(data))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete single MongoDb document for a collection.
func DeleteOneMongoDBDoc(collectionName, id string) (*mongo.DeleteResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	res, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete many MongoDb documents for a collection.
func DeleteManyMongoDBDoc(collectionName string, filter map[string]interface{}) (*mongo.DeleteResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	res, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func CreateUniqueIndex(collName, field string, order int) error {
	collection := defaultMongoHandle.GetCollection(collName)

	indexModel := mongo.IndexModel{
		Keys:    bson.M{field: order},
		Options: options.Index().SetUnique(true),
	}

	indexName, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		fmt.Println("errror creating unique index on email field in users collection")
		return err
	}

	fmt.Printf("%s index on %s collection created successfully\n", indexName, collName)

	return nil
}
