package main

import (
	"log"
	"fmt"
	"time"
	"strconv"

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
			if util.ReadPostJSON(c, &req) != nil {
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

		if !ch.HasPermissions(me, PERM_SEND_MESSAGES) {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_MISSING_PERMISSIONS, "You don't have permission to do that"})
			return
		}

		// Finally construct a minimal message object
		// TODO: the rest lol
		m := &Message{
			Content: req.Content,
			TTS: req.TTS,
			Nonce: fmt.Sprintf("%v", req.Nonce),
			Author: &User{ID: me.ID},
			Timestamp: time.Now().Unix(),
			Embeds: []*MessageEmbed{},
		}

		if req.Embed != nil { m.Embeds = append(m.Embeds, req.Embed) }

		ch.CreateMessage(m)

		util.WriteJSON(c, m.ToAPI())

		return
	}))

	r.GET("/api/v6/channels/:cid/messages", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)

		_ = me

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		if !ch.HasPermissions(me, PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_MISSING_PERMISSIONS, "You don't have permission to do that"})
			return
		}

		around, _ := snowflake.ParseBytes(c.FormValue("around"))
		before, _ := snowflake.ParseBytes(c.FormValue("before"))
		after, _ := snowflake.ParseBytes(c.FormValue("after"))
		limit, err := strconv.Atoi(string(c.FormValue("limit")))
		if err != nil {
			limit = 50
		}

		msgs, err := ch.Messages(around, before, after, limit)

		if err != nil {
			util.WriteJSONStatus(c, 500, &APIResponseError{0, "An internal error occurred"})
			return
		}

		outmsgs := []*APITypeMessage{}
		for _, v := range msgs { outmsgs = append(outmsgs, v.ToAPI()) }

		util.WriteJSON(c, outmsgs)
		return
	}))

	r.GET("/api/v6/channels/:cid/messages/:mid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		if !ch.HasPermissions(me, PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_MISSING_PERMISSIONS, "You don't have permission to do that"})
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg, err := GetMessageByID(msnow)
		if err != nil || msg.ChannelID != csnow {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		_ = me

		util.WriteJSON(c, msg.ToAPI())

		return
	}))

	type APIReqPatchChannelsCidMessagesMid struct {
		Content *string `json:"content"`
		Embed *MessageEmbed `json:"embed"`
		Flags int `json:"flags"` // Ignored
	}

	r.PATCH("/api/v6/channels/:cid/messages/:mid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		var req APIReqPatchChannelsCidMessagesMid
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg, err := GetMessageByID(msnow)
		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		if msg.Author.ID != me.ID {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_CANT_EDIT_MESSAGE, "Can't edit message sent by another user"})
			return
		}

		if req.Content != nil {
			msg.Content = *req.Content
		}

		if req.Embed != nil {
			msg.Embeds = []*MessageEmbed{req.Embed}
		}

		msg.EditedTimestamp = time.Now().Unix()

		msg.Save()

		util.WriteJSON(c, msg.ToAPI())

		return
	}))

	r.DELETE("/api/v6/channels/:cid/messages/:mid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg, err := GetMessageByID(msnow)

		if msg.Author.ID != me.ID {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_CANT_EDIT_MESSAGE, "Can't delete message sent by another user"})
			return
		}

		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg.Deleted = true

		msg.Save()

		c.SetStatusCode(204) // No Content
	}))

	r.PUT("/api/v6/channels/:cid/pins/:mid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		if !ch.HasPermissions(me, PERM_MANAGE_MESSAGES) {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_MISSING_PERMISSIONS, "You don't have permission to do that"})
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg, err := GetMessageByID(msnow)

		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg.Pinned = true

		msg.Save()

		c.SetStatusCode(204)
	}))

	r.GET("/api/v6/channels/:cid/pins", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)

		_ = me

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		if !ch.HasPermissions(me, PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, &APIResponseError{APIERR_MISSING_PERMISSIONS, "You don't have permission to do that"})
			return
		}

		limit, err := strconv.Atoi(string(c.FormValue("limit")))
		if err != nil || limit > 50 {
			limit = 50
		}

		msgs, err := ch.Messages(0, 0, 0, limit, true)

		if err != nil {
			util.WriteJSONStatus(c, 500, &APIResponseError{0, "An internal error occurred"})
			return
		}

		outmsgs := []*APITypeMessage{}
		for _, v := range msgs { outmsgs = append(outmsgs, v.ToAPI()) }

		util.WriteJSON(c, outmsgs)
		return
	}))

	r.PUT("/api/v6/channels/:cid/messages/:mid/reactions/:emoji/:uid", MwTokenAuth(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_CHANNEL, "Channel does not exist"})
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		msg, err := GetMessageByID(msnow)

		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, &APIResponseError{APIERR_UNKNOWN_MESSAGE, "Message does not exist"})
			return
		}

		emoji, err := GetEmojiFromString(c.UserValue("emoji").(string))

		if err != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Nope"})
			return
		}

		_,_,_,_,_ = me, csnow, msnow, msg, emoji

		c.SetStatusCode(204) // No Content
	}, "uid"))

}

