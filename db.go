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
	sess.SetSafe(nil)
	DB.Msg = sess.DB("")
	DB.Core = sess.DB("")
	DB.Files = sess.DB("")
	dbSess = sess

	// Add collections and indexes
	//DB.Core.C("users").EnsureIndex(mgo.Index{Name:"idx_guilds", Key: []string{"guildids"}})
	DB.Core.C("users").EnsureIndex(mgo.Index{Name: "idx_tags", Key: []string{"username", "discriminator"}, Unique: true})
	DB.Core.C("presence").EnsureIndex(mgo.Index{Name: "idx_presence_ttl", Key: []string{"timestamp"}, Unique: false, Background: true, ExpireAfter: 60 * time.Second})

	DB.Core.C("channels").EnsureIndex(mgo.Index{Name: "idx_recipients", Key: []string{"recipients"}, Unique: true})
	DB.Core.C("channels").EnsureIndex(mgo.Index{Name: "idx_channel_deleted", Key: []string{"deleted"}, Sparse: true, Unique: false, ExpireAfter: 60 * time.Second})

	DB.Core.C("guildmembers").EnsureIndex(mgo.Index{Name: "idx_guildmember_id_and_user", Key: []string{"guild_id", "user"}, Unique: true})
	DB.Core.C("guildmembers").EnsureIndex(mgo.Index{Name: "idx_guildmember", Key: []string{"guild_id"}, Unique: false})
	DB.Core.C("guildmembers").EnsureIndex(mgo.Index{Name: "idx_guildmember_deleted", Key: []string{"deleted"}, Sparse: true, Unique: false, ExpireAfter: 60 * time.Second})

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
