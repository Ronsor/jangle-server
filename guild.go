package main

import (
	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

// Represents a currently unavailable guild
// Used in GW_EVT_READY
type UnavailableGuild struct {
	ID snowflake.ID `json:"id"`
	Unavailable bool `json:"unavailable"`
}

type Guild struct {
	ID snowflake.ID `bson:"_id"`
	Name string `bson:"name"`
	Icon string `bson:"icon"`
	Splash string `bson:"splash"`

	OwnerID snowflake.ID `bson:"owner_id"`
	Region string `bson:"region"`
	/* AFK Channels are deprecated and will never be supported */

	EmbedEnabled bool `bson:"embed_enabled"`
	EmbedChannelID snowflake.ID `bson:"embed_channel_id"

	VerificationLevel int `bson:"verification_level"`
	DefaultMessageNotifications int `bson:"default_message_notifications"
	ExplicitContentFilter int `bson:"explicit_content_filter"`

	Roles []*Role `bson:"roles"`
	Emojis []*Emoji `bson:"emojis"`
	Features []string `bson:"features"`
	MFALevel int `bson:"mfa_level"`
	ApplicationID snowflake.ID `bson:"application_id"`

	WidgetEnabled bool `bson:"widget_enabled"`
	WidgetChannelID snowflake.ID `bson:"widget_channel_id"`
	SystemChannelID snowflake.ID `bson:"system_channel_id"`

	Large bool `bson:"large"`

	Description string `bson:"description"`
	Banner string `bson:"banner"`
	PremiumTier int `bson:"premium_tier"`
	PreferredLocale string `bson:"preferred_locale"`
}

func (g *Guild) Members() {
