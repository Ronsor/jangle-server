package main

import (
	"log"

	"github.com/globalsign/mgo"
)

var dbSess *mgo.Session
var DB = struct {
	Core, Msg, Files *mgo.Database
} {}

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
	DB.Core.C("users").EnsureIndex(mgo.Index{Name:"idx_tags", Key: []string{"username","discriminator"}})

	DB.Core.C("channels").EnsureIndex(mgo.Index{Name:"idx_recipients", Key: []string{"recipients"}})


	if *flgStaging {
		InitUserStaging()
		log.Printf("staging: added dummy users")
	}
}
