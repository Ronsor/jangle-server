package main

import (
	"log"

	"jangled/util"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"
	"github.com/bwmarrin/snowflake"
)

func InitRestGuild(r *router.Router) {
	log.Println("Init /guilds Endpoints")

	r.GET("/api/v6/guilds/:gid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
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
		util.WriteJSON(c, g.ToAPI(me.ID))
	}))
}
