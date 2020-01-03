package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo/bson"
)

const GUILD_EVERYONE_DEFAULT_PERMS = PermSet(104324161)

const (
	GUILD_MSG_NOTIFY_ALL           = 0
	GUILD_MSG_NOTIFY_ONLY_MENTIONS = 1
)

const (
	GUILD_EXPLICIT_FILTER_DISABLED = 0
)

// Represents a currently unavailable guild
// Used in GW_EVT_READY
type UnavailableGuild struct {
	ID          snowflake.ID `json:"id"`
	Unavailable bool         `json:"unavailable"`
}

type GuildMember struct {
	UserID       snowflake.ID   `bson:"user"`
	Nick         string         `bson:"nick"`
	Roles        []snowflake.ID `bson:"roles"`
	JoinedAt     int64          `bson:"joined_at"`
	PremiumSince int64          `bson:"premium_since"`
	Deaf         bool           `bson:"deaf"`
	Mute         bool           `bson:"mute"`
}

func (gm *GuildMember) ToAPI() *APITypeGuildMember {
	u, _ := GetUserByID(gm.UserID)
	roles := gm.Roles
	if roles == nil {
		roles = []snowflake.ID{}
	}
	return &APITypeGuildMember{u.ToAPI(true), gm.Nick, roles, time.Unix(gm.JoinedAt, 0), gm.Deaf, gm.Mute, 0}
}

type Role struct {
	ID          snowflake.ID `bson:"id"`
	Name        string       `bson:"name"`
	Color       int          `bson:"color"`
	Hoist       bool         `bson:"hoist"`
	Position    int          `bson:"position"`
	Permissions PermSet      `bson:"permission"`
	Managed     bool         `bson:"managed"`
	Mentionable bool         `bson:"mentionable"`
}

func (r *Role) ToAPI() *APITypeRole {
	x := APITypeRole(*r)
	return &x
}

type Guild struct {
	ID     snowflake.ID `bson:"_id"`
	Name   string       `bson:"name"`
	Icon   string       `bson:"icon"`
	Splash string       `bson:"splash"`

	OwnerID snowflake.ID `bson:"owner_id"`
	Region  string       `bson:"region"`
	/* AFK Channels are deprecated and will never be supported */

	EmbedEnabled   bool         `bson:"embed_enabled"`
	EmbedChannelID snowflake.ID `bson:"embed_channel_id"`

	VerificationLevel           int `bson:"verification_level"`
	DefaultMessageNotifications int `bson:"default_message_notifications"`
	ExplicitContentFilter       int `bson:"explicit_content_filter"`

	Roles         []*Role                       `bson:"roles"`
	Emojis        []*Emoji                      `bson:"emojis"`
	Members       map[string]*GuildMember `bson:"members"`
	Features      []string                      `bson:"features"`
	MfaLevel      int                           `bson:"mfa_level"`
	ApplicationID snowflake.ID                  `bson:"application_id"`

	WidgetEnabled   bool         `bson:"widget_enabled"`
	WidgetChannelID snowflake.ID `bson:"widget_channel_id"`
	SystemChannelID snowflake.ID `bson:"system_channel_id"`

	Large bool `bson:"large"`

	Description     string `bson:"description"`
	Banner          string `bson:"banner"`
	PremiumTier     int    `bson:"premium_tier"`
	PreferredLocale string `bson:"preferred_locale"`
}

func CreateGuild(u *User, g *Guild) (*Guild, error) {
	g.ID = flake.Generate()
	g.OwnerID = u.ID
	g.Members = map[string]*GuildMember{}
	g.Roles = []*Role{
		&Role{
			ID:          g.ID,
			Name:        "@everyone",
			Permissions: GUILD_EVERYONE_DEFAULT_PERMS,
		},
	}
	c := DB.Core.C("guilds")
	err := c.Insert(g)
	if err != nil {
		return nil, err
	}
	g.CreateChannel(&Channel{
		Name: "general",
		Type: CHTYPE_GUILD_TEXT,
	})
	err = g.AddMember(u.ID, false)
	if err != nil {
		return nil, err
	}
	return g, nil
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
	err := c.Find(bson.M{"members." + UserID.String(): bson.M{"$exists": true}}).All(&g2)
	if err != nil {
		return nil, err
	}
	return g2, nil
}

func (g *Guild) AddMember(UserID snowflake.ID, checkBans bool) error {
	_ = checkBans // TODO: use this
	c := DB.Core.C("guilds")
	if _, ok := g.Members[UserID.String()]; ok {
		return &APIResponseError{0, "User has already joined the guild"}
	}
	gm := &GuildMember{UserID: UserID, JoinedAt: time.Now().Unix(), Roles: []snowflake.ID{g.ID}}
	err := c.UpdateId(g.ID, bson.M{"$set": bson.M{"members." + UserID.String(): gm}})
	if err != nil {
		return err
	}
	g.Members[UserID.String()] = gm
	return nil
}

