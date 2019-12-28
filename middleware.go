package main

import (
	"time"

	"jangled/util"

	"github.com/valyala/fasthttp"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

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
	rlc := cache.New(1 * time.Minute, 1 * time.Minute)
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
		lt := getlimiter(factor)
		if !lt.Allow() {
			util.WriteJSONStatus(c, 429, &APIResponseError{0, "Too many requests"})
			return
		}
		orig(c)
	}
}
