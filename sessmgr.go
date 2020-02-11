package main

import (
	"fmt"
	"log"
	"strings"
	//"time"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/cskr/pubsub"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/mitchellh/mapstructure"
)

var SessSub = pubsub.New(64)

func msDecodeBSON(in, out interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "bson", Result: out})
	if err != nil {
		return err
	}
	return decoder.Decode(in)
}

func InitSessionManager() {
	go RunSessionManager(DB.Core.Session, "users", func(dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		gCache.Del(snow)
		if err != nil {
			return err
		}
		switch evt["operationType"].(string) {
		case "update":
			// TODO
		case "replace":
			var pkt User
			err := msDecodeBSON(dm, &pkt)
			if err != nil {
				return err
			}
			SessSub.TryPub(gwPacket{
				Op: GW_OP_DISPATCH,
				Type: GW_EVT_USER_UPDATE,
				Data: pkt.ToAPI(false),
			}, pkt.ID.String())
			psc, err := GetPresenceForUser(pkt.ID)
			if err == nil { SetPresenceForUser(pkt.ID, psc) }
		}
		return nil
	})

	go RunSessionManager(DB.Core.Session, "presence", func(dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		gCache.Del(snow)
		if err != nil {
			return err
		}
		switch evt["operationType"].(string) {
		case "insert":
			var pkt gwPktDataUpdateStatus
			err := msDecodeBSON(dm, &pkt)
			if err != nil {
				return err
			}
			gms, err := GetGuildMembersByUserID(snow)
			if err != nil {
				return err
			}
			usr, err := GetUserByID(snow)
			if err != nil {
				return err
			}
			usrapi := usr.ToAPI(true) // just cache it
			for _, v := range gms {
				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_PRESENCE_UPDATE,
					Data: &gwEvtDataPresenceUpdate{
						User:    usrapi,
						Roles:   v.Roles,
						GuildID: v.GuildID,
						Status:  pkt.Status,
						Nick:    v.Nick,
					},
				}, v.GuildID.String())
			}
		case "update":
			uf := evt["updateDescription"].(bson.M)["updatedFields"].(bson.M)
			if _, ok := uf["notify_typing_start"]; ok && len(uf) == 1 {
				var ts gwEvtDataTypingStart
				err := msDecodeBSON(uf["notify_typing_start"], &ts)
				if err != nil {
					return err
				}
				if ts.GuildID != 0 {
					mem, err := GetGuildMemberByUserAndGuildID(ts.UserID, ts.GuildID)
					if err != nil {
						return err
					}
					ts.Member = mem.ToAPI()
				}
				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_TYPING_START,
					Data: ts,
				}, ts.ChannelID.String())
				return nil
			}
		case "replace":
			var pkt gwPktDataUpdateStatus
			err := msDecodeBSON(dm, &pkt)
			if err != nil {
				return err
			}
			gms, err := GetGuildMembersByUserID(snow)
			if err != nil {
				return err
			}
			usr, err := GetUserByID(snow)
			if err != nil {
				return err
			}
			usrapi := usr.ToAPI(true) // just cache it
			for _, v := range gms {
				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_PRESENCE_UPDATE,
					Data: &gwEvtDataPresenceUpdate{
						User:    usrapi,
						Roles:   v.Roles,
						GuildID: v.GuildID,
						Status:  pkt.Status,
						Nick:    v.Nick,
					},
				}, v.GuildID.String())
			}
		}
		return nil
	})
	go RunSessionManager(DB.Core.Session, "guildmembers", func(dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		gCache.Del(snow)
		if err != nil {
			return err
		}
		switch evt["operationType"].(string) {
		case "insert":
			var gm GuildMember
			err := msDecodeBSON(dm, &gm)
			if err != nil {
				return err
			}
			g, err := GetGuildByID(gm.GuildID)
			if err != nil {
				return err
			}
			pld := gm.ToAPI()
			pld.GuildID = gm.GuildID
			SessSub.TryPub(gwPacket{
				Op:      GW_OP_DISPATCH,
				Type:    GW_EVT_GUILD_MEMBER_ADD,
				Data:    pld,
				PvtData: &gm,
			}, g.ID.String())
			SessSub.TryPub(gwPacket{
				Op:      GW_OP_DISPATCH,
				Type:    GW_EVT_GUILD_CREATE,
				Data:    g.ToAPI(gm.UserID, true),
				PvtData: g,
			}, gm.UserID.String())
		case "replace":
			var gm GuildMember
			err := msDecodeBSON(dm, &gm)
			if err != nil {
				return err
			}
			pld := gm.ToAPI()
			pld.GuildID = gm.GuildID
			SessSub.TryPub(gwPacket{
				Op:      GW_OP_DISPATCH,
				Type:    GW_EVT_GUILD_MEMBER_UPDATE,
				Data:    pld,
				PvtData: &gm,
			}, gm.GuildID.String())
		case "update":
			gm, err := GetGuildMemberByID(snow)
			if err != nil {
				return err
			}
			if gm.Deleted != nil {
				usr, err := GetUserByID(gm.UserID)
				if err != nil {
					return err
				}
				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_GUILD_DELETE,
					Data: bson.M{
						"id":          gm.GuildID.String(),
						"unavailable": true,
					},
				}, gm.UserID.String())

				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_GUILD_MEMBER_REMOVE,
					Data: bson.M{
						"guild_id": gm.GuildID.String(),
						"user":     usr.ToAPI(true),
					},
				}, gm.GuildID.String())
				return nil
			}
			pld := gm.ToAPI()
			pld.GuildID = gm.GuildID
			SessSub.TryPub(gwPacket{
				Op:      GW_OP_DISPATCH,
				Type:    GW_EVT_GUILD_MEMBER_UPDATE,
				Data:    pld,
				PvtData: gm,
			}, gm.GuildID.String())
		}
		return nil
	})
	go RunSessionManager(DB.Core.Session, "guilds", func(dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		gCache.Del(snow)
		if err != nil {
			return err
		}
		switch evt["operationType"].(string) {
		/*case "insert":
		var g Guild
		err := msDecodeBSON(dm, &g)
		if err != nil {
			return err
		}
		SessSub.TryPub(gwPacket{
			Op:      GW_OP_DISPATCH,
			Type:    GW_EVT_GUILD_CREATE,
			Data:    g.ToAPI(g.OwnerID, true),
			PvtData: &g,
		}, g.OwnerID.String())*/
		case "update":
			g, err := GetGuildByID(snow)
			if err != nil {
				return err
			}

			uf := evt["updateDescription"].(bson.M)["updatedFields"].(bson.M)
			for k, v := range uf {
				if strings.HasPrefix(k, "roles.") {
					var role Role
					err := msDecodeBSON(v, &role)
					if err != nil {
						return err // fail fast!
					}
					typ := GW_EVT_GUILD_ROLE_UPDATE
					if role.FirstTime {
						typ = GW_EVT_GUILD_ROLE_CREATE
					}
					SessSub.TryPub(gwPacket{
						Op:   GW_OP_DISPATCH,
						Type: typ,
						Data: bson.M{
							"guild_id": g.ID.String(),
							"role":     role.ToAPI(),
						},
						PvtData: &role,
					}, g.ID.String())
				}
			}
		case "delete":
			SessSub.TryPub(gwPacket{
				Op:   GW_OP_DISPATCH,
				Type: GW_EVT_GUILD_DELETE,
				Data: bson.M{
					"id":          snow.String(),
					"unavailable": true,
				},
			}, snow.String())
		}
		return nil
	})
	go RunSessionManager(DB.Core.Session, "channels", func(dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		gCache.Del(snow)
		if err != nil {
			return err
		}
		switch evt["operationType"].(string) {
		case "insert":
			var c Channel
			err := msDecodeBSON(dm, &c)
			if err != nil {
				return err
			}
			if c.IsGuild() {
				SessSub.TryPub(gwPacket{
					Op:      GW_OP_DISPATCH,
					Type:    GW_EVT_CHANNEL_CREATE,
					Data:    c.ToAPI(),
					PvtData: &c,
				}, c.GuildID.String())
			} else if c.Type == CHTYPE_DM {
				ids := c.RecipientIDs
				SessSub.TryPub(gwPacket{
					Op:      GW_OP_DISPATCH,
					Type:    GW_EVT_CHANNEL_CREATE,
					Data:    c.ToAPI(),
					PvtData: &c,
				}, ids[0].String(), ids[1].String())
			}
		case "update":
			uf := evt["updateDescription"].(bson.M)["updatedFields"].(bson.M)
			if _, ok := uf["last_message_id"]; len(uf) == 1 && ok {
				return nil
			}
			ch, err := GetChannelByID(snow)
			if err != nil {
				return err
			}
			tgt := snow.String()
			if ch.IsGuild() {
				tgt = ch.GuildID.String()
			}
			if ch.Deleted != nil {
				SessSub.TryPub(gwPacket{
					Op:      GW_OP_DISPATCH,
					Type:    GW_EVT_CHANNEL_DELETE,
					Data:    ch.ToAPI(),
					PvtData: ch,
				}, tgt)
				return nil
			}
			SessSub.TryPub(gwPacket{
				Op:      GW_OP_DISPATCH,
				Type:    GW_EVT_CHANNEL_UPDATE,
				Data:    ch.ToAPI(),
				PvtData: ch,
			}, tgt)
		}
		return nil
	})
	go RunSessionManager(DB.Msg.Session, "msgs", func(dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		gCache.Del(snow)
		if err != nil {
			return err
		}
		switch evt["operationType"].(string) {
		case "insert":
			var m Message
			err := msDecodeBSON(dm, &m)
			if err != nil {
				return err
			}
			if m.WebhookID == 0 {
				m.Author, err = GetUserByID(m.Author.ID)
				if err != nil {
					return fmt.Errorf("WARNING: Failed to send MESSAGE_CREATE event: no such Author:", m.Author.ID)
				}
			}
			SessSub.TryPub(gwPacket{
				Op:   GW_OP_DISPATCH,
				Type: GW_EVT_MESSAGE_CREATE,
				Data: m.ToAPI(),
			}, m.ChannelID.String())
			break
		case "update":
			uf := evt["updateDescription"].(bson.M)["updatedFields"].(bson.M)
			m, err := GetMessageByID(snow) // TODO: don't do this!
			if err != nil {
				return err
			}
			if content, ok := uf["content"].(string); ok {
				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_MESSAGE_UPDATE,
					Data: bson.M{
						"id":         m.ID,
						"channel_id": m.ChannelID,
						"content":    content,
					},
				}, m.ChannelID.String())
			}
			if _, ok := uf["pinned"].(bool); ok {
				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_CHANNEL_PINS_UPDATE,
					Data: bson.M{"channel_id": m.ChannelID},
				}, m.ChannelID.String())
			}
			if deleted, ok := uf["deleted"].(bool); ok && deleted {
				payload := bson.M{
					"id":         m.ID,
					"channel_id": m.ChannelID,
				}

				if m.GuildID != 0 {
					payload["guild_id"] = m.GuildID
				}

				SessSub.TryPub(gwPacket{
					Op:   GW_OP_DISPATCH,
					Type: GW_EVT_MESSAGE_DELETE,
					Data: &gwEvtDataMessageDelete{
						ID:        m.ID,
						ChannelID: m.ChannelID,
						GuildID:   m.GuildID,
					},
				}, m.ChannelID.String())
			}
			break
		}
		return nil
	})
	// That's all folks!
}

func RunSessionManager(sess *mgo.Session, col string, fn func(doc bson.M, evt bson.M) error) {
	s2 := sess.Copy().DB("")
	for {
		pipeline := []bson.M{}
		cstream, err := s2.C(col).Watch(pipeline, mgo.ChangeStreamOptions{MaxAwaitTimeMS: 5000})
		if err != nil {
			log.Println("SessionManager:", err)
			continue
		}
		var doc bson.M
		for cstream.Next(&doc) {
			dm, ok := doc["fullDocument"].(bson.M)
			if !ok {
				dm = nil
			}
			(func() {
				defer util.TryRecover()
				err := fn(dm, doc)
				if err != nil {
					log.Println("SessionManager: "+col+": error:", err)
				}
			})()
		}
		if err := cstream.Close(); err != nil {
			log.Println("SessionManager:", err)
		}
	}
	panic("Unreachable")
}
