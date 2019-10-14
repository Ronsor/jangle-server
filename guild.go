package main

import (
	"github.com/bwmarrin/snowflake"
)

// Represents a currently unavailable guild
// Used in GW_EVT_READY
type UnavailableGuild struct {
	ID snowflake.ID `json:"id"`
	Unavailable bool `json:"unavailable"`
}

// TODO: the rest
