package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

type PresenceInternal struct {
	ID        snowflake.ID           `bson:"_id"`
	Timestamp time.Time              `bson:"timestamp"`
	Presence  *gwPktDataUpdateStatus `bson:"presence"`
}

// I dare say it's awful to repurpose a gateway packet data type for this...
func SetPresenceForUser(userID snowflake.ID, presence *gwPktDataUpdateStatus) error {
	c := DB.Core.C("presence")
	dat := &PresenceInternal{userID, time.Now(), presence}
	_, err := c.UpsertId(userID, dat)
	return err
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
