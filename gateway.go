package main

import (
	"log"
	"math/rand"
	"sync"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"github.com/globalsign/mgo/bson"
)

type gwSession struct {
	User     *User
	Identity *gwPktDataIdentify
	Wsc      *util.WsChan
	Seq      int
	ID       snowflake.ID
	Lock     sync.Mutex

	EvtChan chan interface{}
}

func InitGateway(r *router.Router) {
	log.Println("Init Gateway Module")
	gwEndpoint := func(c *fasthttp.RequestCtx) {
		gw := "ws://" + string(c.Host()) + "/gateway_ws6"
		if *flgGatewayUrl != "" {
			gw = *flgGatewayUrl
		}
		util.WriteJSON(c, bson.M{
			"url": gw,
			// It seems discord.js needs this
			"session_start_limit": bson.M{
				"reset_after": 3600,
				"total": 0xFFFF,
				"remaining": 0xFFFF,
			},
		})
	}
	r.GET("/gateway/", gwEndpoint)
	r.GET("/gateway/bot", gwEndpoint)

	r.GET("/gateway_ws6/", MwRl(func(c *fasthttp.RequestCtx) {
		if string(c.FormValue("v")) != "6" {
			util.WriteJSONStatus(c, 400, &responseError{Code: 50041, Message: "We only support version 6 here"})
			return
		}
		if string(c.FormValue("encoding")) != "json" {
			util.WriteJSONStatus(c, 400, &responseError{Code: 0, Message: "Sorry we don't support etf or msgpack yet"})
			return
		}
		err := (&websocket.FastHTTPUpgrader{ReadBufferSize: 4096, WriteBufferSize: 4096,
			CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true }}).Upgrade(c, func(n *websocket.Conn) { InitGatewaySession(n, c) })
		if err != nil {
			util.WriteJSONStatus(c, 400, &responseError{Code: 0, Message: "WebSocket initialization failure"})
		}
	}, RL_NEWOBJ))
}

