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

type APIChannel Channel

func GetChannelByID(ID snowflake.ID) (*Channel, error) {
	var c2 Channel
	c := DB.Core.C("channels")
	e := c.Find(bson.M{"_id": ID}).One(&c2)
	if e != nil {
		return nil, e
	}
	return &c2, nil
}

// TODO: GetChannelByGuild, GetChannelByRecipients, etc.

func (c *Channel) MarshalAPI() *APIChannel {
	c2 := APIChannel(*c)
	if c2.Type == CHTYPE_GUILD_TEXT {
		c2._nsfw = &c.NSFW
		c2._position = &c.Position
	}
	return &c2
}

func (c *Channel) UnmarshalAPI(a *APIChannel) *Channel {
	panic("Unimplemented")
	_ = a
	return c
}
