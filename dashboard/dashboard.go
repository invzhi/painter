package dashboard

import (
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/invzhi/shaker/message"
)

type dashboard struct {
	hub  *hub
	conn *websocket.Conn
	send chan *message.Msg
}

// New allocates a new dashboard and register it to dashboard hub.
func New(h *hub, conn *websocket.Conn, bufferSize int) *dashboard {
	d := &dashboard{
		hub:  h,
		conn: conn,
		send: make(chan *message.Msg, bufferSize),
	}
	h.register <- d
	return d
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// ReadTo read message from dashboard and send to channel
func (d *dashboard) ReadTo(send chan []byte) {
	defer func() {
		d.hub.unregister <- d
		d.conn.Close()
	}()
	pongHandler := func(_ string) error {
		log.Print("pong received.")
		return d.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
	d.conn.SetReadLimit(maxMessageSize)
	d.conn.SetReadDeadline(time.Now().Add(pongWait))
	d.conn.SetPongHandler(pongHandler)

	for {
		_, msg, err := d.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Print("websocket unexpected close error: ", err)
			}
			break
		}
		send <- msg
		log.Print("message to clients: ", string(msg))
	}
}

// Write write message from dashboard's send to conn
func (d *dashboard) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		d.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-d.send:
			d.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				d.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			if err := d.conn.WriteJSON(msg); err != nil {
				log.Printf("cannot write JSON %s to conn: %v", msg, err)
			}
		case <-ticker.C:
			d.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := d.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Print("cannot write ping message: ", err)
				return
			}
		}
	}
}