func InitGatewaySession(ws *websocket.Conn, ctx *fasthttp.RequestCtx) {
	defer util.TryRecover()
	defer ws.Close()
	var sess = &gwSession{}
	var codec util.Codec
	switch string(ctx.FormValue("encoding")) {
	default:
		codec = &util.JsonCodec{}
	}
	codec.Send(ws, mkGwPkt(GW_OP_HELLO, &gwPktDataHello{45000 + (rand.Intn(5) * 1000)}))
	var pkt *gwPacket

	retryHandshake:
	err := codec.Recv(ws, &pkt)
	if err != nil {
		ws.Close()
		return
	}

	switch pkt.Op {
	case GW_OP_IDENTIFY:
		var d gwPktDataIdentify
		pkt.D(&d)
		sess.User, err = GetUserByToken(d.Token)
		if err != nil {
			ws.Close()
			return
		}
		sess.ID = flake.Generate()
		sess.Identity = &d
		sess.EvtChan = SessSub.Sub(sess.User.ID.String())

		guilds, err := GetGuildsByUserID(sess.User.ID)

		if err != nil {
			guilds = []*Guild{}
		}

		ugs := []*UnavailableGuild{}
		for _, g := range guilds {
			ugs = append(ugs, &UnavailableGuild{g.ID, true})
		}

		if sess.Identity.Presence == nil {
			sess.Identity.Presence = &gwPktDataUpdateStatus{}
		}
		if sess.Identity.Presence.Status == STATUS_UNKNOWN {
			sess.Identity.Presence.Status = STATUS_ONLINE
		}
		err = SetPresenceForUser(sess.User.ID, sess.Identity.Presence)
		if err != nil {
			panic(err)
		}

		readyPkt := &gwEvtDataReady{
			Version:         6,
			User:            sess.User.ToAPI(false),
			Guilds:          ugs,
			PrivateChannels: []interface{}{},
		}

		if !sess.User.Bot {
			readyPkt.UserSettings = sess.User.Settings

			chs, err := sess.User.Channels()
			if err != nil {
				panic(err)
			}
			out := make([]*APITypeDMChannel, 0, len(chs))
			for _, v := range chs {
				if v.Type == CHTYPE_DM {
					out = append(out, v.ToAPI(sess.User).(*APITypeDMChannel))
				}
			}

			readyPkt.PrivateChannels = out
			readyPkt.Relationships = []interface{}{}
		}

		codec.Send(ws, &gwPacket{
			GW_OP_DISPATCH,
			readyPkt,
			GW_EVT_READY,
			sess.Seq,
			nil,
		})
		sess.Seq++

		chs, err := sess.User.Channels()
		if err != nil {
			chs = []*Channel{}
		}

		for _, ch := range chs {
			codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, ch.ToAPI(), sess.Seq, GW_EVT_CHANNEL_CREATE))
			SessSub.AddSub(sess.EvtChan, ch.ID.String())
			sess.Seq++
		}

		for _, g := range guilds {
			codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, g.ToAPI(sess.User.ID, true), sess.Seq, GW_EVT_GUILD_CREATE))
			SessSub.AddSub(sess.EvtChan, g.ID.String())
			sess.Seq++
			chans, _ := g.Channels()
			for _, ch := range chans {
				if ch.GetPermissions(sess.User).Has(PERM_VIEW_CHANNEL) {
					//codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, ch.ToAPI(), sess.Seq, GW_EVT_CHANNEL_CREATE))
					SessSub.AddSub(sess.EvtChan, ch.ID.String())
					sess.Seq++
				}
			}
		}
		break
	case GW_OP_RESUME:
		var d gwPktDataResume
		pkt.D(&d)
		s2, ok := sessCache.Get(d.SessionID)
		if !ok {
			codec.Send(ws, mkGwPkt(GW_OP_INVALID_SESSION, false))
			goto retryHandshake
		}

		sess = s2.(*gwSession)
		// TODO: care about d.Seq
		codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, true, sess.Seq, "RESUMED"))
		sess.Seq++
	}
	sessCache.Set(sess.ID, sess)
	log.Printf("Authenticated user: ID=%v", sess.User.ID)
	wsc := util.MakeWsChan(ws, codec)
	sess.Wsc = wsc
	go func() {
		defer util.TryRecover()
		for r := range sess.EvtChan {
			if sess.Wsc.Closed {
				break
			}
			pkt := r.(gwPacket)
			if pkt.Op == GW_OP_DISPATCH {
				switch pkt.Type {
				case GW_EVT_CHANNEL_CREATE:
					ch := pkt.PvtData.(*Channel)
					if ch.GetPermissions(sess.User).Has(PERM_VIEW_CHANNEL) {
						SessSub.AddSub(sess.EvtChan, ch.ID.String())
					}
				case GW_EVT_CHANNEL_UPDATE:
					ch := pkt.PvtData.(*Channel)
					if !ch.GetPermissions(sess.User).Has(PERM_VIEW_CHANNEL) {
						go SessSub.Unsub(sess.EvtChan, ch.ID.String())
					}
				case GW_EVT_GUILD_UPDATE:
					// TODO
				case GW_EVT_CHANNEL_DELETE:
					go SessSub.Unsub(sess.EvtChan, pkt.PvtData.(*Channel).ID.String())
				case GW_EVT_GUILD_CREATE:
					SessSub.AddSub(sess.EvtChan, pkt.PvtData.(*Guild).ID.String())
				case GW_EVT_GUILD_DELETE:
					go SessSub.Unsub(sess.EvtChan, pkt.PvtData.(*Guild).ID.String())
				}
			}
			pkt.Seq = sess.Seq
			sess.Seq++
			wsc.Send(&pkt)
		}
	}()
	for {
		pkt := new(gwPacket)
		err := wsc.Recv(&pkt)
		if err != nil {
			break
		}
		RefreshPresenceForUser(sess.User.ID)
		sessCache.Set(sess.ID, sess)
		switch pkt.Op {
		case GW_OP_HEARTBEAT:
			wsc.Send(&gwPacket{Op: GW_OP_HEARTBEAT_ACK})
			break
		case GW_OP_UPDATE_STATUS:
			var dat gwPktDataUpdateStatus
			pkt.D(&dat)
			SetPresenceForUser(sess.User.ID, &dat)
		}
	}
}
