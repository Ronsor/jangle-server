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
}

func InitGateway(r *router.Router) {
	log.Println("Init Gateway Module")
	r.GET("/api/v6/gateway", func(c *fasthttp.RequestCtx) {
		gw := "ws://" + string(c.Host()) + "/gateway_ws6"
		// TODO handle other cases
		util.WriteJSON(c, &responseGetGateway{URL: gw})
	})

	r.GET("/gateway_ws6/", func(c *fasthttp.RequestCtx) {
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
	defer ws.Close()
	var sess *gwSession
	var codec util.Codec
	switch string(ctx.FormValue("encoding")) {
		default:
			codec = &util.JsonCodec{}
	}
	codec.Send(ws, mkGwPkt(GW_OP_HELLO, &gwPktDataHello{30000}))
	var pkt *gwPacket
	codec.Recv(ws, &pkt)
	switch pkt.Op {
		case GW_OP_IDENTIFY:
			d := pkt.Data.(gwPktDataIdentify)
			sess.User, err := GetUserByToken(d.Token)
			if err != nil {
				ws.Close()
				return
			}
			sess.Identity = d
		default:
			panic("Not supported")
	}
	wsc := util.MakeWsChan(
	for {
	}
}
