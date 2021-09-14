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
	ID          primitive.ObjectID     `json:"_id" bson:"_id"`
	OrgId       string                 `json:"org_id" bson:"org_id"`
	Files       []string               `json:"files" bson:"files"`
	ImageURL    string                 `json:"image_url" bson:"image_url"`
	Name        string                 `json:"name" bson:"name"`
	Email       string                 `json:"email" bson:"email"`
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
	// Socials     Social    `json:"socials" bson:"socials"`
}

type Profile struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	DisplayName string    `json:"display_name" bson:"display_name"`
	Bio         string    `json:"bio" bson:"bio"`
	Pronouns    string    `json:"pronouns" bson:"pronouns"`
	Phone       string    `json:"phone" bson:"phone"`
	TimeZone    string    `json:"time_zone" bson:"time_zone"`
	Socials     [3]string `json:"socials" bson:"socials"`
}

type Preferences struct {
	Notifications    *Notifications    `json:"notifications" bson:"notifications"`
	Themes           *Themes           `json:"themes" bson:"themes"`
	MessagesAndMedia *MessagesAndMedia `json:"messages_and_media" bson:"messages_and_media"`
}

type Notifications struct {
	Section                            string   `json:"section" bson:"section"`
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

type Themes struct {
	Section string `json:"section" bson:"section"`
	Themes  string `json:"themes" bson:"themes"`
	Colors  string `json:"colors" bson:"colors"`
}

type MessagesAndMedia struct {
	Section               string   `json:"section" bson:"section"`
	Theme                 string   `json:"theme" bson:"theme"`
	Names                 string   `json:"names" bson:"names"`
	AdditionalOptions     []string `json:"additional_options" bson:"additional_options"`
	Emoji                 string   `json:"emoji" bson:"emoji"`
	EmojiAsText           bool     `json:"emoji_as_text" bson:"emoji_as_text"`
	ShowJumboMoji         bool     `json:"show_jumbomoji" bson:"show_jumbomoji"`
	FrequentlyUsedEmoji   bool     `json:"frequently_used_emoji" bson:"frequently_used_emoji"`
	Custom                bool     `json:"custom" bson:"custom"`
	InlineMediaAndLinks   []string `json:"inline_media_and_links" bson:"inline_media_and_links"`
	FilesizeBiggerThan2mb bool     `json:"filesize_bigger_than_2mb" bson:"filesize_bigger_than_2mb"`
	BringEmailsIntoZuri   string   `json:"bring_emails_into_zuri bson:"bring_emails_into_zuri"`
}

// type Preferences struct {
// 	// Notifications
// 	NotifyMeAbout                      string   `json:"notify_me_about" bson:"notify_me_about"`
// 	UseDifferentSettingsForMyMobile    string   `json:"use_different_settings_mobile" bson:"use_different_settings_mobile"`
// 	ChannelHurdleNotification          bool     `json:"channel_hurdle_notification" bson:"channel_hurdle_notification"`
// 	ThreadRepliesNotification          bool     `json:"thread_replies_notification" bson:"thread_replies_notification"`
// 	MyKeywords                         string   `json:"my_keywords" bson:"my_keywords"`
// 	NotificationSchedule               string   `json:"notification_schedule" bson:"notification_schedule"`
// 	MessagePreviewInEachNotification   bool     `json:"message_preview_in_each_notification" bson:"message_preview_in_each_notification"`
// 	MuteAllSounds                      bool     `json:"mute_all_sounds" bson:"mute_all_sounds"`
// 	WhenIamNotActiveOnDesktop          string   `json:"when_iam_not_active_on_desktop" bson:"when_iam_not_active_on_desktop"`
// 	EmailNotificationsForMentionsAndDM []string `json:"email_notifications_for_mentions_and_dm" bson:"email_notifications_for_mentions_and_dm"`

// 	// Themes
// 	Themes string `json:"themes" bson:"themes"`
// 	Colors string `json:"colors" bson:"colors"`

// 	// Messages and Media
// 	Theme                 string   `json:"theme" bson:"theme"`
// 	Names                 string   `json:"names" bson:"names"`
// 	AdditionalOptions     []string `json:"additional_options" bson:"additional_options"`
// 	Emoji                 string   `json:"emoji" bson:"emoji"`
// 	EmojiAsText           bool     `json:"emoji_as_text" bson:"emoji_as_text"`
// 	ShowJumboMoji         bool     `json:"show_jumbomoji" bson:"show_jumbomoji"`
// 	FrequentlyUsedEmoji   bool     `json:"frequently_used_emoji" bson:"frequently_used_emoji"`
// 	Custom                bool     `json:"custom" bson:"custom"`
// 	InlineMediaAndLinks   []string `json:"inline_media_and_links" bson:"inline_media_and_links"`
// 	FilesizeBiggerThan2mb bool     `json:"filesize_bigger_than_2mb" bson:"filesize_bigger_than_2mb"`
// 	BringEmailsIntoZuri   string   `json:"bring_emails_into_zuri bson:"bring_emails_into_zuri"`
// }
