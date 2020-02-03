package main

import (
	"fmt"
	"math/rand"
	"time"

	jwt "jangled/sjwt"
	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
	"github.com/valyala/fasthttp"
)

// User flags
const (
	USER_FLAG_NONE         = 0
	USER_FLAG_STAFF        = 1 << 0
	USER_FLAG_PARTNER      = 1 << 1
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
	Locale     string `json:"locale"`
	AfkTimeout int    `json:"afk_timeout"`
	Theme      string `json:"theme"`
	Status     string `json:"status"`
	// TODO: the rest
}

// User is a Discord-compatible structure containing information on a user
// This struct is not safe to be returned by an API call
type User struct {
	ID            snowflake.ID `bson:"_id"`
	Username      string       `bson:"username"`
	Discriminator string       `bson:"discriminator"`
	Email         string       `bson:"email,omitempty"`

	Bot        bool   `bson:"bot"`
	Avatar     string `bson:"avatar"`
	MfaEnabled bool   `bson:"mfa_enabled"`
	Verified   bool   `bson:"verified"`

	Flags        int    `bson:"flags"`
	PremiumType  int    `bson:"premium_type"`
	PremiumSince int    `bson:"premium_since"`
	Phone        string `bson:"phone"`

	LastSession int `bson:"last_session"`

	PasswordHash string        `bson:"password_hash"`
	Settings     *UserSettings `bson:"user_settings"`

	Presence       *gwPktDataUpdateStatus        `bson:"presence"`
	LastMessageIDs map[snowflake.ID]snowflake.ID `bson:"read_last_message_ids"`
	JwtSecret      string                        `bson:"jwt_secret"`
}

// CreateUser creates a user
func CreateUser(username, email, password string) (*User, error) {
	c := DB.Core.C("users")

	usr := &User{
		ID:           flake.Generate(),
		Username:     username,
		Email:        email,
		PasswordHash: util.CryptPass(password),
		Settings: &UserSettings{
			Locale: "en-US",
		},
	}

	dint := 1 + rand.Intn(9998)

	var err error
	for tries := 0; tries < 100; tries++ {
		usr.Discriminator = fmt.Sprintf("%04d", dint)
		err = c.Insert(usr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}
	return usr, nil
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

func GetUserByEmail(email string) (*User, error) {
	var u2 User
	c := DB.Core.C("users")
	err := c.Find(bson.M{"email": email}).One(&u2)
	if err != nil {
		return nil, err
	}
	return &u2, nil
}

// GetUserByToken returns a user using an authentication token
func GetUserByToken(token string) (*User, error) {
	if *flgStaging {
		// TODO: snowflake.ParseString
		var i snowflake.ID
		n, err := fmt.Sscanf(token, "%d", &i)
		if n != 1 {
			return nil, fmt.Errorf("Bad ID")
		}
		if err != nil {
			return nil, err
		}
		return GetUserByID(i)
	} else {
		claims, err := jwt.Parse(token)
		if err != nil {
			return nil, err
		}
		if err := claims.Validate(); err != nil {
			return nil, err
		}
		subj, err := claims.GetSubject()
		if err != nil {
			return nil, err
		}
		uid, err := snowflake.ParseString(subj)
		if err != nil {
			return nil, err
		}
		user, err := GetUserByID(uid)
		if err != nil {
			return nil, err
		}
		if !jwt.Verify(token, user.GetTokenSecret()) {
			return nil, fmt.Errorf("Invalid token")
		}
		return user, nil
	}
	return nil, fmt.Errorf("Not implemented")
}

// GetUserByHttpRequest returns a user using a fasthttp.RequestCtx.
// Specifically, it attempts to authorize the request using a token.
func GetUserByHttpRequest(c *fasthttp.RequestCtx, ctxvar string) (*User, error) {
	b := c.Request.Header.Peek("Authorization")
	if b == nil {
		return nil, fmt.Errorf("No authorization token supplied")
	}
	a := string(b)
	user, err := GetUserByToken(a)
	if err != nil {
		return nil, err
	}
	if ctxvar != "" {
		uid2 := c.UserValue(ctxvar).(string) // You're going to pass a string, or I'll panic()
		if uid2 == "" || uid2 == "@me" {
			return user, nil
		}
		snow, err := snowflake.ParseString(uid2)
		if err != nil {
			return nil, err
		}
		user, err = GetUserByID(snow)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (u *User) GetTokenSecret() []byte {
	return []byte(u.PasswordHash + u.JwtSecret)
}

func (u *User) IssueToken(duration time.Duration) string {
	c := jwt.New()
	c.SetSubject(u.ID.String())
	c.Set("tag", u.Username+"#"+u.Discriminator)
	c.SetIssuedAt(time.Now())
	c.SetExpiresAt(time.Now().Add(duration))
	c.SetTokenID()
	return c.Generate(u.GetTokenSecret())
}

// ToAPI returns a version of the User struct that can be returned by API calls
func (u *User) ToAPI(safe bool) *APITypeUser {
	u2 := &APITypeUser{
		ID:            u.ID,
		Username:      u.Username,
		Discriminator: u.Discriminator,
		AvatarHash:    u.Avatar,
		Bot:           u.Bot,
		MfaEnabled:    true,
		Flags:         u.Flags,
		PremiumType:   u.PremiumType,
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

func (u *User) StartTyping(c *Channel) error {
	return StartTypingForUser(u.ID, &gwEvtDataTypingStart{
		ChannelID: c.ID,
		GuildID:   c.GuildID,
		UserID:    u.ID,
		Timestamp: time.Now().Unix(),
	})
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
	err := DB.Core.C("channels").Find(bson.M{"recipient_ids": u.ID}).All(&ch)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (u *User) Save() error {
	c := DB.Core.C("users")
	return c.UpdateId(u.ID, bson.M{"$set": u})
}

func (u *User) SetTag(username, discriminator string) error {
	c := DB.Core.C("users")
	if discriminator == "" {
		dint := 1 + rand.Intn(9998)
		var err error
		for tries := 0; tries < 100; tries++ {
			dscm := fmt.Sprintf("%04d", dint)
			err = c.UpdateId(u.ID, bson.M{"username":username,"discriminator":dscm})
			if err == nil {
				u.Discriminator = dscm
				u.Username = username
				return nil
			}
		}
		return err
	} else {
		err := c.UpdateId(u.ID, bson.M{"username":username,"discriminator":discriminator})
		if err == nil {
			u.Username = username
			u.Discriminator = discriminator
		}
		return err
	}
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
		ID:            42,
		Username:      "test1",
		Discriminator: "1234",
		Email:         "test@localhost",
		PasswordHash:  util.CryptPass("hello"),
		Flags:         USER_FLAG_STAFF | USER_FLAG_EARLYADOPTER,
		Settings: &UserSettings{
			Locale: "en-US",
		},
	})
	c.Insert(&User{
		ID:            43,
		Username:      "hello",
		Discriminator: "4242",
		Email:         "test2@localhost",
		PasswordHash:  util.CryptPass("hello"),
		Flags:         USER_FLAG_EARLYADOPTER,
		Settings: &UserSettings{
			Locale: "en-US",
		},
	})
}
