package object

import "nhooyr.io/websocket"

type GuildFeature string
type Dispatch string
type ChannelType int

const (
	AnimatedBanner                        GuildFeature = "ANIMATED_BANNER"
	AnimatedIcon                          GuildFeature = "ANIMATED_ICON"
	ApplicationCommandPermissionsV2       GuildFeature = "APPLICATION_COMMAND_PERMISSIONS_V2"
	AutoModeration                        GuildFeature = "AUTO_MODERATION"
	Banner                                GuildFeature = "BANNER"
	Community                             GuildFeature = "COMMUNITY"
	CreatorMonetizableProvisional         GuildFeature = "CREATOR_MONETIZABLE_PROVISIONAL"
	CreatorStorePage                      GuildFeature = "CREATOR_STORE_PAGE"
	DeveloperSupportServer                GuildFeature = "DEVELOPER_SUPPORT_SERVER"
	Discoverable                          GuildFeature = "DISCOVERABLE"
	Featureable                           GuildFeature = "FEATURABLE"
	InvitesDisabled                       GuildFeature = "INVITES_DISABLED"
	InviteSplash                          GuildFeature = "INVITE_SPLASH"
	MemberVerificationGateEnabled         GuildFeature = "MEMBER_VERIFICATION_GATE_ENABLED"
	MoreStickers                          GuildFeature = "MORE_STICKERS"
	News                                  GuildFeature = "NEWS"
	Partnered                             GuildFeature = "PARTNERED"
	PreviewEnabled                        GuildFeature = "PREVIEW_ENABLED"
	RoleIcons                             GuildFeature = "ROLE_ICONS"
	RoleSubscriptionsAvailableForPurchase GuildFeature = "ROLE_SUBSCRIPTIONS_AVAILABLE_FOR_PURCHASE"
	RoleSubscriptionsEnabled              GuildFeature = "ROLE_SUBSCRIPTIONS_ENABLED"
	TicketedEventsEnabled                 GuildFeature = "TICKETED_EVENTS_ENABLED"
	VanityURL                             GuildFeature = "VANITY_URL"
	Verified                              GuildFeature = "VERIFIED"
	VIPRegions                            GuildFeature = "VIP_REGIONS"
	WelcomeScreenEnabled                  GuildFeature = "WELCOME_SCREEN_ENABLED"

	ReadyType             Dispatch = "READY"
	GuildCreateType       Dispatch = "GUILD_CREATE"
	MessageCreateType     Dispatch = "MESSAGE_CREATE"
	VoiceStateUpdateType  Dispatch = "VOICE_STATE_UPDATE"
	VoiceServerUpdateType Dispatch = "VOICE_SERVER_UPDATE"

	GuildText         ChannelType = 0
	DM                ChannelType = 1
	GuildVoice        ChannelType = 2
	GroupDM           ChannelType = 3
	GuildCategory     ChannelType = 4
	GuildAccouncement ChannelType = 5

	UnknownError         int = 4000
	UnknownOpcode        int = 4001
	DecodeError          int = 4002
	NotAuthenticated     int = 4003
	AuthenticationFailed int = 4004
	AlreadyAuthenticated int = 4005
	InvalidSeq           int = 4007
	RateLimited          int = 4008
	SessionTimedOut      int = 4009
	InvalidShard         int = 4010
	ShardingRequired     int = 4011
	InvalidAPIVersion    int = 4012
	InvalidIntent        int = 4013
	DisallowedIntent     int = 4014
	StatusNormalClosure  int = int(websocket.StatusNormalClosure)
)

var ReconnectOnError map[int]bool = map[int]bool{
	UnknownError:         true,
	UnknownOpcode:        true,
	DecodeError:          true,
	NotAuthenticated:     true,
	AuthenticationFailed: false,
	AlreadyAuthenticated: true,
	InvalidSeq:           true,
	RateLimited:          true,
	SessionTimedOut:      true,
	InvalidShard:         false,
	ShardingRequired:     false,
	InvalidAPIVersion:    false,
	InvalidIntent:        false,
	DisallowedIntent:     false,
	StatusNormalClosure:  true,
}

type Event[T any] struct {
	Op int     `json:"op"`
	D  T       `json:"d,omitempty"`
	S  *int    `json:"s,omitempty"`
	T  *string `json:"t,omitempty"`
}

type Resume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  int    `json:"seq"`
}

type Identify struct {
	Token      string         `json:"token"`
	Properties map[string]any `json:"properties"`
	Compress   bool           `json:"compress"`
	Intents    int            `json:"intents"`
}

