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
	ID snowflake.ID `bson:"_id"`
	Type int `bson:"type"`

	// Text Channel only
	LastMessageID snowflake.ID `bson:"last_message_id"`

	// DM/Group DM only
	RecipientIDs []snowflake.ID `bson:"recipient_ids"`

	// Group DM only
	OwnerID snowflake.ID `bson:"owner_id"`
	Icon string `bson:"icon"`

	// Guild only
	GuildID snowflake.ID `bson:"guild_id"`
	Position int `bson:"position"`
	Name string `bson:"name"`
	ParentID snowflake.ID `bson:"parent_id"`
	PermissionOverwrites []interface{} `bson:"permission_overwrites"`

	// Guild Text Channel only
	Topic string `bson:"topic"`
	NSFW bool `bson:"nsfw"`
	RateLimitPerUser int `bson:"rate_limit_per_user"`

	// Guild Voice Channel only
	Bitrate int `bson:"bitrate"`
	UserLimit int `bson:"user_limit"`

	Deleted bool `bson:"deleted"`
}

func CreateDMChannel(party1, party2 snowflake.ID) (*Channel, error) {
	var c2 Channel
	c := DB.Core.C("channels")
	e := c.Find(bson.M{"recipients": bson.M{"$all": []snowflake.ID{party1,party2}}, "type": CHTYPE_DM}).One(&c2)
	if e != nil {
		c2.ID = flake.Generate()
		c2.RecipientIDs = []snowflake.ID{party1,party2}
		c2.Type = CHTYPE_DM
		err := c.Insert(&c2)
		if err != nil { return nil, err }
	}
	return &c2, nil
}

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

func (c *Channel) ToAPI() APITypeAnyChannel {
	if c.Type == CHTYPE_DM {
		rcp := []*APITypeUser{}
		for _, v := range c.RecipientIDs {
			u, err := GetUserByID(v)
			if err != nil {
				rcp = append(rcp, &APITypeUser{ID:v,Discriminator:"0000",Username:"Unknown"})
			} else {
				rcp = append(rcp, u.ToAPI(true))
			}
		}
		return &APITypeDMChannel{
			ID: c.ID,
			Type: c.Type,
			Recipients: rcp,
			LastMessageID: c.LastMessageID,
		}
	}
	return nil
}

func (c *Channel) CreateMessage(m *Message) error {
	d := DB.Msg.C("msgs")
	m.ID = flake.Generate()
	m.ChannelID = c.ID
	err := d.Insert(&m)
	if err == nil {
		c.LastMessageID = m.ID
		return c.Save()
	}
	return err
}

func (c *Channel) Save() error {
	d := DB.Core.C("channels")
	return d.UpdateId(c.ID, bson.M{"$set":c})
}

func InitChannelStaging() {
	c := DB.Core.C("channels")
	c.Insert(&Channel{
		ID: 1,
		Type: CHTYPE_DM,
		RecipientIDs: []snowflake.ID{42,43},
	})
}
