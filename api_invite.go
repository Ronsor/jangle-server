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
		code, err := snowflake.ParseBase32([]byte(c.UserValue("code").(string)))
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

	r.POST("/invites/:code", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		code, err := snowflake.ParseBase32([]byte(c.UserValue("code").(string)))
		if err != nil {
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		inv, err := GetInviteByID(code)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_BAD_REQUEST)
			return
		}

		g, err := GetGuildByID(inv.GuildID)
		if err != nil {
			panic(err)
		}

		if _, err := g.GetMember(me.ID); err == nil {
			util.NoContentJSON(c)
			return
		}

		err = g.AddMember(me.ID, true)
		if err != nil {
			if err.Error() == "BANNED" {
				util.WriteJSONStatus(c, 403, APIERR_YOUREBANNEDCREEP)
				return
			}
			panic(err)
		}

		util.WriteJSON(c, inv.ToAPI())
	}, RL_NEWOBJ)))

	r.DELETE("/invites/:code", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		code, err := snowflake.ParseBase32([]byte(c.UserValue("code").(string)))
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
	}, RL_DELOBJ)))
}
