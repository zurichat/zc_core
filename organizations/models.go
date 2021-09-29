package organizations

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

const (
	OrganizationCollectionName     = "organizations"
	InstalledPluginsCollectionName = "installed_plugins"
	OrganizationInviteCollection   = "organizations_invites"
	OrganizationSettings           = "organizations_settings"
	MemberCollectionName           = "members"
	UserCollectionName             = "users"
)

const (
	CreateOrganizationMember         = "CreateOrganizationMember"
	UpdateOrganizationName           = "UpdateOrganizationName"
	UpdateOrganizationMemberPic      = "UpdateOrganizationMemberPic"
	UpdateOrganizationUrl            = "UpdateOrganizationUrl"
	UpdateOrganizationLogo           = "UpdateOrganizationUrl"
	DeactivateOrganizationMember     = "DeactivateOrganizationMember"
	ReactivateOrganizationMember     = "ReactivateOrganizationMember"
	UpdateOrganizationMemberStatus   = "UpdateOrganizationMemberStatus"
	UpdateOrganizationMemberProfile  = "UpdateOrganizationMemberProfile"
	UpdateOrganizationMemberPresence = "UpdateOrganizationMemberPresence"
	UpdateOrganizationMemberSettings = "UpdateOrganizationMemberSettings"
)

const (
	OwnerRole  = "owner"
	AdminRole  = "admin"
	EditorRole = "editor"
	MemberRole = "member"
	GuestRole  = "guest"
)

const (
	FreeVersion = "free"
	ProVersion  = "pro"
)

const NairaToTokenRate = 0.01

type MemberPassword struct {
	MemberID string `bson:"member_id"`
	Password string `bson:"password"`
}

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
	Tokens       float64                  `json:"tokens" bson:"tokens"`
}