type VoiceIdentify struct {
	ServerID  string `json:"server_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
}

type VoiceState struct {
	GuildID   *string `json:"guild_id,omitempty"`
	ChannelID *string `json:"channel_id,omitempty"`
	UserID    string  `json:"user_id"`
	SessionID string  `json:"session_id"`
	SelfMute  *bool   `json:"self_mute,omitempty"`
	SelfDeaf  *bool   `json:"self_deaf,omitempty"`
}

type Message struct {
	ID              string  `json:"id"`
	ChannelID       string  `json:"channel_id"`
	Author          *User   `json:"author,omitempty"`
	Content         *string `json:"content,omitempty"`
	Timestamp       string  `json:"timestamp"`
	EditedTimestamp *string `json:"edited_timestamp,omitempty"`
	TTS             bool    `json:"tts"`
	MentionEveryone bool    `json:"mention_everyone"`
	Mentions        []User  `json:"mentions"`
}

type Ready struct {
	V                int          `json:"v"`
	User             *User        `json:"user"`
	SessionID        string       `json:"session_id"`
	Guilds           []Guild      `json:"guilds"`
	ResumeGatewayURL string       `json:"resume_gateway_url"`
	Shard            *[2]int      `json:"shard,omitempty"`
	Application      *Application `json:"application"`
}

type VoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

type Application struct {
	ID                             string         `json:"id"`
	Name                           string         `json:"name"`
	Icon                           *string        `json:"icon"`
	Description                    string         `json:"description"`
	RPCOrigins                     *[]string      `json:"rpc_origins,omitempty"`
	BotPublic                      bool           `json:"bot_public"`
	BotRequireCodeGrant            bool           `json:"bot_require_code_grant"`
	TermsOfServiceURL              *string        `json:"terms_of_service,omitempty"`
	PrivacyPolicyURL               *string        `json:"privacy_policy_url,omitempty"`
	Owner                          *User          `json:"owner,omitempty"`
	VerifyKey                      string         `json:"verify_key"`
	Team                           *Team          `json:"team"`
	GuildID                        *string        `json:"guild_id,omitempty"`
	PrimarySkuID                   *string        `json:"primary_sku_id,omitempty"`
	Slug                           *string        `json:"slug,omitempty"`
	CoverImage                     *string        `json:"cover_image,omitempty"`
	Flags                          *int           `json:"flags,omitempty"`
	Tags                           *[]string      `json:"tags,omitempty"`
	InstallParams                  *InstallParams `json:"install_params,omitempty"`
	CustomInstallURL               *string        `json:"custom_install_url,omitempty"`
	RoleConnectionsVerificationURL *string        `json:"role_connections_verification_url,omitempty"`
}

type InstallParams struct {
	Scopes      []string `json:"scopes"`
	Permissions string   `json:"permissions"`
}

type Team struct {
	ID          string   `json:"id"`
	Icon        string   `json:"icon"`
	Members     []Member `json:"members"`
	Name        string   `json:"name"`
	OwnerUserID string   `json:"owner_user_id"`
}

type Member struct {
	MembershipState int    `json:"membership_state"`
	Permissions     []int  `json:"permissions"`
	TeamID          string `json:"team_id"`
	User            User   `json:"user"`
}

type User struct {
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Discriminator string  `json:"discriminator"`
	Avatar        *string `json:"avatar"`
	Bot           *bool   `json:"bot,omitempty"`
	System        *bool   `json:"system,omitempty"`
	MfaEnabled    *bool   `json:"mfa_enabled,omitempty"`
	Banner        *string `json:"banner,omitempty"`
	AccentColor   *int    `json:"accent_color,omitempty"`
	Locale        *string `json:"locale,omitempty"`
	Verified      *bool   `json:"verified,omitempty"`
	Email         *string `json:"email,omitempty"`
	Flags         *int    `json:"flags,omitempty"`
	PremiumType   *int    `json:"premium_type,omitempty"`
	PublicFlags   *int    `json:"public_flags,omitempty"`
}

type Guild struct {
	ID                          string         `json:"id"`
	Name                        string         `json:"name"`
	Icon                        *string        `json:"icon"`
	IconHash                    *string        `json:"icon_hash,omitempty"`
	Splash                      *string        `json:"splash"`
	DiscoverySplash             *string        `json:"discovery_splash"`
	Owner                       *bool          `json:"owner,omitempty"`
	OwnerID                     string         `json:"owner_id"`
	Permissions                 *string        `json:"permissions,omitempty"`
	Region                      *string        `json:"region,omitempty"`
	AfkChannelID                *string        `json:"afk_channel_id"`
	AfkTimeout                  int            `json:"afk_timeout"`
	WidgetEnabled               *bool          `json:"widget_enabled,omitempty"`
	WidgetChannelID             *string        `json:"widget_channel_id"`
	VerificationLevel           int            `json:"verification_level"`
	DefaultMessageNotifications int            `json:"default_message_notifications"`
	ExplicitContentFilter       int            `json:"explicit_content_filter"`
	Members                     *[]Member      `json:"members,omitempty"`
	Channels                    *[]Channel     `json:"channels,omitempty"`
	Roles                       []Role         `json:"roles"`
	Emojis                      []Emoji        `json:"emoji"`
	Features                    []GuildFeature `json:"features"`
	MFALevel                    int            `json:"mfa_level"`
	ApplicationID               *string        `json:"application_id"`
	SystemChannelID             *string        `json:"system_channel_id"`
	RulesChannelID              *string        `json:"rules_channel_id"`
	MaxPresences                *int           `json:"max_presences"`
	MaxMembers                  *int           `json:"max_members"`
	VanityURLCode               *string        `json:"vanity_url_code"`
	Description                 string         `json:"description"`
	Banner                      *string        `json:"banner"`
	PremiumTier                 int            `json:"premium_tier"`
	PremiumSubscriptionCount    *int           `json:"premium_subscription_count,omitempty"`
	PreferredLocale             string         `json:"preferred_locale"`
	PublicUpdatesChannelID      *string        `json:"public_updates_channel_id"`
	MaxVideoChannelUsers        *int           `json:"max_video_channel_users,omitempty"`
	ApproximateMemberCount      *int           `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount    *int           `json:"approximate_presence_count,omitempty"`
	WelcomeScreen               *WelcomeScreen `json:"welcome_screen,omitempty"`
	NSFWLevel                   int            `json:"nsfw_level"`
	Stickers                    *[]Sticker     `json:"stickers,omitempty"`
	PremiumProgressBarEnabled   bool           `json:"premium_progress_bar_enabled"`
	VoiceStates                 *[]VoiceState  `json:"voice_states,omitempty"`
}

