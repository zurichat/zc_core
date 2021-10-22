package organizations

import (
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
	CardCollectionName             = "cards"
	UserCollectionName             = "users"
	PluginCollection               = "plugins"
)

const (
	CreateOrganizationMember              = "CreateOrganizationMember"
	UpdateOrganizationName                = "UpdateOrganizationName"
	UpdateOrganizationMemberPic           = "UpdateOrganizationMemberPic"
	UpdateOrganizationURL                 = "UpdateOrganizationUrl"
	UpdateOrganizationLogo                = "UpdateOrganizationLogo"
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
const StatusHistoryLimit = 6

var ExpiryTime = make(chan int64, 1)
var ClearOld = make(chan bool, 1)

var RequestData = make(map[string]string)

const(
	logoWidth = 111
	logoHeight = 74
	imageWidth = 170
	imageHeight = 170
)

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
	Customize    Customize                `json:"customize" bson:"customize"`
	LogoURL      string                   `json:"logo_url" bson:"logo_url"`
	WorkspaceURL string                   `json:"workspace_url" bson:"workspace_url"`
	CreatedAt    time.Time                `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at" bson:"updated_at"`
	Tokens       float64                  `json:"tokens" bson:"tokens"`
	Version      string                   `json:"version" bson:"version"`
	Billing      Billing                  `json:"billing" bson:"billing"`
}

type Billing struct {
	Settings 	BillingSetting 	`json:"billing_setting" bson:"setting"`
	Contact 	BillingContact	`json:"billing_contact" bson:"contact"`
}

type BillingContact struct {
	ToDefaultEmail      bool 		`json:"to_default_email" bson:"to_default_email"`
	Contact 			[]Contact	`json:"contacts" bson:"contacts"`
}

type Contact struct {
	Email         string `json:"email" bson:"email"`
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
	OrgID         string    `json:"org_id" bson:"org_id"`
	Currency      string    `json:"currency" bson:"currency"`
	Token         float64   `json:"token" bson:"token"`
	Type          string    `json:"type" bson:"type"`
	Description   string    `json:"description" bson:"description"`
	Amount        float64   `json:"amount" bson:"amount"`
	Time          time.Time `json:"time" bson:"time"`
	TransactionID string    `json:"transaction_id" bson:"transaction_id"`
}

type Invite struct {
	ID          string `json:"_id,omitempty" bson:"_id,omitempty"`
	OrgID       string `json:"org_id" bson:"org_id"`
	UUID        string `json:"uuid" bson:"uuid"`
	Email       string `json:"email" bson:"email"`
	HasAccepted bool   `json:"has_accepted" bson:"has_accepted"`
}
type SendInviteResponse struct {
	InvalidEmails []interface{}
	InviteIDs     []interface{}
}

type OrgPluginBody struct {
	PluginID string `json:"plugin_id"`
	UserID   string `json:"user_id"`
}

