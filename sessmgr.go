package main

import (
	"log"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/cskr/pubsub"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	_ "github.com/bwmarrin/snowflake"
)

var SessSub = pubsub.New(16)

func msDecodeBSON(in, out interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName:"bson", Result: out})
	if err != nil { return err }
	return decoder.Decode(in)
}

func InitSessionManager() {
	go RunSessionManager("msgs", func (dm bson.M, evt bson.M) error {
		log.Println("Yo: ", evt)
		id := fmt.Sprintf("%v", dm["channel_id"])
		switch evt.operationType.(string) {
			case "insert":
				var m Message
				err := msDecodeBSON(dm, &m)
				if err != nil { return err }
				log.Println("SessionManager: Msg:", id, m)
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
				}, id)
			break
			case "update":
				// TODO
			break
		}
		return nil
	})
	// That's all folks!
}

func RunSessionManager(col string, fn func (doc bson.M, evt bson.M) error) {
	for {
		s2 := DB.Msg.Session.New().DB("")
		pipeline := []bson.M{}
		cstream, err := s2.C(col).Watch(pipeline, mgo.ChangeStreamOptions{MaxAwaitTimeMS:time.Second*5})
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
