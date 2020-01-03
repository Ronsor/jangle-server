package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/globalsign/mgo/bson"
	"github.com/valyala/fasthttp"
)

func InitRestChannel(r *router.Router) {
	log.Println("Init /channels Endpoints")

	r.GET("/api/v6/channels/:cid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		snow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}
		ch, err := GetChannelByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_VIEW_CHANNEL) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		util.WriteJSON(c, ch.ToAPI())
	}, RL_GETINFO)))

	type APIReqPutPatchChannelsCid struct {
		Name                 *string                       `json:"name" validate:"min=1,max=64"`
		Position             *int                          `json:"position"`
		Topic                *string                       `json:"topic" validate:"max=256"`
		NSFW                 *bool                         `json:"nsfw"`
		RateLimitPerUser     *int                          `json:"rate_limit_per_user"`
		PermissionOverwrites []*APITypePermissionOverwrite `json:"permission_overwrites"`
		ParentID             *snowflake.ID                 `json:"parent_id"`
	}

	type APIReqPostChannelsCidMessages struct {
		Content     string        `json:"content" validate:"required"`
		Nonce       interface{}   `json:"nonce"`
		TTS         bool          `json:"tts"`
		Embed       *MessageEmbed `json:"embed"`
		PayloadJson string        `json:"payload_json"`
	}

	// Why is this so convoluted Discord? multipart/form-data, application/json, "payload_json"????
	r.POST("/api/v6/channels/:cid/messages", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}
		ch, err := GetChannelByID(snow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_SEND_MESSAGES) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		// Finally construct a minimal message object
		// TODO: the rest lol
		m := &Message{
			Content:   req.Content,
			TTS:       req.TTS,
			Nonce:     fmt.Sprintf("%v", req.Nonce),
			Author:    &User{ID: me.ID},
			Timestamp: time.Now().Unix(),
			Embeds:    []*MessageEmbed{},
		}

		if req.Embed != nil {
			m.Embeds = append(m.Embeds, req.Embed)
		}

		if req.Embed == nil && req.Content == "" {
			util.WriteJSONStatus(c, 400, APIERR_EMPTY_MESSAGE)
			return
		}

		err = ch.CreateMessage(m)

		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, m.ToAPI())

		return
	}, RL_SENDMSG)))

	r.GET("/api/v6/channels/:cid/messages", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)

		_ = me

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
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
			panic(err)
		}

		outmsgs := []*APITypeMessage{}
		for _, v := range msgs {
			outmsgs = append(outmsgs, v.ToAPI())
		}

		util.WriteJSON(c, outmsgs)
		return
	}, RL_RECVMSG)))

	r.GET("/api/v6/channels/:cid/messages/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg, err := GetMessageByID(msnow)
		if err != nil || msg.ChannelID != csnow {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		_ = me

		util.WriteJSON(c, msg.ToAPI())

		return
	}, RL_RECVMSG)))

	r.POST("/api/v6/channels/:cid/messages/:mid/ack", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg, err := GetMessageByID(msnow)
		if err != nil || msg.ChannelID != csnow {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		_ = me

		util.WriteJSON(c, bson.M{"token": nil})

		return
	}, RL_RECVMSG)))

	type APIReqPatchChannelsCidMessagesMid struct {
		Content *string       `json:"content"`
		Embed   *MessageEmbed `json:"embed"`
		Flags   int           `json:"flags"` // Ignored
	}

	r.PATCH("/api/v6/channels/:cid/messages/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg, err := GetMessageByID(msnow)
		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		if msg.Author.ID != me.ID {
			util.WriteJSONStatus(c, 403, APIERR_CANT_EDIT_MESSAGE)
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
	}, RL_SENDMSG)))

	r.DELETE("/api/v6/channels/:cid/messages/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg, err := GetMessageByID(msnow)

		if msg.Author.ID != me.ID {
			util.WriteJSONStatus(c, 403, APIERR_CANT_EDIT_MESSAGE)
			return
		}

		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg.Deleted = true

		msg.Save()

		c.SetStatusCode(204) // No Content
	}, RL_DELMSG)))

	r.PUT("/api/v6/channels/:cid/pins/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_MANAGE_MESSAGES) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg, err := GetMessageByID(msnow)

		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg.Pinned = true

		msg.Save()

		c.SetStatusCode(204)
	}, RL_SETINFO)))

	r.GET("/api/v6/channels/:cid/pins", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)

		_ = me

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		if !ch.GetPermissions(me).Has(PERM_VIEW_CHANNEL | PERM_READ_MESSAGE_HISTORY) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		limit, err := strconv.Atoi(string(c.FormValue("limit")))
		if err != nil || limit > 50 {
			limit = 50
		}

		msgs, err := ch.Messages(0, 0, 0, limit, true)

		if err != nil {
			panic(err)
		}

		outmsgs := []*APITypeMessage{}
		for _, v := range msgs {
			outmsgs = append(outmsgs, v.ToAPI())
		}

		util.WriteJSON(c, outmsgs)
		return
	}, RL_RECVMSG)))

	r.PUT("/api/v6/channels/:cid/messages/:mid/reactions/:emoji/:uid", MwTkA(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		mid := c.UserValue("mid").(string)

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		msnow, err := snowflake.ParseString(mid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		msg, err := GetMessageByID(msnow)

		if err != nil || msg.ChannelID != csnow || msg.Deleted {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_MESSAGE)
			return
		}

		emoji, err := GetEmojiFromString(c.UserValue("emoji").(string))

		if err != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Nope"})
			return
		}

		_, _, _, _, _ = me, csnow, msnow, msg, emoji

		c.SetStatusCode(204) // No Content
	}, "uid"))

}
