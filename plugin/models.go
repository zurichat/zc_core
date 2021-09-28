package plugin

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	InstallCount   int64              `json:"install_count,omitempty" bson:"install_count"`
	Approved       bool               `json:"approved" bson:"approved"`
	Deleted        bool               `json:"deleted" bson:"deleted"`
	Images         []string           `json:"images,omitempty" bson:"images,omitempty"`
	Version        string             `json:"version" bson:"version"`
	Category       string             `json:"category" bson:"category"`
	Tags           []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	ApprovedAt     string             `json:"approved_at" bson:"approved_at"`
	CreatedAt      string             `json:"created_at" bson:"created_at"`
	UpdatedAt      string             `json:"updated_at" bson:"updated_at"`
	DeletedAt      string             `json:"deleted_at" bson:"deleted_at"`
}

type PluginPatch struct {
	Name        *string  `json:"name,omitempty" bson:"name,omitempty"`
	Description *string  `json:"description,omitempty"  bson:"description,omitempty"`
	Images      []string `json:"images,omitempty" bson:"images,omitempty"`
	Tags        []string `json:"tags,omitempty"  bson:"tags,omitempty"`
	Version     *string  `json:"version,omitempty"  bson:"version,omitempty"`
	SidebarURL  *string  `json:"sidebar_url,omitempty"  bson:"sidebar_url,omitempty"`
	InstallURL  *string  `json:"install_url,omitempty"  bson:"install_url,omitempty"`
	TemplateURL *string  `json:"template_url,omitempty"  bson:"template_url,omitempty"`
}

func CreatePlugin(ctx context.Context, p *Plugin) error {
	p.Approved = false
	p.CreatedAt = time.Now().String()
	p.UpdatedAt = time.Now().String()
	collection := utils.GetCollection(PluginCollectionName)
	res, err := collection.InsertOne(ctx, p)
	p.ID = res.InsertedID.(primitive.ObjectID)
	return err
}

func FindPluginByID(ctx context.Context, id string) (*Plugin, error) {
	p := new(Plugin)
	objID, _ := primitive.ObjectIDFromHex(id)
	collection := utils.GetCollection(PluginCollectionName)
	res := collection.FindOne(ctx, bson.M{"_id": objID, "deleted": M{"$ne": true}})
	if err := res.Decode(p); err != nil {
		return nil, err
	}
	return p, nil
}

func FindPlugins(ctx context.Context, filter bson.M) ([]*Plugin, error) {
	ps := []*Plugin{}
	collection := utils.GetCollection(PluginCollectionName)
	cursor, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &ps); err != nil {
		return nil, err
	}
	return ps, nil
}

func updatePlugin(ctx context.Context, id string, pp PluginPatch) error {
	collection := utils.GetCollection(PluginCollectionName)
	objId, _ := primitive.ObjectIDFromHex(id)
	update := M{}
	if pp.Name != nil {
		update["name"] = *(pp.Name)
	}
	if pp.Description != nil {
		update["description"] = *(pp.Description)
	}
	if pp.SidebarURL != nil {
		update["sidebar_url"] = *(pp.SidebarURL)
	}
	if pp.InstallURL != nil {
		update["install_url"] = *(pp.SidebarURL)
	}
	if pp.TemplateURL != nil {
		update["template_url"] = *(pp.Description)
	}

	if pp.Version != nil {
		update["version"] = *(pp.Version)
	}

	if pp.Images != nil {

		update["$push"] = bson.M{"images": bson.M{"$each": pp.Images}}
	}

	if pp.Tags != nil {
		update["$push"] = bson.M{"tags": bson.M{"$each": pp.Tags}}
	}
	_, err := collection.UpdateOne(ctx, M{"_id": objId}, update)
	return err
}
