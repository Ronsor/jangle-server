package main

import (
	"log"
	"fmt"
//	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/cskr/pubsub"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/bwmarrin/snowflake"
)

var SessSub = pubsub.New(16)

func msDecodeBSON(in, out interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName:"bson", Result: out})
	if err != nil { return err }
	return decoder.Decode(in)
}

func InitSessionManager() {
	go RunSessionManager("guilds", func (dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		_ = snow
		if err != nil { return err }
		switch evt["operationType"].(string) {
			case "insert":
				var g Guild
				err := msDecodeBSON(dm, &g)
				if err != nil { return err }
				SessSub.TryPub(gwPacket{
					Op: GW_OP_DISPATCH,
					Type: GW_EVT_GUILD_CREATE,
					Data: g.ToAPI(g.OwnerID),
					PvtData: &g,
				}, g.OwnerID.String())
			case "update":
				uf := evt["updateDescription"].(bson.M)["updatedFields"].(bson.M)
				log.Println("Jnk", uf)
		}
		return nil
	})
	go RunSessionManager("channels", func (dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		_ = snow
		if err != nil { return err }
		switch evt["operationType"].(string) {
			case "insert":
				var c Channel
				err := msDecodeBSON(dm, &c)
				if err != nil { return err }
				if c.IsGuild() {
					SessSub.TryPub(gwPacket{
						Op: GW_OP_DISPATCH,
						Type: GW_EVT_CHANNEL_CREATE,
						Data: c.ToAPI(),
						PvtData: &c,
					}, c.GuildID.String())
				} else if c.Type == CHTYPE_DM {
					ids := c.RecipientIDs
					SessSub.TryPub(gwPacket{
						Op: GW_OP_DISPATCH,
						Type: GW_EVT_CHANNEL_CREATE,
						Data: c.ToAPI(),
						PvtData: &c,
					}, ids[0].String(), ids[1].String())
				}
		}
		return nil
	})
	go RunSessionManager("msgs", func (dm bson.M, evt bson.M) error {
		log.Println(evt)
		id := fmt.Sprintf("%v", evt["documentKey"].(bson.M)["_id"])
		snow, err := snowflake.ParseString(id)
		if err != nil { return err }
		switch evt["operationType"].(string) {
			case "insert":
				var m Message
				err := msDecodeBSON(dm, &m)
				if err != nil { return err }
				if m.WebhookID == 0 {
					m.Author, err = GetUserByID(m.Author.ID)
					if err != nil {
						return fmt.Errorf("WARNING: Failed to send MESSAGE_CREATE event: no such Author:", m.Author.ID)
					}
				}
				SessSub.TryPub(gwPacket{
					Op: GW_OP_DISPATCH,
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
						Op: GW_OP_DISPATCH,
						Type: GW_EVT_MESSAGE_UPDATE,
						Data: bson.M{
							"id": m.ID,
							"channel_id": m.ChannelID,
							"content": content,
						},
					}, m.ChannelID.String())
				}
				if deleted, ok := uf["deleted"].(bool); ok && deleted {
					payload := bson.M{
						"id": m.ID,
						"channel_id": m.ChannelID,
					}

					if m.GuildID != 0 { payload["guild_id"] = m.GuildID }

					SessSub.TryPub(gwPacket{
						Op: GW_OP_DISPATCH,
						Type: GW_EVT_MESSAGE_DELETE,
						Data: &gwEvtDataMessageDelete{
							ID: m.ID,
							ChannelID: m.ChannelID,
							GuildID: m.GuildID,
						},
					}, m.ChannelID.String())
				}
			break
		}
		return nil
	})
	// That's all folks!
}

func RunSessionManager(col string, fn func (doc bson.M, evt bson.M) error) {
	s2 := DB.Msg.Session.Copy().DB("")
	for {
		pipeline := []bson.M{}
		cstream, err := s2.C(col).Watch(pipeline, mgo.ChangeStreamOptions{MaxAwaitTimeMS:5000})
		if err != nil { log.Println("SessionManager:", err); continue }
		var doc bson.M
		for cstream.Next(&doc) {
			dm, ok := doc["fullDocument"].(bson.M)
			if !ok { dm = nil }
			err := fn(dm, doc)
			if err != nil {
				log.Println("SessionManager: " + col + ": error:", err)
			}
		}
		if err := cstream.Close(); err != nil {
			log.Println("SessionManager:", err)
		}
	}
	panic("Unreachable")
}
