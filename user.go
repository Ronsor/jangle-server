package main

import (
	"fmt"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/valyala/fasthttp"
	"github.com/globalsign/mgo/bson"
)

// User flags
const (
	USER_FLAG_NONE = 0
	USER_FLAG_STAFF = 1 << 0
	USER_FLAG_PARTNER = 1 << 1
	USER_FLAG_EARLYADOPTER = 1 << 24
	// The rest are unused
)

// User premium subscription type
const (
	USER_PREMIUM_NONE = 0
	// USER_PREMIUM_NITRO_CLASSIC = 1
	USER_PREMIUM_BOLT = 2
)

// UserSettings is a Discord-compatible structure containing a user's settings
// This struct is safe to be returned by an API call
type UserSettings struct {
	Locale string `bson:"locale"`
	AfkTimeout int `bson:"afk_timeout"`
	// TODO: the rest
}

// User is a Discord-compatible structure containing information on a user
// This struct is not safe to be returned by an API call
type User struct {
	ID snowflake.ID `bson:"_id"`
	Username string `bson:"username"`
	Discriminator string `bson:"discriminator"`
	Email string `bson:"email,omitempty"`

	Bot bool `bson:"bot"`
	Avatar string `bson:"avatar"`
	MfaEnabled bool `bson:"mfa_enabled"`
	Verified bool `bson:"verified"`

	Flags int `bson:"flags"`
	PremiumType int `bson:"premium_type"`
	PremiumSince int `bson:"premium_since"`
	Phone string `bson:"phone"`

	LastSession int `bson:"last_session"`

	PasswordHash string `bson:"password_hash"`
	Settings *UserSettings `bson:"user_settings"`

	Presence *gwPktDataUpdateStatus `bson:"presence"`
	LastMessageIDs map[snowflake.ID]snowflake.ID `bson:"read_last_message_ids"`
}

// GetUserByID returns a user by their unique ID
func GetUserByID(ID snowflake.ID) (u *User, e error) {
	var u2 User
	c := DB.Core.C("users")
	e = c.Find(bson.M{"_id": ID}).One(&u2)
	u2.ID = ID
	u = &u2
	return
}

// GetUserByToken returns a user using an authentication token
func GetUserByToken(token string) (*User, error) {
	if *flgStaging {
		var i snowflake.ID
		n, err := fmt.Sscanf(token, "%d", &i)
		if n != 1 { return nil, fmt.Errorf("Bad ID") }
		if err != nil { return nil, err }
		return GetUserByID(i)
	}
	return nil, fmt.Errorf("Not implemented") // TODO: implement actual auth
}

// GetUserByHttpRequest returns a user using a fasthttp.RequestCtx.
// Specifically, it attempts to authorize the request using a token.
func GetUserByHttpRequest(c *fasthttp.RequestCtx, ctxvar string) (*User, error) {
	b := c.Request.Header.Peek("Authorization")
	if b == nil { return nil, fmt.Errorf("No authorization token supplied") }
	a := string(b)
	user, err := GetUserByToken(a)
	if err != nil { return nil, err }
	if ctxvar != "" {
		uid2 := c.UserValue(ctxvar).(string) // You're going to pass a string, or I'll panic()
		if uid2 == "" || uid2 == "@me" {
			return user, nil
		}
		// TODO
	}
	return user, nil
}

// ToAPI returns a version of the User struct that can be returned by API calls
func (u *User) ToAPI(safe bool) *APITypeUser {
	u2 := &APITypeUser{
		ID: u.ID,
		Username: u.Username,
		Discriminator: u.Discriminator,
		AvatarHash: u.Avatar,
		Bot: u.Bot,
		MfaEnabled: true,
		Flags: u.Flags,
		PremiumType: u.PremiumType,
	}
	if u.Settings != nil {
		u2.Locale = u.Settings.Locale
	}
	if u.PremiumType != USER_PREMIUM_NONE {
		u2.Premium = true
	}
	if !safe {
		if u.Phone != "" {
			u2.Phone = &u.Phone
			u2.Mobile = true
		}
		u2.Email = u.Email
		u2.Verified = &u.Verified
	}
	return u2
}

func (u *User) MarkRead(cid, mid snowflake.ID) {
	if u.LastMessageIDs != nil {
		u.LastMessageIDs = map[snowflake.ID]snowflake.ID{}
	}
	u.LastMessageIDs[cid] = mid
}

// DMChannels returns the DM channels a user is in
func (u *User) Channels() ([]*Channel, error) {
	ch := []*Channel{}
	err := DB.Core.C("channels").Find(bson.M{"recipient_ids":u.ID}).All(&ch)
	if err != nil { return nil, err }
	return ch, nil
}

func (u *User) Guilds() ([]*Guild, error) {
	return GetGuildsByUserID(u.ID)
}

/*
	The following code is for testing.
	It is not used in production.
*/

// Initialize dummy users in database
func InitUserStaging() {
	c := DB.Core.C("users")
	c.Insert(&User{
		ID: 42,
		Username: "test1",
		Discriminator: "1234",
		Email: "test@localhost",
		PasswordHash: util.CryptPass("hello"),
		Flags: USER_FLAG_STAFF | USER_FLAG_EARLYADOPTER,
		Settings: &UserSettings{
			Locale: "en-US",
		},
	})
	c.Insert(&User{
		ID: 43,
		Username: "hello",
		Discriminator: "4242",
		Email: "test2@localhost",
		PasswordHash: util.CryptPass("hello"),
		Flags: USER_FLAG_EARLYADOPTER,
		Settings: &UserSettings{
			Locale: "en-US",
		},
	})
}
