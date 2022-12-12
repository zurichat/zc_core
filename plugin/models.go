package plugin

import (
	"context"

	"github.com/mitchellh/mapstructure"
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
	SyncRequestURL string             `json:"sync_request_url" bson:"sync_request_url"`
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
	SyncRequestURL *string  `json:"sync_request_url" bson:"sync_request_url"`
}

func FindPluginByID(ctx context.Context, id string) (*Plugin, error) {
	var (
		p  *Plugin
		bp *Plugin
	)

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	res, err := utils.GetMongoDBDoc(PluginCollectionName, bson.M{"_id": objID})

	if err != nil {
		return nil, err
	}

	bsonBytes, err := bson.Marshal(res)

	if err != nil {
		return nil, err
	}

	err = bson.Unmarshal(bsonBytes, &p)
	if err != nil {
		return nil, err
	}

	if err := mapstructure.Decode(res, &bp); err != nil {
		return nil, err
	}

	p.Queue = bp.Queue

	return p, nil
}

func FindPlugins(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*Plugin, error) {
	ps := []*Plugin{}

	cursor, err := utils.GetMongoDBDocs(PluginCollectionName, filter, opts...)

	if err != nil {
		return nil, err
	}

	for _, plng := range cursor {
		var (
			nps *Plugin
			bp  *Plugin
		)

		bsonBytes, err := bson.Marshal(plng)
		if err != nil {
			return nil, err
		}

		err = bson.Unmarshal(bsonBytes, &nps)
		if err != nil {
			return nil, err
		}

		if err := mapstructure.Decode(plng, &bp); err != nil {
			return nil, err
		}

		nps.Queue = bp.Queue
		ps = append(ps, nps)
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

func FindPluginByTemplateURL(ctx context.Context, url string) (*Plugin, error) {
	var (
		p  *Plugin
		bp *Plugin
	)

	res, err := utils.GetMongoDBDoc(PluginCollectionName, bson.M{"deleted": false, "template_url": url})

	if err != nil {
		return nil, err
	}

	bsonBytes, err := bson.Marshal(res)

	if err != nil {
		return nil, err
	}

	err = bson.Unmarshal(bsonBytes, &p)
	if err != nil {
		return nil, err
	}

	if err := mapstructure.Decode(res, &bp); err != nil {
		return nil, err
	}

	p.Queue = bp.Queue

	return p, nil
}

type SyncUpdateRequest struct {
	ID int `json:"id" bson:"id" validate:"required"`
}

type MessageModel struct {
	ID      int         `json:"id" bson:"id"`
	Event   string      `json:"event" bson:"event"`
	Message interface{} `json:"message" bson:"message"`
}
