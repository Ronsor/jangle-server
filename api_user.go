package main

import (
	"log"

	"jangled/util"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"
	"github.com/bwmarrin/snowflake"
)

func InitRestUser(r *router.Router) {
	log.Println("Init /users Endpoints")

	r.GET("/api/v6/users/:uid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		uid := c.UserValue("uid").(string)
		if uid == "@me" {
			util.WriteJSON(c, me.ToAPI(false))
		} else {
			usnow, err := snowflake.ParseString(uid)
			if err != nil {
				util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_USER, "User does not exist"})
				return
			}
			user, err := GetUserByID(usnow)
			if err != nil {
				util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_USER, "User does not exist"})
				return
			}
			util.WriteJSON(c, user.ToAPI(true))
		}
	}))

	r.GET("/api/v6/users/:uid/settings", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		util.WriteJSON(c, me.Settings)
	}, "uid"))

	type APIReqPostUsersUidChannels struct {
		RecipientID snowflake.ID `json:"recipient_id"`
	}

	r.POST("/api/v6/users/:uid/channels", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		var req APIReqPostUsersUidChannels
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}
		rcp, err := GetUserByID(req.RecipientID)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_USER, "User does not exist"})
			return
		}

		ch, err := CreateDMChannel(me.ID, rcp.ID)
		if err != nil {
			util.WriteJSONStatus(c, 500, &APIResponseError{0, "Unknown error"})
			return
		}
		util.WriteJSON(c, ch.ToAPI().(*APITypeDMChannel))
	}, "uid"))

	r.GET("/api/v6/users/:uid/channels", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		// Something *should* be done here
		chs, err := me.Channels()
		if err != nil {
			util.WriteJSONStatus(c, 500, &APIResponseError{0, "Unknown error"})
		}
		out := make([]*APITypeDMChannel, 0, len(chs))
		for _, v := range chs {
			if v.Type == CHTYPE_DM {
				out = append(out, v.ToAPI().(*APITypeDMChannel))
			}
		}
		util.WriteJSON(c, out)
	}, "uid"))

	r.GET("/api/v6/users/:uid/guilds", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		guilds, err := me.Guilds()
		if err != nil {
			util.WriteJSONStatus(c, 500, &APIResponseError{0, "Unknown error"})
			return
		}
		out := make([]*APITypeGuild, 0, len(guilds))
		for _, v := range guilds {
			out = append(out, v.ToAPI(me.ID, false))
		}
		util.WriteJSON(c, out)
	}, "uid"))
}
