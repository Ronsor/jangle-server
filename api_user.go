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

	r.GET("/users/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	type APIReqPatchUsersUid struct {
		Username *string `json:"username" validate:"omitempty,min=2,max=32"`
		Discriminator *string `json:"discriminator" validate:"omitempty,len=4"`
		Password string `json:"password" validate:"min=1"`

		Email *string `json:"email" validate:"omitempty,email"`
		NewPassword *string `json:"new_password" validate:"omitempty,min=6"`
		Avatar *string `json:"avatar" validate:"omitempty,datauri"`
	}

	r.PATCH("/users/:uid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)

		var req APIReqPatchUsersUid
		if err := util.ReadPostJSON(c, &req); err != nil {
			println(err.Error())
			util.WriteJSONStatus(c, 400, APIERR_BAD_REQUEST)
			return
		}

		if !util.VerifyPass(me.PasswordHash, req.Password) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_ACCESS)
			return
		}

		if req.Username != nil {
			err := me.SetTag(*req.Username, "")
			if err != nil {
				panic(err)
			}
		}

		if req.Email != nil {
			me.Email = *req.Email
		}

		if req.NewPassword != nil {
			me.PasswordHash = util.CryptPass(*req.NewPassword)
		}

		if req.Avatar != nil {
			err := me.SetAvatar(*req.Avatar)
			if err != nil {
				panic(err)
			}
		}

		err := me.Save()
		if err != nil { panic(err) }

		util.WriteJSON(c, me.ToAPI(false))
	}, RL_SETINFO), "uid"))

	r.GET("/users/:uid/settings", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		util.WriteJSON(c, me.Settings)
	}, RL_GETINFO), "uid"))

	type APIReqPostUsersUidChannels struct {
		RecipientID snowflake.ID `json:"recipient_id"`
	}

	r.POST("/users/:uid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.GET("/users/:uid/channels", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		// Something *should* be done here... and now I forget what
		chs, err := me.Channels()
		if err != nil {
			panic(err)
		}
		out := make([]*APITypeDMChannel, 0, len(chs))
		for _, v := range chs {
			if v.Type == CHTYPE_DM {
				out = append(out, v.ToAPI(me).(*APITypeDMChannel))
			}
		}
		util.WriteJSON(c, out)
	}, RL_GETINFO), "uid"))

	r.GET("/users/:uid/guilds", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.DELETE("/users/:uid/guilds/:gid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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
