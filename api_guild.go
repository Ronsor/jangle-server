package main

import (
	"log"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func InitRestGuild(r *router.Router) {
	log.Println("Init /guilds Endpoints")

	r.GET("/api/v6/guilds/:gid", MwTkA(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Invalid snowflake ID"})
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_GUILD, "The guild specified does not exist"})
			return
		}
		util.WriteJSON(c, g.ToAPI(me.ID, false))
	}))

	r.GET("/api/v6/guilds/:gid/channels", MwTkA(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Invalid snowflake ID"})
			return
		}
		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_GUILD, "The guild specified does not exist"})
			return
		}
		util.WriteJSON(c, g.ToAPI(me.ID).Channels)
	}))

	type APIReqPostGuildsGidChannels struct {
		Name                 string                        `json:"name"`
		Type                 int                           `json:"type"`
		Topic                string                        `json:"topic"`
		Position             int                           `json:"position"`
		PermissionOverwrites []*APITypePermissionOverwrite `json:"permission_overwrites"`
		ParentID             snowflake.ID                  `json:"parent_id"`
		NSFW                 bool                          `json:"nsfw"`
	}

	r.POST("/api/v6/guilds/:gid/channels", MwTkA(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPostGuildsGidChannels
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}

		gid := c.UserValue("gid").(string)
		snow, err := snowflake.ParseString(gid)
		if err != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Invalid snowflake ID"})
			return
		}

		g, err := GetGuildByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_GUILD, "The guild specified does not exist"})
			return
		}
		if !g.GetPermissions(me).Has(PERM_MANAGE_CHANNELS) {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_MISSING_PERMISSIONS, "Missing MANAGE_CHANNELS permission"})
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
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Missing channel name"})
			return
		}

		if ch.Type == 0 {
			ch.Type = CHTYPE_GUILD_TEXT
		} else if ch.Type != CHTYPE_GUILD_TEXT && ch.Type != CHTYPE_GUILD_CATEGORY {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Unsupported channel type"})
			return
		}
		if ch.ParentID != 0 {
			pch, err := GetChannelByID(ch.ParentID)
			if err != nil {
				util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Parent channel does not exist"})
				return
			}
			if pch.GuildID != g.ID {
				util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Parent channel does not exist"})
				return
			}
			if pch.Type != CHTYPE_GUILD_CATEGORY {
				util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Parent channel does not exist"})
				return
			}
		}

		ch, err = g.CreateChannel(ch)

		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, ch.ToAPI())
	}))

}
