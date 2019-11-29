package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

const (
	MSGTYPE_DEFAULT = 0
	MSGTYPE_DUMMY = 0x42424242
	// TODO: the rest
)

type MessageEmbed struct {

}

type MessageReaction struct {
	Count int `bson:"count"`
	Emoji Emoji `bson:"emoji"`
}

type Message struct {
	ID snowflake.ID `bson:"_id"`
	ChannelID snowflake.ID `bson:"channel_id"`
	GuildID snowflake.ID `bson:"guild_id"`

	Author *User `bson:"author"`
	Member interface{} `bson:"member"`

	Content string `bson:"content"`
	Timestamp int64 `bson:"timestamp"`
	EditedTimestamp int64 `bson:"edited_timestamp"`

	TTS bool `bson:"tts"`

	MentionEveryone bool `bson:"mention_everyone"`
	Mentions []*User `bson:"mentions"`
	MentionRoles []snowflake.ID `bson:"mention_roles"`
	MentionChannels []interface{} `bson:"mention_channels"`

	Attachments []interface{} `bson:"attachments"`
	Embeds []*MessageEmbed `bson:"embeds"`
	Reactions []*MessageReaction `bson:"reactions"`

	Nonce string `bson:"nonce"`
	Pinned bool `bson:"pinned"`
	WebhookID snowflake.ID `bson:"webhook_id"`

	Type int `bson:"type"`
	Flags int `bson:"flags"`

	MiscData interface{} `bson:"misc_data"`
}

func (m *Message) ToAPI() (ret *APITypeMessage) {
	ret = &APITypeMessage{
		ID: m.ID,
		ChannelID: m.ChannelID,
		GuildID: m.GuildID,
		Member: m.Member,
		Content: m.Content,
		Timestamp: time.Unix(m.Timestamp, 0),
		EditedTimestamp: time.Unix(m.EditedTimestamp, 0),
		TTS: m.TTS,
		MentionEveryone: m.MentionEveryone,
		MentionRoles: m.MentionRoles,
		MentionChannels: m.MentionChannels,
		Attachments: m.Attachments,
		Embeds: m.Embeds,
		Nonce: m.Nonce,
		Pinned: m.Pinned,
		Type: m.Type,
		Flags: m.Flags,
	}

	ret.Author = m.Author.ToAPI(true)
	ret.Mentions = []*APITypeUser{}
	for _, v := range m.Mentions {
		ret.Mentions = append(ret.Mentions, v.ToAPI(true))
	}

	return
}
