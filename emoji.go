package main

import (
	"github.com/bwmarrin/snowflake"
)

/* Discord-compatible Emoji structure */

type Emoji struct {
	ID snowflake.ID `bson:"id"`
	Name string `bson:"name"`
	Roles []snowflake.ID `bson:"roles"`
	User *User `bson:"user"`
	RequireColons bool `bson:"require_colons"`
	Managed /* ? */ bool `bson:"managed"`
	Animated bool `bson:"animated"`
}

func (e *Emoji) ToAPI(lite bool) *APITypeEmoji {
	e2 := *e
	ret := &APITypeEmoji{}
	if e2.ID != 0 { ret.ID = &e2.ID }
	ret.Name = &e2.Name
	if !lite {
		ret.Roles = e2.Roles
		if e2.User != nil {
			ret.User = e2.User.ToAPI(true)
		}
		ret.RequireColons = &e2.RequireColons
		ret.Managed = &e2.Managed
		ret.Animated = &e2.Animated
	}
	return ret
}
