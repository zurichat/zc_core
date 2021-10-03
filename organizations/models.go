package organizations

import (
	"encoding/json"

	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

const (
	OrganizationCollectionName     = "organizations"
	TokenTransactionCollectionName = "token_transaction"
	InstalledPluginsCollectionName = "installed_plugins"
	OrganizationInviteCollection   = "organizations_invites"
	MemberCollectionName           = "members"
	UserCollectionName             = "users"
)

const (
	CreateOrganizationMember              = "CreateOrganizationMember"
	UpdateOrganizationName                = "UpdateOrganizationName"
	UpdateOrganizationMemberPic           = "UpdateOrganizationMemberPic"
	UpdateOrganizationUrl                 = "UpdateOrganizationUrl"
	UpdateOrganizationLogo                = "UpdateOrganizationUrl"
	DeactivateOrganizationMember          = "DeactivateOrganizationMember"
	ReactivateOrganizationMember          = "ReactivateOrganizationMember"
	UpdateOrganizationMemberStatus        = "UpdateOrganizationMemberStatus"
	UpdateOrganizationMemberProfile       = "UpdateOrganizationMemberProfile"
	UpdateOrganizationMemberPresence      = "UpdateOrganizationMemberPresence"
	UpdateOrganizationMemberSettings      = "UpdateOrganizationMemberSettings"
	UpdateOrganizationMemberRole          = "UpdateOrganizationMemberRole"
	UpdateOrganizationMemberStatusCleared = "UpdateOrganizationMemberStatusCleared"
)

const (
	OwnerRole  = "owner"
	AdminRole  = "admin"
	EditorRole = "editor"
	MemberRole = "member"
	GuestRole  = "guest"
	Bot        = "bot"
)

var Roles = map[string]string{
	OwnerRole:  OwnerRole,
	AdminRole:  AdminRole,
	EditorRole: EditorRole,
	MemberRole: MemberRole,
	GuestRole:  GuestRole,
}

const (
	FreeVersion = "free"
	ProVersion  = "pro"
)

const ProSubscriptionRate = 10

var RequestData = make(map[string]string)

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
	Settings     OrganizationPreference   `json:"settings" bson:"settings"`
	LogoURL      string                   `json:"logo_url" bson:"logo_url"`
	WorkspaceURL string                   `json:"workspace_url" bson:"workspace_url"`
	CreatedAt    time.Time                `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at" bson:"updated_at"`
	Tokens       float64                  `json:"tokens" bson:"tokens"`
	Version      string                   `json:"version" bson:"version"`
	Billing      Billing                  `json:"billing" bson:"billing"`
}

type Billing struct {
	Settings BillingSetting
}

type BillingSetting struct {
	Country         string `json:"country" bson:"country"`
	CompanyName     string `json:"company_name" bson:"company_name"`
	StreetAddress   string `json:"street_address" bson:"street_address" `
	Suite           string `json:"suite" bson:"suite"`
	City            string `json:"city" bson:"city"`
	State           string `json:"state" bson:"state"`
	PostalCode      string `json:"postal_code" bson:"postal_code"`
	AdditionalNotes string `json:"additional_notes" bson:"additional_notes"`
}

