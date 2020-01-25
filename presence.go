package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

type PresenceInternal struct {
	ID        snowflake.ID           `bson:"_id"`
	Timestamp time.Time              `bson:"timestamp"`
	Presence  *gwPktDataUpdateStatus `bson:"presence"`
	Typing *gwEvtDataTypingStart `bson:"notify_typing_start"`
}

// I dare say it's awful to repurpose a gateway packet data type for this...
func SetPresenceForUser(userID snowflake.ID, presence *gwPktDataUpdateStatus) error {
	c := DB.Core.C("presence")
	if presence.Status != STATUS_ONLINE && presence.Status != STATUS_DND &&
		presence.Status != STATUS_INVISIBLE && presence.Status != STATUS_OFFLINE {
		return fmt.Errorf("Bad status")
	}
	dat := &PresenceInternal{userID, time.Now(), presence, nil}
	_, err := c.UpsertId(userID, dat)
	return err
}

func StartTypingForUser(userID snowflake.ID, typing *gwEvtDataTypingStart) error {
	c := DB.Core.C("presence")
	return c.UpdateId(userID, bson.M{"$set": bson.M{"notify_typing_start": typing}})
}

func RefreshPresenceForUser(userID snowflake.ID) error {
	c := DB.Core.C("presence")
	err := c.UpdateId(userID, bson.M{"$set": bson.M{"timestamp": time.Now()}})
	return err
}

func GetPresenceForUser(userID snowflake.ID) (*gwPktDataUpdateStatus, error) {
	var dat PresenceInternal
	c := DB.Core.C("presence")
	err := c.Find(bson.M{"_id": userID}).One(&dat)
	if err != nil {
		return &gwPktDataUpdateStatus{Status: STATUS_OFFLINE}, nil
		//return nil, err
	}
	return dat.Presence, nil
}
