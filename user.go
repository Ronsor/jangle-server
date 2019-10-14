package main

import (
	"github.com/bwmarrin/snowflake"
)

// User flags
const (
	USER_FLAG_NONE = 0
	USER_FLAG_EMPLOYEE = 1 << 0
	USER_FLAG_PARTNER = 1 << 1
	// The rest are unused
)

// User premium subscription type
const (
	USER_PREMIUM_NONE = 0
	USER_PREMIUM_BOLT = 2
)

// UserInternal contains data on users that should never be exposed
type UserInternal struct {
	Password string
}

// User is a Discord-compatible structure containing information on a user
type User struct {
	ID snowflake.ID `json:"id"`
	Username string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar string `json:"avatar"`

	Bot bool `json:"bot"`
	Owner string `json:"bot_owner,omitempty"`
	Flags int `json:"flags"`
	PremiumType int `json:"premium_type"`
	Verified bool `json:"verified"`

	Locale string /* en_US */ `json:"locale"`
	Email string `json:"email,omitempty"`
	MfAuth bool `json:"mfa_enabled"`

	Internal *UserInternal `json:"-"`
}

// Safe() filters out information that should not be exposed to other users
func (u *User) Safe() *User {
	u2 := *u
	u2.Email = ""

	return &u2
}

