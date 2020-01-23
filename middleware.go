package main

import (
	"time"

	"jangled/util"

	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

// Rate limit classes
type RateLimitClass struct {
	Limit rate.Limit
	Burst int
}

var (
	RL_GETINFO = &RateLimitClass{5, 1}
	RL_SETINFO = &RateLimitClass{5, 1}

	RL_SENDMSG = &RateLimitClass{5, 10}
	RL_RECVMSG = &RateLimitClass{5, 10}
	RL_DELMSG  = &RateLimitClass{10, 10}

	RL_NEWOBJ = &RateLimitClass{3, 1}
	RL_DELOBJ = &RateLimitClass{3, 3}
)

// Access-control-allow-whatever middleware
func MwAccCtl(orig func(c *fasthttp.RequestCtx), allow string) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		rh := &c.Response.Header
		rh.Set("Access-Control-Allow-Origin", "*")
		rh.Set("Access-Control-Allow-Credentials", "true")
		rh.Set("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, DELETE, OPTIONS")
		rh.Set("Access-Control-Expose-Headers", "*")
		rh.Set("Access-Control-Allow-Headers", "User-Agent, Authorization, X-Jangle-Meta, X-Jangle-Client-Version, Content-Type, Expires, Cache-Control")
		if string(c.Request.Header.Method()) != "OPTIONS" {
			orig(c)
		}
	}
}

// Token-based authorization middleware
func MwTkA(orig func(c *fasthttp.RequestCtx), uservar ...string) func(c *fasthttp.RequestCtx) {
	uservar = append(uservar, "")
	return func(c *fasthttp.RequestCtx) {
		user, err := GetUserByHttpRequest(c, uservar[0])
		if err != nil {
			util.WriteJSONStatus(c, 401, APIERR_UNAUTHORIZED)
			return
		}
		c.SetUserValue("m:user", user)
		orig(c)
	}
}

// Rate-limiting middleware
func MwRl(orig func(c *fasthttp.RequestCtx), clazz *RateLimitClass) func(c *fasthttp.RequestCtx) {
	rlc := cache.New(15*time.Second, 1*time.Minute)
	getlimiter := func(factor string) *rate.Limiter {
		lt, ok := rlc.Get(factor)
		if !ok {
			lt = rate.NewLimiter(clazz.Limit, clazz.Burst)
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
