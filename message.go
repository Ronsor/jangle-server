package main

import (
	"github.com/bwmarrin/snowflake"
)

type Message struct {
	ID snowflake.ID `json:"id" bson:"_id"`
	ChannelID snowflake.ID `json:"channel_id"`
	GuildID snowflake.ID `json:"guild_id,omitempty"`

	Author *User `json:"user"`
	MemberRef
	Member interface{} `json:"member,omitempty"
}
