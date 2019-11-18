package main

import (
	"fmt"

	"server/util"

	"github.com/bwmarrin/snowflake"
	"github.com/valyala/fasthttp"
	"github.com/globalsign/mgo/bson"
)

// User flags
const (
	USER_FLAG_NONE = 0
	USER_FLAG_EMPLOYEE = 1 << 0
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
type UserSettings struct {
	Locale string `json:"locale"`
	AfkTimeout int `json:"afk_timeout"`
	// TODO: the rest
}

// User is a Discord-compatible structure containing information on a user
type User struct {
	ID snowflake.ID `json:"id,string" bson:"_id"`
	Username string `json:"username"`
	Discriminator string `json:"discriminator"`
	Email string `json:"email,omitempty"`

	Bot bool `json:"bot"`
	Avatar string `json:"avatar"`
	MfaEnabled bool `json:"mfa_enabled"`
	Verified bool `json:"verified"`

	Flags int `json:"flags"`
	PremiumType int `json:"premium_type"`
	PremiumSince int `json:"premium_since"`
	Phone string `json:"-" bson:"phone"` // We use _phone instead, for reasons

	// Gonna send this anyway

	LastSession int `json:"last_session"`

	// This is never sent in a user structure

	PasswordHash string `json:"-" bson:"password_hash"`
	Settings *UserSettings `json:"-" bson:"user_settings"`
	GuildIDs []snowflake.ID `json:"-" bson:"guildids"`

	// Extra stuff for "compatibility" (ick!)

	mobile bool `json:"mobile" bson:"-"`   // Why not just check if `phone` is null?
	locale string `json:"locale" bson:"-"` // Technically in UserSettings
	premium bool `json:"premium" bson:"-"` // Why?!? We already have PremiumType and PremiumSince
	_phone *string `json:"phone" bson:"-"` // ....
}

// GetUserByID returns a user by their unique ID
func GetUserByID(ID snowflake.ID) (u *User, e error) {
	var u2 User
	c := DB.Core.C("users")
	e = c.Find(bson.M{"_id": ID}).One(&u2)
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

// GetUserByHttpCtx returns a user using a fasthttp.RequestCtx
func GetUserByHttpCtx(c *fasthttp.RequestCtx, ctxvar string) (*User, error) {
	b := c.Request.Header.Peek("Authorization")
	if b == nil { return nil, fmt.Errorf("No authorization token supplied") }
	a := string(b)
	user, err := GetUserByToken(a)
	if err != nil { return nil, err }
	if ctxvar != "" {
		uid2 := c.UserValue(ctxvar)
		if uid2 == "" || uid2 == "@me" {
			return user, nil
		}
		// TODO
	}
	return user, nil
}

// MarshalAPI returns a version of the User struct that can be safely returned
func (u *User) MarshalAPI(includeEmail bool) *User {
	u2 := *u
	if u2.Phone != "" {
		u2.mobile = true
		u2._phone = &u2.Phone
	}
	if u2.PremiumType != 0 {
		u2.premium = true
	}
	if !includeEmail {
		u2.Email = ""
	}
	u2.locale = u2.Settings.Locale
	return &u2
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
		Settings: &UserSettings{
			Locale: "en-US",
		},
	})
}
