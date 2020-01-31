package main

import (
	"log"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func InitRestUser(r *router.Router) {
	log.Println("Init /users Endpoints")

	r.GET("/api/v6/users/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		uid := c.UserValue("uid").(string)
		if uid == "@me" {
			util.WriteJSON(c, me.ToAPI(false))
		} else {
			usnow, err := snowflake.ParseString(uid)
			if err != nil {
				util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_USER)
				return
			}
			user, err := GetUserByID(usnow)
			if err != nil {
				util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_USER)
				return
			}
			util.WriteJSON(c, user.ToAPI(true))
		}
	}, RL_GETINFO)))

	r.GET("/api/v6/users/:uid/settings", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		util.WriteJSON(c, me.Settings)
	}, RL_GETINFO), "uid"))

	type APIReqPostUsersUidChannels struct {
		RecipientID snowflake.ID `json:"recipient_id"`
	}

	r.POST("/api/v6/users/:uid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		var req APIReqPostUsersUidChannels
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}
		rcp, err := GetUserByID(req.RecipientID)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_USER)
			return
		}

		ch, err := CreateDMChannel(me.ID, rcp.ID)
		if err != nil {
			panic(err)
		}
		util.WriteJSON(c, ch.ToAPI().(*APITypeDMChannel))
	}, RL_NEWOBJ), "uid"))

	r.GET("/api/v6/users/:uid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		// Something *should* be done here... and now I forget what
		chs, err := me.Channels()
		if err != nil {
			panic(err)
		}
		out := make([]*APITypeDMChannel, 0, len(chs))
		for _, v := range chs {
			if v.Type == CHTYPE_DM {
				out = append(out, v.ToAPI().(*APITypeDMChannel))
			}
		}
		util.WriteJSON(c, out)
	}, RL_GETINFO), "uid"))

	r.GET("/api/v6/users/:uid/guilds", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		guilds, err := me.Guilds()
		if err != nil {
			panic(err)
		}
		out := make([]*APITypeGuild, 0, len(guilds))
		for _, v := range guilds {
			out = append(out, v.ToAPI(me.ID, false))
		}
		util.WriteJSON(c, out)
	}, RL_GETINFO), "uid"))

	r.DELETE("/api/v6/users/:uid/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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
		// I have no idea what Discord's behavior is on this?
		// Delete the server or return an error?
		if g.OwnerID == me.ID {
			panic("I don't know what to do here")
		}

		err = g.DelMember(me.ID)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MEMBER)
		}

	}, RL_DELOBJ), "uid"))
}
