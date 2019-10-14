package main

import (
	"net/http"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/gorilla/websocket"
)

var gatewayUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func InitGateway(e *echo.Echo) {
	e.GET("/api/v6/gateway", func (c echo.Context) error {
		gw := "http://" + c.Request().Host + "/gateway_ws6"
		// TODO handle other cases
		return c.JSON(http.StatusOK, &responseGetGateway{URL: gw})
	})

	e.GET("/gateway_ws6", func (c echo.Context) error {
		log.Println(c.QueryParams())
		if c.QueryParam("v") != "6" {
			return c.JSON(400, &responseError{Code: 50041, Message: "We only support version 6 here"})
		}
		if c.QueryParam("encoding") != "json" {
			return c.JSON(400, &responseError{Code: 0, Message: "Sorry we don't support etf or msgpack yet"})
		}
		conn, err := gatewayUpgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if err != nil {
			return c.JSON(400, &responseError{Code: 0, Message: "Error initializing gateway websocket"})
		}
		InitSession(conn, c)
		return nil
	})
}

type gwSession struct {
	// This is nil until we get the OP_IDENTIFY packet
	User *User
	// This is nil until we get the OP_IDENTIFY packet
	PktDataIdentify *gwPktDataIdentify
	// Do not expect this!
	// It may be `nil` at any time
	Conn *websocket.Conn
	// Codec
	CodecName string // usually "json"

	recvPkt chan *gwPacket
	sendPkt chan *gwPacket
}

func InitSession(conn *websocket.Conn, ctx echo.Context) {
	s := &gwSession{Conn: conn, CodecName: ctx.QueryParam("encoding")}
	s.rawSendPkt(&gwPacket{
		Op: GW_OP_HELLO,
		Data: &gwPktDataHello{
			HeartbeatInterval: 30000,
		},
	})
	pkt, err := s.rawRecvPkt()
	if err != nil {
		conn.Close()
		return
	}
	switch pkt.Op {
		case GW_OP_IDENTIFY:
			d := pkt.Data.(gwPktDataIdentify)
			log.Println(d)
			s.Ready(d)
			s.Main()
		default:
			conn.Close()
			return
	}
}

func (s *gwSession) Ready(p *gwPktDataIdentify) error {
	s.PktDataIdentify = p
	s.User = &User{
		Username: "Test",
		Discriminator: "1234",
	}
	s.InitIO()
	return nil
}

func (s *gwSession) rawSendPkt(p *gwPacket) error {
	return s.Conn.WriteJSON(p)
}

func (s *gwSession) rawRecvPkt() (p *gwPacket, e error) {
	e = s.Conn.ReadJSON(&p)
	return
}

func (s *gwSession) InitIO() {
	go func() {
		for {
			pkt, err := s.rawRecvPkt()
			if err != nil { break }
			s.recvPkt <- 
		}
	} ()
}

func (s *gwSession) Main() {
	//s.InitIO()
	for {

	}
}
