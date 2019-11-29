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
	User *User
	Identity *gwPktDataIdentify
	Wsc *util.WsChan
	Seq int

	EvtChan chan interface{}
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
	var sess = &gwSession{}
	var codec util.Codec
	switch string(ctx.FormValue("encoding")) {
		default:
			codec = &util.JsonCodec{}
	}
	codec.Send(ws, mkGwPkt(GW_OP_HELLO, &gwPktDataHello{60000}))
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
			sess.EvtChan = SessSub.Sub()
			codec.Send(ws, &gwPacket{
				GW_OP_DISPATCH,
				&gwEvtDataReady{
					Version: 6,
					User: sess.User.ToAPI(false),
					Guilds: []*UnavailableGuild{},
					PrivateChannels: []interface{}{},
				},
				GW_EVT_READY,
				sess.Seq,
				nil,
			})
			sess.Seq++
			for _, ch := range sess.User.Channels() {
				codec.Send(ws, mkGwPkt(GW_OP_DISPATCH, ch.ToAPI(), sess.Seq, GW_EVT_CHANNEL_CREATE))
				log.Println(ch.ID.String())
				SessSub.AddSub(sess.EvtChan, ch.ID.String())
				sess.Seq++
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
				log.Println("E")
				break
			}
			pkt := r.(gwPacket)
			pkt.Seq = sess.Seq
			log.Println(pkt.Data)
			wsc.Send(&pkt)
			sess.Seq++
		}
	} ()
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

