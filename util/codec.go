package util

import (
	"github.com/gorilla/websocket"
)

type Codec interface {
	Send(*websocket.Conn, interface{}) error
	Recv(*websocket.Conn, interface{}) error
}

type JsonCodec struct {}

func (j *JsonCodec) Send(ws *websocket.Conn, i interface{}) error {
	return ws.WriteJSON(i)
}

func (j *JsonCodec) Recv(ws *websocket.Conn, i interface{}) error {
	return ws.ReadJSON(i)
}

type WsChan struct {
	Conn *websocket.Conn
	Codec Codec
	Send chan interface{}
	Recv chan interface{}
	Close chan interface{}
	Closed bool
}

func MakeWsChan(ws *websocket.Conn, c Codec) *WsChan {
	w := &WsChan{
		Conn: ws,
		Codec: c,
		Send: make(chan interface{}, 32),
		Recv: make(chan interface{}, 4),
	}
	go func() {
		for i := range w.Send {
			w.Codec.Send(w.Conn, i)
		}
	} ()
	return w
}
