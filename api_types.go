package main

import (
	"time"
	"fmt"

	"github.com/bwmarrin/snowflake"
)

// API codes

const (
	APIERR_UNKNOWN_USER = 10013
	APIERR_UNKNOWN_CHANNEL = 10003
	APIERR_UNKNOWN_MEMBER = 10007
	APIERR_UNKNOWN_MESSAGE = 10008
	APIERR_UNAUTHORIZED = 40001
	APIERR_CANT_EDIT_MESSAGE = 50005
	APIERR_MISSING_PERMISSIONS = 50013
	// TODO fill in the rest of the magic numbers
)

// API call error
type APIResponseError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func (a *APIResponseError) Error() string {
	return fmt.Sprintf("Error code %d: %s", a.Code, a.Message)
}

type responseError APIResponseError // TODO: get rid of this

// "Safe" User type to be returned for API calls
type APITypeUser struct {
	ID snowflake.ID `json:"id,string"`
	Username string `json:"username"`
	Discriminator string `json:"discriminator"`
	AvatarHash string `json:"avatar"`

	Bot bool `json:"bot,omitempty"`
	MfaEnabled bool `json:"mfa_enabled,omitempty"`
	Locale string `json:"locale,omitempty"`
	Verified *bool `json:"verified,omitempty"`
	Email string `json:"email,omitempty"`

	Flags int `json:"flags,omitempty"`
	PremiumType int `json:"premium_type,omitempty"`

	Premium bool `json:"premium"`
	Mobile bool `json:"mobile"`
	Phone *string `json:"phone,omitempty"`
}

// "Safe" Channel type that represents any channel
type APITypeAnyChannel interface {
	// There's nothing here
	// Should there be?
}

// "Safe" DM Channel type
type APITypeDMChannel struct {
	ID snowflake.ID `json:"id,string"`
	Type int `json:"type"`
	LastMessageID snowflake.ID `json:"last_message_id,string"`
	Recipients []*APITypeUser `json:"recipients"`
}

// "Safe" Guild Text Channel type
type APITypeGuildTextChannel struct {
	ID snowflake.ID `json:"id,string"`
	GuildID snowflake.ID `json:"guild_id,string"`
	Type int `json:"type"`
	LastMessageID snowflake.ID `json:"last_message_id,string"`

	Name string `json:"name"`
	Topic string `json:"topic"`
	NSFW bool `json:"nsfw"`
	Position int `json:"position"`

	PermissionOverwrites []interface{} `json:"permission_overwrites"`
	RateLimitPerUser int `json:"rate_limit_per_user,omitempty"`
	LastPinTimestamp time.Time `json:"last_pin_timestamp"`
}

// "Safe" MessageReaction type
type APITypeMessageReaction struct {
	Emoji *APITypeEmoji `json:"emoji"`
	Count int `json:"count"`
	Me bool `json:"me"`
	user *User `json:"-"`
}

// "Safe" Message type
// Good grief Discord that's a lot of fields
type APITypeMessage struct {
	ID snowflake.ID `json:"id,string"`
	ChannelID snowflake.ID `json:"channel_id,string"`
	GuildID snowflake.ID `json:"guild_id,string,omitempty"`

	Author *APITypeUser `json:"author"`
	Member interface{} `json:"member,omitempty"`

	Content string `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	EditedTimestamp *time.Time `json:"edited_timestamp,omitempty"`

	TTS bool `json:"tts"`

	MentionEveryone bool `json:"mention_everyone"`
	Mentions []*APITypeUser `json:"mentions"`
	MentionRoles []snowflake.ID `json:"mention_roles"`
	MentionChannels []interface{} `json:"mention_channels"`

	Attachments []interface{} `json:"attachments"`
	Embeds []*MessageEmbed `json:"embeds"`
	Reactions []*APITypeMessageReaction `json:"reactions,omitempty"`

	Nonce string `json:"nonce,omitempty"`
	Pinned bool `json:"pinned"`
	WebhookID snowflake.ID `json:"webhook_id,omitempty"`

	Type int `json:"type"`
	Flags int `json:"flags,omitempty"`
}

// "Safe" Emoji type
// The amount of pointer ~~types~~fields in this struct is awful
type APITypeEmoji struct {
	ID *snowflake.ID `json:"id,string"` // Can be null
	Name *string `json:"name,omitempty"` // Can be null
	Roles []snowflake.ID `json:"roles,omitempty"`
	User *APITypeUser `json:"user,omitempty"`
	RequireColons *bool `json:"require_colons,omitempty"`
	Managed *bool `json:"managed,omitempty"`
	Animated *bool `json:"animated,omitempty"`
}

// "Safe" GuildMember type
type APITypeGuildMember struct {
	User *APITypeUser `json:"user"`
	Nick string `json:"nick,omitempty"`
	Roles []snowflake.ID `json:"roles,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
	Deaf bool `json:"deaf"`
	Mute bool `json:"mute"`
}

type APITypeRole struct{}

// "Safe" Guild type
type APITypeGuild struct {
	ID snowflake.ID `json:"id,string"`
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
	Splash string `json:"splash,omitempty"`

	Owner bool `json:"bool,omitempty"`
	OwnerID snowflake.ID `json:"owner_id,string"`

	Permissions PermSet `json:"permissions"`
	Region string `json:"region"`

	EmbedEnabled bool `json:"embed_enabled,omitempty"`
	EmbedChannelID snowflake.ID `json:"embed_channel_id,omitempty"`

	VerificationLevel int `json:"verification_level"`
	DefaultMessageNotifications int `json:"default_message_notifications"`
	ExplicitContentFilter int `json:"explicit_content_filter"`

	Roles []*APITypeRole `json:"roles"`
	Emojis []*APITypeEmoji `json:"emojis"`

	Features []string `json:"features"`
	MfaLevel int `json:"mfa_level"`
	ApplicationID snowflake.ID `json:"application_id,string,omitempty"`

	WidgetEnabled bool `json:"widget_enabled,omitempty"`
	WidgetChannelID snowflake.ID `json:"widget_channel_id,omitempty"`

	SystemChannelID snowflake.ID `json:"system_channel_id,omitempty"`
	JoinedAt time.Time `json:"joined_at,omitempty"`
	Large bool `json:"large,omitempty"`
	Unavailable bool `json:"unavailable,omitempty"`

	MemberCount int `json:"member_count,omitempty"`
	Members []*APITypeGuildMember `json:"members"`
	Channels []APITypeAnyChannel `json:"channels"`
	Presences []interface{} `json:"presences"`
	MaxPresences int `json:"max_presences,omitempty"` // Should be ~5k (maybe 10k)?

	Description string `json:"description,omitempty"`
	Banner string `json:"banner,omitempty"`

	PremiumTier int `json:"premium_tier"`
	PreferredLocale string `json:"preferred_locale"`
}
