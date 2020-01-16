package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
)

// REST API call error
type APIResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// API errors
var (
	APIERR_BAD_REQUEST = &APIResponseError{0, "Unknown malformed request"}
	APIERR_UNKNOWN_CHANNEL     = &APIResponseError{10003, "Unknown channel"}
	APIERR_UNKNOWN_GUILD       = &APIResponseError{10004, "Unknown guild"}
	APIERR_UNKNOWN_MEMBER      = &APIResponseError{10007, "Unknown guild member"}
	APIERR_UNKNOWN_MESSAGE     = &APIResponseError{10008, "Unknown message"}
	APIERR_UNKNOWN_ROLE 	   = &APIResponseError{10011, "Unknown guild role"}
	APIERR_UNKNOWN_USER        = &APIResponseError{10013, "Unknown user"}
	APIERR_UNAUTHORIZED        = &APIResponseError{40001, "Unauthoized"}
	APIERR_MAX_GUILDS          = &APIResponseError{30001, "Maximum number of guilds reached"}
	APIERR_MAX_FRIENDS         = &APIResponseError{30002, "Maximum number of friends reached"}
	APIERR_MAX_PINS            = &APIResponseError{30003, "Maximum number of message pins reached"}
	APIERR_MAX_GUILD_ROLES     = &APIResponseError{30005, "Maximum number of guild roles reached"}
	APIERR_MAX_REACTIONS       = &APIResponseError{30010, "Maximum number of message reactions reached"}
	APIERR_MAX_GUILD_CHANNELS  = &APIResponseError{30013, "Maximum number of guild channels reached"}
	APIERR_MAX_INVITES         = &APIResponseError{30016, "Maximum number of guild invites reached"}
	APIERR_MISSING_ACCESS      = &APIResponseError{50001, "Access denied"}
	APIERR_EMPTY_MESSAGE       = &APIResponseError{50006, "Message is empty"}
	APIERR_CANT_EDIT_MESSAGE   = &APIResponseError{50005, "Can't edit message"}
	APIERR_MISSING_PERMISSIONS = &APIResponseError{50013, "Missing required permissions"}
	// TODO fill in the rest of the magic numbers
)

func (a *APIResponseError) Error() string {
	return fmt.Sprintf("Error code %d: %s", a.Code, a.Message)
}

type responseError APIResponseError // TODO: get rid of this

// "Safe" User type to be returned for API calls
type APITypeUser struct {
	ID            snowflake.ID `json:"id,string"`
	Username      string       `json:"username"`
	Discriminator string       `json:"discriminator"`
	AvatarHash    string       `json:"avatar"`

	Bot        bool   `json:"bot,omitempty"`
	MfaEnabled bool   `json:"mfa_enabled,omitempty"`
	Locale     string `json:"locale,omitempty"`
	Verified   *bool  `json:"verified,omitempty"`
	Email      string `json:"email,omitempty"`

	Flags       int `json:"flags,omitempty"`
	PremiumType int `json:"premium_type,omitempty"`

	Premium bool    `json:"premium"`
	Mobile  bool    `json:"mobile"`
	Phone   *string `json:"phone,omitempty"`
}

// "Safe" Channel type that represents any channel
type APITypeAnyChannel interface {
	// There's nothing here
}

// "Safe" Permission Overwrite type
type APITypePermissionOverwrite PermissionOverwrite

// "Safe" DM Channel type
type APITypeDMChannel struct {
	ID            snowflake.ID   `json:"id,string"`
	Type          int            `json:"type"`
	LastMessageID snowflake.ID   `json:"last_message_id,string"`
	Recipients    []*APITypeUser `json:"recipients"`
}

// "Safe" Guild Text Channel type
type APITypeGuildTextChannel struct {
	ID            snowflake.ID `json:"id,string"`
	GuildID       snowflake.ID `json:"guild_id,string"`
	Type          int          `json:"type"`
	LastMessageID snowflake.ID `json:"last_message_id,string"`

	Name     string `json:"name"`
	Topic    string `json:"topic"`
	NSFW     bool   `json:"nsfw"`
	Position int    `json:"position"`

	PermissionOverwrites []*APITypePermissionOverwrite `json:"permission_overwrites"`
	RateLimitPerUser     int                           `json:"rate_limit_per_user,omitempty"`
	LastPinTimestamp     time.Time                     `json:"last_pin_timestamp"`
}

// "Safe" MessageReaction type
type APITypeMessageReaction struct {
	Emoji *APITypeEmoji `json:"emoji"`
	Count int           `json:"count"`
	Me    bool          `json:"me"`
	user  *User         `json:"-"`
}

