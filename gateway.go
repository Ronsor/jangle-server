package main

import (
	"log"

	"github.com/valyala/fasthttp"
	//"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"

	"jangled/util"
)

type gwSession struct {
	User     *User
	Identity *gwPktDataIdentify
	Wsc      *util.WsChan
	Seq      int

	EvtChan chan interface{}
}

func InitGateway(r *router.Router) {
	log.Println("Init Gateway Module")
	r.GET("/api/v6/gateway", func(c *fasthttp.RequestCtx) {
		gw := "ws://" + string(c.Host()) + "/gateway_ws6"
		if *flgGatewayUrl != "" {
			gw = *flgGatewayUrl
		}
		// TODO handle other cases
		util.WriteJSON(c, &responseGetGateway{URL: gw})
	})

	r.GET("/gateway_ws6/", MwRl(func(c *fasthttp.RequestCtx) {
		if string(c.FormValue("v")) != "6" {
			util.WriteJSONStatus(c, 400, &responseError{Code: 50041, Message: "We only support version 6 here"})
			return
		}
		if string(c.FormValue("encoding")) != "json" {
			util.WriteJSONStatus(c, 400, &responseError{Code: 0, Message: "Sorry we don't support etf or msgpack yet"})
			return
		}
		err := (&websocket.FastHTTPUpgrader{ReadBufferSize: 4096, WriteBufferSize: 4096}).Upgrade(c, func(n *websocket.Conn) { InitGatewaySession(n, c) })
		if err != nil {
			util.WriteJSONStatus(c, 400, &responseError{Code: 0, Message: "WebSocket initialization failure"})
		}
	}, RL_NEWOBJ))
}

func InitGatewaySession(ws *websocket.Conn, ctx *fasthttp.RequestCtx) {
	defer ws.Close()
	var sess = &gwSession{}
	var codec util.Codec
	switch string(ctx.FormValue("encoding")) {
	default:
		codec = &util.JsonCodec{}
	}
	codec.Send(ws, mkGwPkt(GW_OP_HELLO, &gwPktDataHello{45000}))
	var pkt *gwPacket
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

		codec.Send(ws, &gwPacket{
			GW_OP_DISPATCH,
			&gwEvtDataReady{
				Version:         6,
				User:            sess.User.ToAPI(false),
				Guilds:          ugs,
				PrivateChannels: []interface{}{},
			},
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
			codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, g.ToAPI(sess.User.ID), sess.Seq, GW_EVT_GUILD_CREATE))
			SessSub.AddSub(sess.EvtChan, g.ID.String())
			sess.Seq++
			chans, _ := g.Channels()
			for _, ch := range chans {
				if ch.GetPermissions(sess.User).Has(PERM_VIEW_CHANNEL) {
					codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, ch.ToAPI(), sess.Seq, GW_EVT_CHANNEL_CREATE))
					SessSub.AddSub(sess.EvtChan, ch.ID.String())
					sess.Seq++
				}
			}
		}
		break
	// TODO: resuming sessions
	default:
		panic("Not supported")
	}
	log.Printf("Authenticated user: ID=%v", sess.User.ID)
	wsc := util.MakeWsChan(ws, codec)
	sess.Wsc = wsc
	go func() {
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
				case GW_EVT_CHANNEL_DELETE:
					go SessSub.Unsub(sess.EvtChan, pkt.PvtData.(*Channel).ID.String())
				case GW_EVT_GUILD_CREATE:
					SessSub.AddSub(sess.EvtChan, pkt.PvtData.(*Guild).ID.String())
				}
			}
			pkt.Seq = sess.Seq
			wsc.Send(&pkt)
			sess.Seq++
		}
	}()
	for {
		pkt := new(gwPacket)
		err := wsc.Recv(&pkt)
		if err != nil {
			break
		}
		RefreshPresenceForUser(sess.User.ID)
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
	ws.Close()
}
