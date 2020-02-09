package main

import (
	"fmt"
	"time"
	"path"
	"strings"
	"crypto/md5"

	"github.com/bwmarrin/snowflake"
	"github.com/globalsign/mgo"
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
	ID           snowflake.ID   `bson:"_id"`
	GuildID      snowflake.ID   `bson:"guild_id"`
	UserID       snowflake.ID   `bson:"user"`
	Nick         string         `bson:"nick"`
	Roles        []snowflake.ID `bson:"roles"`
	JoinedAt     int64          `bson:"joined_at"`
	PremiumSince int64          `bson:"premium_since"`
	Deaf         bool           `bson:"deaf"`
	Mute         bool           `bson:"mute"`
	Deleted      *time.Time     `bson:"deleted,omitempty"`
}

func GetGuildMembersByUserID(userID snowflake.ID) ([]*GuildMember, error) {
	var gm1 []*GuildMember
	gmc := DB.Core.C("guildmembers")
	err := gmc.Find(bson.M{"user": userID}).All(&gm1)
	if err != nil {
		return nil, err
	}
	return gm1, nil
}

func GetGuildMemberByID(id snowflake.ID) (*GuildMember, error) {
	var gm GuildMember
	c := DB.Core.C("guildmembers")
	err := c.Find(bson.M{"_id": id}).One(&gm)
	if err != nil {
		return nil, err
	}
	return &gm, nil
}

func GetGuildMemberByUserAndGuildID(userID, guildID snowflake.ID) (*GuildMember, error) {
	var gm GuildMember
	c := DB.Core.C("guildmembers")
	err := c.Find(bson.M{"user": userID, "guild_id": guildID}).One(&gm)
	if err != nil {
		return nil, err
	}
	return &gm, nil
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
	if id == gm.GuildID {
		return
	}
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
	FirstTime   bool         `bson:"first_time"`
}

func CanSetRoles(has []*Role, set []*Role) bool {
	var bestA, bestB *Role
	for _, v := range has {
		if v.Position > bestA.Position {
			bestA = v
		}
	}
	for _, v := range set {
		if v.Position > bestB.Position {
			bestB = v
		}
	}
	if bestA == nil {
		return false
	}
	if bestB == nil {
		return true
	}
	return bestA.Position > bestB.Position
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

	Roles  map[string]*Role `bson:"roles"`
	Emojis []*Emoji         `bson:"emojis"`
	//Members       map[string]*GuildMember `bson:"members"` // We don't use this now as it is too inefficient
	Features      []string     `bson:"features"`
	MfaLevel      int          `bson:"mfa_level"`
	ApplicationID snowflake.ID `bson:"application_id"`

	WidgetEnabled   bool         `bson:"widget_enabled"`
	WidgetChannelID snowflake.ID `bson:"widget_channel_id"`
	SystemChannelID snowflake.ID `bson:"system_channel_id"`

	Large bool `bson:"large"`

	Description     string `bson:"description"`
	Tags []string `bson:"tags"`
	Banner          string `bson:"banner"`
	PremiumTier     int    `bson:"premium_tier"`
	PreferredLocale string `bson:"preferred_locale"`

	NSFW bool `bson:"nsfw"`
}

