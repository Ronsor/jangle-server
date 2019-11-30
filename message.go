package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

const (
	MSGTYPE_DEFAULT = 0
	// TODO: the rest
	MSGTYPE_PATCH_REACT = 0x42424201 // wrong on so many levels
)

type MessageEmbed struct {
	// There will be something here
}

/*
type MessageReaction struct {
	Emoji *Emoji `bson:"emoji"`
	Users []*User `bson:"users"`
}

type MessageReactions []*MessageReaction

func (mr MessageReactions) ToAPI(curuser *User) (ret []*APITypeMessageReaction) {
	ret = make([]*APITypeMessageReaction, 0, len(mr))
	for k, v := range mr {
		ret[k] = &APITypeMessageReaction{
			Emoji: v.Emoji,
			Count: len(v.Users),
		}
		if curuser != nil {
			for _, v2 := range v.Users {
				if v2.ID == curuser.ID {
					ret[k].Me = true
				}
			}
		}
	}
	return
}

func (mr MessageReactions) React(curuser *User, e *Emoji, addnew bool) error {
	for _, v := range mr {
		if v.Emoji.String() == e.String() {
			for _, u := range v.Users {
				if u.ID == curuser.ID { return nil }
			}
			v.Users = append(v.Users, curuser)
			return nil
		}
	}
	if addnew {
		mr = append(mr, &MessageReaction{e, Users: []*User{&User{ID:curuser.ID}}})
	}
}*/

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
	Reactions interface{} `bson:"reactions"`

	Nonce string `bson:"nonce"`
	Pinned bool `bson:"pinned"`
	WebhookID snowflake.ID `bson:"webhook_id"`

	Type int `bson:"type"`
	Flags int `bson:"flags"`

	MiscData interface{} `bson:"misc_data"`
	Deleted bool `bson:"deleted"`
}

func GetMessageByID(i snowflake.ID) (*Message, error) {
	var m Message
	c := DB.Msg.C("msgs")
	err := c.Find(bson.M{"_id":i}).One(&m)
	if err != nil { return nil, err }
	return &m, nil
}

func (m *Message) Save() error {
	c := DB.Msg.C("msgs")
	return c.UpdateId(m.ID, bson.M{"$set":m})
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

	author, err := GetUserByID(m.Author.ID)
	if err == nil {
		ret.Author = author.ToAPI(true)
	} else {
		ret.Author = m.Author.ToAPI(true)
	}
	ret.Mentions = []*APITypeUser{}
	ret.Reactions = []*APITypeMessageReaction{}
	for _, v := range m.Mentions {
		ret.Mentions = append(ret.Mentions, v.ToAPI(true))
	}

	return
}
