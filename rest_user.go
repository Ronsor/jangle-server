package main

import (
	//"net/http"
	"log"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"

	"server/util"
)

func InitRestUser(r *router.Router) {
	log.Println("Init Gateway Module")
	r.GET("/api/v6/users/:uid/settings", func(c *fasthttp.RequestCtx) {
		defer util.TryRecover()
		user, err := GetUserByHttpCtx(c, "uid")
		if err != nil {
			// TODO: something
		}
		util.WriteJSON(c, user.Settings)
	})
}
