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

	r.GET("/api/v6/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.DELETE("/api/v6/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		if g.OwnerID != me.ID {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		err = g.Delete()
		if err != nil {
			panic(err)
		}
		c.SetStatusCode(204)
	}, RL_DELOBJ)))

	r.GET("/api/v6/guilds/:gid/members", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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
		mem, err := g.ListMembers(limit, after)
		if err != nil {
			panic(err)
		}
		o := []*APITypeGuildMember{}
		for _, v := range mem {
			o = append(o, v.ToAPI())
		}
		util.WriteJSON(c, o)
	}, RL_GETINFO)))

	r.GET("/api/v6/guilds/:gid/members/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.GET("/api/v6/guilds/:gid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.PATCH("/api/v6/guilds/:gid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.POST("/api/v6/guilds/:gid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPostGuildsGidChannels
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
	}

	r.POST("/api/v6/guilds", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		g, err := CreateGuild(me, &Guild{
			Name: req.Name,
		})

		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, g.ToAPI(me.ID, true))
	}, RL_NEWOBJ)))
}
