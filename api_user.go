package main

import (
	"log"

	"jangled/util"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"
	"github.com/bwmarrin/snowflake"
)

type APIReqPostUsersUidChannels struct {
	RecipientID snowflake.ID `json:"recipient_id"`
}

func InitRestUser(r *router.Router) {
	log.Println("Init /users Endpoints")
	r.GET("/api/v6/users/:uid/settings", MiddleTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		util.WriteJSON(c, me.Settings)
	}, "uid"))

	r.POST("/api/v6/users/:uid/channels", MiddleTokenAuth(func(c *fasthttp.RequestCtx) {
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
		_,_=rcp,me
		panic("we did it guys! now we can rest")
	}, "uid"))
}