// "Safe" Message type
// Good grief Discord that's a lot of fields
type APITypeMessage struct {
	ID        snowflake.ID `json:"id,string"`
	ChannelID snowflake.ID `json:"channel_id,string"`
	GuildID   snowflake.ID `json:"guild_id,string,omitempty"`

	Author *APITypeUser        `json:"author"`
	Member *APITypeGuildMember `json:"member,omitempty"`

	Content         string     `json:"content"`
	Timestamp       time.Time  `json:"timestamp"`
	EditedTimestamp *time.Time `json:"edited_timestamp,omitempty"`

	TTS bool `json:"tts"`

	MentionEveryone bool           `json:"mention_everyone"`
	Mentions        []*APITypeUser `json:"mentions"`
	MentionRoles    []snowflake.ID `json:"mention_roles"`
	MentionChannels []interface{}  `json:"mention_channels"`

	Attachments []interface{}             `json:"attachments"`
	Embeds      []*MessageEmbed           `json:"embeds"`
	Reactions   []*APITypeMessageReaction `json:"reactions,omitempty"`

	Nonce     string       `json:"nonce,omitempty"`
	Pinned    bool         `json:"pinned"`
	WebhookID snowflake.ID `json:"webhook_id,omitempty"`

	Type  int `json:"type"`
	Flags int `json:"flags,omitempty"`
}

// "Safe" Emoji type
// The amount of pointer ~~types~~fields in this struct is awful
type APITypeEmoji struct {
	ID            *snowflake.ID  `json:"id,string"`      // Can be null
	Name          *string        `json:"name,omitempty"` // Can be null
	Roles         []snowflake.ID `json:"roles,omitempty"`
	User          *APITypeUser   `json:"user,omitempty"`
	RequireColons *bool          `json:"require_colons,omitempty"`
	Managed       *bool          `json:"managed,omitempty"`
	Animated      *bool          `json:"animated,omitempty"`
}

// "Safe" GuildMember type
type APITypeGuildMember struct {
	User     *APITypeUser   `json:"user,omitempty"`
	Nick     string         `json:"nick,omitempty"`
	Roles    []snowflake.ID `json:"roles"`
	JoinedAt time.Time      `json:"joined_at"`
	Deaf     bool           `json:"deaf"`
	Mute     bool           `json:"mute"`
	GuildID  snowflake.ID `json:"guild_id,string,omitempty"`
}

// "Safe" Role type
type APITypeRole struct {
	ID          snowflake.ID `json:"id,string"`
	Name        string       `json:"name"`
	Color       int          `json:"color"`
	Hoist       bool         `json:"hoist"`
	Position    int          `json:"position"`
	Permissions PermSet      `json:"permission"`
	Managed     bool         `json:"managed"`
	Mentionable bool         `json:"mentionable"`
}

// "Safe" Guild type
type APITypeGuild struct {
	ID     snowflake.ID `json:"id,string"`
	Name   string       `json:"name"`
	Icon   string       `json:"icon,omitempty"`
	Splash string       `json:"splash,omitempty"`

	Owner   bool         `json:"owner,omitempty"`
	OwnerID snowflake.ID `json:"owner_id,string"`

	Permissions PermSet `json:"permissions"`
	Region      string  `json:"region"`

	EmbedEnabled   bool         `json:"embed_enabled,omitempty"`
	EmbedChannelID snowflake.ID `json:"embed_channel_id,omitempty"`

	VerificationLevel           int `json:"verification_level"`
	DefaultMessageNotifications int `json:"default_message_notifications"`
	ExplicitContentFilter       int `json:"explicit_content_filter"`

	Roles  []*APITypeRole  `json:"roles"`
	Emojis []*APITypeEmoji `json:"emojis"`

	Features      []string     `json:"features"`
	MfaLevel      int          `json:"mfa_level"`
	ApplicationID snowflake.ID `json:"application_id,string,omitempty"`

	WidgetEnabled   bool         `json:"widget_enabled,omitempty"`
	WidgetChannelID snowflake.ID `json:"widget_channel_id,omitempty"`

	SystemChannelID snowflake.ID `json:"system_channel_id,omitempty"`
	JoinedAt        time.Time    `json:"joined_at,omitempty"`
	Large           bool         `json:"large,omitempty"`
	Unavailable     bool         `json:"unavailable,omitempty"`

	MemberCount  int                      `json:"member_count,omitempty"`
	Members      []*APITypeGuildMember    `json:"members,omitempty"`
	Channels     []APITypeAnyChannel      `json:"channels,omitempty"`
	Presences    []*APITypePresenceUpdate `json:"presences,omitempty"`
	MaxPresences int                      `json:"max_presences,omitempty"` // Should be ~5k (maybe 10k)?
	VoiceStates  []interface{}            `json:"voice_states"`

	Description string `json:"description,omitempty"`
	Banner      string `json:"banner,omitempty"`

	PremiumTier     int    `json:"premium_tier"`
	PreferredLocale string `json:"preferred_locale"`
}

// "Safe" Presence Update type
type APITypePresenceUpdate gwEvtDataPresenceUpdate
