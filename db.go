package main

import (
	"log"
	"time"

	"github.com/globalsign/mgo"
)

var dbSess *mgo.Session
var DB = struct {
	Core, Msg, Files *mgo.Database
}{}

func InitDB() {
	sess, err := mgo.Dial(*flgMongoDB)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: separate DB server for messages, and files
	DB.Msg = sess.DB("")
	DB.Core = sess.DB("")
	DB.Files = sess.DB("")
	dbSess = sess

	// Add collections and indexes
	//DB.Core.C("users").EnsureIndex(mgo.Index{Name:"idx_guilds", Key: []string{"guildids"}})
	DB.Core.C("users").EnsureIndex(mgo.Index{Name: "idx_tags", Key: []string{"username", "discriminator"}})
	DB.Core.C("presence").EnsureIndex(mgo.Index{Name: "idx_presence_ttl", Key: []string{"timestamp"}, Unique: false, Background: true, ExpireAfter: 60 * time.Second})

	DB.Core.C("channels").EnsureIndex(mgo.Index{Name: "idx_recipients", Key: []string{"recipients"}})

	DB.Msg.C("msgs").EnsureIndex(mgo.Index{Name: "idx_pinned", Key: []string{"channel_id", "pinned"}})

	if *flgStaging {
		InitUserStaging()
		log.Printf("staging: added dummy users")
		InitChannelStaging()
		log.Printf("staging: added dummy channels")
		InitGuildStaging()
		log.Printf("staging: added dummy guilds")
	}
}
