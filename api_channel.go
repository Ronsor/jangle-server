package main

import (
	"log"
	"fmt"
	"time"

	"jangled/util"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"
	"github.com/bwmarrin/snowflake"
)

func InitRestChannel(r *router.Router) {
	log.Println("Init /channels Endpoints")

	type APIReqPostChannelsCidMessages struct {
		Content string `json:"content"`
		Nonce interface{} `json:"nonce"`
		TTS bool `json:"tts"`
		Embed *MessageEmbed `json:"embed"`
		PayloadJson string `json:"payload_json"`
	}

	// Why is this so convoluted Discord? multipart/form-data, application/json, "payload_json"????
	r.POST("/api/v6/channels/:cid/messages", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		var req APIReqPostChannelsCidMessages
		cid := c.UserValue("cid").(string)
		ct := string(c.Request.Header.Peek("Content-Type"))
		if ct == "multipart/form-data" {
			// TODO: something
			// This is gonna be horrible to implement
			panic("TODO")
		} else {
			err := util.ReadPostJSON(c, &req)
			if err != nil {
				util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
				return
			}
		}

		snow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}
		ch, err := GetChannelByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		// Finally construct a minimal message object
		// TODO: the rest lol
		m := &Message{
			Content: req.Content,
			TTS: req.TTS,
			Nonce: fmt.Sprintf("%v", req.TTS),
			Author: &User{ID: me.ID},
			Timestamp: time.Now().Unix(),
			Embeds: []*MessageEmbed{req.Embed},
		}

		ch.CreateMessage(m)

		return
	}))
}

