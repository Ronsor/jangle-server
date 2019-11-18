package main

import (
	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

// Channel types
const (
	CHTYPE_GUILD_TEXT = 0
	CHTYPE_DM = 1
	// TODO: the rest
)

// Channel is a Discord-compatible structure representing any type of channel
type Channel struct {
	ID snowflake.ID `json:"id" bson:"_id"`
	Type int `json:"type"`

	// Text Channel only
	LastMessageID snowflake.ID `json:"last_message_id"`

	// DM/Group DM only
	Recipients []snowflake.ID `json:"recipients,omitempty"`

	// Group DM only
	OwnerID snowflake.ID `json:"owner_id,omitempty"`
	Icon string `json:"icon,omitempty"`

	// Guild only
	GuildID snowflake.ID `json:"guild_id,omitempty"`
	Position int `json:"-" bson:"position"`
	Name string `json:"name,omitempty"`
	ParentID snowflake.ID `json:"parent_id,omitempty"`
	PermissionOverwrites []interface{} `json:"permission_overwrites,omitempty"`

	// Guild Text Channel only
	Topic string `json:"topic,omitempty"`
	NSFW bool `json:"-" bson:"nsfw"`
	RateLimitPerUser int `json:"rate_limit_per_user,omitempty"`

	// Guild Voice Channel only
	Bitrate int `json:"bitrate,omitempty"`
	UserLimit int `json:"user_limit,omitempty"`

	// """COMPATIBILITY"""
	_position *int `json:"position,omitempty" bson:"-"`
	_nsfw *bool `json:"nsfw,omitempty" bson:"-"`
}

func GetChannelByID(i snowflake.ID) (*Channel, error) {
	var x Channel
	c := DB.Core.C("channels")
	err := c.Find(bson.M{"_id": i}).One(&x)
	return &x, err
}

func (c *Channel) MarshalAPI() *Channel {
	c2 := *c
	if c.Type == CHTYPE_GUILD_TEXT {
		c2._position = &c.Position
		c2._nsfw = &c.NSFW
	}
	return &c2
}