type InstalledPlugin struct {
	ID          string                 `json:"id" bson:"_id"`
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

type Social struct {
	URL   string `json:"url" bson:"url"`
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
	Tag           string          `json:"tag" bson:"tag"`
	Text          string          `json:"text" bson:"text"`
	ExpiryTime    string          `json:"expiry_time" bson:"expiry_time"`
	StatusHistory []StatusHistory `json:"status_history" bson:"status_history"`
}

type StatusHistory struct {
	TagHistory    string `json:"tag_history" bson:"tag_history"`
	TextHistory   string `json:"text_history" bson:"text_history"`
	ExpiryHistory string `json:"expiry_history" bson:"expiry_history"`
}

type Member struct {
	ID          string    `json:"_id,omitempty" bson:"_id,omitempty"`
	OrgID       string    `json:"org_id" bson:"org_id"`
	Files       []string  `json:"files" bson:"files"`
	ImageURL    string    `json:"image_url" bson:"image_url"`
	FirstName   string    `json:"first_name" bson:"first_name"`
	LastName    string    `json:"last_name" bson:"last_name"`
	Email       string    `json:"email" bson:"email"`
	UserName    string    `bson:"user_name" json:"user_name"`
	DisplayName string    `json:"display_name" bson:"display_name"`
	Bio         string    `json:"bio" bson:"bio"`
	Status      Status    `json:"status" bson:"status"`
	Presence    string    `json:"presence" bson:"presence"`
	Pronouns    string    `json:"pronouns" bson:"pronouns"`
	Phone       string    `json:"phone" bson:"phone"`
	TimeZone    string    `json:"time_zone" bson:"time_zone"`
	Role        string    `json:"role" bson:"role"`
	JoinedAt    time.Time `json:"joined_at" bson:"joined_at"`
	Settings    *Settings `json:"settings" bson:"settings"`
	Deleted     bool      `json:"deleted" bson:"deleted"`
	DeletedAt   time.Time `json:"deleted_at" bson:"deleted_at"`
	Socials     []Social  `json:"socials" bson:"socials"`
	Language    string    `json:"language" bson:"language"`
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
	Themes              UserThemes           `json:"themes" bson:"themes"`
	MessagesAndMedia    MessagesAndMedia    `json:"messages_and_media" bson:"messages_and_media"`
	ChatSettings        ChatSettings        `json:"chat_settings" bson:"chat_settings"`
	LanguagesAndRegions LanguagesAndRegions `json:"languages_and_regions" bson:"languages_and_regions"`
	Accessibility       Accessibility       `json:"accessibility" bson:"accessibility"`
	Advanced            Advanced            `json:"advanced" bson:"advanced"`
	AudioAndVideo       AudioAndVideo       `json:"audio_and_video" bson:"audio_and_video"`
	PluginSettings      []PluginSettings    `json:"plugin_settings" bson:"plugin_settings"`
}

type Customize struct {
	Prefixes       []ChannelPrefixes `json:"prefixes" bson:"prefixes"`
	AddCustomEmoji []CustomEmoji     `json:"addcustomemoji" bson:"addcustomemoji"`
	SlackBot       []SlackBot        `json:"slackbot" bson:"slackbot"`
}

type SlackBot struct {
	WhenSomeOneSays string `json:"whensomeonesays" bson:"whensomeonesays"`
	SlackResponds   string `json:"slackresponds" bson:"slackresponds"`
}

type ChannelPrefixes struct {
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}

type CustomEmoji struct {
	Name      string    `json:"name" bson:"name"`
	ImageURL  string    `json:"imageurl" bson:"imageurl"`
	User      string    `json:"user" bson:"user"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
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
	WorkspaceURL       string                 `json:"workspacename" bson:"workspacename"`
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
	ChannelHurdleNotification        bool                   `json:"channel_hurdle_notification" bson:"channel_hurdle_notification"`
	NotificationSchedule             NotificationSchedule   `json:"notification_schedule" bson:"notification_schedule"`
	CustomNotificationSchedule       []NotificationSchedule `json:"custom_notification_schedule" bson:"custom_notification_schedule"`
	MessagePreviewInEachNotification bool                   `json:"message_preview_in_each_notification" bson:"message_preview_in_each_notification"`
	SetMessageNotificationsRight     string                 `json:"set_message_notifications_right" bson:"set_message_notifications_right"`
	SetLoungeNotificationsRight      string                 `json:"set_lounge_notifications_right" bson:"set_lounge_notifications_right"`
	MuteAllSounds                    bool                   `json:"mute_all_sounds" bson:"mute_all_sounds"`
}

type NotificationSchedule struct {
	Day  string `json:"day" bson:"day"`
	From string `json:"from" bson:"from"`
	To   string `json:"to" bson:"to"`
}

type Sidebar struct {
	AlwaysShowInTheSidebar        []string `json:"always_show_in_the_sidebar" bson:"always_show_in_the_sidebar"`
	ShowAllTheFollowing           string   `json:"show_all_the_following" bson:"show_all_the_following"`
	SidebarSort                   string   `json:"sidebar_sort" bson:"sidebar_sort"`
	ShowProfilePictureNextToDM    bool     `json:"show_profile_picture_next_to_dm" bson:"show_profile_picture_next_to_dm"`
	ListPrivateChannelsSeperately bool     `json:"list_private_channels_separately" bson:"list_private_channels_separately"`
	OrganizeExternalConversations bool     `json:"organize_external_conversations" bson:"organize_external_conversations"`
	ShowConversations             string   `json:"show_conversations" bson:"show_conversations"`
}

type Themes struct {
	SyncWithOsSetting                bool   `json:"sync_with_os_setting" bson:"sync_with_os_setting"`
	DirectMessagesMentionsAndNetwork bool   `json:"direct_messages_mentions_and_networks" bson:"direct_messages_mentions_and_networks"`
	Themes                           string `json:"themes" bson:"themes"`
	Colors                           string `json:"colors" bson:"colors"`
}
type UserThemes struct {
	Mode	string `json:"mode"`
	Colors	string `json:"colors"`
}
const (
	ThemeClean   = "clean"
	ThemeCompact = "compact"
	NameFull     = "full & display names"
	NameDisplay  = "just display names"
	EmojiTone1   = "EmojiTone1"
	EmojiTone2   = "EmojiTone2"
	EmojiTone3   = "EmojiTone3"
	EmojiTone4   = "EmojiTone4"
	EmojiTone5   = "EmojiTone5"
)

var MsgMedias = map[string]string{
	ThemeClean:   ThemeClean,
	ThemeCompact: ThemeCompact,
	NameFull:     NameFull,
	NameDisplay:  NameDisplay,
	EmojiTone1:   EmojiTone1,
	EmojiTone2:   EmojiTone2,
	EmojiTone3:   EmojiTone3,
	EmojiTone4:   EmojiTone4,
	EmojiTone5:   EmojiTone5,
}

type AdditionalOption struct {
	CurrentlyTyping bool `json:"currently_typing" bson:"currently_typing"`
	Clock           bool `json:"clock" bson:"clock"`
	ColorSwatches   bool `json:"color_swatches" bson:"color_swatches"`
}

type InlineMediaAndLinks struct {
	ShowImagesAndFilesUploaded  bool `json:"show_images_and_files_uploaded_to_zurichat" bson:"show_images_and_files_uploaded_to_zurichat"`
	ShowImagesAndFilesFromSites bool `json:"show_images_and_files_from_linked_websites" bson:"show_images_and_files_from_linked_websites"`
	LargerThan2MB               bool `json:"larger_than_2_mb" bson:"larger_than_2_mb"`
	ShowTextPreviews            bool `json:"show_text_previews_of_linked_websites" bson:"show_text_previews_of_linked_websites"`
}

type MessagesAndMedia struct {
	Theme                    string              `json:"theme" bson:"theme"`
	Names                    string              `json:"names" bson:"names"`
	AdditionalOptions        AdditionalOption    `json:"additional_options" bson:"additional_options"`
	Emoji                    string              `json:"emoji" bson:"emoji"`
	EmojiAsText              bool                `json:"emoji_as_text" bson:"emoji_as_text"`
	ShowJumboMoji            bool                `json:"show_jumbomoji" bson:"show_jumbomoji"`
	ConvertEmoticonsToEmoji  bool                `json:"convert_emoticons_to_emoji" bson:"convert_emoticons_to_emoji"`
	MessagesOneClickReaction []string            `json:"messages_one_click_reaction" bson:"messages_one_click_reaction"`
	FrequentlyUsedEmoji      bool                `json:"frequently_used_emoji" bson:"frequently_used_emoji"`
	Custom                   bool                `json:"custom" bson:"custom"`
	InlineMediaAndLinks      InlineMediaAndLinks `json:"inline_media_and_links" bson:"inline_media_and_links"`
	BringEmailsIntoZuri      string              `json:"bring_emails_into_zuri" bson:"bring_emails_into_zuri"`
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

const (
	FocusOnLastMessage = "focus_on_last_message"
	EditLastMessage    = "edit_last_message"
)

var EmptyMessageFields = map[string]string{
	FocusOnLastMessage: FocusOnLastMessage,
	EditLastMessage:    EditLastMessage,
}

type DirectMessageAnnouncement struct {
	ReceiveSound bool `json:"receive_sound" bson:"receive_sound"`
	SendSound    bool `json:"send_sound" bson:"send_sound"`
	ReadMessage  bool `json:"read_message" bson:"read_message"`
}

type Accessibility struct {
	Links                     bool                      `json:"links" bson:"links"`
	Animation                 bool                      `json:"animation" bson:"animation"`
	DirectMessageAnnouncement DirectMessageAnnouncement `json:"direct_message_announcement" bson:"direct_message_announcement"`
	PressEmptyMessageField    string                    `json:"press_empty_message_field" bson:"press_empty_message_field"`
}

type InputOption struct {
	DontSendWithEnter bool `json:"dont_send_with_enter" bson:"dont_send_with_enter"`
	FormatMessages    bool `json:"format_messages" bson:"format_messages"`
}

type SearchOption struct {
	StartSlackSearch   bool `json:"start_slack_search" bson:"start_slack_search"`
	StartQuickSwitcher bool `json:"start_quick_switcher" bson:"start_quick_switcher"`
}

type OtherOption struct {
	KeyScrollMessages bool `json:"key_scroll_messages" bson:"key_scroll_messages"`
	ToggleAwayStatus  bool `json:"toggle_away_status"  bson:"toggle_away_status"`
	SendSurvey        bool `json:"send_survey" bson:"send_survey"`
	WarnAgainstLinks  bool `json:"warn_against_links" bson:"warn_against_links"`
	WarnAgainstFiles  bool `json:"warn_against_files" bson:"warn_against_files"`
}

const (
	SendMessage  = "send_message"
	StartNewLine = "start_new_line"
)

var EnterActions = map[string]string{
	SendMessage:  SendMessage,
	StartNewLine: StartNewLine,
}

type Advanced struct {
	InputOption      InputOption  `json:"input_option" bson:"input_option"`
	PressEnterTo     string       `json:"press_enter_to" bson:"press_enter_to"`
	SearchOption     SearchOption `json:"search_option" bson:"search_option"`
	ExcludedChannels []string     `json:"excluded_channels" bson:"excluded_channels"`
	OtherOption      OtherOption  `json:"other_option" bson:"other_option"`
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

type updateParam struct {
	orgFilterKey   string
	requestDataKey string
	eventKey       string
	successMessage string
}

type Card struct {
	NameOnCard string `json:"name_on_card" bson:"name_on_card"`
	OrgID      string `json:"org_id" bson:"org_id"`
	MemberID   string `json:"member_id" bson:"member_id"`
	Type       string `json:"type" bson:"type"`
	ExpMonth   int    `json:"exp_month" bson:"exp_month"`
	ExpYear    int    `json:"exp_year" bson:"exp_year"`
	CardNumber string `json:"card_number" bson:"card_number"`
	Country    string `json:"country,omitempty" bson:"country,omitempty"`
	CVCCheck   string `json:"cvc_check" bson:"cvc_check"`
}

type EnterLeaveMessage struct {
	OrganizationID string `json:"organization_id" bson:"organization_id"`
	MemberID       string `json:"member_id" bson:"member_id"`
}

type MemberIDS struct {
	IDList []string `json:"id_list" bson:"id_list" validate:"required"`
}

type HandleMemberSearchResponse struct {
	Memberinfo Member
	Err        error
}
