package main

import (
	//	"net/http"
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"server/util"
)

func InitGateway(e *echo.Echo) {
	log.Println("Init Gateway Module")
	e.GET("/api/v6/gateway", func(c echo.Context) error {
		gw := "ws://" + c.Request().Host + "/gateway_ws6"
		// TODO handle other cases
		return c.JSON(200, &responseGetGateway{URL: gw})
	})

	e.GET("/gateway_ws6/", func(c echo.Context) error {
		if c.QueryParam("v") != "6" {
			return c.JSON(400, &responseError{Code: 50041, Message: "We only support version 6 here"})
		}
		if c.QueryParam("encoding") != "json" {
			return c.JSON(400, &responseError{Code: 0, Message: "Sorry we don't support etf or msgpack yet"})
		}
		ws, err := (&websocket.Upgrader{ReadBufferSize: 4096, WriteBufferSize: 4096}).Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(400, &responseError{Code: 0, Message: "WebSocket initialization failure"})
		}
		InitGatewaySession(ws, c)
		return nil
	})
}

func InitGatewaySession(ws *websocket.Conn, ctx echo.Context) {
	defer ws.Close()
	_ = ctx
	codec := &util.JsonCodec{}
	log.Println("ok")
	codec.Send(ws, &gwPktMini{GW_OP_HELLO, &gwPktDataHello{30000}})
	var pkt *gwPacket
	codec.Recv(ws, &pkt)
	switch pkt.Op {
	case GW_OP_IDENTIFY:
		log.Println(pkt)
	default:
		panic("Not supported")
	}
	for {
	}
}
