package main

import (
	"log"
	"strconv"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func InitRestGuild(r *router.Router) {
	log.Println("Init /guilds Endpoints")

	r.GET("/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		util.WriteJSON(c, g.ToAPI(me.ID, false))
	}, RL_GETINFO)))

	type APIReqPatchGuildsGid struct {
		Name                        *string       `json:"name" validate:"min=2,max=100"`
		Region                      *string       `json:"region,omitempty"` // Ignored
		Icon                        *string       `json:"icon,omitempty" validate:"omitempty,datauri"`
		VerificationLevel           *int          `json:"verification_level,omitempty"`
		DefaultMessageNotifications *int          `json:"default_message_notifications,omitempty" validate:"min=0,max=1"`
		ExplicitContentFilter       *int          `json:"explicit_content_filter" validate:"min=0,max=0"`
		Public                      *bool         `json:"public"`
		NSFW                        *bool         `json:"nsfw"`
		OwnerID                     *snowflake.ID `json:"owner_id"`
		SystemChannelID             *snowflake.ID `json:"system_channel_id"`
		Features                    *[]string     `json:"features"`
	}

	r.PATCH("/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPatchGuildsGid
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}

		if req.Name != nil {
			g.Name = *req.Name
		}
		if req.VerificationLevel != nil {
			g.VerificationLevel = *req.VerificationLevel
		}
		if req.DefaultMessageNotifications != nil {
			g.DefaultMessageNotifications = *req.DefaultMessageNotifications
		}
		if req.NSFW != nil {
			g.NSFW = *req.NSFW
		}
		if req.Public != nil {
			if *req.Public {
				if err := g.AddFeature(GUILD_FEATURE_DISCOVERABLE); err != nil {
					panic(err)
				}
			} else {
				if err := g.DelFeature(GUILD_FEATURE_DISCOVERABLE); err != nil {
					panic(err)
				}
			}
		}
		if req.OwnerID != nil {
			if !g.HasMember(*req.OwnerID) {
				util.WriteJSONStatus(c, 400, APIERR_UNKNOWN_MEMBER)
				return
			}
			g.OwnerID = *req.OwnerID
		}
		if req.SystemChannelID != nil {
			if ch, err := GetChannelByID(*req.SystemChannelID); err != nil || ch.GuildID != snow {
				util.WriteJSONStatus(c, 400, APIERR_UNKNOWN_CHANNEL)
				return
			}
			g.SystemChannelID = *req.SystemChannelID
		}
		if req.Features != nil {
			if me.Flags&USER_FLAG_STAFF == 0 {
				util.WriteJSONStatus(c, 403, APIERR_MISSING_ACCESS)
				return
			}
			g.Features = *req.Features
		}

		err = g.Save()
		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, g)

	}, RL_SETINFO)))

	r.DELETE("/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}

		if g.OwnerID != me.ID && me.Flags&USER_FLAG_STAFF == 0 {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		err = g.Delete()
		if err != nil {
			panic(err)
		}
		c.SetStatusCode(204)
	}, RL_DELOBJ)))

	r.GET("/guilds/:gid/members", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		//me := c.UserValue("m:user").(*User)
		//TODO: check if user is in guild
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		after, _ := snowflake.ParseBytes(c.FormValue("after"))
		limit, _ := strconv.Atoi(string(c.FormValue("limit")))
		mem, err := g.Members(limit, after)
		if err != nil {
			panic(err)
		}
		o := []*APITypeGuildMember{}
		for _, v := range mem {
			o = append(o, v.ToAPI())
		}
		util.WriteJSON(c, o)
	}, RL_GETINFO)))

	r.GET("/guilds/:gid/members/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		gid := c.UserValue("gid").(string)
		uid := c.UserValue("uid").(string)
		usnow, err := snowflake.ParseString(uid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		p, err := g.GetMember(usnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MEMBER)
			return
		}
		util.WriteJSON(c, p.ToAPI())
	}, RL_GETINFO)))

	r.DELETE("/guilds/:gid/members/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		if !g.GetPermissions(me).Has(PERM_KICK_MEMBERS) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		err = g.DelMember(snow)

		if err != nil {
			panic(err)
		}
		c.SetStatusCode(204)
	}, RL_DELOBJ)))

	type APIReqPutGuildsGidMembersUid struct {
		// TODO: accept arguments here
	}

	r.PUT("/guilds/:gid/members/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		if !g.HasFeature(GUILD_FEATURE_DISCOVERABLE) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}
		if _, err := g.GetMember(me.ID); err != nil {
			c.SetStatusCode(204)
			return
		}
		err = g.AddMember(me.ID, true)
		if err != nil {
			panic(err)
		}
		mem, err := g.GetMember(me.ID)
		if err != nil {
			panic(err)
		}
		c.SetStatusCode(201)
		util.WriteJSON(c, mem.ToAPI())
	}, RL_NEWOBJ), "uid"))

	type APIReqPatchGuildsGidMembersUid struct {
		Nick  *string         `json:"nick,omitempty" validate:"min=0,max=32,omitempty"`
		Roles *[]snowflake.ID `json:"roles,omitempty" validate:"max=250,omitempty"`
	}

	r.PATCH("/guilds/:gid/members/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPatchGuildsGidMembersUid
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		gid := c.UserValue("gid").(string)
		uid := c.UserValue("uid").(string)

		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}

		usnow, err := snowflake.ParseString(uid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		mem, err := g.GetMember(usnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MEMBER)
			return
		}

		if req.Nick != nil && !g.GetPermissions(me).Has(PERM_MANAGE_NICKNAMES) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		} else if req.Nick != nil {
			mem.Nick = *req.Nick
		}

		if req.Roles != nil && !g.GetPermissions(me).Has(PERM_MANAGE_ROLES) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		} else if req.Roles != nil {
			// TODO: insecure
			mem.Roles = *req.Roles
		}

		err = g.SetMember(mem)

		if err != nil {
			panic(err)
		}

		c.SetStatusCode(204)
	}, RL_SETINFO)))

	type APIReqPatchGuildsGidMembersMeNick struct {
		Nick string `json:"nick" validate:"min=0,max=32"`
	}

	r.PATCH("/guilds/:gid/members/:uid/nick", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPatchGuildsGidMembersMeNick
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		if !g.GetPermissions(me).Has(PERM_CHANGE_NICKNAME) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		mem, err := g.GetMember(me.ID)
		if err != nil {
			panic(err)
		}
		mem.Nick = req.Nick
		err = g.SetMember(mem)
		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, req)
	}, RL_SETINFO), "uid"))

	r.GET("/guilds/:gid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		util.WriteJSON(c, g.ToAPI(me.ID).Channels)
	}, RL_GETINFO)))

	type APIReqPatchGuildsGidChannels []struct {
		ID       snowflake.ID `json:"id,string"`
		Position int          `json:"position"`
	}

	r.PATCH("/guilds/:gid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPatchGuildsGidChannels
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		if !g.GetPermissions(me).Has(PERM_MANAGE_CHANNELS) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		for _, v := range req {
			ch, err := GetChannelByID(v.ID)
			if err != nil || ch.GuildID != g.ID {
				util.WriteJSONStatus(c, 403, APIERR_UNKNOWN_CHANNEL)
				return
			}
			ch.Position = v.Position
			ch.Save()
		}

		c.SetStatusCode(204)
	}, RL_SETINFO)))

	type APIReqPostGuildsGidChannels struct {
		Name                 string                        `json:"name" validate:"min=1,max=64"`
		Type                 int                           `json:"type"`
		Topic                string                        `json:"topic" validate:"max=256"`
		Position             int                           `json:"position"`
		PermissionOverwrites []*APITypePermissionOverwrite `json:"permission_overwrites"`
		ParentID             snowflake.ID                  `json:"parent_id"`
		NSFW                 bool                          `json:"nsfw"`
	}

	r.POST("/guilds/:gid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPostGuildsGidChannels
		if err := util.ReadPostJSON(c, &req); err != nil {
			log.Println(err);
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_GUILD)
			return
		}
		if !g.GetPermissions(me).Has(PERM_MANAGE_CHANNELS) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		ch := &Channel{
			Name:     req.Name,
			Topic:    req.Topic,
			NSFW:     req.NSFW,
			Type:     req.Type,
			Position: req.Position,
			ParentID: req.ParentID,
		}

		if ch.Name == "" {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		if ch.Type == 0 {
			ch.Type = CHTYPE_GUILD_TEXT
		} else if ch.Type != CHTYPE_GUILD_TEXT && ch.Type != CHTYPE_GUILD_CATEGORY {
			util.WriteJSONStatus(c, 400, APIERR_UNKNOWN_CHANNEL)
			return
		}
		if ch.ParentID != 0 {
			pch, err := GetChannelByID(ch.ParentID)
			if err != nil {
				util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
				return
			}
			if pch.GuildID != g.ID {
				util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
				return
			}
			if pch.Type != CHTYPE_GUILD_CATEGORY {
				util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
				return
			}
		}

		curchs, err := g.Channels()

		if err != nil {
			panic(err)
		}

		if len(curchs) > LIMIT_MAX_GUILD_CHANNELS {
			util.WriteJSONStatus(c, 400, APIERR_MAX_GUILD_CHANNELS)
			return
		}

		ch, err = g.CreateChannel(ch)

		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, ch.ToAPI())
	}, RL_NEWOBJ)))

	type APIReqPostGuilds struct {
		Name                        string              `json:"name" validate:"min=2,max=100"`
		Region                      string              `json:"region,omitempty"` // Ignored
		Icon                        string              `json:"icon,omitempty" validate:"omitempty,datauri"`
		VerificationLevel           int                 `json:"verification_level,omitempty"`
		DefaultMessageNotifications int                 `json:"default_message_notifications,omitempty" validate:"min=0,max=1"`
		ExplicitContentFilter       int                 `json:"explicit_content_filter" validate:"min=0,max=0"`
		Roles                       []*APITypeRole      `json:"roles,omitempty"`    // Ignored
		Channels                    []APITypeAnyChannel `json:"channels,omitempty"` // Ignored
		Public                      bool                `json:"public,omitempty"`
	}

	r.POST("/guilds", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPostGuilds
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		gds, err := me.Guilds()

		if len(gds) > LIMIT_MAX_GUILDS {
			util.WriteJSONStatus(c, 400, APIERR_MAX_GUILDS)
		}

		feat := []string{}
		if req.Public {
			feat = append(feat, GUILD_FEATURE_DISCOVERABLE)
		}

		g, err := CreateGuild(me, &Guild{
			Name:     req.Name,
			Features: feat,
		})

		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, g.ToAPI(me.ID, true))
	}, RL_NEWOBJ)))
}
