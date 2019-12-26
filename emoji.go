package main

import (
	"github.com/bwmarrin/snowflake"
	"fmt"

	"strings"
)

// Discord-compatible Emoji structure
type Emoji struct {
	ID snowflake.ID `bson:"_id"`
	Name string `bson:"name"`
	Roles []snowflake.ID `bson:"roles"`
	User *User `bson:"user"`
	RequireColons bool `bson:"require_colons"`
	Managed /* what's this for? */ bool `bson:"managed"`
	Animated bool `bson:"animated"`
}

func GetEmojiFromString(s string) (*Emoji, error) {
	sa := strings.Split(s, ":")
	if len(sa) == 1 {
		return &Emoji{Name: sa[0]}, nil
	}
	return nil, fmt.Errorf("Unsupported")
}

func (e *Emoji) String() string {
	if e.ID == 0 {
		return e.Name
	} else {
		return e.Name + ":" + e.ID.String()
	}
	panic("unreachable")
	return ""
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