func (g *Guild) GetMember(UserID snowflake.ID) (*GuildMember, error) {
	m, ok := g.Members[UserID.String()]
	if !ok {
		return nil, APIERR_UNKNOWN_MEMBER
	}
	return m, nil
}

func (g *Guild) DelMember(UserID snowflake.ID) error {
	c := DB.Core.C("guilds")
	err := c.UpdateId(g.ID, bson.M{"$unset": bson.M{"members." + UserID.String(): ""}})
	if err != nil {
		return err
	}
	delete(g.Members, UserID.String())
	return nil
}

func (g *Guild) GetPermissions(u *User) PermSet {
	if g.OwnerID == u.ID {
		return PERM_EVERYTHING // PERM_ADMINISTRATOR?
	}
	mem, err := g.GetMember(u.ID)
	if err != nil {
		return 0
	}
	var uRoles = map[snowflake.ID]bool{}
	for _, v := range mem.Roles {
		uRoles[v] = true
	}
	var perm PermSet
	for _, v := range g.Roles {
		if uRoles[v.ID] || v.ID == g.ID { // (v.ID == g.ID) is @everyone. This is a special case.
			perm |= v.Permissions
		}
	}
	return perm // TODO: the rest
}

func (g *Guild) Channels() ([]*Channel, error) {
	return GetChannelsByGuild(g.ID)
}

func (g *Guild) CreateChannel(ch *Channel) (*Channel, error) {
	ch.ID = flake.Generate()
	ch.GuildID = g.ID
	c := DB.Core.C("channels")
	err := c.Insert(&ch)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (g *Guild) ToAPI(options ...interface{} /* UserID snowflake.ID, forCreateEvent bool */) *APITypeGuild {
	var oUid snowflake.ID
	var forCreateEvent = true
	if len(options) > 0 {
		oUid = options[0].(snowflake.ID)
	}
	if len(options) > 1 {
		forCreateEvent = options[1].(bool)
	}
	out := &APITypeGuild{
		ID:                          g.ID,
		Name:                        g.Name,
		Icon:                        g.Icon,
		Splash:                      g.Splash,
		Owner:                       oUid == g.OwnerID,
		OwnerID:                     g.OwnerID,
		Permissions:                 0, // TODO
		Region:                      g.Region,
		DefaultMessageNotifications: g.DefaultMessageNotifications,
		ExplicitContentFilter:       g.ExplicitContentFilter,
		Features:                    g.Features,
		MfaLevel:                    g.MfaLevel,
		ApplicationID:               g.ApplicationID,
		SystemChannelID:             g.SystemChannelID,
		Description:                 g.Description,
		Banner:                      g.Banner,
		PremiumTier:                 g.PremiumTier,
		PreferredLocale:             g.PreferredLocale,
		MemberCount:                 len(g.Members),
		MaxPresences:                5000,
		VoiceStates:                 []interface{}{},
	}
	if oUid != 0 && g.Members[oUid.String()] != nil {
		out.JoinedAt = time.Unix(g.Members[oUid.String()].JoinedAt, 0)
	}

	out.Members = []*APITypeGuildMember{}
	out.Presences = []*APITypePresenceUpdate{}
	if forCreateEvent {
		for _, v := range g.Members {
			mem := v.ToAPI()
			out.Members = append(out.Members, mem)
			psn, err := GetPresenceForUser(v.UserID)
			if err == nil {
				out.Presences = append(out.Presences, &APITypePresenceUpdate{
					User:    mem.User,
					Roles:   mem.Roles,
					GuildID: g.ID,
					Status:  psn.Status,
					Game:    nil,
					Nick:    mem.Nick,
				})
			}
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
	}

	out.Emojis = []*APITypeEmoji{}
	for _, v := range g.Emojis {
		out.Emojis = append(out.Emojis, v.ToAPI(false))
	}

	out.Roles = []*APITypeRole{}
	for _, v := range g.Roles {
		out.Roles = append(out.Roles, v.ToAPI())
	}

	return out
}

func InitGuildStaging() {
	guilds := DB.Core.C("guilds")
	guilds.Insert(&Guild{
		ID:      84,
		Name:    "A test",
		OwnerID: 42,
		Roles: []*Role{
			&Role{
				ID:          84,
				Name:        "@everyone",
				Permissions: GUILD_EVERYONE_DEFAULT_PERMS,
			},
		},
		Members: map[string]*GuildMember{
			"42": &GuildMember{UserID: 42, Roles: []snowflake.ID{84}},
			"43": &GuildMember{UserID: 43, Roles: []snowflake.ID{84}},
		},
	})
	chans := DB.Core.C("channels")
	chans.Insert(&Channel{
		ID:      85,
		GuildID: 84,
		Name:    "nonsense-chat",
		Topic:   "Ain't nothin' worth seein' here",
		NSFW:    false,
	})
	chans.Insert(&Channel{
		ID:      86,
		GuildID: 84,
		Name:    "silo-zone",
		Topic:   "Shhhhh",
		NSFW:    true,
	})
}
