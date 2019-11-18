package util

import (
	"errors"

	"github.com/fasthttp/websocket"
	"github.com/mitchellh/mapstructure"
)

const WSCHAN_NOMSG = "WsChan: no message in queue"

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
	sendChan chan interface{}
	recvChan chan interface{}
	getError error
	Closed bool
}

func MakeWsChan(ws *websocket.Conn, c Codec) *WsChan {
	w := &WsChan{
		Conn: ws,
		Codec: c,
		sendChan: make(chan interface{}, 32),
		recvChan: make(chan interface{}, 4),
	}
	go func() {
		for i := range w.sendChan {
			w.Codec.Send(w.Conn, i)
		}
	} ()
	go func() {
		var err error
		for {
			var i map[string]interface{}
			err = w.Codec.Recv(w.Conn, &i)
			if err != nil { break }
			w.recvChan <- i
		}
		w.getError = err
		w.recvChan <- nil
		w.Conn.Close()
	} ()
	return w
}

func (w *WsChan) Send(i interface{}) {
	w.sendChan <- i
}

func (w *WsChan) Recv(i interface{}) error {
	if w.getError != nil {
		return w.getError
	}
	n := <- w.recvChan
	return mapstructure.Decode(n, i)
}

func (w *WsChan) TryRecv(i interface{}) error {
	if w.getError != nil {
		return w.getError
	}
	select {
		case n := <- w.recvChan:
			return mapstructure.Decode(n, i)
		default:
			return errors.New(WSCHAN_NOMSG)
	}
	panic("unreachable")
	return nil
}
