package organizations

import (
	"encoding/json"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const (
	OrganizationCollectionName     = "organizations"
	InstalledPluginsCollectionName = "installed_plugins"
	OrganizationSettings           = "organizations_settings"
)

type Organization struct {
	ID           string                   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name         string                   `json:"name" bson:"name"`
	CreatorEmail string                   `json:"creator_email" bson:"creator_email"`
	CreatorID    string                   `json:"creator_id" bson:"creator_id"`
	Plugins      []map[string]interface{} `json:"plugins" bson:"plugins"`
	Admins       []string                 `json:"admins" bson:"admins"`
	Settings     map[string]interface{}   `json:"settings" bson:"settings"`
	LogoURL      string                   `json:"logo_url" bson:"logo_url"`
	WorkspaceURL string                   `json:"workspace_url" bson:"workspace_url"`
	CreatedAt    time.Time                `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at" bson:"updated_at"`
}

func (o *Organization) OrgPlugins() []map[string]interface{} {
	orgCollectionName := GetOrgPluginCollectionName(o.ID)

	orgPlugins, _ := utils.GetMongoDbDocs(orgCollectionName, nil)

	var pluginsMap []map[string]interface{}
	pluginJson, _ := json.Marshal(orgPlugins)
	json.Unmarshal(pluginJson, &pluginsMap)

	return pluginsMap
}

type OrgPluginBody struct {
	PluginId string `json:"plugin_id"`
	UserId   string `json:"user_id"`
}

type InstalledPlugin struct {
	_id         string                 `json:"id" bson:"_id"`
	PluginID    string                 `json:"plugin_id" bson:"plugin_id"`
	Plugin      map[string]interface{} `json:"plugin" bson:"plugin"`
	AddedBy     string                 `json:"added_by" bson:"added_by"`
	ApprovedBy  string                 `json:"approved_by" bson:"approved_by"`
	InstalledAt time.Time              `json:"installed_at" bson:"installed_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
}

type OrganizationAdmin struct {
	ID             primitive.ObjectID `bson:"id"`
	OrganizationID string             `bson:"organization_id"`
	UserID         string             `bson:"user_id"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

func GetOrgPluginCollectionName(orgName string) string {
	return strings.ToLower(orgName) + "_" + InstalledPluginsCollectionName
}

// type Social struct {
// 	ID    primitive.ObjectID `json:"id" bson:"id"`
// 	url   string             `json:"url" bson:"url"`
// 	title string             `json:"title" bson:"title"`
// }

type Member struct {
	ID          primitive.ObjectID     `json:"_id" bson:"_id"`
	OrgId       string                 `json:"org_id" bson:"org_id"`
	Files       []string               `json:"files" bson:"files"`
	ImageURL    string                 `json:"image_url" bson:"image_url"`
	FirstName   string                 `json:"first_name" bson:"first_name"`
	LastName    string                 `json:"last_name" bson:"last_name"`
	Email       string                 `json:"email" bson:"email"`
	UserName    string                 `bson:"user_name" json:"user_name"`
	DisplayName string                 `json:"display_name" bson:"display_name"`
	Bio         string                 `json:"bio" bson:"bio"`
	Status      string                 `json:"status" bson:"status"`
	Presence    string                 `json:"presence" bson:"presence"`
	Pronouns    string                 `json:"pronouns" bson:"pronouns"`
	Phone       string                 `json:"phone" bson:"phone"`
	TimeZone    string                 `json:"time_zone" bson:"time_zone"`
	Role        string                 `json:"role" bson:"role"`
	JoinedAt    time.Time              `json:"joined_at" bson:"joined_at"`
	Settings    map[string]interface{} `json:"settings" bson:"settings"`
	Deleted     bool                   `json:"deleted" bson:"deleted"`
	DeletedAt   time.Time              `json:"deleted_at" bson:"deleted_at"`
	Socials     map[string]string      `json:"socials" bson:"socials"`
}

type Profile struct {
	ID          string            `json:"id" bson:"_id"`
	Name        string            `json:"name" bson:"name"`
	DisplayName string            `json:"display_name" bson:"display_name"`
	Bio         string            `json:"bio" bson:"bio"`
	Pronouns    string            `json:"pronouns" bson:"pronouns"`
	Phone       string            `json:"phone" bson:"phone"`
	TimeZone    string            `json:"time_zone" bson:"time_zone"`
	Socials     map[string]string `json:"socials" bson:"socials"`
}
