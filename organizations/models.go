package organizations

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrganizationCollectionName     = "organizations"
	InstalledPluginsCollectionName = "installed_plugins"
)

type Organization struct {
	ID           string                   `json:"id" bson:"_id"`
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
	ID          string             `json:"id" bson:"_id"`
	OrgId       primitive.ObjectID `json:"org_id" bson:"org_id"`
	Files       []string           `json:"files" bson:"files"`
	ImageURL    string             `json:"image_url" bson:"image_url"`
	Name        string             `json:"name" bson:"name"`
	Email       string             `json:"email" bson:"email"`
	DisplayName string             `json:"display_name" bson:"display_name"`
	Bio         string             `json:"bio" bson:"bio"`
	Status      string             `json:"status" bson:"status"`
	Pronouns    string             `json:"pronouns" bson:"pronouns"`
	Phone       string             `json:"phone" bson:"phone"`
	TimeZone    string             `json:"time_zone" bson:"time_zone"`
	Role        string             `json:"role" bson:"role"`
	JoinedAt    time.Time          `json:"joined_at" bson:"joined_at"`
	// Socials     Social    `json:"socials" bson:"socials"`
}
