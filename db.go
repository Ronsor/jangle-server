package main

import (
	"log"

	"github.com/globalsign/mgo"
)

var dbSess *mgo.Session
var DB = struct {
	Core, Msg *mgo.Database
} {}
func InitDB() {
	sess, err := mgo.Dial(*flgMongoDB)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: separate DB server for messages
	DB.Msg = sess.DB("")
	DB.Core = sess.DB("")
	dbSess = sess

	// Add collections and indexes
	DB.Core.C("users").EnsureIndex(mgo.Index{Name:"idx_guilds", Key: []string{"guildids"}})
	DB.Core.C("users").EnsureIndex(mgo.Index{Name:"idx_tags", Key: []string{"username","discriminator"}})

	if *flgStaging {
		InitUserStaging()
		log.Printf("staging: added dummy users")
	}
}
