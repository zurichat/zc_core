package plugin

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const (
	PluginCollectionName            = "plugins"
	PluginCollectionsCollectionName = "plugin_collections"
)

type Plugin struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	Name           string             `json:"name" bson:"name" validate:"required"`
	Description    string             `json:"description" bson:"description" validate:"required"`
	DeveloperName  string             `json:"developer_name" bson:"developer_name" validate:"required"`
	DeveloperEmail string             `json:"developer_email" bson:"developer_email" validate:"required"`
	TemplateURL    string             `json:"template_url" bson:"template_url" validate:"required"`
	SidebarURL     string             `json:"sidebar_url" bson:"sidebar_url" validate:"required"`
	InstallURL     string             `json:"install_url" bson:"install_url" validate:"required"`
	IconURL        string             `json:"icon_url" bson:"icon_url"`
	InstallCount   int64              `json:"install_count,omitempty" bson:"install_count"`
	Approved       bool               `json:"-" bson:"approved"`
	ApprovedAt     time.Time          `json:"approved_at" bson:"approved_at"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// PluginCollections is used internally to keep track collections a plugin created.
type PluginCollections struct {
	ID             primitive.ObjectID `bson:"_id"`
	PluginID       string             `bson:"plugin_id"`
	OrganizationID string             `bson:"organization_id"`
	CollectionName string             `bson:"collection_name"`
	CreatedAt      time.Time          `bson:"created_at"`
}

func CreatePlugin(p *Plugin) error {
	p.CreatedAt = time.Now()
	doc, _ := utils.StructToMap(p, "bson")
	res, err := utils.CreateMongoDbDoc(PluginCollectionName, doc)
	p.ID = res.InsertedID.(primitive.ObjectID)
	return err
}

func FindPluginByID(id string) (*Plugin, error) {
	p := &Plugin{}
	objID, _ := primitive.ObjectIDFromHex(id)
	doc, err := utils.GetMongoDbDoc(PluginCollectionName, M{"_id": objID})
	if err != nil {
		return nil, err
	}
	if err := MapToStruct(doc, p); err != nil {
		return nil, err
	}
	return p, nil
}

func FindPlugins(filter bson.M) ([]*Plugin, error) {
	docs, err := utils.GetMongoDbDocs(PluginCollectionName, filter)
	if err != nil {
		return nil, err
	}
	plugins := make([]*Plugin, len(docs))
	for i, doc := range docs {
		p := &Plugin{}
		if err := MapToStruct(doc, p); err != nil {
			return nil, err
		}
		plugins[i] = p
	}
	return plugins, nil
}

func MapToStruct(m map[string]interface{}, v interface{}) error {
	return mapstructure.Decode(m, v)
}
