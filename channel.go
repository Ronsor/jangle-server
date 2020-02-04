package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

// Channel types
const (
	CHTYPE_GUILD_TEXT     = 0
	CHTYPE_DM             = 1
	CHTYPE_GUILD_VOICE    = 2
	CHTYPE_GROUP_DM       = 3
	CHTYPE_GUILD_CATEGORY = 4
	// WONTFIX: GUILD_NEWS, GUILD_STORE
)

type PermissionOverwrite struct {
	ID    snowflake.ID `json:"id"`
	Type  string       `json:"type"` // Either "role" or "member"
	Allow PermSet      `json:"allow"`
	Deny  PermSet      `json:"deny"`
}

// Channel is a Discord-compatible structure representing any type of channel
type Channel struct {
	ID   snowflake.ID `bson:"_id"`
	Type int          `bson:"type"`

	// Text Channel only
	LastMessageID snowflake.ID `bson:"last_message_id"`

	// DM/Group DM only
	RecipientIDs []snowflake.ID `bson:"recipient_ids,omitempty"`

	// Group DM only
	OwnerID snowflake.ID `bson:"owner_id"`
	Icon    string       `bson:"icon"`

	// Guild only
	GuildID              snowflake.ID                    `bson:"guild_id"`
	Position             int                             `bson:"position"`
	Name                 string                          `bson:"name"`
	ParentID             snowflake.ID                    `bson:"parent_id"`
	PermissionOverwrites map[string]*PermissionOverwrite `bson:"permission_overwrites"`

	// Guild Text Channel only
	Topic            string `bson:"topic"`
	NSFW             bool   `bson:"nsfw"`
	RateLimitPerUser int    `bson:"rate_limit_per_user"`

	// Guild Voice Channel only
	Bitrate   int `bson:"bitrate"`
	UserLimit int `bson:"user_limit"`

	Deleted *time.Time `bson:"deleted,omitempty"`
}

func CreateDMChannel(party1, party2 snowflake.ID) (*Channel, error) {
	var c2 Channel
	c := DB.Core.C("channels")
	e := c.Find(bson.M{"recipient_ids": bson.M{"$all": []snowflake.ID{party1, party2}}, "type": CHTYPE_DM}).One(&c2)
	if e != nil {
		c2.ID = flake.Generate()
		c2.RecipientIDs = []snowflake.ID{party1, party2}
		c2.Type = CHTYPE_DM
		err := c.Insert(&c2)
		if err != nil {
			return nil, err
		}
	}
	return &c2, nil
}

func GetChannelByID(id snowflake.ID) (*Channel, error) {
	var c2 Channel
	c := DB.Core.C("channels")
	e := c.Find(bson.M{"_id": id}).One(&c2)
	if e != nil {
		return nil, e
	}
	return &c2, nil
}

func GetChannelsByGuild(guildID snowflake.ID) ([]*Channel, error) {
	var c2 []*Channel
	c := DB.Core.C("channels")
	e := c.Find(bson.M{"guild_id": guildID}).All(&c2)
	if e != nil {
		return nil, e
	}
	return c2, nil
}

func GetChannelsByParentID(parentID snowflake.ID) ([]*Channel, error) {
	var c2 []*Channel
	c := DB.Core.C("channels")
	e := c.Find(bson.M{"parent_id": parentID}).All(&c2)
	if e != nil {
		return nil, e
	}
	return c2, nil
}

func (c *Channel) CreateMessage(m *Message) error {
	d := DB.Msg.C("msgs")
	m.ID = flake.Generate()
	if c.IsGuild() {
		m.GuildID = c.GuildID
	}
	m.ChannelID = c.ID
	err := d.Insert(&m)
	if err == nil {
		c.LastMessageID = m.ID
		return c.Save()
	}
	return err
}