func CreateGuild(u *User, g *Guild) (*Guild, error) {
	g.ID = flake.Generate()
	g.OwnerID = u.ID
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
	gCache.Set(g.ID, *g)
	ch, err := g.CreateChannel(&Channel{
		Name: "general",
		Type: CHTYPE_GUILD_TEXT,
	})
	if err != nil {
		return nil, err
	}
	g.SystemChannelID = ch.ID
	g.Save()
	err = g.AddMember(u.ID, false)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func GetGuildByID(id snowflake.ID) (*Guild, error) {
	var g2 Guild
	if g, ok := gCache.Get(id); ok {
		g2 = g.(Guild)
		return &g2, nil
	}
	c := DB.Core.C("guilds")
	err := c.Find(bson.M{"_id": id}).One(&g2)
	if err != nil {
		return nil, err
	}
	gCache.Set(id, g2)
	return &g2, nil
}

type GuildSearchQuery struct {
	MustFeatures []string
	MustTags []string
	Before snowflake.ID
	Limit int
	SortBy string
}

func GetGuildsBySearchQuery(gsq GuildSearchQuery) ([]*Guild, error) {
	c := DB.Core.C("guilds")
	realquery := bson.M{
		"features": bson.M{"$all": append([]string{}, gsq.MustFeatures...)},
		"tags": bson.M{"$all": append([]string{}, gsq.MustTags...)},
	}

	if gsq.Before != 0 {
		realquery["_id"] = bson.M{"$lt": gsq.Before}
	}

	if len(gsq.MustTags) == 0 { delete(realquery, "tags") }
	if len(gsq.MustFeatures) == 0 { delete(realquery, "features") }

	fmt.Println(realquery)

	found := c.Find(realquery)

	if gsq.Limit != 0 {
		found = found.Limit(gsq.Limit)
	} else {
		found = found.Limit(20)
	}

	if gsq.SortBy != "" {
		found = found.Sort("-" + gsq.SortBy)
	} else {
		found = found.Sort("-_id")
	}

	var out []*Guild

	err := found.All(&out)

	if err != nil && err != mgo.ErrNotFound { return nil, err }

	return out, nil
}

func GetGuildsByUserID(UserID snowflake.ID) ([]*Guild, error) {
	var gm1 []*GuildMember
	gmc := DB.Core.C("guildmembers")
	err := gmc.Find(bson.M{"user": UserID}).All(&gm1)
	if err != nil {
		return nil, err
	}
	ids := []snowflake.ID{}
	for _, v := range gm1 {
		ids = append(ids, v.GuildID)
	}
	var g2 []*Guild
	c := DB.Core.C("guilds")
	err = c.Find(bson.M{"_id": bson.M{"$in": ids}}).All(&g2)
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
	gmc := DB.Core.C("guildmembers")
	_, err = gmc.RemoveAll(bson.M{"guild_id": g.ID})
	if err != nil {
		return err
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

func (g *Guild) DelFeature(feat string) error {
	for k, v := range g.Features {
		if v != feat {
			continue
		}
		g.Features[k] = g.Features[len(g.Features)-1]
		g.Features = g.Features[:len(g.Features)-1]
	}
	return g.Save(false, true)
}

func (g *Guild) HasFeature(feat string) bool {
	for _, v := range g.Features {
		if v == feat {
			return true
		}
	}
	return false
}

func (g *Guild) SetMember(extra *GuildMember, opts ...interface{} /* checkBans, noUpsert, announce bool */) error {
	c := DB.Core.C("guildmembers")
	//var om GuildMember
	//if err := c.Find(bson.M{"user": extra.UserID, "guild_id": g.ID}).One(&om); err == nil {
	//	return fmt.Errorf("Member already joined")
	//}
	gm := extra
	if extra.ID == 0 {
		gm = &GuildMember{ID: flake.Generate(), GuildID: g.ID, UserID: extra.UserID, JoinedAt: time.Now().Unix(), Roles: []snowflake.ID{g.ID}}
	}
	var err error
	// TODO: actually check bans if requested
	if len(opts) > 1 && opts[1].(bool) {
		err = c.Remove(bson.M{"user": gm.UserID, "guild_id": g.ID, "deleted": bson.M{"$exists": true}})
		if err == nil || err == mgo.ErrNotFound {
			err = c.Insert(gm)
		}
	} else {
		_, err = c.Upsert(bson.M{"user": gm.UserID, "guild_id": g.ID}, gm)
	}
	if len(opts) > 2 && opts[2].(bool) && err == nil && g.SystemChannelID > 0 {
		ch, err := GetChannelByID(g.SystemChannelID)
		if err == nil {
			ch.CreateMessage(&Message{
				Author:  &User{ID: gm.UserID},
				Type:    MSGTYPE_GUILD_MEMBER_JOIN,
				Content: "@self has joined",
			})
		}
	}
	return err
}

func (g *Guild) AddMember(userID snowflake.ID, checkBans bool) error {
	return g.SetMember(&GuildMember{UserID: userID}, checkBans, true, true)
}

func (g *Guild) HasMember(userID snowflake.ID) bool {
	_, err := g.GetMember(userID)
	return err == nil
}

func (g *Guild) GetMember(userID snowflake.ID) (*GuildMember, error) {
	c := DB.Core.C("guildmembers")
	var m GuildMember
	err := c.Find(bson.M{"user": userID, "guild_id": g.ID, "deleted": bson.M{"$exists": false}}).One(&m)
	if m.Deleted != nil { panic("We can't do this!") }
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (g *Guild) Members(limit int, after snowflake.ID) ([]*GuildMember, error) {
	var o []*GuildMember
	wholequery := bson.M{"guild_id": g.ID}
	if after != 0 {
		wholequery["user"] = bson.M{"$gt": after}
	}
	c := DB.Core.C("guildmembers")
	resp := c.Find(wholequery).Sort("user")
	if limit == 0 {
		limit = 100
	}
	if limit > 0 {
		resp = resp.Limit(limit)
	}
	err := resp.All(&o)
	if err != nil {
		return nil, err
	}
	return o, err
}

func (g *Guild) CountMembers() int {
	c := DB.Core.C("guildmembers")
	ct, _ := c.Find(bson.M{"guild_id": g.ID}).Count()
	return ct
}

func (g *Guild) DelMember(UserID snowflake.ID) error {
	c := DB.Core.C("guildmembers")
	err := c.Update(bson.M{"user": UserID, "guild_id": g.ID}, bson.M{"$set":bson.M{"deleted": time.Now()}})
	return err
}

func (g *Guild) GetPermissions(u *User) PermSet {
	if g.OwnerID == u.ID || u.Flags&USER_FLAG_STAFF != 0 {
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
	err := c.Insert(ch)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (g *Guild) SetIcon(dataURL string) error {
	c := DB.Core.C("guilds")
	imgFp := fmt.Sprintf("%x", md5.Sum([]byte(dataURL)))
	fullpath, err := ImageDataURLUpload(gFileStore, "/icons/" + g.ID.String() + "/" + imgFp + ".png", dataURL, ImageUploadOptions{MaxWidth: 1024, MaxHeight: 1024, ForcePNG: true})
	if err != nil { return err }
	bp := path.Base(fullpath)
	g.Icon = strings.TrimRight(bp, path.Ext(bp))
	c.UpdateId(g.ID, bson.M{"$set":bson.M{"icon": bp}})
	return nil
}

func (g *Guild) AddRole(r *Role) error {
	if r.ID == 0 {
		r.ID = flake.Generate()
		r.FirstTime = true // a stupid hack, but who cares. I have to be done with this.
	} else {
		r.FirstTime = false
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

func (g *Guild) CanSetRoles(has []snowflake.ID, set []snowflake.ID) bool {
	var has2, set2 []*Role
	for _, v := range has {
		role, err := g.GetRole(v)
		if err != nil {
			return false
		}
		has2 = append(has2, role)
	}
	for _, v := range set {
		role, err := g.GetRole(v)
		if err != nil {
			return false
		}
		set2 = append(set2, role)
	}
	return CanSetRoles(has2, set2)
}

func (g *Guild) DelRole(id snowflake.ID) error {
	// probably not the best idea, but the fastest way to remove a dead role
	// alternatively drop it on next update, but meh
	if id == g.ID {
		return APIERR_UNKNOWN_ROLE
	}
	gmc := DB.Core.C("guildmembers")
	_, err := gmc.UpdateAll(bson.M{"guild_id": g.ID}, bson.M{"$pull": bson.M{"roles": id}})
	if err != nil {
		return err
	}
	c := DB.Core.C("guilds")
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
		panic("This option is no longer supported.")
	}
	if len(flags) > 1 && flags[1] {
		err := c.UpdateId(g.ID, bson.M{"$set": bson.M{"features": g.Features}})
		if err != nil {
			return err
		}
	}
	if len(flags) == 0 {
		return c.UpdateId(g.ID, bson.M{"$set": g})
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
		NSFW:                        g.NSFW,
		Region:                      g.Region,
		DefaultMessageNotifications: g.DefaultMessageNotifications,
		ExplicitContentFilter:       g.ExplicitContentFilter,
		Features:                    g.Features,
		MfaLevel:                    g.MfaLevel,
		ApplicationID:               g.ApplicationID,
		SystemChannelID:             g.SystemChannelID,
		Description:                 g.Description,
		Tags: g.Tags,
		Banner:                      g.Banner,
		PremiumTier:                 g.PremiumTier,
		PreferredLocale:             g.PreferredLocale,
		MemberCount:                 g.CountMembers(),
		MaxPresences:                4000,
		VoiceStates:                 []interface{}{},
	}

	if out.Tags == nil { g.Tags = []string{} }

	if oUid != 0 {
		if memb, err := g.GetMember(oUid); err == nil {
			out.JoinedAt = time.Unix(memb.JoinedAt, 0)
		}
	}

	out.Presences = []*APITypePresenceUpdate{}
	if forCreateEvent {
		outmems := []*APITypeGuildMember{}
		mems, err := g.Members(-1, 0)
		if err != nil {
			panic("Unexpected error: can't access list of guild members")
		}
		for _, v := range mems {
			mem := v.ToAPI()
			outmems = append(outmems, mem)
			psn, _ := GetPresenceForUser(v.UserID)
			if psn != nil {
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
		out.Members = &outmems
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
	g := &Guild{
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
		//Members: map[string]*GuildMember{
		//	"42": &GuildMember{UserID: 42, Roles: []snowflake.ID{84}},
		//	"43": &GuildMember{UserID: 43, Roles: []snowflake.ID{84}},
		//},
		Features: []string{GUILD_FEATURE_DISCOVERABLE},
	}
	guilds.Insert(g)
	g.AddMember(42, false)
	g.AddMember(43, false)
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
