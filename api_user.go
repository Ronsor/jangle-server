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

	type APIRespGetUsersUidSettings *UserSettings

	r.GET("/api/v6/users/:uid/settings", MiddleTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		util.WriteJSON(c, APIRespGetUsersUidSettings(me.Settings))
	}, "uid"))

	type APIReqPostUsersUidChannels struct {
		RecipientID snowflake.ID `json:"recipient_id"`
	}


	/*r.POST("/api/v6/users/:uid/channels", MiddleTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		var req APIReqPostUsersUidChannels
		if util.PostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}
		rcp, err := GetUserByID(req.RecipientID)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_USER, "User does not exist"})
			return
		}
		// Something *should* be done here
		chs := rcp.Channels()
		out := make([]*APITypeDMChannel, 0, len(chs))
		for k, v := range chs {
			if v.Type == CHTYPE_DM {
				out = append(out, v.ToAPI().(*APITypeDMChannel))
			}
		}
		util.WriteJSON(c, out)
	}, "uid"))*/

	r.GET("/api/v6/users/:uid/channels", MiddleTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		// Something *should* be done here
		chs := me.Channels()
		out := make([]*APITypeDMChannel, 0, len(chs))
		for _, v := range chs {
			if v.Type == CHTYPE_DM {
				out = append(out, v.ToAPI().(*APITypeDMChannel))
			}
		}
		util.WriteJSON(c, out)
	}, "uid"))
}
