package main

import (
	"log"

_	"jangled/util"

	"github.com/fasthttp/router"
_	"github.com/valyala/fasthttp"
)

func InitRestAuth(r *router.Router) {
	log.Println("Init auth = [/register, /login, /verify] endpoints")

	type APIReqPostRegister struct {
		Username string `json:"username" validate:"min=2,max=32"`
		Password string `json:"password" validate:"min=6"`
		Email string `json:"email" validate:"email"`
	}
}
