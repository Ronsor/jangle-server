package main

import (
	"time"
	"regexp"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

var (
	MSGFMT_USER = regexp.MustCompile(`<@([0-9]+)>`)
	MSGFMT_USER_NICK = regexp.MustCompile(`<@!([0-9]+)>`)
	MSGFMT_CHANNEL = regexp.MustCompile(`<@#([0-9]+)>`)
	MSGFMT_ROLE = regexp.MustCompile(`<@&([0-9]+)>`)
)

const (
	MSGTYPE_DEFAULT           = 0
	MSGTYPE_GUILD_MEMBER_JOIN = 7
	// TODO: the rest
	MSGTYPE_PATCH_REACT = 0x42424201 // wrong on so many levels
)

func reExtSnowflakes(r *regexp.Regexp, text string) []snowflake.ID {
	out := []snowflake.ID{}
	s := r.FindAllStringSubmatch(text, -1)
	for _, v := range s {
		snow, _ := snowflake.ParseString(v[1])
		out = append(out, snow)
	}
	return out
}

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
	ID        snowflake.ID `bson:"_id"`
	ChannelID snowflake.ID `bson:"channel_id"`
	GuildID   snowflake.ID `bson:"guild_id"`

	Author *User `bson:"author"`
	//Member *GuildMember `bson:"member"`

	Content         string `bson:"content"`
	Timestamp       int64  `bson:"timestamp"`
	EditedTimestamp int64  `bson:"edited_timestamp"`

	TTS bool `bson:"tts"`

	MentionEveryone bool           `bson:"mention_everyone"`
	Mentions        []snowflake.ID       `bson:"mentions"`
	MentionRoles    []snowflake.ID `bson:"mention_roles"`
	MentionChannels []snowflake.ID `bson:"mention_channels"`

	Attachments []interface{}   `bson:"attachments"`
	Embeds      []*MessageEmbed `bson:"embeds"`
	Reactions   interface{}     `bson:"reactions"`

	Nonce     string       `bson:"nonce"`
	Pinned    bool         `bson:"pinned"`
	WebhookID snowflake.ID `bson:"webhook_id"`

	Type  int `bson:"type"`
	Flags int `bson:"flags"`

	MiscData interface{} `bson:"misc_data"`
	Deleted  bool        `bson:"deleted"`
}

func GetMessageByID(i snowflake.ID) (*Message, error) {
	var m Message
	c := DB.Msg.C("msgs")
	err := c.Find(bson.M{"_id": i}).One(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Message) Member() (*GuildMember, error) {
	gd, err := GetGuildByID(m.GuildID)
	if err != nil {
		return nil, err
	}
	return gd.GetMember(m.Author.ID)
}

func (m *Message) Save() error {
	c := DB.Msg.C("msgs")
	return c.UpdateId(m.ID, bson.M{"$set": m})
}

func (m *Message) ParseContent(usr *User, ch *Channel) error {
	if m.MentionEveryone && ch.IsGuild() {
		g, err := ch.Guild()
		if err != nil { return err }
		if g.GetPermissions(usr).Has(PERM_MENTION_EVERYONE) && strings.Contains(m.Content, "@everyone") {
			m.MentionEveryone = true
		}
	}
	m.MentionRoles = reExtSnowflakes(MSGFMT_ROLE, m.Content)
	m.Mentions = append(reExtSnowflakes(MSGFMT_USER, m.Content), reExtSnowflakes(MSGFMT_USER_NICK, m.Content)...)
	m.MentionChannels = reExtSnowflakes(MSGFMT_CHANNEL, m.Content)
	return nil
}

func (m *Message) ToAPI() (ret *APITypeMessage) {
	ret = &APITypeMessage{
		ID:              m.ID,
		ChannelID:       m.ChannelID,
		GuildID:         m.GuildID,
		Content:         m.Content,
		Timestamp:       time.Unix(m.Timestamp, 0),
		TTS:             m.TTS,
		MentionEveryone: m.MentionEveryone,
		MentionRoles:    m.MentionRoles,
		Attachments:     m.Attachments,
		Nonce:           m.Nonce,
		Pinned:          m.Pinned,
		Type:            m.Type,
		Flags:           m.Flags,
	}

	ret.Embeds = append([]*MessageEmbed{}, m.Embeds...)

	if mem, err := m.Member(); err == nil {
		ret.Member = mem.ToAPI()
		ret.Member.User = nil
	}

	author, err := GetUserByID(m.Author.ID)
	if err == nil {
		ret.Author = author.ToAPI(true)
	} else {
		ret.Author = m.Author.ToAPI(true)
	}
	ret.Attachments = []interface{}{}
	ret.Mentions = []*APITypeUser{}
	if m.GuildID != 0 {
		g, err := GetGuildByID(m.GuildID)
		if err == nil {
			for _, v := range m.Mentions {
				mem, err := g.GetMember(v)
				if err == nil {
					memapi := mem.ToAPI()
					usrapi := memapi.User
					usrapi.Member = memapi
					memapi.User = nil
					ret.Mentions = append(ret.Mentions, usrapi)
				}
			}
		}
	}
	ret.MentionChannels = []*APITypeChannelMention{}
	for _, v := range m.MentionChannels {
		ch, err := GetChannelByID(v)
		if err == nil {
			ret.MentionChannels = append(ret.MentionChannels, &APITypeChannelMention{ch.ID, ch.GuildID, ch.Type, ch.Name})
		}
	}

	ret.Reactions = []*APITypeMessageReaction{}

	if m.EditedTimestamp != 0 {
		ts := time.Unix(m.EditedTimestamp, 0)
		ret.EditedTimestamp = &ts
	}

	return
}
