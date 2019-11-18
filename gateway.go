package main

import (
	//"net/http"
	"log"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"

	"server/util"
)

type gwSession struct {
	User *User
	Identity *gwPktDataIdentify
	Wsc *util.WsChan
	Seq int
}

func InitGateway(r *router.Router) {
	log.Println("Init Gateway Module")
	r.GET("/api/v6/gateway", func(c *fasthttp.RequestCtx) {
		defer util.TryRecover()
		gw := "ws://" + string(c.Host()) + "/gateway_ws6"
		// TODO handle other cases
		util.WriteJSON(c, &responseGetGateway{URL: gw})
	})

	r.GET("/gateway_ws6/", func(c *fasthttp.RequestCtx) {
		defer util.TryRecover()
		if string(c.FormValue("v")) != "6" {
			util.WriteJSONStatus(c, 400, &responseError{Code: 50041, Message: "We only support version 6 here"}); return
		}
		if string(c.FormValue("encoding")) != "json" {
			util.WriteJSONStatus(c, 400, &responseError{Code: 0, Message: "Sorry we don't support etf or msgpack yet"}); return
		}
		err := (&websocket.FastHTTPUpgrader{ReadBufferSize: 4096, WriteBufferSize: 4096}).Upgrade(c, func (n *websocket.Conn) { InitGatewaySession(n, c) })
		if err != nil {
			util.WriteJSONStatus(c, 400, &responseError{Code: 0, Message: "WebSocket initialization failure"})
		}
	})
}

func InitGatewaySession(ws *websocket.Conn, ctx *fasthttp.RequestCtx) {
	defer util.TryRecover()
	defer ws.Close()
	var sess = new(gwSession)
	var codec util.Codec
	switch string(ctx.FormValue("encoding")) {
		default:
			codec = &util.JsonCodec{}
	}
	codec.Send(ws, mkGwPkt(GW_OP_HELLO, &gwPktDataHello{1000}))
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
			guilds := make([]*UnavailableGuild, len(sess.User.GuildIDs))
			for k, v := range sess.User.GuildIDs { guilds[k] = &UnavailableGuild{v, true} }
			codec.Send(ws, &gwPacket{
				GW_OP_DISPATCH,
				&gwEvtDataReady{
					Version: 6,
					User: sess.User.MarshalAPI(true),
					Guilds: guilds,
					PrivateChannels: []interface{}{},
				},
				GW_EVT_READY,
				sess.Seq,
			})
			sess.Seq++
			break
		// TODO: resuming sessions
		default:
			panic("Not supported")
	}
	log.Printf("Authenticated user: ID=%v", sess.User.ID)
	wsc := util.MakeWsChan(ws, codec)
	sess.Wsc = wsc
	for {
		pkt := new(gwPacket)
		err := wsc.Recv(&pkt)
		if err != nil {
			break
		}
		log.Println("got pkt:", pkt)
		switch pkt.Op {
			case GW_OP_HEARTBEAT:
				wsc.Send(&gwPacket{Op: GW_OP_HEARTBEAT_ACK})
			break
		}
	}
	ws.Close()
}