func (c *Channel) Messages(around, before, after snowflake.ID, limit int, extra ...interface{} /* pinned bool */) ([]*Message, error) {
	d := DB.Msg.C("msgs")
	idquery := bson.M{}
	wholequery := bson.M{"channel_id": c.ID}
	if around != 0 {
		idquery["$gt"] = around - 0xFFFFFFFF
		idquery["$lt"] = around + 0xFFFFFFFF
		wholequery["_id"] = idquery
	} else {
		if before != 0 {
			idquery["$lt"] = before
			wholequery["_id"] = idquery
		} else if after != 0 {
			idquery["$gt"] = after
			wholequery["_id"] = idquery
		}
	}
	if len(extra) > 0 {
		if pinned, ok := extra[0].(bool); ok && pinned {
			wholequery["pinned"] = true
		}
	}
	out := []*Message{}
	err := d.Find(wholequery).Sort("-timestamp").Limit(limit).All(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Channel) IsGuild() bool {
	return c.Type == CHTYPE_GUILD_TEXT || c.Type == CHTYPE_GUILD_CATEGORY // TODO VOICE: the other types
}

func (c *Channel) Delete() error {
	if !c.IsGuild() {
		return nil // We don't actually let you delete DM channels
	}
	//chcol := DB.Core.C("channels")
	if c.Type == CHTYPE_GUILD_TEXT {
		msgcol := DB.Msg.C("msgs")
		_, err := msgcol.RemoveAll(bson.M{"channel_id": c.ID})
		if err != nil {
			return err
		}
	}
	if c.Type == CHTYPE_GUILD_CATEGORY {
		chs, err := GetChannelsByParentID(c.ID)
		if err != nil {
			return err
		}
		for _, v := range chs {
			// This is probably a race condition
			// Channel update (PUT/PATCH /channels/:cid) and (DELETE /channels/:parentcid)  Channel parent ID set to 0
			// Still, this is probably not going to happen so <shrug>
			v.ParentID = 0
			v.Save()
		}
	}

	t := time.Now()
	c.Deleted = &t
	err := c.Save()
	if err != nil {
		return err
	}
	//err = chcol.RemoveId(c.ID)
	// Goodbye!
	return err
}

func (c *Channel) ToAPI() APITypeAnyChannel {
	if c.Type == CHTYPE_DM {
		rcp := []*APITypeUser{}
		for _, v := range c.RecipientIDs {
			u, err := GetUserByID(v)
			if err != nil {
				rcp = append(rcp, &APITypeUser{ID: v, Discriminator: "0000", Username: "Unknown"})
			} else {
				rcp = append(rcp, u.ToAPI(true))
			}
		}
		return &APITypeDMChannel{
			ID:            c.ID,
			Type:          c.Type,
			Recipients:    rcp,
			LastMessageID: c.LastMessageID,
		}
	} else if c.Type == CHTYPE_GUILD_TEXT {
		ovw := []*APITypePermissionOverwrite{}
		for _, v := range c.PermissionOverwrites {
			x := APITypePermissionOverwrite(*v)
			ovw = append(ovw, &x)
		}
		return &APITypeGuildTextChannel{
			ID:                   c.ID,
			GuildID:              c.GuildID,
			Type:                 c.Type,
			LastMessageID:        c.LastMessageID,
			Name:                 c.Name,
			Topic:                c.Topic,
			NSFW:                 c.NSFW,
			Position:             c.Position,
			PermissionOverwrites: ovw,
		}
	}
	return nil
}

func (c *Channel) Guild() (*Guild, error) {
	if !c.IsGuild() {
		return nil, fmt.Errorf("This is not a guild channel")
	}
	return GetGuildByID(c.GuildID)
}

func (c *Channel) SetPermissionOverwrites(po []*PermissionOverwrite, u *User) error {
	// TODO: actually check stuff
	if c.IsGuild() {
		g, err := c.Guild()
		if err != nil {
			return err
		}
		roles := []snowflake.ID{}
		for _, v := range po {
			roles = append(roles, v.ID)
		}
		mem, err := g.GetMember(u.ID)
		if err != nil {
			return err
		}
		if g.OwnerID != u.ID && !g.CanSetRoles(mem.Roles, roles) {
			return fmt.Errorf("Permission denied")
		}
	}
	c.PermissionOverwrites = map[string]*PermissionOverwrite{}
	for _, v := range po {
		c.PermissionOverwrites[v.ID.String()] = v
	}
	return c.Save()
}

func (c *Channel) GetPermissions(u *User) PermSet {
	if c.Type == CHTYPE_DM {
		return PERM_EVERYTHING
	}
	if c.Type == CHTYPE_GUILD_TEXT {
		gd, err := c.Guild()
		if err != nil {
			return 0
		}
		perm := gd.GetPermissions(u)
		if perm.Has(PERM_ADMINISTRATOR) || c.PermissionOverwrites == nil {
			return perm
		}
		mem, err := gd.GetMember(u.ID)
		if err != nil {
			return 0
		}
		var allow, deny PermSet
		for _, v := range mem.Roles {
			ovw, ok := c.PermissionOverwrites[v.String()]
			if !ok {
				continue
			}
			allow |= ovw.Allow
			deny |= ovw.Deny
		}
		perm &= ^deny
		perm |= allow
		uOvw := c.PermissionOverwrites[u.ID.String()]
		if uOvw != nil {
			perm &= ^uOvw.Deny
			perm |= uOvw.Allow
		}
		return perm
	}
	return 0
}

// HasPermissions() is deprecated; Use GetPermissions().Has()
func (c *Channel) HasPermissions(u *User, p PermSet) bool {
	if c.Type == CHTYPE_DM {
		for _, v := range c.RecipientIDs {
			if v == u.ID {
				return true
			}
		}
		return false
	}
	return false
}

func (c *Channel) Save() error {
	d := DB.Core.C("channels")
	return d.UpdateId(c.ID, bson.M{"$set": c})
}

func InitChannelStaging() {
	//	c := DB.Core.C("channels")
	/*	c.Insert(&Channel{
		ID: 1,
		Type: CHTYPE_DM,
		RecipientIDs: []snowflake.ID{42,43},
	})*/
}