type TokenTransaction struct {
	OrgId         string    `json:"org_id" bson:"org_id"`
	Currency      string    `json:"currency" bson:"currency"`
	Token         float64   `json:"token" bson:"token"`
	Type          string    `json:"type" bson:"type"`
	Description   string    `json:"description" bson:"description"`
	Amount        float64   `json:"amount" bson:"amount"`
	Time          time.Time `json:"time" bson:"time"`
	TransactionId string    `json:"transaction_id" bson:"transaction_id"`
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

const (
	DontClear  = "dont_clear"
	ThirtyMins = "thirty_mins"
	OneHr      = "one_hour"
	FourHrs    = "four_hours"
	Today      = "today"
	ThisWeek   = "this_week"
)

var StatusExpiryTime = map[string]string{
	DontClear:  DontClear,
	ThirtyMins: ThirtyMins,
	OneHr:      OneHr,
	FourHrs:    FourHrs,
	Today:      Today,
	ThisWeek:   ThisWeek,
}

type Status struct {
	Tag        string `json:"tag" bson:"tag"`
	Text       string `json:"text" bson:"text"`
	ExpiryTime string `json:"expiry_time" bson:"expiry_time"`
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
	Status      Status             `json:"status" bson:"status"`
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
	Notifications       Notifications       `json:"notifications" bson:"notifications"`
	Sidebar             Sidebar             `json:"sidebar" bson:"sidebar"`
	Themes              Themes              `json:"themes" bson:"themes"`
	MessagesAndMedia    MessagesAndMedia    `json:"messages_and_media" bson:"messages_and_media"`
	ChatSettings        ChatSettings        `json:"chat_settings" bson:"chat_settings"`
	LanguagesAndRegions LanguagesAndRegions `json:"languages_and_regions" bson:"languages_and_regions"`
	Accessibility       Accessibility       `json:"accessibility" bson:"accessibility"`
	MarkAsRead          MarkAsRead          `json:"mark_as_read" bson:"mark_as_read"`
	AudioAndVideo       AudioAndVideo       `json:"audio_and_video" bson:"audio_and_video"`
	PluginSettings      []PluginSettings    `json:"plugin_settings" bson:"plugin_settings"`
}

type OrganizationPreference struct {
	Settings       OrgSettings       `json:"settings" bson:"settings"`
	Permissions    OrgPermissions    `json:"permissions" bson:"permissions"`
	Authentication OrgAuthentication `json:"authentication" bson:"authentication"`
}

type OrgAuthentication struct {
	AuthenticationMethod                 map[string]interface{} `json:"authenticationmethod" bson:"authenticationmethod"`
	WorkspaceWideTwoFactorAuthentication map[string]interface{} `json:"workspacewidetwofactorauthentication" bson:"workspacewidetwofactorauthentication"`
	SessionDuration                      string                 `json:"sessionduration" bson:"sessionduration"`
	ForcedPasswordReset                  map[string]interface{} `json:"forcedpasswordreset" bson:"forcedpasswordreset"`
	AutomaticallyOpen                    map[string]interface{} `json:"automaticallyopen" bson:"automaticallyopen"`
}

type OrgSettings struct {
	OrganizationIcon   string                 `json:"workspaceicon" bson:"workspaceicon"`
	DeleteOrganization map[string]interface{} `json:"deleteorganization" bson:"deleteorganization"`
	WorkspaceLanguage  string                 `json:"workspacelanguage" bson:"workspacelanguage"`
	DefaultChannels    []string               `json:"defaultchannels" bson:"defaultchannels"`
	ShowDisplayName    bool                   `json:"showdisplayname" bson:"showdisplayname"`
	DisplayEmail       bool                   `json:"displayemail" bson:"displayemail"`
	DisplayPronouns    bool                   `json:"displaypronouns" bson:"displaypronouns"`
	NotifyOfNewUsers   bool                   `json:"notifyofnewusers" bson:"notifyofnewusers"`
	WorkspaceUrl       string                 `json:"workspacename" bson:"workspacename"`
}

type OrgPermissions struct {
	Messaging         map[string]interface{} `json:"messaging" bson:"messaging"`
	Invitations       bool                   `json:"invitations" bson:"invitations"`
	MessageSettings   MessageSettings        `json:"messagesettings" bson:"messagesettings"`
	CustomEmoji       map[string]interface{} `json:"customemoji" bson:"customemoji"`
	PublicFileSharing bool                   `json:"publicfilesharing" bson:"publicfilesharing"`
}

type MessageSettings struct {
	MessageEditing  bool `json:"messageediting" bson:"messageediting"`
	MessageDeleting bool `json:"messagedeleting" bson:"messagedeleting"`
}

type Notifications struct {
	NotifyMeAbout                    string   `json:"notify_me_about" bson:"notify_me_about"`
	UseDifferentSettingsForMyMobile  bool     `json:"use_different_settings_mobile" bson:"use_different_settings_mobile"`
	ChannelHurdleNotification        bool     `json:"channel_hurdle_notification" bson:"channel_hurdle_notification"`
	MeetingRepliesNotification       bool     `json:"meeting_replies_notification" bson:"meeting_replies_notification"`
	ThreadRepliesNotification        bool     `json:"thread_replies_notification" bson:"thread_replies_notification"`
	MyKeywords                       []string `json:"my_keywords" bson:"my_keywords"`
	NotificationSchedule             string   `json:"notification_schedule" bson:"notification_schedule"`
	MessagePreviewInEachNotification bool     `json:"message_preview_in_each_notification" bson:"message_preview_in_each_notification"`
	SetMessageNotificationsRight     string   `json:"set_message_notifications_right" bson:"set_message_notifications_right"`
	SetLoungeNotificationsRight      string   `json:"set_lounge_notifications_right" bson:"set_lounge_notifications_right"`
	MuteAllSounds                    bool     `json:"mute_all_sounds" bson:"mute_all_sounds"`
	FlashWindowWhenNotificationComes string   `json:"flash_window_when_notification_comes" bson:"flash_window_when_notification_comes"`
	DeliverNotificationsVia          string   `json:"deliver_notifications_via" bson:"deliver_notifications_via"`
	WhenIamNotActiveOnDesktop        string   `json:"when_iam_not_active_on_desktop" bson:"when_iam_not_active_on_desktop"`
	EmailNotificationsForMentions    bool     `json:"email_notifications_for_mentions" bson:"email_notifications_for_mentions"`
}

type Sidebar struct {
	AlwaysShowInTheSidebar        []string `json:"always_show_in_the_sidebar" bson:"always_show_in_the_sidebar"`
	ShowAllTheFollowing           string   `json:"show_all_the_following" bson:"show_all_the_following"`
	SidebarSort                   string   `json:"sidebar_sort" bson:"sidebar_sort"`
	ShowProfilePictureNextToDM    bool     `json:"show_profile_picture_next_to_dm" bson:"show_profile_picture_next_to_dm"`
	ListPrivateChannelsSeperately bool     `json:"list_private_channels_seperately" bson:"list_private_channels_seperately"`
	OrganizeExternalConversations bool     `json:"organize_external_conversations" bson:"organize_external_conversations"`
	ShowConversations             string   `json:"show_conversations" bson:"show_conversations"`
}

type Themes struct {
	SyncWithOsSetting                bool   `json:"sync_with_os_setting" bson:"sync_with_os_setting"`
	DirectMessagesMentionsAndNetwork bool   `json:"direct_messages_mentions_and_networks" bson:"direct_messages_mentions_and_networks"`
	Themes                           string `json:"themes" bson:"themes"`
	Colors                           string `json:"colors" bson:"colors"`
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

type LanguagesAndRegions struct {
	Language                      string   `json:"language" bson:"language"`
	TimeZone                      string   `json:"time_zone" bson:"time_zone"`
	SetTimeZoneAutomatically      bool     `json:"set_time_zone_automatically" bson:"set_time_zone_automatically"`
	SpellCheck                    bool     `json:"spell_check" bson:"spell_check"`
	LanguagesZuriShouldSpellCheck []string `json:"languages_zuri_should_spell_check" bson:"languages_zuri_should_spell_check"`
}

type Accessibility struct {
	Animation                 bool     `json:"animation" bson:"animation"`
	DirectMessageAnnouncement []string `json:"direct_message_announcement" bson:"direct_message_announcement"`
	PressEmptyMessageField    string   `json:"press_empty_message_field" bson:"press_empty_message_field"`
}

type MarkAsRead struct {
	WhenIViewAChannel string `json:"when_i_view_a_channel" bson:"when_i_view_a_channel"`
	PromptToConfirm   bool   `json:"prompt_to_confirm" bson:"language"`
}
type AudioAndVideo struct {
	IntegratedWebcam           string   `json:"integrated_webcam" bson:"integrated_webcam"`
	Microphone                 string   `json:"microphone" bson:"microphone"`
	EnableAutomaticGainControl bool     `json:"enable_automatic_gain_control" bson:"enable_automatic_gain_control"`
	Speaker                    string   `json:"speaker" bson:"speaker"`
	WhenJoiningAZuriChatCall   []string `json:"when_joining_a_zuri_chat_call" bson:"when_joining_a_zuri_chat_call"`
	WhenJoiningAHuddle         []string `json:"when_joining_a_huddle" bson:"when_joining_a_huddle"`
	WhenSlackIsInTheBackground []string `json:"when_slack_is_in_the_background" bson:"when_slack_is_in_the_background"`
}
type PluginSettings struct {
	Plugin      string `json:"plugin" bson:"plugin" validate:"required"`
	AccessLevel string `json:"access_level" bson:"access_level" validate:"required"`
}
type OrganizationHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}
