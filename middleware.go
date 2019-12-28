package main

import (
	"jangled/util"

	"github.com/valyala/fasthttp"
)

func MwTokenAuth(orig func(c *fasthttp.RequestCtx), uservar ...string) func(c *fasthttp.RequestCtx) {
	uservar = append(uservar, "")
	return func(c *fasthttp.RequestCtx) {
		user, err := GetUserByHttpRequest(c, uservar[0])
		if err != nil {
			util.WriteJSONStatus(c, 401, &APIResponseError{APIERR_UNAUTHORIZED, err.Error()})
			return
		}
		c.SetUserValue("m:user", user)
		orig(c)
	}
}