type Channel struct {
	ID        string  `json:"id"`
	Type      int     `json:"type"`
	GuildID   *string `json:"guild_id,omitempty"`
	Position  *int    `json:"position,omitempty"`
	Name      *string `json:"name,omitempty"`
	Bitrate   *int    `json:"bitrate,omitempty"`
	RTCRegion *string `json:"rtc_region,omitempty"`
}

type WelcomeScreen struct {
	Description     *string                `json:"description"`
	WelcomeChannels []WelcomeScreenChannel `json:"welcome_channels"`
}

type WelcomeScreenChannel struct {
	ChannelID   string  `json:"channel_id"`
	Description string  `json:"description"`
	EmojiID     *string `json:"emoji_id"`
	EmojiName   *string `json:"emoji_name"`
}

type Emoji struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Roles         *[]Role `json:"roles,omitempty"`
	User          *User   `json:"user,omitempty"`
	RequireColons *bool   `json:"require_colons,omitempty"`
	Managed       *bool   `json:"managed,omitempty"`
	Animated      *bool   `json:"aminated,omitempty"`
	Available     *bool   `json:"available,omitempty"`
}

type Sticker struct {
	ID          string  `json:"id"`
	PackID      *string `json:"pack_id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Tags        string  `json:"tags"`
	Asset       *string `json:"asset,omitempty"`
	Type        int     `json:"type"`
	FormatType  int     `json:"format_type"`
	Available   *bool   `json:"available,omitempty"`
	GuildID     *string `json:"guild_id,omitempty"`
	User        *User   `json:"user,omitempty"`
	SortValue   *int    `json:"sort_value,omitempty"`
}

type Role struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Color        int       `json:"color"`
	Hoist        bool      `json:"hoist"`
	Icon         *string   `json:"icon,omitempty"`
	UnicodeEmoji *string   `json:"unicode_emoji,omitempty"`
	Position     int       `json:"position"`
	Permissions  string    `json:"permissions"`
	Managed      bool      `json:"managed"`
	Mentionable  bool      `json:"mentionable"`
	Tags         *RoleTags `json:"tags,omitempty"`
}

type RoleTags struct {
	BotID                 *string `json:"bot_id,omitempty"`
	IntegrationID         *string `json:"integration_id,omitempty"`
	PremiumSubscriber     any     `json:"premium_subscriber,omitempty"`
	SubscriptionListingID *string `json:"subscription_listing_id,omitempty"`
	AvailableForPurchase  *bool   `json:"available_for_purchase,omitempty"`
	GuildConnections      any     `json:"guild_connections,omitempty"`
}
