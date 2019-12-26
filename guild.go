package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

// Represents a currently unavailable guild
// Used in GW_EVT_READY
type UnavailableGuild struct {
	ID snowflake.ID `json:"id"`
	Unavailable bool `json:"unavailable"`
}

type GuildMember struct {
	UserID snowflake.ID `bson:"user"`
	Nick string `bson:"nick"`
	Roles []snowflake.ID `bson:"roles"`
	JoinedAt int64 `bson:"joined_at"`
	PremiumSince int64 `bson:"premium_since"`
	Deaf bool `bson:"deaf"`
	Mute bool `bson:"mute"`
}

func (gm *GuildMember) ToAPI() *APITypeGuildMember {
	u, _ := GetUserByID(gm.UserID)
	return &APITypeGuildMember{u.ToAPI(true), gm.Nick, gm.Roles, time.Unix(gm.JoinedAt, 0), gm.Deaf, gm.Mute}
}

type Role struct {
	ID snowflake.ID `bson:"id"`
	Name string `bson:"name"`
	Color int `bson:"color"`
	Hoist bool `bson:"hoist"`
	Position int `bson:"position"`
	Permissions PermSet `bson:"permission"`
	Managed bool `bson:"managed"`
	Mentionable bool `bson:"mentionable"`
}

type Guild struct {
	ID snowflake.ID `bson:"_id"`
	Name string `bson:"name"`
	Icon string `bson:"icon"`
	Splash string `bson:"splash"`

	OwnerID snowflake.ID `bson:"owner_id"`
	Region string `bson:"region"`
	/* AFK Channels are deprecated and will never be supported */

	EmbedEnabled bool `bson:"embed_enabled"`
	EmbedChannelID snowflake.ID `bson:"embed_channel_id"`

	VerificationLevel int `bson:"verification_level"`
	DefaultMessageNotifications int `bson:"default_message_notifications"`
	ExplicitContentFilter int `bson:"explicit_content_filter"`

	Roles []*Role `bson:"roles"`
	Emojis []*Emoji `bson:"emojis"`
	Members map[snowflake.ID]*GuildMember `bson:"members"`
	Features []string `bson:"features"`
	MfaLevel int `bson:"mfa_level"`
	ApplicationID snowflake.ID `bson:"application_id"`

	WidgetEnabled bool `bson:"widget_enabled"`
	WidgetChannelID snowflake.ID `bson:"widget_channel_id"`
	SystemChannelID snowflake.ID `bson:"system_channel_id"`

	Large bool `bson:"large"`

	Description string `bson:"description"`
	Banner string `bson:"banner"`
	PremiumTier int `bson:"premium_tier"`
	PreferredLocale string `bson:"preferred_locale"`

}

func GetGuildByID(ID snowflake.ID) (*Guild, error) {
	var g2 Guild
	c := DB.Core.C("guilds")
	err := c.Find(bson.M{"_id": ID}).One(&g2)
	if err != nil {
		return nil, err
	}
	return &g2, nil
}

func GetGuildsByUserID(UserID snowflake.ID) ([]*Guild, error) {
	var g2 []*Guild
	c := DB.Core.C("guilds")
	err := c.Find(bson.M{"members.user": UserID}).All(&g2)
	if err != nil {
		return nil, err
	}
	return g2, nil
}

func (g *Guild) AddMember(UserID snowflake.ID, checkBans bool) error {
	_ = checkBans // TODO: use this
	c := DB.Core.C("guilds")
	if _, ok := g.Members[UserID]; ok {
		return &APIResponseError{0, "User has already joined guild"}
	}
	gm := &GuildMember{UserID: UserID, JoinedAt: time.Now().Unix()}
	err := c.UpdateId(g.ID, bson.M{"$set": bson.M{"members." + UserID.String(): gm}})
	if err != nil {
		return err
	}
	g.Members[UserID] = gm
	return nil
}

func (g *Guild) GetMember(UserID snowflake.ID) (*GuildMember, error) {
	m, ok := g.Members[UserID]
	if !ok {
		return nil, &APIResponseError{APIERR_UNKNOWN_MEMBER, "The member specified does not exist"}
	}
	return m, nil
}

func (g *Guild) Channels() ([]*Channel, error) {
	return GetChannelsByGuild(g.ID)
}

func (g *Guild) ToAPI(options ...interface{} /* UserID snowflake.ID */) *APITypeGuild {
	var oUid snowflake.ID
	if len(options) > 0 {
		oUid = options[0].(snowflake.ID)
	}
	out := &APITypeGuild{
		ID: g.ID,
		Name: g.Name,
		Icon: g.Icon,
		Splash: g.Splash,
		Owner: oUid == g.OwnerID,
		OwnerID: g.OwnerID,
		Permissions: 0, // TODO
		Region: g.Region,
		DefaultMessageNotifications: g.DefaultMessageNotifications,
		ExplicitContentFilter: g.ExplicitContentFilter,
		Features: g.Features,
		MfaLevel: g.MfaLevel,
		ApplicationID: g.ApplicationID,
		SystemChannelID: g.SystemChannelID,
		Description: g.Description,
		Banner: g.Banner,
		PremiumTier: g.PremiumTier,
		PreferredLocale: g.PreferredLocale,
		MaxPresences: 5000,
	}
	if oUid != 0 {
		out.JoinedAt = time.Unix(g.Members[oUid].JoinedAt, 0)
	}
	out.Members = []*APITypeGuildMember{}
	for _, v := range g.Members {
		out.Members = append(out.Members, v.ToAPI())
	}

	out.Channels = []APITypeAnyChannel{}
	gchs, err := g.Channels()
	if err != nil {
		// ????
		panic("Unexpected error: can't access list of guild channels!")
	}
	for _, v := range gchs {
		out.Channels = append(out.Channels, v.ToAPI())
	}

	return out
}

func InitGuildStaging() {
	c := DB.Core.C("guilds")
	c.Insert(&Guild{
		ID: 84,
		Name: "A test",
		OwnerID: 42,
	})
}
