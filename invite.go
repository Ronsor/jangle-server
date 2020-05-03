package main

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/bwmarrin/snowflake"
)

type Invite struct {
	ID snowflake.ID `bson:"_id"`
	GuildID snowflake.ID `bson:"guild_id"`
	ChannelID snowflake.ID `bson:"channel_id"`
	Inviter snowflake.ID `bson:"inviter"`
	CreatedAt int64 `bson:"created_at"`

	MaxUses int `bson:"max_uses"`
	Uses int `bson:"uses"`
	MaxAge *time.Time `bson:"max_age,omitempty"`
}

func GetInviteByID(id snowflake.ID) (*Invite, error) {
	var i Invite
	c := DB.Core.C("guildinvites")
	e := c.Find(bson.M{"_id": id}).One(&i)
	if e != nil {
		return nil, e
	}
	return &i, nil
}

func GetInvitesByGuild(guildID snowflake.ID) ([]*Invite, error) {
	i := []*Invite{}
	c := DB.Core.C("guildinvites")
	e := c.Find(bson.M{"guild_id": guildID}).All(&i)
	if e != nil {
		return i, e
	}
	return i, nil
}

func (i *Invite) Delete() error {
	c := DB.Core.C("guildinvites")
	return c.RemoveId(i.ID)
}

func (i *Invite) IncrUses() error {
	c := DB.Core.C("guildinvites")
	e := c.UpdateId(i.ID, bson.M{"$inc":bson.M{"uses":1}})
	if e != nil { return e }
	i.Uses++
	return nil
}

func (i *Invite) ToAPI() *APITypeInvite {
	out := &APITypeInvite{
		Code: i.ID.Base32(),
		CreatedAt: time.Unix(i.CreatedAt, 0),
		Uses: i.Uses,
		MaxUses: i.MaxUses,
	}

	if i.MaxAge != nil {
		out.MaxAge = int(i.MaxAge.Unix() - i.CreatedAt)
	}

	if gd, err := GetGuildByID(i.GuildID); err == nil && gd != nil { out.Guild = gd.ToAPI(0, false) }
	if ch, err := GetChannelByID(i.ChannelID); err == nil && ch != nil { out.Channel = ch.ToAPI() }
	if user, err := GetUserByID(i.Inviter); err == nil && user != nil { out.Inviter = user.ToAPI(true) }

	return out
}
