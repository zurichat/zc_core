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
	_id       string                   `json:"id" bson:"_id"`
	Name      string                   `json:"name" bson:"name"`
	Email     string                   `json:"email" bson:"email"`
	CreatorID string                   `json:"creator_id" bson:"creator_id"`
	Plugins   []map[string]interface{} `json:"plugins" bson:"plugins"`
	Admins    []string                 `json:"admins" bson:"admins"`
	Settings  map[string]interface{}   `json:"settings" bson:"settings"`
	ImageURL  string                   `json:"image_url" bson:"image_url"`
	URL       string                   `json:"url" bson:"url"`
	CreatedAt time.Time                `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time                `json:"updated_at" bson:"updated_at"`
}

type InstalledPlugin struct {
	ID          primitive.ObjectID     `bson:"_id"`
	PluginID    string                 `bson:"plugin_id"`
	Plugin      map[string]interface{} `bson:"plugin"`
	AddedBy     string                 `bson:"added_by"`
	ApprovedBy  string                 `bson:"approved_by"`
	InstalledAt time.Time              `bson:"installed_at"`
	UpdatedAt   time.Time              `bson:"updated_at"`
}

type OrganizationAdmin struct {
	ID             primitive.ObjectID `bson:"id"`
	OrganizationID string             `bson:"organization_id"`
	UserID         string             `bson:"user_id"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

type OrganizationMember struct {
    ID          string                   `json:"id" bson:"_id"`
    Files         []string                  `json:"files" bson:"files"`
    Logo          string                   `json:"logo" bson:"logo"`
	Name          string                   `json:"name" bson:"name"`
	Email         string                   `json:"email" bson:"email"`
    DisplayName   string                   `json:"display_name" bson:"display_name"`
    Bio           string                   `json:"bio" bson:"bio"`
	Pronouns      []string                   `json:"pronouns" bson:"pronouns"`
	PhoneNumber   string                   `json:"phone_number" bson:"phone_number"`
	TimeZone      string                   	`json:"time_zone" bson:"time_zone"`
	Socials   	[]Social 					`json:"socials" bson:"s						ocials"`
	Preferences   []map[string]interface{} `json:"preferences" bson:"preferences"`
	IsActive	bool		`json:"is_active" bson:"is_active"`
}

type Social struct {
	 ID          	string                  `json:"id" bson:"_id"`
    Title         	string                  `json:"title" bson:"title"`
    Url          	string                   `json:"url" bson:"url"`
}

func GetOrgPluginCollectionName(orgName string) string {
	return strings.ToLower(orgName) + "_" + InstalledPluginsCollectionName
}