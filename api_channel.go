package main

import (
	"fmt"
	"log"
	"mime/multipart"
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

	r.GET("/channels/:cid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		util.WriteJSON(c, ch.ToAPI(me))
	}, RL_GETINFO)))

	r.DELETE("/channels/:cid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		if !ch.GetPermissions(me).Has(PERM_MANAGE_CHANNELS) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		resp := ch.ToAPI(me)

		err = ch.Delete()
		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, resp)
		// It was nice knowing ya
	}, RL_DELOBJ)))

	type APIReqPutPatchChannelsCid struct {
		Name                 *string                       `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
		Position             *int                          `json:"position,omitempty" validate:"omitempty,min=0"`
		Topic                *string                       `json:"topic,omitempty" validate:"omitempty,max=256"`
		NSFW                 *bool                         `json:"nsfw,omitempty"`
		RateLimitPerUser     *int                          `json:"rate_limit_per_user,omitempty" validate:"omitempty,min=0 max=86400"`
		PermissionOverwrites []*APITypePermissionOverwrite `json:"permission_overwrites,omitempty"`
		ParentID             *snowflake.ID                 `json:"parent_id,string,omitempty"`
	}

	APIReqPutPatchChannelsCidFn := MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		var req APIReqPutPatchChannelsCid
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}
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

		if !ch.GetPermissions(me).Has(PERM_MANAGE_CHANNELS) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		if req.Name != nil {
			ch.Name = *req.Name
		}
		if req.Position != nil {
			ch.Position = *req.Position
		}
		if req.Topic != nil {
			ch.Topic = *req.Topic
		}
		if req.NSFW != nil {
			ch.NSFW = *req.NSFW
		}
		if req.RateLimitPerUser != nil {
			ch.RateLimitPerUser = *req.RateLimitPerUser
		}
		if req.ParentID != nil {
			pid := *req.ParentID
			pch, err := GetChannelByID(pid)
			if err != nil || pch.GuildID != ch.GuildID || pch.Type != CHTYPE_GUILD_CATEGORY || ch.Type == CHTYPE_GUILD_CATEGORY { // this is loaded
				util.WriteJSONStatus(c, 400, APIERR_UNKNOWN_CHANNEL)
			}
			ch.ParentID = pid
		}
		if req.PermissionOverwrites != nil {
			po := []*PermissionOverwrite{}
			for _, v := range req.PermissionOverwrites {
				x := PermissionOverwrite(*v)
				po = append(po, &x)
			}
			err := ch.SetPermissionOverwrites(po, me)
			if err != nil {
				util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
				return
			}
		}

		ch.Save()

		util.WriteJSON(c, ch.ToAPI())
	}, RL_SETINFO))

	r.PATCH("/channels/:cid", APIReqPutPatchChannelsCidFn)
	r.PUT("/channels/:cid", APIReqPutPatchChannelsCidFn)

	r.POST("/channels/:cid/typing", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)

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

		if !ch.GetPermissions(me).Has(PERM_SEND_MESSAGES) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		err = me.StartTyping(ch)
		if err != nil {
			panic(err)
		}

		util.NoContentJSON(c)
	}, RL_SENDMSG)))

	type APIReqPostChannelsCidMessages struct {
		Content     string                `json:"content" validate:"required_without_all=Embed File,max=3072"`
		Nonce       interface{}           `json:"nonce"`
		TTS         bool                  `json:"tts"`
		Embed       *MessageEmbed         `json:"embed"`
		PayloadJson string                `json:"payload_json"`
		File        *multipart.FileHeader `json:"file"`
	}

	// Why is this so convoluted Discord? multipart/form-data, application/json, "payload_json"????
	r.POST("/channels/:cid/messages", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		defer c.Request.RemoveMultipartFormFiles()
		me := c.UserValue("m:user").(*User)
		var req APIReqPostChannelsCidMessages
		cid := c.UserValue("cid").(string)
		if err := util.ReadPostAny(c, &req); err != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Bad request:" + err.Error()})
			return
		}

		if str, ok := req.Nonce.(string); ok && len(str) > 100 {
			panic("Denial of service attack detected.")
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

		m.ParseContent(me, ch)

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

	r.GET("/channels/:cid/messages", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.GET("/channels/:cid/messages/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.POST("/channels/:cid/messages/:mid/ack", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	r.PATCH("/channels/:cid/messages/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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
		ch, err := GetChannelByID(csnow)
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
			msg.ParseContent(me, ch)
		}

		if req.Embed != nil {
			msg.Embeds = []*MessageEmbed{req.Embed}
		}

		msg.EditedTimestamp = time.Now().Unix()

		msg.Save()

		util.WriteJSON(c, msg.ToAPI())

		return
	}, RL_SENDMSG)))

	r.DELETE("/channels/:cid/messages/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		util.NoContentJSON(c) // No Content
	}, RL_DELMSG)))

	r.PUT("/channels/:cid/pins/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		util.NoContentJSON(c)
	}, RL_SETINFO)))

	r.DELETE("/channels/:cid/pins/:mid", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

		msg.Pinned = false

		msg.Save()

		util.NoContentJSON(c)
	}, RL_SETINFO)))

	r.GET("/channels/:cid/pins", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
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

	type APIReqPostChannelsCidInvites struct {
		MaxAge int `json:"max_age"`
		MaxUses int `json:"max_uses"`
	}

	r.POST("/channels/:cid/invites", MwTkA(MwRl(func(c *fasthttp.RequestCtx) {
		me := c.UserValue("m:user").(*User)
		cid := c.UserValue("cid").(string)
		var req APIReqPostChannelsCidInvites
		if util.ReadPostJSON(c, &req) != nil {
			util.WriteJSONStatus(c, 400, &APIResponseError{0, "Malformed request body"})
			return
		}

		csnow, err := snowflake.ParseString(cid)
		if err != nil {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		ch, err := GetChannelByID(csnow)
		if err != nil || !ch.IsGuild() {
			util.WriteJSONStatus(c, 404, APIERR_UNKNOWN_CHANNEL)
			return
		}

		gd, err := ch.Guild()
		if err != nil {
			panic(err)
		}

		if !ch.GetPermissions(me).Has(PERM_CREATE_INVITE) {
			util.WriteJSONStatus(c, 403, APIERR_MISSING_PERMISSIONS)
			return
		}

		inv := &Invite{
			ChannelID: ch.ID,
			MaxUses: req.MaxUses,
		}

		if req.MaxAge != 0 {
			tm := time.Unix(time.Now().Unix() + int64(req.MaxAge), 0)
			inv.MaxAge = &tm
		}

		inv, err = gd.CreateInvite(inv)
		if err != nil {
			panic(err)
		}

		util.WriteJSON(c, inv.ToAPI())
	}, RL_NEWOBJ)))

	r.PUT("/channels/:cid/messages/:mid/reactions/:emoji/:uid", MwTkA(func(c *fasthttp.RequestCtx) {
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

		util.NoContentJSON(c) // No Content
	}, "uid"))

}
