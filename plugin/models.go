package plugin

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zuri.chat/zccore/utils"
)

const (
	PluginCollectionName = "plugins"
)

type Plugin struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name" validate:"required"`
	Description    string             `json:"description" bson:"description" validate:"required"`
	DeveloperName  string             `json:"developer_name" bson:"developer_name" validate:"required"`
	DeveloperEmail string             `json:"developer_email" bson:"developer_email" validate:"required"`
	TemplateURL    string             `json:"template_url" bson:"template_url" validate:"required"`
	SidebarURL     string             `json:"sidebar_url" bson:"sidebar_url" validate:"required"`
	InstallURL     string             `json:"install_url" bson:"install_url" validate:"required"`
	IconURL        string             `json:"icon_url" bson:"icon_url"`
	InstallCount   int64              `json:"install_count" bson:"install_count"`
	Approved       bool               `json:"approved" bson:"approved"`
	Images         []string           `json:"images,omitempty" bson:"images,omitempty"`
	Version        string             `json:"version" bson:"version"`
	Category       string             `json:"category" bson:"category"`
	Tags           []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	ApprovedAt     string             `json:"approved_at" bson:"approved_at"`
	CreatedAt      string             `json:"created_at" bson:"created_at"`
	UpdatedAt      string             `json:"updated_at" bson:"updated_at"`
	DeletedAt      string             `json:"deleted_at" bson:"deleted_at"`
	SyncRequestUrl string             `json:"sync_request_url" bson:"sync_request_url"`
	Queue          []MessageModel     `json:"queue" bson:"queue"`
	QueuePID       int                `json:"queuepid" bson:"queuepid"`
}

type Patch struct {
	Name           *string  `json:"name,omitempty" bson:"name,omitempty"`
	Description    *string  `json:"description,omitempty"  bson:"description,omitempty"`
	Images         []string `json:"images,omitempty" bson:"images,omitempty"`
	Tags           []string `json:"tags,omitempty"  bson:"tags,omitempty"`
	Version        *string  `json:"version,omitempty"  bson:"version,omitempty"`
	SidebarURL     *string  `json:"sidebar_url,omitempty"  bson:"sidebar_url,omitempty"`
	InstallURL     *string  `json:"install_url,omitempty"  bson:"install_url,omitempty"`
	TemplateURL    *string  `json:"template_url,omitempty"  bson:"template_url,omitempty"`
	SyncRequestUrl *string  `json:"sync_request_url" bson:"sync_request_url"`
}

func CreatePlugin(ctx context.Context, p *Plugin) error {
	p.Approved = false
	p.CreatedAt = time.Now().String()
	p.UpdatedAt = time.Now().String()
	collection := utils.GetCollection(PluginCollectionName)
	res, err := collection.InsertOne(ctx, p)

	if err != nil {
		return err
	}

	value, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return err
	}

	p.ID = value

	return err
}

func FindPluginByID(ctx context.Context, id string) (*Plugin, error) {
	p := &Plugin{}

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	collection := utils.GetCollection(PluginCollectionName)
	res := collection.FindOne(ctx, bson.M{"_id": objID })

	if err := res.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

func FindPlugins(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*Plugin, error) {
	ps := []*Plugin{}

    collection := utils.GetCollection(PluginCollectionName)
	cursor, err := collection.Find(ctx, filter, opts...)

	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &ps); err != nil {
		return nil, err
	}

	return ps, nil
}

func SortPlugins(ctx context.Context, filter bson.M, sort bson.D) ([]*Plugin, error) {
	ps := []*Plugin{}
	collection := utils.GetCollection(PluginCollectionName)

	findOptions := options.Find()
	findOptions.SetSort(sort)

	cursor, err := collection.Find(ctx, filter, findOptions)

	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &ps); err != nil {
		return nil, err
	}

	return ps, nil
}

func updatePlugin(ctx context.Context, id string, pp *Patch) error {
	collection := utils.GetCollection(PluginCollectionName)
	objID, _ := primitive.ObjectIDFromHex(id)
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
		set["template_url"] = *(pp.Description)
	}

	if pp.Version != nil {
		set["version"] = *(pp.Version)
	}
	if pp.SyncRequestUrl != nil {
		set["sync_request_url"] = *(pp.SyncRequestUrl)
	}

	if pp.Images != nil {
		push["images"] = bson.M{"$each": pp.Images}
	}

	if pp.Tags != nil {
		push["tags"] = bson.M{"$each": pp.Tags}
	}

	_, err := collection.UpdateOne(ctx, M{"_id": objID}, bson.M{
		"$set":  set,
		"$push": push,
	})

	return err
}

type SyncUpdateRequest struct {
	ID int `json:"id" bson:"id" validate:"required"`
}

type MessageModel struct {
	Id      int         `json:"id" bson:"id"`
	Event   string      `json:"event" bson:"event"`
	Message interface{} `json:"message" bson:"message"`
}
