package main

import (
	"github.com/bwmarrin/snowflake"
)

type Message struct {
	ID snowflake.ID `json:"id" bson:"_id"`
}
