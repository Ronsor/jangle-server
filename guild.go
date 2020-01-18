package main

import (
	"fmt"
	"sort"
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

const (
	GUILD_FEATURE_BANNER        = "BANNER"
	GUILD_FEATURE_INVITE_SPLASH = "INVITE_SPLASH"
	GUILD_FEATURE_DISCOVERABLE  = "DISCOVERABLE"
	GUILD_FEATURE_PARTNERED     = "PARTNERED"
	GUILD_FEATURE_VANITY_URL    = "VANITY_URL"
	GUILD_FEATURE_VERIFIED      = "VERIFIED"
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

func (gm *GuildMember) AddRole(id snowflake.ID) {
	for _, v := range gm.Roles {
		if v == id {
			return
		}
	}
	gm.Roles = append(gm.Roles, id)
}

func (gm *GuildMember) HasRole(id snowflake.ID) bool {
	for _, v := range gm.Roles {
		if v == id {
			return true
		}
	}
	return false
}

func (gm *GuildMember) DelRole(id snowflake.ID) {
	for k, v := range gm.Roles {
		if v != id {
			continue
		}
		gm.Roles[k] = gm.Roles[len(gm.Roles)-1]
		gm.Roles = gm.Roles[:len(gm.Roles)-1]
	}
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

	Roles         map[string]*Role        `bson:"roles"`
	Emojis        []*Emoji                `bson:"emojis"`
	Members       map[string]*GuildMember `bson:"members"`
	Features      []string                `bson:"features"`
	MfaLevel      int                     `bson:"mfa_level"`
	ApplicationID snowflake.ID            `bson:"application_id"`

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
	g.Roles = map[string]*Role{
		g.ID.String(): &Role{
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

func (g *Guild) Delete() error {
	chs, err := g.Channels()
	if err != nil {
		return err
	}
	for _, v := range chs {
		err := v.Delete()
		if err != nil {
			return err
		}
	}
	c := DB.Core.C("guilds")
	err = c.RemoveId(g.ID)
	if err != nil {
		return err
	}
	return nil
}

func (g *Guild) AddFeature(feat string) error {
	for _, v := range g.Features {
		if v == feat {
			return nil
		}
	}
	g.Features = append(g.Features, feat)
	return g.Save(false, true /* featureList */)
}

func (g *Guild) HasFeature(feat string) bool {
	for _, v := range g.Features {
		if v == feat {
			return true
		}
	}
	return false
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

func (g *Guild) ListMembers(limit int, after snowflake.ID) ([]*GuildMember, error) {
	// TODO: this is super inefficient
	if limit == 0 {
		limit = 1
	}
	x := make([]*GuildMember, 0, len(g.Members))
	for _, v := range g.Members {
		x = append(x, v)
	}
	sort.SliceStable(x, func(i, j int) bool { return x[i].UserID < x[j].UserID })
	if after == 0 {
		if limit > len(x) {
			limit = len(x)
		}
		return x[:limit], nil
	}
	for k, v := range x {
		if v.UserID == after {
			x = x[k+1:]
			if limit > len(x) {
				limit = len(x)
			}
			return x[:limit], nil
		}
	}
	return nil, fmt.Errorf("ListMembers: ?")
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

func (g *Guild) AddRole(r *Role) error {
	if r.ID == 0 {
		r.ID = flake.Generate()
	}
	c := DB.Core.C("guilds")
	err := c.UpdateId(g.ID, bson.M{"$set": bson.M{"roles." + r.ID.String(): r}})
	if err != nil {
		return err
	}
	g.Roles[r.ID.String()] = r
	return nil
}

func (g *Guild) GetRole(id snowflake.ID) (*Role, error) {
	r, ok := g.Roles[id.String()]
	if !ok {
		return nil, APIERR_UNKNOWN_ROLE
	}
	return r, nil
}

func (g *Guild) DelRole(id snowflake.ID) error {
	c := DB.Core.C("guilds")
	for _, v := range g.Members {
		v.DelRole(id)
	}
	err := g.Save()
	if err != nil {
		return err
	}
	err = c.UpdateId(g.ID, bson.M{"$unset": bson.M{"roles." + id.String(): ""}})
	if err != nil {
		return err
	}
	delete(g.Roles, id.String())
	return nil
}

func (g *Guild) Save(flags ...bool /* membersOnly, featureList bool */) error {
	c := DB.Core.C("guilds")
	if len(flags) > 0 && flags[0] {
		err := c.UpdateId(g.ID, bson.M{"$set": bson.M{"members": g.Members}})
		if err != nil {
			return err
		}
	}
	if len(flags) > 0 && flags[1] {
		err := c.UpdateId(g.ID, bson.M{"$set": bson.M{"features": g.Features}})
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Guild) ToAPI(options ...interface{} /* UserID snowflake.ID, forCreateEvent bool */) *APITypeGuild {
	var oUid snowflake.ID
	var forCreateEvent = true
	var perm = PermSet(0)
	if len(options) > 0 {
		oUid = options[0].(snowflake.ID)
		perm = g.GetPermissions(&User{ID: oUid}) // This is an awful hack
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
		Permissions:                 perm, // TODO
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
			// Is panic() reasonable?
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
		Roles: map[string]*Role{
			"84": &Role{
				ID:          84,
				Name:        "@everyone",
				Permissions: GUILD_EVERYONE_DEFAULT_PERMS,
			},
		},
		Members: map[string]*GuildMember{
			"42": &GuildMember{UserID: 42, Roles: []snowflake.ID{84}},
			"43": &GuildMember{UserID: 43, Roles: []snowflake.ID{84}},
		},
		Features: []string{GUILD_FEATURE_DISCOVERABLE},
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
