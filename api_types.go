package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// API codes

const (
	APIERR_UNKNOWN_USER = 10013
	APIERR_UNAUTHORIZED = 40001
)

// API call error
type APIResponseError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type responseError APIResponseError // TODO: get rid of this

// "Safe" User type to be returned for API calls
type APITypeUser struct {
	ID snowflake.ID `json:"id,string"`
	Username string `json:"username"`
	Discriminator int `json:"discriminator,string"`
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
	Phone *string `json:"phone"`
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
	Recipients []snowflake.ID `json:"recipients"`
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
	EditedTimestamp time.Time `json:"edited_timestamp,omitempty"`

	TTS bool `json:"tts"`

	MentionEveryone bool `json:"mention_everyone"`
	Mentions []*APITypeUser `json:"mentions"`
	MentionRoles []snowflake.ID `json:"mention_roles"`
	MentionChannels []interface{} `json:"mention_channels"`

	Attachments []interface{} `json:"attachments"`
	Embeds []interface{} `json:"embeds"`
	Reactions []interface{} `json:"reactions,omitempty"`

	Nonce string `json:"nonce,omitempty"`
	Pinned bool `json:"pinned"`
	WebhookID snowflake.ID `json:"webhook_id,omitempty"`

	Type int `json:"type"`
	Flags int `json:"flags,omitempty"`
}
