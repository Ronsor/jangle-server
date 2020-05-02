package main

import (
	"log"

	"jangled/util"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/bwmarrin/snowflake"
)

func InitRestInvite(r *router.Router) {
	log.Println("Init /invites endpoints")

	r.GET("/invites/:code", MwRl(func(c *fasthttp.RequestCtx) {
		code, err := snowflake.ParseString(c.UserValue("code").(string))
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		inv, err := GetInviteByID(code)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_BAD_REQUEST)
			return
		}

		util.WriteJSON(c, inv.ToAPI())
	}, RL_GETINFO))

	r.DELETE("/invites/:code", MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		code, err := snowflake.ParseString(c.UserValue("code").(string))
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		inv, err := GetInviteByID(code)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_BAD_REQUEST)
			return
		}

		if inv.Inviter != me.ID && me.Flags & USER_FLAG_STAFF == 0 {
			gd, err := GetGuildByID(inv.GuildID)
			if err != nil || !gd.GetPermissions(me).Has(PERM_MANAGE_GUILD) {
				util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
				return
			}
		}

		err = inv.Delete()
		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, inv.ToAPI())
	}, RL_DELOBJ))

}