type Invite struct {
	ID    string `json:"_id,omitempty" bson:"_id,omitempty"`
	OrgID string `json:"org_id" bson:"org_id"`
	Uuid  string `json:"uuid" bson:"uuid"`
	Email string `json:"email" bson:"email"`
}
type SendInviteResponse struct {
	InvalidEmails []interface{}
	InviteIDs     []interface{}
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

type SendInviteBody struct {
	Emails []string `json:"emails" bson:"emails"`
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

type Social struct {
	Url   string `json:"url" bson:"url"`
	Title string `json:"title" bson:"title"`
}

type Member struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	OrgId       string             `json:"org_id" bson:"org_id"`
	Files       []string           `json:"files" bson:"files"`
	ImageURL    string             `json:"image_url" bson:"image_url"`
	FirstName   string             `json:"first_name" bson:"first_name"`
	LastName    string             `json:"last_name" bson:"last_name"`
	Email       string             `json:"email" bson:"email"`
	UserName    string             `bson:"user_name" json:"user_name"`
	DisplayName string             `json:"display_name" bson:"display_name"`
	Bio         string             `json:"bio" bson:"bio"`
	Status      string             `json:"status" bson:"status"`
	Presence    string             `json:"presence" bson:"presence"`
	Pronouns    string             `json:"pronouns" bson:"pronouns"`
	Phone       string             `json:"phone" bson:"phone"`
	TimeZone    string             `json:"time_zone" bson:"time_zone"`
	Role        string             `json:"role" bson:"role"`
	JoinedAt    time.Time          `json:"joined_at" bson:"joined_at"`
	Settings    *Settings          `json:"settings" bson:"settings"`
	Deleted     bool               `json:"deleted" bson:"deleted"`
	DeletedAt   time.Time          `json:"deleted_at" bson:"deleted_at"`
	Socials     []Social           `json:"socials" bson:"socials"`
	Language    string             `json:"language" bson:"language"`
}
type Profile struct {
	ID          string   `json:"id" bson:"_id"`
	FirstName   string   `json:"first_name" bson:"first_name"`
	LastName    string   `json:"last_name" bson:"last_name"`
	DisplayName string   `json:"display_name" bson:"display_name"`
	Bio         string   `json:"bio" bson:"bio"`
	Pronouns    string   `json:"pronouns" bson:"pronouns"`
	Phone       string   `json:"phone" bson:"phone"`
	TimeZone    string   `json:"time_zone" bson:"time_zone"`
	Socials     []Social `json:"socials" bson:"socials"`
	Language    string   `json:"language" bson:"language"`
}

type Settings struct {
	Notifications    Notifications    `json:"notifications" bson:"notifications"`
	Sidebar          Sidebar          `json:"sidebar" bson:"sidebar"`
	Themes           Themes           `json:"themes" bson:"themes"`
	MessagesAndMedia MessagesAndMedia `json:"messages_and_media" bson:"messages_and_media"`
	ChatSettings     ChatSettings     `json:"chat_settings" bson:"chat_settings"`
}

type Notifications struct {
	NotifyMeAbout                      string   `json:"notify_me_about" bson:"notify_me_about"`
	UseDifferentSettingsForMyMobile    string   `json:"use_different_settings_mobile" bson:"use_different_settings_mobile"`
	ChannelHurdleNotification          bool     `json:"channel_hurdle_notification" bson:"channel_hurdle_notification"`
	ThreadRepliesNotification          bool     `json:"thread_replies_notification" bson:"thread_replies_notification"`
	MyKeywords                         string   `json:"my_keywords" bson:"my_keywords"`
	NotificationSchedule               string   `json:"notification_schedule" bson:"notification_schedule"`
	MessagePreviewInEachNotification   bool     `json:"message_preview_in_each_notification" bson:"message_preview_in_each_notification"`
	MuteAllSounds                      bool     `json:"mute_all_sounds" bson:"mute_all_sounds"`
	WhenIamNotActiveOnDesktop          string   `json:"when_iam_not_active_on_desktop" bson:"when_iam_not_active_on_desktop"`
	EmailNotificationsForMentionsAndDM []string `json:"email_notifications_for_mentions_and_dm" bson:"email_notifications_for_mentions_and_dm"`
}

type Sidebar struct {
	AlwaysShowInTheSidebar        []string `json:"always_show_in_the_sidebar" bson:"always_show_in_the_sidebar"`
	SidebarSort                   string   `json:"sidebar_sort" bson:"sidebar_sort"`
	ShowProfilePictureNextToDM    bool     `json:"show_profile_picture_next_to_dm" bson:"show_profile_picture_next_to_dm"`
	ListPrivateChannelsSeperately bool     `json:"list_private_channels_seperately" bson:"list_private_channels_seperately"`
	OrganizeExternalConversations bool     `json:"organize_external_conversations" bson:"organize_external_conversations"`
	ShowConversations             string   `json:"show_conversations" bson:"show_conversations"`
}

type Themes struct {
	Themes string `json:"themes" bson:"themes"`
	Colors string `json:"colors" bson:"colors"`
}

type MessagesAndMedia struct {
	Theme                    string   `json:"theme" bson:"theme"`
	Names                    string   `json:"names" bson:"names"`
	AdditionalOptions        []string `json:"additional_options" bson:"additional_options"`
	Emoji                    string   `json:"emoji" bson:"emoji"`
	EmojiAsText              bool     `json:"emoji_as_text" bson:"emoji_as_text"`
	ShowJumboMoji            bool     `json:"show_jumbomoji" bson:"show_jumbomoji"`
	ConvertEmoticonsToEmoji  bool     `json:"convert_emoticons_to_emoji" bson:"convert_emoticons_to_emoji"`
	MessagesOneClickReaction []string `json:"messages_one_click_reaction" bson:"messages_one_click_reaction"`
	FrequentlyUsedEmoji      bool     `json:"frequently_used_emoji" bson:"frequently_used_emoji"`
	Custom                   bool     `json:"custom" bson:"custom"`
	InlineMediaAndLinks      []string `json:"inline_media_and_links" bson:"inline_media_and_links"`
	BringEmailsIntoZuri      string   `json:"bring_emails_into_zuri" bson:"bring_emails_into_zuri"`
}

type ChatSettings struct {
	Theme           string `json:"theme" bson:"theme"`
	Wallpaper       string `json:"wallpaper" bson:"wallpaper"`
	EnterIsSend     bool   `json:"enter_is_send" bson:"enter_is_send"`
	MediaVisibility bool   `json:"media_visibility" bson:"media_visibility"`
	FontSize        string `json:"font_size" bson:"font_size"`
}
type OrganizationHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

func NewOrganizationHandler(c *utils.Configurations, mail service.MailService) *OrganizationHandler {
	return &OrganizationHandler{configs: c, mailService: mail}
}

// gets the details of a member in a workspace using parameters such as email, username etc
// returns parameters based on the member struct
func FetchMember(filter map[string]interface{}) (*Member, error) {
	member_collection := MemberCollectionName
	member := &Member{}
	memberCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), member_collection)
	if err != nil {
		return member, err
	}
	result := memberCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&member)
	return member, err
}
