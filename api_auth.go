package main

import (
	"log"
	"time"

	"jangled/util"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/globalsign/mgo/bson"
)

func InitRestAuth(r *router.Router) {
	log.Println("Init /auth endpoints")

	type APIReqPostRegister struct {
		Username string `json:"username" validate:"min=2,max=32"`
		Password string `json:"password" validate:"min=6"`
		Email string `json:"email" validate:"email"`
	}

	r.POST("/api/v6/auth/register", MwRl(func (c *fasthttp.RequestCtx) {
		var req APIReqPostRegister
		if err := util.ReadPostJSON(c, &req); err != nil {
			util.WriteJSONStatus(c, 400, bson.M{"email": "Invalid email address", "_raw": err.Error()})
			return
		}
		user, err := CreateUser(req.Username, req.Email, req.Password)
		if err != nil {
			panic(err)
		}
		util.WriteJSON(c, bson.M{"token": user.IssueToken(72 * time.Hour)})
	}, &RateLimitClass{1, 1}))

	type APIReqPostLogin struct {
		Email string `json:"email"`
		Password string `json:"password"`
		Lifetime int `json:"lifetime" validate:"omitempty,max=168"`
	}

	r.POST("/api/v6/auth/login", MwRl(func (c *fasthttp.RequestCtx) {
		var req APIReqPostLogin
		if err := util.ReadPostJSON(c, &req); err != nil {
			util.WriteJSONStatus(c, 400, bson.M{"email": "Invalid email address", "_raw": err.Error()})
			return
		}
		user, err := GetUserByEmail(req.Email)
		if err != nil {
			util.WriteJSONStatus(c, 401, bson.M{"email": "User not found."})
			return
		}

		if req.Lifetime == 0 { req.Lifetime = 168 }

		if !util.VerifyPass(user.PasswordHash, req.Password) {
			util.WriteJSONStatus(c, 401, bson.M{"password": "Incorrect password."})
			return
		}

		util.WriteJSON(c, bson.M{"token": user.IssueToken(time.Duration(req.Lifetime) * time.Hour)})
	}, &RateLimitClass{1, 1}))
}
