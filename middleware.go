package main

import (
	"time"

	"jangled/util"

	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

// Token-based authorization middleware
func MwTkA(orig func(c *fasthttp.RequestCtx), uservar ...string) func(c *fasthttp.RequestCtx) {
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

// Rate-limiting middleware
func MwRl(orig func(c *fasthttp.RequestCtx), lmt rate.Limit, burst int, options ...interface{}) func(c *fasthttp.RequestCtx) {
	rlc := cache.New(15*time.Second, 1*time.Minute)
	getlimiter := func(factor string) *rate.Limiter {
		lt, ok := rlc.Get(factor)
		if !ok {
			lt = rate.NewLimiter(rate.Limit(lmt), burst)
			rlc.Set(factor, lt, cache.DefaultExpiration)
		}
		return lt.(*rate.Limiter)
	}
	return func(c *fasthttp.RequestCtx) {
		factor := c.RemoteIP().String()
		if usr, ok := c.UserValue("m:user").(*User); ok {
			factor = usr.ID.String()
		}
		lt := getlimiter(factor)
		if !lt.Allow() {
			util.WriteJSONStatus(c, 429, &APIResponseError{0, "Too many requests"})
			return
		}
		orig(c)
	}
}
