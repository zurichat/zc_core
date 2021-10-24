package plugin

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	Create(ctx context.Context, p *Plugin) error
	FindOne(ctx context.Context, f interface{}) (*Plugin, error)
	FindMany(ctx context.Context, f interface{}) ([]*Plugin, error)
	Update(ctx context.Context, f interface{}, pp Patch) error
	Delete(ctx context.Context, f interface{}) error
}


type mongoService struct {
	c *mongo.Client
	dbName string
}

func (m *mongoService) Create(ctx context.Context, p *Plugin) error {
	db := m.database()
	res, err := db.Collection("plugins").InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (m *mongoService) FindOne(ctx context.Context, f interface{}) (*Plugin, error) {
	p := &Plugin{}
	db := m.database()
	res := db.Collection("plugins").FindOne(ctx, f)
	return p, res.Decode(p)
}

func (m *mongoService) FindMany(ctx context.Context, f interface{}) (ps []*Plugin, _ error) {
	db := m.database()
	cursor, err := db.Collection("plugins").Find(ctx, f)

	if err != nil {
		return nil, err
	}

	return ps, cursor.All(ctx, &ps)
}

func (m *mongoService) Update(ctx context.Context, f interface{}, pp Patch) error {
	set := bson.M{}
	push := bson.M{}

	if pp.Name != nil {
		set["name"] = *(pp.Name)
	}

	if pp.Description != nil {
		set["description"] = *(pp.Description)
	}

	if pp.SidebarURL != nil {
		set["sidebar_url"] = *(pp.SidebarURL)
	}

	if pp.InstallURL != nil {
		set["install_url"] = *(pp.InstallURL)
	}

	if pp.TemplateURL != nil {
		set["template_url"] = *(pp.TemplateURL)
	}

	if pp.Version != nil {
		set["version"] = *(pp.Version)
	}

	if pp.SyncRequestURL != nil {
		set["sync_request_url"] = *(pp.SyncRequestURL)
	}

	if pp.Images != nil {
		push["images"] = bson.M{"$each": pp.Images}
	}

	if pp.Tags != nil {
		push["tags"] = bson.M{"$each": pp.Tags}
	}

	db := m.database()
	res, err := db.Collection("plugins").UpdateOne(ctx, f, bson.M{
		"$set":  set,
		"$push": push,
	})

	if res.MatchedCount < 1 {
		return Errorf(ENOENT, "no plugin matches the query")
	}

	return err
}

func (m *mongoService) Delete(ctx context.Context, f interface{}) error {
	db := m.database()
	_, err := db.Collection("plugins").DeleteOne(ctx, f)

	return err
}

func (m *mongoService) database() *mongo.Database {
	return m.c.Database(m.dbName)
}

func NewMongoService(c *mongo.Client) Service {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "zurichat"
	}
	return &mongoService{c, dbName}
}
